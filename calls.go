//go:build desktop

package main

import (
	"encoding/hex"
	"errors"
	"fmt"

	"github.com/ISI0007/pulseroute-network-lab/internal/network"
	wr "github.com/wailsapp/wails/v2/pkg/runtime"
)

type CallEnvelope struct {
	Seq   uint64 `json:"seq"`
	From  string `json:"from"`
	Peer  string `json:"peer"`
	ID    string `json:"id"`
	Text  string `json:"text"`
	Error string `json:"error,omitempty"`
}

// SDP and ICE are ephemeral; only bounded in-memory signaling reaches the UI.
func (a *App) pushCall(e CallEnvelope)                             { a.publishCall(nil, e) }
func (a *App) pushCallFrom(source *network.Client, e CallEnvelope) { a.publishCall(source, e) }
func (a *App) publishCall(source *network.Client, e CallEnvelope) {
	a.mu.Lock()
	if a.closing || (source != nil && a.client != source) || (source == nil && e.ID == "" && e.Error != "" && a.client != nil) {
		a.mu.Unlock()
		return
	}
	a.callSeq++
	e.Seq = a.callSeq
	a.state.CallSignals = append(a.state.CallSignals, e)
	if len(a.state.CallSignals) > 128 {
		a.state.CallSignals = append([]CallEnvelope{}, a.state.CallSignals[len(a.state.CallSignals)-128:]...)
	}
	ctx := a.ctx
	a.mu.Unlock()
	if ctx != nil {
		wr.EventsEmit(ctx, "pulseroute:call", e)
	}
}
func (a *App) SendCallSignal(to, callID, text string) error {
	if err := a.beginOperation(); err != nil {
		return err
	}
	defer a.wg.Done()
	a.mu.Lock()
	client := a.client
	a.mu.Unlock()
	if client == nil {
		return errors.New("connect to a PulseRoute hub before calling")
	}
	return client.SendSignal(to, callID, text)
}
func (a *App) RecordCallSummary(id, peer, mode, status string, durationSeconds int) error {
	if err := a.beginOperation(); err != nil {
		return err
	}
	defer a.wg.Done()
	decoded, err := hex.DecodeString(id)
	if err != nil || len(decoded) != 16 || len(peer) > 64 || (mode != "audio" && mode != "video") || durationSeconds < 0 || durationSeconds > 86400 {
		return errors.New("invalid call summary")
	}
	switch status {
	case "ended", "rejected", "failed", "missed":
	default:
		return errors.New("invalid call outcome")
	}
	return a.record(network.Event{Kind: "call", ID: id, Peer: peer, Name: mode, Status: status, Text: fmt.Sprintf("%s call; %d seconds; %s", mode, durationSeconds, status)})
}
