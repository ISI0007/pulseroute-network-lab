package nearby

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"mime/multipart"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/ISI0007/pulseroute-network-lab/internal/network"
)

func startTest(t *testing.T, paths []string, sink network.Sink, options Options) (*Server, string, string) {
	t.Helper()
	dir := t.TempDir()
	server, err := StartWithOptions("127.0.0.1:0", dir, paths, sink, options)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = server.Close() })
	return server, "http://" + server.Addr(), dir
}
func request(t *testing.T, method, url, token string, body io.Reader, contentType, digest string) *http.Response {
	t.Helper()
	r, err := http.NewRequest(method, url, body)
	if err != nil {
		t.Fatal(err)
	}
	if token != "" {
		r.Header.Set("X-PulseRoute-Token", token)
	}
	if contentType != "" {
		r.Header.Set("Content-Type", contentType)
	}
	if digest != "" {
		r.Header.Set("X-PulseRoute-SHA256", digest)
	}
	response, err := (&http.Client{Timeout: 3 * time.Second}).Do(r)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { response.Body.Close() })
	return response
}
func multipartBody(t *testing.T, name string, data []byte, extra bool) (*bytes.Buffer, string) {
	t.Helper()
	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", name)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = part.Write(data); err != nil {
		t.Fatal(err)
	}
	if extra {
		if err = writer.WriteField("unexpected", "field"); err != nil {
			t.Fatal(err)
		}
	}
	if err = writer.Close(); err != nil {
		t.Fatal(err)
	}
	return body, writer.FormDataContentType()
}
func noFiles(t *testing.T, dir string) {
	t.Helper()
	files, err := os.ReadDir(dir)
	if err != nil || len(files) != 0 {
		t.Fatalf("unexpected received/partial files: %v, %v", files, err)
	}
}
func TestCapabilitySessionAndSafeDownloads(t *testing.T) {
	source := filepath.Join(t.TempDir(), "document.bin")
	data := []byte{0, 255, 42, 17}
	if err := os.WriteFile(source, data, 0600); err != nil {
		t.Fatal(err)
	}
	s, url, _ := startTest(t, []string{source}, nil, Options{})
	if len(s.Token()) != 64 {
		t.Fatal("capability token is not 256-bit hexadecimal")
	}
	public := request(t, "GET", url+"/", "", nil, "", "")
	if public.StatusCode != 200 || public.Header.Get("Referrer-Policy") != "no-referrer" || !strings.Contains(public.Header.Get("Content-Security-Policy"), "frame-ancestors 'none'") {
		t.Fatal("missing portal safety headers")
	}
	for _, endpoint := range []string{"/api/session", "/api/session?token=" + s.Token()} {
		response := request(t, "GET", url+endpoint, "", nil, "", "")
		if response.StatusCode != 401 {
			t.Fatal("session metadata accessible without header capability", response.StatusCode)
		}
	}
	response := request(t, "GET", url+"/api/session", s.Token(), nil, "", "")
	raw, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatal(err)
	}
	if response.StatusCode != 200 || bytes.Contains(raw, []byte(source)) || bytes.Contains(raw, []byte(s.Token())) {
		t.Fatal("private path/token leaked in metadata", string(raw))
	}
	var metadata struct {
		Files     []SharedFile `json:"files"`
		MaxUpload int64        `json:"maxUploadBytes"`
	}
	if err = json.Unmarshal(raw, &metadata); err != nil {
		t.Fatal(err)
	}
	if len(metadata.Files) != 1 || metadata.Files[0].Path != "" || metadata.MaxUpload != MaxUpload {
		t.Fatal(metadata)
	}
	downloaded := request(t, "GET", url+"/api/files/"+metadata.Files[0].ID+"?token="+s.Token(), "", nil, "", "")
	got, err := io.ReadAll(downloaded.Body)
	if err != nil || downloaded.StatusCode != 200 || !bytes.Equal(got, data) || !strings.HasPrefix(downloaded.Header.Get("Content-Disposition"), "attachment") {
		t.Fatal("download mismatch", downloaded.StatusCode, err)
	}
	for _, bad := range []string{"../document.bin", "..%2fdocument.bin", "not-a-selected-id"} {
		response = request(t, "GET", url+"/api/files/"+bad+"?token="+s.Token(), "", nil, "", "")
		if response.StatusCode != 404 {
			t.Fatal("unselected path accepted", bad, response.StatusCode)
		}
	}
	if err = os.WriteFile(source, []byte("changed"), 0600); err != nil {
		t.Fatal(err)
	}
	response = request(t, "GET", url+"/api/files/"+metadata.Files[0].ID+"?token="+s.Token(), "", nil, "", "")
	if response.StatusCode != 409 {
		t.Fatal("changed file was offered as original metadata")
	}
}
func TestUploadIntegrityDuplicateNamesAndEmptyFile(t *testing.T) {
	var mu sync.Mutex
	var events []network.Event
	s, url, dir := startTest(t, nil, func(e network.Event) error {
		if e.Kind == "transfer" {
			mu.Lock()
			events = append(events, e)
			mu.Unlock()
		}
		return nil
	}, Options{})
	for _, data := range [][]byte{nil, {0, 255, 42, 17}, {0, 255, 42, 17}} {
		body, kind := multipartBody(t, "photo.bin", data, false)
		sum := sha256.Sum256(data)
		response := request(t, "POST", url+"/api/upload", s.Token(), body, kind, hex.EncodeToString(sum[:]))
		var receipt struct {
			SHA256, Status string
			Compared       bool `json:"senderDigestCompared"`
			Bytes          int64
		}
		if err := json.NewDecoder(response.Body).Decode(&receipt); err != nil {
			t.Fatal(err)
		}
		if response.StatusCode != 201 || receipt.Status != "verified" || !receipt.Compared || receipt.Bytes != int64(len(data)) {
			t.Fatal(response.StatusCode, receipt)
		}
		mu.Lock()
		event := events[len(events)-1]
		mu.Unlock()
		got, err := os.ReadFile(event.Path)
		if err != nil || !bytes.Equal(data, got) || event.SHA256 != hex.EncodeToString(sum[:]) {
			t.Fatal("received bytes differ", err)
		}
	}
	files, err := os.ReadDir(dir)
	if err != nil || len(files) != 3 {
		t.Fatal("duplicate filename overwrote received data", files, err)
	}
	for _, file := range files {
		if strings.HasPrefix(file.Name(), ".nearby-partial-") {
			t.Fatal("committed upload retained partial")
		}
	}
}
func TestReceiverOnlyDigestIsNotReportedSenderVerified(t *testing.T) {
	s, url, _ := startTest(t, nil, nil, Options{})
	body, kind := multipartBody(t, "document.bin", []byte("content"), false)
	response := request(t, "POST", url+"/api/upload", s.Token(), body, kind, "")
	var receipt struct {
		Status   string
		Compared bool `json:"senderDigestCompared"`
		SHA256   string
	}
	if err := json.NewDecoder(response.Body).Decode(&receipt); err != nil {
		t.Fatal(err)
	}
	if response.StatusCode != 201 || receipt.Status != "received" || receipt.Compared || len(receipt.SHA256) != 64 {
		t.Fatal(receipt)
	}
}
func TestMalformedCorruptOversizedAndExtraPartRejected(t *testing.T) {
	for _, test := range []struct {
		name, filename, digest string
		data                   []byte
		extra                  bool
		status                 int
	}{
		{"traversal", "../escape.bin", "", []byte{1}, false, 400},
		{"corrupt", "file.bin", strings.Repeat("0", 64), []byte{1}, false, 422},
		{"invalid digest", "file.bin", "bad", []byte{1}, false, 400},
		{"oversized", "file.bin", "", bytes.Repeat([]byte{1}, 65), false, 413},
		{"extra part", "file.bin", "", []byte{1}, true, 400},
	} {
		t.Run(test.name, func(t *testing.T) {
			s, url, dir := startTest(t, nil, nil, Options{MaxUploadBytes: 64})
			body, kind := multipartBody(t, test.filename, test.data, test.extra)
			response := request(t, "POST", url+"/api/upload", s.Token(), body, kind, test.digest)
			if response.StatusCode != test.status {
				raw, _ := io.ReadAll(response.Body)
				t.Fatal(response.StatusCode, string(raw))
			}
			_ = s.Close()
			noFiles(t, dir)
		})
	}
}
func TestHistoryFailureRollsBackReceivedFile(t *testing.T) {
	s, url, dir := startTest(t, nil, func(e network.Event) error {
		if e.Status == "received" || e.Status == "verified" {
			return errors.New("private/internal/failure")
		}
		return nil
	}, Options{})
	body, kind := multipartBody(t, "file.bin", []byte{1, 2, 3}, false)
	response := request(t, "POST", url+"/api/upload", s.Token(), body, kind, "")
	raw, _ := io.ReadAll(response.Body)
	if response.StatusCode != 500 || bytes.Contains(raw, []byte("private/internal")) {
		t.Fatal("history failure response", response.StatusCode, string(raw))
	}
	_ = s.Close()
	noFiles(t, dir)
}
func TestRejectCrossOriginMutation(t *testing.T) {
	s, url, dir := startTest(t, nil, nil, Options{})
	body, kind := multipartBody(t, "file.bin", []byte{1}, false)
	r, err := http.NewRequest("POST", url+"/api/upload", body)
	if err != nil {
		t.Fatal(err)
	}
	r.Header.Set("X-PulseRoute-Token", s.Token())
	r.Header.Set("Content-Type", kind)
	r.Header.Set("Origin", "https://unrelated.example")
	response, err := http.DefaultClient.Do(r)
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	if response.StatusCode != 403 {
		t.Fatal("cross-origin request accepted")
	}
	noFiles(t, dir)
}
func TestSessionExpiryAndCloseRevokeCapability(t *testing.T) {
	s, _, _ := startTest(t, nil, nil, Options{TTL: 40 * time.Millisecond})
	deadline := time.Now().Add(2 * time.Second)
	for !s.Closed() && time.Now().Before(deadline) {
		time.Sleep(5 * time.Millisecond)
	}
	if !s.Closed() {
		t.Fatal("session did not expire")
	}
	recorder := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "http://localhost/api/session", nil)
	r.Header.Set("X-PulseRoute-Token", s.Token())
	s.handle(recorder, r)
	if recorder.Code != 410 {
		t.Fatal("closed session capability remained active")
	}
	if c, err := net.DialTimeout("tcp", s.Addr(), 100*time.Millisecond); err == nil {
		c.Close()
		t.Fatal("expired listener remained open")
	}
	_ = s.Close() // idempotent
}
func TestRejectNonRegularSelectionAndLimits(t *testing.T) {
	if s, err := Start("127.0.0.1:0", t.TempDir(), []string{t.TempDir()}, nil); err == nil {
		s.Close()
		t.Fatal("directory selection accepted")
	}
	if s, err := StartWithOptions("127.0.0.1:0", t.TempDir(), nil, nil, Options{TTL: -time.Second}); err == nil {
		s.Close()
		t.Fatal("negative expiry accepted")
	}
	if s, err := StartWithOptions("127.0.0.1:0", t.TempDir(), nil, nil, Options{MaxUploadBytes: MaxUpload + 1}); err == nil {
		s.Close()
		t.Fatal("excessive upload limit accepted")
	}
}

