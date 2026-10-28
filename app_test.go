//go:build desktop

package main

import (
	"context"
	"fmt"
	"github.com/ISI0007/pulseroute-network-lab/internal/network"
	"github.com/ISI0007/pulseroute-network-lab/internal/store"
	"sync"
	"testing"
	"time"
)

func TestProgressRetentionAndShutdown(t *testing.T) {
	a, err := newApp(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 200; i++ {
		id := fmt.Sprint(i)
		if err = a.record(network.Event{Kind: "progress", ID: id, Status: "sending", At: time.Now()}); err != nil {
			t.Fatal(err)
		}
		if err = a.record(network.Event{Kind: "transfer", ID: id, Status: "sent-verified", At: time.Now()}); err != nil {
			t.Fatal(err)
		}
	}
	snapshot, err := a.Snapshot()
	if err != nil || len(snapshot.Progress) != 128 {
		t.Fatalf("progress retained: %d, error: %v", len(snapshot.Progress), err)
	}
	a.shutdown(context.Background())
	if _, err = a.Snapshot(); err == nil {
		t.Fatal("snapshot allowed after database shutdown")
	}
	if err = a.SaveSettings(store.DefaultSettings()); err == nil {
		t.Fatal("settings write allowed after shutdown")
	}
}
func TestDesktopSettingsRestart(t *testing.T) {
	dir := t.TempDir()
	a, err := newApp(dir)
	if err != nil {
		t.Fatal(err)
	}
	settings := store.DefaultSettings()
	settings.Name = "saved-name"
	if err = a.SaveSettings(settings); err != nil {
		t.Fatal(err)
	}
	a.shutdown(context.Background())
	a, err = newApp(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer a.shutdown(context.Background())
	snapshot, err := a.Snapshot()
	if err != nil || snapshot.Settings.Name != "saved-name" {
		t.Fatal(snapshot.Settings, err)
	}
}

func TestConcurrentOperationsDuringShutdown(t *testing.T) {
	a, err := newApp(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	var workers sync.WaitGroup
	start := make(chan struct{})
	for i := 0; i < 8; i++ {
		workers.Add(1)
		go func() {
			defer workers.Done()
			<-start
			for j := 0; j < 20; j++ {
				_, _ = a.Snapshot()
				_ = a.SaveSettings(store.DefaultSettings())
				_ = a.SetTheme("dark")
			}
		}()
	}
	close(start)
	a.shutdown(context.Background())
	workers.Wait()
	if _, err = a.Snapshot(); err == nil {
		t.Fatal("database reopened after shutdown")
	}
}

func TestDesktopConfirmedMessageHistory(t *testing.T) {
	a, err := newApp(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	defer a.shutdown(context.Background())
	server, err := network.Start("127.0.0.1:0", t.TempDir(), a.record)
	if err != nil {
		t.Fatal(err)
	}
	defer server.Close()
	if err = a.Connect(server.Addr(), "operator"); err != nil {
		t.Fatal(err)
	}
	id, err := a.SendMessage("persist confirmed status")
	if err != nil {
		t.Fatal(err)
	}
	snapshot, err := a.Snapshot()
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, e := range snapshot.History {
		if e.Kind == "delivery" && e.ID == id && e.Status == "accepted" && e.Count == 1 {
			found = true
		}
	}
	if !found {
		t.Fatal("missing persisted server acknowledgment", snapshot.History)
	}
}

func TestThemePalettesPersistAndRejectInvalidValues(t *testing.T) {
	dir := t.TempDir()
	a, err := newApp(dir)
	if err != nil {
		t.Fatal(err)
	}
	snapshot, err := a.Snapshot()
	if err != nil || snapshot.Theme != "forest" {
		t.Fatal("fresh theme", snapshot.Theme, err)
	}
	for _, theme := range []string{"dark", "light", "forest", "ocean"} {
		if err = a.SetTheme(theme); err != nil {
			t.Fatal(theme, err)
		}
		snapshot, err = a.Snapshot()
		if err != nil || snapshot.Theme != theme {
			t.Fatal("theme not applied", theme, snapshot.Theme, err)
		}
	}
	if err = a.SetTheme("unknown-palette"); err == nil {
		t.Fatal("invalid palette accepted")
	}
	a.shutdown(context.Background())
	a, err = newApp(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer a.shutdown(context.Background())
	snapshot, err = a.Snapshot()
	if err != nil || snapshot.Theme != "ocean" {
		t.Fatal("theme not restored or invalid write changed it", snapshot.Theme, err)
	}
}
