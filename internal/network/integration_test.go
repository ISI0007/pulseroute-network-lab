package network

import (
	"bytes"
	"context"
	"crypto/sha256"
	"fmt"
	"github.com/ISI0007/pulseroute-network-lab/internal/wire"
	"net"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
)

func next(t *testing.T, c *Client, op string) wire.Header {
	t.Helper()
	deadline := time.After(3 * time.Second)
	for {
		select {
		case h, ok := <-c.Events:
			if !ok {
				t.Fatal("closed")
			}
			if h.Op == op {
				return h
			}
		case <-deadline:
			t.Fatalf("waiting for %s", op)
		}
	}
}
func TestMultiClientReconnect(t *testing.T) {
	s, e := Start("127.0.0.1:0", t.TempDir(), nil)
	if e != nil {
		t.Fatal(e)
	}
	defer s.Close()
	var clients []*Client
	for i := 0; i < 8; i++ {
		c, e := Connect(context.Background(), s.Addr(), fmt.Sprint("peer", i))
		if e != nil {
			t.Fatal(e)
		}
		defer c.Close()
		clients = append(clients, c)
	}
	id, e := clients[0].Send("binary-safe text: 世界")
	if e != nil {
		t.Fatal(e)
	}
	for _, c := range clients {
		h := next(t, c, "message")
		if h.ID != id {
			t.Fatal(h)
		}
	}
	if h := next(t, clients[0], "ack"); h.Count != 8 {
		t.Fatal(h)
	}
	clients[7].Close()
	c, e := Connect(context.Background(), s.Addr(), "reconnected")
	if e != nil {
		t.Fatal(e)
	}
	defer c.Close()
	clients[0].Send("after reconnect")
	if h := next(t, c, "message"); h.Text != "after reconnect" {
		t.Fatal(h)
	}
}
func TestVerifiedFiles(t *testing.T) {
	var mu sync.Mutex
	var saved []Event
	s, e := Start("127.0.0.1:0", t.TempDir(), func(e Event) error {
		if e.Kind == "transfer" {
			mu.Lock()
			saved = append(saved, e)
			mu.Unlock()
		}
		return nil
	})
	if e != nil {
		t.Fatal(e)
	}
	defer s.Close()
	for _, size := range []int{0, 1, 65537, 16 * 1024 * 1024} {
		t.Run(fmt.Sprint(size), func(t *testing.T) {
			p := filepath.Join(t.TempDir(), "sample.bin")
			data := bytes.Repeat([]byte{0, 255, 42, 17}, (size+3)/4)[:size]
			if e := os.WriteFile(p, data, 0600); e != nil {
				t.Fatal(e)
			}
			for i := 0; i < 2; i++ {
				if _, e := SendFile(context.Background(), s.Addr(), p, nil); e != nil {
					t.Fatal(e)
				}
				mu.Lock()
				ev := saved[len(saved)-1]
				mu.Unlock()
				got, e := os.ReadFile(ev.Path)
				if e != nil || !bytes.Equal(got, data) {
					t.Fatal("file mismatch", e)
				}
			}
		})
	}
}
func TestCorruptInterruptedAndTraversal(t *testing.T) {
	dir := t.TempDir()
	s, e := Start("127.0.0.1:0", dir, nil)
	if e != nil {
		t.Fatal(e)
	}
	defer s.Close()
	for _, name := range []string{"../escape", "CON", "ok.bin"} {
		if name == "CON" {
			continue
		}
		c, e := net.Dial("tcp", s.Addr())
		if e != nil {
			t.Fatal(e)
		}
		c.SetDeadline(time.Now().Add(time.Second))
		wire.Send(c, wire.Header{Op: "put", ID: ID(), Name: name, Size: 1, SHA256: strings.Repeat("0", 64)})
		h, e := wire.Receive(c)
		if e != nil {
			t.Fatal(e)
		}
		if name == "ok.bin" {
			if h.Op != "ready" {
				t.Fatal(h)
			}
			wire.Write(c, wire.Chunk, []byte{1})
			wire.Write(c, wire.End, nil)
			h, e = wire.Receive(c)
			if e != nil {
				t.Fatal(e)
			}
		}
		if h.Op != "error" {
			t.Fatal(h)
		}
		c.Close()
	}
	c, _ := net.Dial("tcp", s.Addr())
	sum := sha256.Sum256([]byte{1})
	wire.Send(c, wire.Header{Op: "put", ID: ID(), Name: "partial.bin", Size: 1, SHA256: fmt.Sprintf("%x", sum)})
	wire.Receive(c)
	c.Close()
	s.Close()
	entries, e := os.ReadDir(dir)
	if e != nil || len(entries) != 0 {
		t.Fatal("left partial files", entries, e)
	}
}
func TestCancellationAndClose(t *testing.T) {
	s, e := Start("127.0.0.1:0", t.TempDir(), nil)
	if e != nil {
		t.Fatal(e)
	}
	p := filepath.Join(t.TempDir(), "cancel.bin")
	os.WriteFile(p, []byte{1}, 0600)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, e = SendFile(ctx, s.Addr(), p, nil); e == nil {
		t.Fatal("accepted canceled transfer")
	}
	c, e := Connect(context.Background(), s.Addr(), "idle")
	if e != nil {
		t.Fatal(e)
	}
	defer c.Close()
	done := make(chan struct{})
	go func() { s.Close(); close(done) }()
	select {
	case <-done:
	case <-time.After(3 * time.Second):
		t.Fatal("shutdown blocked")
	}
}
