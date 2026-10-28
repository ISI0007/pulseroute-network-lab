package network

import (
	"bytes"
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestLargeFileStream128MiB(t *testing.T) {
	const size = int64(128 * 1024 * 1024)
	source := filepath.Join(t.TempDir(), "large-video.bin")
	f, err := os.Create(source)
	if err != nil {
		t.Fatal(err)
	}
	block := bytes.Repeat([]byte{0, 255, 17, 42}, 16*1024)
	expected := sha256.New()
	for written := int64(0); written < size; written += int64(len(block)) {
		if _, err = f.Write(block); err != nil {
			f.Close()
			t.Fatal(err)
		}
		expected.Write(block)
	}
	if err = f.Close(); err != nil {
		t.Fatal(err)
	}
	received := make(chan Event, 1)
	s, err := Start("127.0.0.1:0", t.TempDir(), func(e Event) error {
		if e.Kind == "transfer" && e.Status == "verified" {
			received <- e
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	header, err := SendFile(ctx, s.Addr(), source, nil)
	if err != nil {
		t.Fatal(err)
	}
	ev := <-received
	if header.Size != size || ev.Bytes != size {
		t.Fatal("incorrect streamed length")
	}
	actual := sha256.New()
	saved, err := os.Open(ev.Path)
	if err != nil {
		t.Fatal(err)
	}
	defer saved.Close()
	n, err := io.Copy(actual, saved)
	if err != nil || n != size || !bytes.Equal(actual.Sum(nil), expected.Sum(nil)) {
		t.Fatal("128 MiB received file differs from source", n, err)
	}
	if ev.SHA256 != fmt.Sprintf("%x", expected.Sum(nil)) {
		t.Fatal("wrong stored digest")
	}
}

func TestCancelActiveUploadRemovesPartial(t *testing.T) {
	dir := t.TempDir()
	s, err := Start("127.0.0.1:0", dir, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	source := filepath.Join(t.TempDir(), "cancel.bin")
	if err = os.WriteFile(source, bytes.Repeat([]byte{42}, 1024*1024), 0600); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	reachedSend := false
	_, err = SendFile(ctx, s.Addr(), source, func(e Event) {
		if e.Status == "sending" {
			reachedSend = true
			cancel()
		}
	})
	if !reachedSend || err == nil {
		t.Fatal("active upload did not cancel", err)
	}
	s.Close() // Wait for receiver cleanup before inspecting the directory.
	entries, err := os.ReadDir(dir)
	if err != nil || len(entries) != 0 {
		t.Fatal("canceled upload left files", entries, err)
	}
}

func TestHistoryFailureRejectsFileAndCleansDestination(t *testing.T) {
	dir := t.TempDir()
	s, err := Start("127.0.0.1:0", dir, func(e Event) error {
		if e.Kind == "transfer" && e.Status == "verified" {
			return errors.New("disk unavailable")
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	source := filepath.Join(t.TempDir(), "data.bin")
	if err = os.WriteFile(source, []byte{0, 255, 42}, 0600); err != nil {
		t.Fatal(err)
	}
	if _, err = SendFile(context.Background(), s.Addr(), source, nil); err == nil {
		t.Fatal("sender reported success despite history failure")
	}
	s.Close()
	entries, err := os.ReadDir(dir)
	if err != nil || len(entries) != 0 {
		t.Fatal("failed history left committed file", entries, err)
	}
}

func TestCanceledHandshakeDoesNotWaitForDeadline(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()
	accepted := make(chan net.Conn, 1)
	go func() {
		c, err := ln.Accept()
		if err == nil {
			accepted <- c
		}
	}()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan error, 1)
	go func() {
		c, err := Connect(ctx, ln.Addr().String(), "waiting")
		if c != nil {
			c.Close()
		}
		done <- err
	}()
	select {
	case c := <-accepted:
		defer c.Close()
	case <-time.After(3 * time.Second):
		t.Fatal("no connection")
	}
	cancel()
	select {
	case err := <-done:
		if err == nil {
			t.Fatal("canceled handshake succeeded")
		}
	case <-time.After(time.Second):
		t.Fatal("handshake ignored cancellation")
	}
}
