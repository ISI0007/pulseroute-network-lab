package network

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"sort"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"github.com/ISI0007/pulseroute-network-lab/internal/wire"
)

// MaxSignalBytes leaves enough room for worst-case nested JSON escaping in a
// 128 KiB control frame. Media bytes never pass through the signaling hub.
const MaxSignalBytes = 60 * 1024
const maxSignalsPerSecond = 64

type ICECandidate struct {
	Candidate        string  `json:"candidate"`
	SDPMid           *string `json:"sdpMid,omitempty"`
	SDPMLineIndex    *uint16 `json:"sdpMLineIndex,omitempty"`
	UsernameFragment *string `json:"usernameFragment,omitempty"`
}

type SignalPayload struct {
	Type      string        `json:"type"`
	Mode      string        `json:"mode,omitempty"`
	SDP       string        `json:"sdp,omitempty"`
	Candidate *ICECandidate `json:"candidate,omitempty"`
	Reason    string        `json:"reason,omitempty"`
}

func validIdentifier(id string) bool {
	if len(id) != 32 {
		return false
	}
	decoded, err := hex.DecodeString(id)
	return err == nil && len(decoded) == 16
}

// ValidateSignal bounds and checks the signaling envelope without interpreting
// SDP or claiming peer authentication. Browser WebRTC validates media syntax.
func ValidateSignal(text string) error {
	if len(text) == 0 || len(text) > MaxSignalBytes || !utf8.ValidString(text) {
		return errors.New("signal payload must contain 1..61440 UTF-8 bytes")
	}
	var p SignalPayload
	decoder := json.NewDecoder(strings.NewReader(text))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&p); err != nil {
		return errors.New("invalid signal JSON")
	}
	if err := decoder.Decode(new(any)); err != io.EOF {
		return errors.New("signal must contain one JSON object")
	}
	if len(p.Reason) > 160 {
		return errors.New("signal reason is too long")
	}
	switch p.Type {
	case "offer", "answer":
		if p.SDP == "" || p.Candidate != nil || p.Reason != "" {
			return errors.New("invalid session description signal")
		}
		if p.Type == "offer" && p.Mode != "audio" && p.Mode != "video" {
			return errors.New("offer mode must be audio or video")
		}
		if p.Mode != "" && p.Mode != "audio" && p.Mode != "video" {
			return errors.New("invalid call mode")
		}
	case "ice":
		if p.Candidate == nil || p.SDP != "" || p.Mode != "" || p.Reason != "" {
			return errors.New("invalid ICE signal")
		}
		c := p.Candidate
		if len(c.Candidate) > 4096 || (c.Candidate != "" && !strings.HasPrefix(c.Candidate, "candidate:")) ||
			(c.SDPMid != nil && len(*c.SDPMid) > 256) || (c.UsernameFragment != nil && len(*c.UsernameFragment) > 256) {
			return errors.New("invalid ICE candidate")
		}
	case "reject", "hangup", "busy":
		if p.SDP != "" || p.Mode != "" || p.Candidate != nil {
			return errors.New("invalid call control signal")
		}
	default:
		return errors.New("unsupported signal type")
	}
	return nil
}

func (c *Client) SendSignal(to, callID, text string) error {
	if c.ID == "" {
		return errors.New("server does not support calling")
	}
	if !validIdentifier(to) || !validIdentifier(callID) || to == c.ID {
		return errors.New("invalid signal recipient or call ID")
	}
	if err := ValidateSignal(text); err != nil {
		return err
	}
	return c.p.send(wire.Header{Op: "signal", ID: callID, To: to, Text: text})
}

func (s *Server) routeSignal(sender *session, h wire.Header) error {
	if !validIdentifier(h.ID) || !validIdentifier(h.To) || h.To == sender.id {
		return errors.New("invalid signal recipient or call ID")
	}
	if err := ValidateSignal(h.Text); err != nil {
		return err
	}
	now := time.Now()
	if sender.signalWindow.IsZero() || now.Sub(sender.signalWindow) >= time.Second {
		sender.signalWindow, sender.signalCount = now, 0
	}
	sender.signalCount++
	if sender.signalCount > maxSignalsPerSecond {
		return errors.New("signal rate limit exceeded")
	}
	s.mu.Lock()
	var recipient *session
	for p := range s.clients {
		if p.id == h.To {
			recipient = p
			break
		}
	}
	s.mu.Unlock()
	if recipient == nil {
		return errors.New("call recipient is offline")
	}
	// Rebuild rather than forward client-supplied identity/roster fields.
	if err := recipient.send(wire.Header{Op: "signal", ID: h.ID, From: sender.id, To: recipient.id, Peer: sender.name, Text: h.Text}); err != nil {
		recipient.conn.Close()
		return errors.New("call recipient is unavailable")
	}
	return nil
}

func (s *Server) broadcastRoster() {
	// Serialize snapshots so that an older roster can never follow a newer one.
	// Writes are parallel and bounded by existing per-session deadlines.
	s.rosterMu.Lock()
	defer s.rosterMu.Unlock()
	s.mu.Lock()
	if s.closed {
		s.mu.Unlock()
		return
	}
	recipients := make([]*session, 0, len(s.clients))
	roster := make([]wire.Peer, 0, len(s.clients))
	for p := range s.clients {
		recipients = append(recipients, p)
		roster = append(roster, wire.Peer{ID: p.id, Name: p.name})
	}
	revision := s.rosterRevision
	s.mu.Unlock()
	sort.Slice(roster, func(i, j int) bool { return roster[i].ID < roster[j].ID })
	message := wire.Header{Op: "peers", Peers: roster, Revision: revision}
	var writes sync.WaitGroup
	for _, p := range recipients {
		writes.Add(1)
		go func(p *session) {
			defer writes.Done()
			if p.send(message) != nil {
				p.conn.Close()
			}
		}(p)
	}
	writes.Wait()
}
