package network

import (
	"context"
	"errors"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestConfirmedMessagesConcurrent(t *testing.T) {
	s, err := Start("127.0.0.1:0", t.TempDir(), nil)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	sender, err := Connect(context.Background(), s.Addr(), "sender")
	if err != nil {
		t.Fatal(err)
	}
	defer sender.Close()
	receiver, err := Connect(context.Background(), s.Addr(), "receiver")
	if err != nil {
		t.Fatal(err)
	}
	defer receiver.Close()
	var wg sync.WaitGroup
	ids := make(chan string, 16)
	for i := 0; i < 16; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			ack, err := sender.SendConfirmed(ctx, "confirmed")
			if err != nil || ack.Count != 2 {
				t.Errorf("confirmation: %+v, %v", ack, err)
				return
			}
			ids <- ack.ID
		}()
	}
	wg.Wait()
	close(ids)
	expected := map[string]bool{}
	for id := range ids {
		expected[id] = true
	}
	if len(expected) != 16 {
		t.Fatalf("got %d distinct confirmations", len(expected))
	}
	for i := 0; i < 16; i++ {
		message := next(t, receiver, "message")
		if !expected[message.ID] {
			t.Fatal("confirmation did not correlate to received message", message)
		}
		delete(expected, message.ID)
	}
}

func TestConfirmedMessageRejectsHistoryFailure(t *testing.T) {
	s, err := Start("127.0.0.1:0", t.TempDir(), func(e Event) error {
		if e.Kind == "message" {
			return errors.New("history unavailable")
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	c, err := Connect(context.Background(), s.Addr(), "sender")
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	ack, err := c.SendConfirmed(ctx, "must not claim success")
	if err == nil || !strings.Contains(err.Error(), "server rejected") || ack.Op != "error" {
		t.Fatal("history rejection not propagated", ack, err)
	}
	c.pendingMu.Lock()
	n := len(c.pending)
	c.pendingMu.Unlock()
	if n != 0 {
		t.Fatal("pending confirmation leaked")
	}
}

func TestConfirmedMessageCanceledBeforeSend(t *testing.T) {
	s, err := Start("127.0.0.1:0", t.TempDir(), nil)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	c, err := Connect(context.Background(), s.Addr(), "sender")
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err = c.SendConfirmed(ctx, "canceled"); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
}
