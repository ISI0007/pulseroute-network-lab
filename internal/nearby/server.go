// Package nearby provides an explicitly enabled, temporary LAN browser file portal.
// It is not Bluetooth or Apple's AirDrop protocol. HTTP is for trusted LANs only.
package nearby

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"embed"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"mime"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/ISI0007/pulseroute-network-lab/internal/network"
)

const MaxUpload = int64(512 * 1024 * 1024)
const maxFiles = 128

//go:embed web/index.html web/app.js web/style.css
var assets embed.FS

type SharedFile struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Size int64  `json:"size"`
	Path string `json:"-"`
	info os.FileInfo
}

type Options struct {
	TTL            time.Duration
	MaxUploadBytes int64
}

type Server struct {
	listener    net.Listener
	http        *http.Server
	dir         string
	token       string
	expires     time.Time
	files       []SharedFile
	sink        network.Sink
	maxUpload   int64
	slots       chan struct{}
	mu          sync.Mutex
	closed      bool
	active      sync.WaitGroup
	done        chan struct{}
	stop        context.CancelFunc
	expiryTimer *time.Timer
	closeOnce   sync.Once
}

// Start snapshots the list of operator-selected paths; remote clients never choose paths.
func Start(addr, dir string, paths []string, sink network.Sink) (*Server, error) {
	return StartWithOptions(addr, dir, paths, sink, Options{})
}
func StartWithOptions(addr, dir string, paths []string, sink network.Sink, options Options) (*Server, error) {
	if options.TTL == 0 {
		options.TTL = 15 * time.Minute
	}
	if options.TTL < 0 || options.TTL > time.Hour {
		return nil, errors.New("nearby lifetime must be positive and at most one hour")
	}
	if options.MaxUploadBytes == 0 {
		options.MaxUploadBytes = MaxUpload
	}
	if options.MaxUploadBytes < 1 || options.MaxUploadBytes > MaxUpload {
		return nil, errors.New("nearby upload limit must be 1..512 MiB")
	}
	if len(paths) > maxFiles {
		return nil, fmt.Errorf("select at most %d nearby files", maxFiles)
	}
	files := make([]SharedFile, 0, len(paths))
	for _, path := range paths {
		info, err := os.Lstat(path)
		if err != nil {
			return nil, err
		}
		if !info.Mode().IsRegular() || !network.ValidName(info.Name()) || info.Size() > network.MaxFile {
			return nil, errors.New("select regular files with supported names and sizes")
		}
		files = append(files, SharedFile{ID: network.ID(), Name: info.Name(), Size: info.Size(), Path: path, info: info})
	}
	if err := os.MkdirAll(dir, 0700); err != nil {
		return nil, err
	}
	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		return nil, err
	}
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithCancel(context.Background())
	s := &Server{listener: listener, dir: dir, token: hex.EncodeToString(key), expires: time.Now().Add(options.TTL).UTC(), files: files, sink: sink, maxUpload: options.MaxUploadBytes, slots: make(chan struct{}, 4), done: make(chan struct{}), stop: cancel}
	s.http = &http.Server{Handler: http.HandlerFunc(s.handle), ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 15 * time.Minute, WriteTimeout: 15 * time.Minute, IdleTimeout: 30 * time.Second, MaxHeaderBytes: 16 * 1024, ErrorLog: log.New(io.Discard, "", 0), BaseContext: func(net.Listener) context.Context { return ctx }}
	go func() { defer close(s.done); _ = s.http.Serve(listener) }()
	s.mu.Lock()
	s.expiryTimer = time.AfterFunc(options.TTL, func() { _ = s.Close() })
	s.mu.Unlock()
	return s, nil
}
func (s *Server) Addr() string         { return s.listener.Addr().String() }
func (s *Server) Token() string        { return s.token }
func (s *Server) ExpiresAt() time.Time { return s.expires }
func (s *Server) Files() []SharedFile {
	out := make([]SharedFile, len(s.files))
	copy(out, s.files)
	return out
}
func (s *Server) Closed() bool { s.mu.Lock(); defer s.mu.Unlock(); return s.closed }
func (s *Server) Close() error {
	s.closeOnce.Do(func() {
		s.mu.Lock()
		s.closed = true
		timer := s.expiryTimer
		s.mu.Unlock()
		if timer != nil {
			timer.Stop()
		}
		s.stop()
		_ = s.http.Close()
		<-s.done
		s.active.Wait()
	})
	return nil
}
func (s *Server) emit(event network.Event) error {
	if s.sink == nil {
		return nil
	}
	event.At = time.Now().UTC()
	return s.sink(event)
}
func (s *Server) handle(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	if s.closed {
		s.mu.Unlock()
		http.Error(w, "Nearby session closed", http.StatusGone)
		return
	}
	s.active.Add(1)
	s.mu.Unlock()
	defer s.active.Done()
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Referrer-Policy", "no-referrer")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("X-Frame-Options", "DENY")
	w.Header().Set("Cross-Origin-Resource-Policy", "same-origin")
	w.Header().Set("Permissions-Policy", "camera=(), microphone=(), geolocation=()")
	w.Header().Set("Content-Security-Policy", "default-src 'none'; script-src 'self'; style-src 'self'; img-src 'self' data:; connect-src 'self'; base-uri 'none'; form-action 'self'; frame-ancestors 'none'")
	if !time.Now().Before(s.expires) {
		http.Error(w, "Nearby session expired", http.StatusGone)
		return
	}
	if r.URL.Path == "/" || r.URL.Path == "/app.js" || r.URL.Path == "/style.css" {
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		name := "index.html"
		kind := "text/html; charset=utf-8"
		if r.URL.Path == "/app.js" {
			name = "app.js"
			kind = "text/javascript; charset=utf-8"
		}
		if r.URL.Path == "/style.css" {
			name = "style.css"
			kind = "text/css; charset=utf-8"
		}
		data, err := assets.ReadFile("web/" + name)
		if err != nil {
			http.Error(w, "Portal unavailable", 500)
			return
		}
		w.Header().Set("Content-Type", kind)
		if r.Method != http.MethodHead {
			_, _ = w.Write(data)
		}
		return
	}
	token := r.Header.Get("X-PulseRoute-Token")
	if r.Method == http.MethodGet && strings.HasPrefix(r.URL.Path, "/api/files/") {
		token = r.URL.Query().Get("token")
	}
	if subtle.ConstantTimeCompare([]byte(token), []byte(s.token)) != 1 {
		http.Error(w, "This nearby link is invalid or expired", http.StatusUnauthorized)
		return
	}
	if origin := r.Header.Get("Origin"); origin != "" && origin != "http://"+r.Host {
		http.Error(w, "Cross-origin requests are not allowed", http.StatusForbidden)
		return
	}
	switch {
	case r.URL.Path == "/api/session" && r.Method == http.MethodGet:
		jsonResponse(w, http.StatusOK, map[string]any{"files": s.Files(), "expiresAt": s.expires, "maxUploadBytes": s.maxUpload, "transport": "HTTP on trusted LAN; not encrypted", "integrity": "Uploads compare a sender digest only when one is supplied. Downloads are attachments, not receipt-verified."})
	case r.URL.Path == "/api/upload" && r.Method == http.MethodPost:
		s.upload(w, r)
	case strings.HasPrefix(r.URL.Path, "/api/files/") && r.Method == http.MethodGet:
		s.download(w, r, strings.TrimPrefix(r.URL.Path, "/api/files/"))
	default:
		http.NotFound(w, r)
	}
}
func jsonResponse(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}
func (s *Server) download(w http.ResponseWriter, r *http.Request, id string) {
	var selected *SharedFile
	for i := range s.files {
		if s.files[i].ID == id {
			selected = &s.files[i]
			break
		}
	}
	if selected == nil {
		http.NotFound(w, r)
		return
	}
	info, err := os.Lstat(selected.Path)
	if err != nil || !info.Mode().IsRegular() || !os.SameFile(info, selected.info) || info.Size() != selected.Size || !info.ModTime().Equal(selected.info.ModTime()) {
		http.Error(w, "Shared file changed or is no longer available; restart sharing", http.StatusConflict)
		return
	}
	file, err := os.Open(selected.Path)
	if err != nil {
		http.Error(w, "Shared file is unavailable", http.StatusGone)
		return
	}
	defer file.Close()
	opened, err := file.Stat()
	if err != nil || !opened.Mode().IsRegular() || !os.SameFile(opened, info) || opened.Size() != selected.Size || !opened.ModTime().Equal(selected.info.ModTime()) {
		http.Error(w, "Shared file is unavailable", http.StatusConflict)
		return
	}
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Disposition", mime.FormatMediaType("attachment", map[string]string{"filename": selected.Name}))
	w.Header().Set("Content-Length", fmt.Sprint(selected.Size))
	// Stream only the selected initial size. A changed file is not claimed verified.
	_, _ = io.CopyN(w, file, selected.Size)
}
func (s *Server) upload(w http.ResponseWriter, r *http.Request) {
	if r.ContentLength > s.maxUpload+64*1024 {
		http.Error(w, "File exceeds nearby upload limit", http.StatusRequestEntityTooLarge)
		return
	}
	select {
	case s.slots <- struct{}{}:
		defer func() { <-s.slots }()
	default:
		w.Header().Set("Retry-After", "2")
		http.Error(w, "Four uploads are already active; retry shortly", http.StatusTooManyRequests)
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, s.maxUpload+64*1024)
	expected := strings.ToLower(r.Header.Get("X-PulseRoute-SHA256"))
	if expected != "" {
		decoded, err := hex.DecodeString(expected)
		if err != nil || len(decoded) != 32 {
			http.Error(w, "Expected SHA-256 must contain 64 hexadecimal characters", http.StatusBadRequest)
			return
		}
	}
	reader, err := r.MultipartReader()
	if err != nil {
		http.Error(w, "Expected a multipart file upload", http.StatusBadRequest)
		return
	}
	part, err := reader.NextPart()
	if err != nil {
		http.Error(w, "Select one file to upload", http.StatusBadRequest)
		return
	}
	defer part.Close()
	_, params, err := mime.ParseMediaType(part.Header.Get("Content-Disposition"))
	name := params["filename"]
	if err != nil || part.FormName() != "file" || !network.ValidName(name) {
		http.Error(w, "Use a file name of 1..120 ASCII letters, digits, spaces, dots, hyphens or underscores", http.StatusBadRequest)
		return
	}
	id := network.ID()
	event := network.Event{Kind: "transfer", ID: id, Peer: "Nearby browser", Name: name}
	completed := false
	failure := "Upload interrupted or could not be stored"
	defer func() {
		if !completed {
			event.Status = "failed"
			if r.Context().Err() != nil {
				event.Status = "cancelled"
			}
			event.Text = failure
			_ = s.emit(event)
		}
	}()
	file, err := os.CreateTemp(s.dir, ".nearby-partial-")
	if err != nil {
		http.Error(w, "Unable to store file on this computer", http.StatusInsufficientStorage)
		return
	}
	temporary := file.Name()
	defer os.Remove(temporary)
	defer file.Close()
	hash := sha256.New()
	limited := io.LimitReader(part, s.maxUpload+1)
	buffer := make([]byte, 32*1024)
	var total int64
	last := time.Now()
	for {
		if r.Context().Err() != nil {
			failure = "Upload cancelled"
			return
		}
		n, readErr := limited.Read(buffer)
		if n > 0 {
			total += int64(n)
			event.Bytes = total
			if total > s.maxUpload {
				failure = "File exceeds nearby upload limit"
				http.Error(w, failure, http.StatusRequestEntityTooLarge)
				return
			}
			if _, err = file.Write(buffer[:n]); err != nil {
				failure = "Unable to store file on this computer"
				http.Error(w, failure, http.StatusInsufficientStorage)
				return
			}
			_, _ = hash.Write(buffer[:n])
			if time.Since(last) > 150*time.Millisecond {
				_ = s.emit(network.Event{Kind: "progress", ID: id, Peer: "Nearby browser", Name: name, Bytes: total, Status: "receiving"})
				last = time.Now()
			}
		}
		if readErr == io.EOF {
			break
		}
		if readErr != nil {
			http.Error(w, failure, http.StatusBadRequest)
			return
		}
	}
	if _, err = reader.NextPart(); err != io.EOF {
		failure = "Upload exactly one file per request"
		http.Error(w, failure, http.StatusBadRequest)
		return
	}
	digest := hex.EncodeToString(hash.Sum(nil))
	if expected != "" && digest != expected {
		failure = "SHA-256 did not match the sender"
		http.Error(w, failure, http.StatusUnprocessableEntity)
		return
	}
	if err = file.Sync(); err != nil {
		failure = "Unable to sync received file"
		http.Error(w, failure, http.StatusInsufficientStorage)
		return
	}
	if err = file.Close(); err != nil {
		http.Error(w, failure, http.StatusInsufficientStorage)
		return
	}
	destination := filepath.Join(s.dir, network.ID()+"-"+name)
	if err = os.Rename(temporary, destination); err != nil {
		http.Error(w, failure, http.StatusInsufficientStorage)
		return
	}
	event.Path = destination
	event.SHA256 = digest
	event.Total = total
	event.Status = "received"
	if expected != "" {
		event.Status = "verified"
	} else {
		event.Text = "Receiver SHA-256 calculated; no sender digest supplied for comparison"
	}
	if err = s.emit(event); err != nil {
		_ = os.Remove(destination)
		failure = "History unavailable; upload not committed"
		http.Error(w, failure, http.StatusInternalServerError)
		return
	}
	completed = true
	jsonResponse(w, http.StatusCreated, map[string]any{"id": id, "name": name, "bytes": total, "sha256": digest, "senderDigestCompared": expected != "", "status": event.Status})
}
