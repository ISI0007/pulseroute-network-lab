//go:build desktop

package main

import (
	"context"
	"fmt"
	"testing"
)

func TestCallSignalRetentionAndSnapshotIsolation(t *testing.T) {
	a, err := newApp(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	defer a.shutdown(context.Background())
	for i := 0; i < 200; i++ {
		a.pushCall(CallEnvelope{ID: fmt.Sprint(i), Text: "ephemeral SDP"})
	}
	snapshot, err := a.Snapshot()
	if err != nil {
		t.Fatal(err)
	}
	if len(snapshot.CallSignals) != 128 || snapshot.CallSignals[0].Seq != 73 || snapshot.CallSignals[127].Seq != 200 {
		t.Fatal("incorrect bounded signaling queue")
	}
	snapshot.CallSignals[0].Text = "modified snapshot"
	again, err := a.Snapshot()
	if err != nil || again.CallSignals[0].Text != "ephemeral SDP" {
		t.Fatal("snapshot exposed mutable queue", err)
	}
	if err = a.SendCallSignal("", "", ""); err == nil {
		t.Fatal("calling without a connection succeeded")
	}
}

func TestCallSummaryValidation(t *testing.T) {
	a, err := newApp(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	defer a.shutdown(context.Background())
	id := "0123456789abcdef0123456789abcdef"
	if err = a.RecordCallSummary(id, "peer", "audio", "ended", 12); err != nil {
		t.Fatal(err)
	}
	for _, test := range []struct {
		id, mode, status string
		seconds          int
	}{
		{"bad", "audio", "ended", 1}, {id, "unknown", "ended", 1}, {id, "video", "connected", 1}, {id, "audio", "ended", -1}, {id, "video", "ended", 86401},
	} {
		if a.RecordCallSummary(test.id, "peer", test.mode, test.status, test.seconds) == nil {
			t.Fatal("invalid call summary accepted", test)
		}
	}
}
