package nearby

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"
)

// Opt-in integration: real HTTP server, real browser UI, no bridge or API mocks.
func TestBrowserPortal(t *testing.T) {
	if os.Getenv("PULSEROUTE_BROWSER_TEST") != "1" {
		t.Skip("requires installed Chrome and npm dependencies; enabled in Windows CI")
	}
	root, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	selected := filepath.Join(t.TempDir(), "offered.txt")
	if err = os.WriteFile(selected, []byte("PulseRoute browser download fixture\n"), 0600); err != nil {
		t.Fatal(err)
	}
	var received []string
	// Independently verify files on disk after the browser completes.
	server, err := Start("127.0.0.1:0", t.TempDir(), []string{selected}, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer server.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	cmd := exec.CommandContext(ctx, "node", filepath.Join(root, "scripts", "check-nearby.cjs"))
	cmd.Env = append(os.Environ(), "PULSEROUTE_TEST_URL=http://"+server.Addr()+"/#"+server.Token())
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("browser integration: %v\n%s", err, output)
	}
	t.Log(string(output))
	entries, err := os.ReadDir(server.dir)
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		if !entry.IsDir() {
			received = append(received, filepath.Join(server.dir, entry.Name()))
		}
	}
	if len(received) != 1 {
		t.Fatalf("expected exactly one completed upload, got %d", len(received))
	}
	data, err := os.ReadFile(received[0])
	if err != nil {
		t.Fatal(err)
	}
	want := []byte("PulseRoute browser upload fixture\n")
	if !bytes.Equal(data, want) {
		t.Fatal("uploaded bytes differ")
	}
	hash := sha256.Sum256(data)
	t.Log("PASS: independent received-file SHA-256", hex.EncodeToString(hash[:]))
}