func TestCloseCancelsActiveUploadsAndCleansPartials(t *testing.T) {
	s, url, dir := startTest(t, nil, nil, Options{})
	var writers []*io.PipeWriter
	done := make(chan error, 4)
	for i := 0; i < 4; i++ {
		reader, writer := io.Pipe()
		writers = append(writers, writer)
		form := multipart.NewWriter(writer)
		r, err := http.NewRequest("POST", url+"/api/upload", reader)
		if err != nil {
			t.Fatal(err)
		}
		r.Header.Set("Content-Type", form.FormDataContentType())
		r.Header.Set("X-PulseRoute-Token", s.Token())
		go func() {
			part, err := form.CreateFormFile("file", "partial.bin")
			if err == nil {
				_, err = part.Write(bytes.Repeat([]byte{1}, 32*1024))
			}
			if err != nil {
				_ = writer.CloseWithError(err)
			}
		}()
		go func() {
			response, err := (&http.Client{Timeout: 5 * time.Second}).Do(r)
			if response != nil {
				response.Body.Close()
			}
			done <- err
		}()
	}
	defer func() {
		for _, w := range writers {
			_ = w.Close()
		}
	}()
	deadline := time.Now().Add(3 * time.Second)
	count := 0
	for time.Now().Before(deadline) {
		files, _ := os.ReadDir(dir)
		count = len(files)
		if count == 4 {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	if count != 4 {
		t.Fatalf("expected four active streaming uploads, got %d", count)
	}
	body, kind := multipartBody(t, "fifth.bin", []byte{1}, false)
	response := request(t, "POST", url+"/api/upload", s.Token(), body, kind, "")
	if response.StatusCode != 429 {
		t.Fatal("concurrency limit not enforced", response.StatusCode)
	}
	closed := make(chan struct{})
	go func() { _ = s.Close(); close(closed) }()
	select {
	case <-closed:
	case <-time.After(2 * time.Second):
		t.Fatal("close did not cancel active uploads")
	}
	for _, w := range writers {
		_ = w.Close()
	}
	for i := 0; i < 4; i++ {
		select {
		case <-done:
		case <-time.After(2 * time.Second):
			t.Fatal("upload request leaked")
		}
	}
	noFiles(t, dir)
}
