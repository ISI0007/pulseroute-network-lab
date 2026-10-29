package network

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/ISI0007/pulseroute-network-lab/internal/wire"
)

func rosterSize(t *testing.T, c *Client, count int) wire.Header {
	t.Helper()
	deadline := time.After(5 * time.Second)
	for {
		select {
		case h, ok := <-c.Events:
			if !ok {
				t.Fatal("connection closed while awaiting roster")
			}
			if h.Op == "peers" && len(h.Peers) == count {
				found := false
				for i, p := range h.Peers {
					if !validIdentifier(p.ID) {
						t.Fatal("invalid roster identity", p)
					}
					if i > 0 && h.Peers[i-1].ID >= p.ID {
						t.Fatal("roster IDs are not unique/sorted", h.Peers)
					}
					if p.ID == c.ID {
						found = true
					}
				}
				if !found {
					t.Fatal("own session missing from roster")
				}
				return h
			}
		case <-deadline:
			t.Fatalf("waiting for roster size %d", count)
		}
	}
}

func noSignalBeforeMessage(t *testing.T, c *Client, id string) {
	t.Helper()
	deadline := time.After(3 * time.Second)
	for {
		select {
		case h, ok := <-c.Events:
			if !ok {
				t.Fatal("connection closed")
			}
			if h.Op == "signal" {
				t.Fatal("directed signaling leaked to another session", h)
			}
			if h.Op == "message" && h.ID == id {
				return
			}
		case <-deadline:
			t.Fatal("message barrier not received")
		}
	}
}

func TestDirectedSignalIdentityAndRoster(t *testing.T) {
	var historyMu sync.Mutex
	var history []Event
	s, err := Start("127.0.0.1:0", t.TempDir(), func(e Event) error { historyMu.Lock(); history = append(history, e); historyMu.Unlock(); return nil })
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	a, err := Connect(context.Background(), s.Addr(), "Alice")
	if err != nil {
		t.Fatal(err)
	}
	defer a.Close()
	b, err := Connect(context.Background(), s.Addr(), "Bob")
	if err != nil {
		t.Fatal(err)
	}
	defer b.Close()
	c, err := Connect(context.Background(), s.Addr(), "Carol")
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	first := rosterSize(t, a, 3)
	rosterSize(t, b, 3)
	rosterSize(t, c, 3)
	if a.ID == b.ID || b.ID == c.ID || a.ID == c.ID {
		t.Fatal("duplicate session identity")
	}
	callID := ID()
	offer := `{"type":"offer","mode":"video","sdp":"v=0\r\n"}`
	// A malicious caller cannot impersonate another session or attach a roster.
	err = a.p.send(wire.Header{Op: "signal", ID: callID, From: c.ID, Peer: "Carol", To: b.ID, Text: offer, Peers: []wire.Peer{{ID: c.ID, Name: "forged"}}})
	if err != nil {
		t.Fatal(err)
	}
	received := next(t, b, "signal")
	if received.ID != callID || received.From != a.ID || received.Peer != "Alice" || received.To != b.ID || received.Text != offer || len(received.Peers) != 0 {
		t.Fatal("untrusted identity forwarded", received)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	barrier, err := a.SendConfirmed(ctx, "after directed signal")
	if err != nil {
		t.Fatal(err)
	}
	noSignalBeforeMessage(t, c, barrier.ID)
	noSignalBeforeMessage(t, a, barrier.ID)
	historyMu.Lock()
	if len(history) != 1 || history[0].Kind != "message" || history[0].Text != "after directed signal" {
		t.Error("signaling or roster was persisted", history)
	}
	historyMu.Unlock()
	c.Close()
	after := rosterSize(t, a, 2)
	if after.Revision <= first.Revision {
		t.Fatal("roster revision did not advance after departure")
	}
	for _, peer := range after.Peers {
		if peer.ID == c.ID {
			t.Fatal("departed peer retained")
		}
	}
	rosterSize(t, b, 2)
}

func TestSignalsRejectMalformedOfflineAndSelfTargets(t *testing.T) {
	s, err := Start("127.0.0.1:0", t.TempDir(), nil)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	a, err := Connect(context.Background(), s.Addr(), "same-name")
	if err != nil {
		t.Fatal(err)
	}
	defer a.Close()
	b, err := Connect(context.Background(), s.Addr(), "same-name")
	if err != nil {
		t.Fatal(err)
	}
	defer b.Close()
	rosterSize(t, a, 2)
	rosterSize(t, b, 2)
	invalid := []string{
		`{`, `null`, `{"type":"unknown"}`, `{"type":"offer","mode":"video"}`,
		`{"type":"offer","mode":"telephone","sdp":"v=0"}`,
		`{"type":"answer","sdp":"v=0","unexpected":"field"}`,
		`{"type":"hangup"} {"type":"hangup"}`,
		`{"type":"ice","candidate":{"candidate":"not-an-ICE-candidate"}}`,
		`{"type":"ice","candidate":{"candidate":"candidate:1","sdpMLineIndex":999999}}`,
		`{"type":"busy","sdp":"must not be relayed"}`,
		strings.Repeat(" ", MaxSignalBytes+1),
	}
	for _, payload := range invalid {
		id := ID()
		if err = a.p.send(wire.Header{Op: "signal", ID: id, To: b.ID, Text: payload}); err != nil {
			t.Fatal(err)
		}
		reply := next(t, a, "error")
		if reply.ID != id {
			t.Fatal("error did not preserve call correlation", reply)
		}
	}
	if err = a.SendSignal(a.ID, ID(), `{"type":"hangup"}`); err == nil {
		t.Fatal("self call accepted")
	}
	if err = a.SendSignal("../bad", ID(), `{"type":"hangup"}`); err == nil {
		t.Fatal("malformed target accepted")
	}
	if err = a.SendSignal(b.ID, "not-a-call-id", `{"type":"hangup"}`); err == nil {
		t.Fatal("malformed call ID accepted")
	}
	if err = a.SendSignal(b.ID, ID(), invalid[0]); err == nil {
		t.Fatal("malformed local payload accepted")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	barrier, err := a.SendConfirmed(ctx, "invalid signals were not delivered")
	if err != nil {
		t.Fatal(err)
	}
	noSignalBeforeMessage(t, b, barrier.ID)
	b.Close()
	rosterSize(t, a, 1)
	id := ID()
	if err = a.SendSignal(b.ID, id, `{"type":"hangup"}`); err != nil {
		t.Fatal(err)
	}
	reply := next(t, a, "error")
	if reply.ID != id || !strings.Contains(reply.Error, "offline") {
		t.Fatal("missing offline rejection", reply)
	}
}

func TestSignalEnvelopeBoundsAndCandidateValidation(t *testing.T) {
	valid := []string{
		`{"type":"offer","mode":"audio","sdp":"v=0\r\n"}`,
		`{"type":"answer","sdp":"v=0\r\n"}`,
		`{"type":"ice","candidate":{"candidate":"candidate:1 1 UDP 1 127.0.0.1 5000 typ host","sdpMid":"0","sdpMLineIndex":0,"usernameFragment":"abcd"}}`,
		`{"type":"ice","candidate":{"candidate":""}}`,
		`{"type":"reject","reason":"declined"}`, `{"type":"hangup"}`, `{"type":"busy"}`,
	}
	for _, text := range valid {
		if err := ValidateSignal(text); err != nil {
			t.Fatal(text, err)
		}
	}
	text := strings.Repeat("\n", MaxSignalBytes-len(valid[0])) + valid[0]
	if err := ValidateSignal(text); err != nil {
		t.Fatal(err)
	}
	var buffer bytes.Buffer
	if err := wire.Send(&buffer, wire.Header{Op: "signal", ID: ID(), From: ID(), To: ID(), Peer: strings.Repeat("\x00", 64), Text: text}); err != nil {
		t.Fatal("nested JSON envelope exceeds bounded frame", err)
	}
	decoded, err := wire.Receive(&buffer)
	if err != nil || decoded.Text != text {
		t.Fatal("signal frame did not round-trip", err)
	}
	var tooLong SignalPayload
	tooLong.Type = "ice"
	tooLong.Candidate = &ICECandidate{Candidate: "candidate:" + strings.Repeat("x", 4096)}
	raw, _ := json.Marshal(tooLong)
	if ValidateSignal(string(raw)) == nil {
		t.Fatal("oversized candidate accepted")
	}
}

func TestConcurrentJoinHandshakeWhileChatActive(t *testing.T) {
	s, err := Start("127.0.0.1:0", t.TempDir(), nil)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	first, err := Connect(context.Background(), s.Addr(), "existing")
	if err != nil {
		t.Fatal(err)
	}
	defer first.Close()
	go func() {
		for range first.Events {
		}
	}()
	var joins sync.WaitGroup
	for i := 0; i < 16; i++ {
		joins.Add(1)
		go func() {
			defer joins.Done()
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			c, err := Connect(ctx, s.Addr(), "joining")
			if err != nil {
				t.Error("handshake interleaved with roster/chat", err)
				return
			}
			if !validIdentifier(c.ID) {
				t.Error("no assigned identity")
			}
			c.Close()
		}()
	}
	for i := 0; i < 8; i++ {
		if _, err = first.Send("during join"); err != nil {
			t.Fatal(err)
		}
	}
	joins.Wait()
}

func TestLegacyWelcomeKeepsChatAndDisablesCalling(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()
	release := make(chan struct{})
	defer close(release)
	serverError := make(chan error, 1)
	go func() {
		conn, err := ln.Accept()
		if err != nil {
			serverError <- err
			return
		}
		defer conn.Close()
		h, err := wire.Receive(conn)
		if err != nil {
			serverError <- err
			return
		}
		if h.Op != "hello" {
			serverError <- fmt.Errorf("expected hello")
			return
		}
		if err = wire.Send(conn, wire.Header{Op: "welcome", Peer: h.Peer}); err != nil {
			serverError <- err
			return
		}
		h, err = wire.Receive(conn)
		if err != nil {
			serverError <- err
			return
		}
		if h.Op != "message" || h.Text != "legacy chat" {
			serverError <- fmt.Errorf("unexpected message")
			return
		}
		if err = wire.Send(conn, wire.Header{Op: "ack", ID: h.ID, Count: 1}); err != nil {
			serverError <- err
			return
		}
		serverError <- nil
		<-release
	}()
	c, err := Connect(context.Background(), ln.Addr().String(), "compatible")
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	if c.ID != "" {
		t.Fatal("legacy server unexpectedly assigned call identity")
	}
	if err = c.SendSignal(ID(), ID(), `{"type":"hangup"}`); err == nil || !strings.Contains(err.Error(), "does not support calling") {
		t.Fatal("legacy call not rejected", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	ack, err := c.SendConfirmed(ctx, "legacy chat")
	if err != nil || ack.Count != 1 {
		t.Fatal("legacy chat failed", ack, err)
	}
	if err = <-serverError; err != nil {
		t.Fatal(err)
	}
}
