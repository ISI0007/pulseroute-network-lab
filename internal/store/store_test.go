package store

import (
	"github.com/ISI0007/pulseroute-network-lab/internal/network"
	"path/filepath"
	"testing"
)

func TestPersistenceAndMigrations(t *testing.T) {
	p := filepath.Join(t.TempDir(), "history.db")
	s, e := Open(p)
	if e != nil {
		t.Fatal(e)
	}
	if e = s.Record(network.Event{Kind: "message", ID: "test", Text: "'; DROP TABLE events; --"}); e != nil {
		t.Fatal(e)
	}
	s.Setting("theme", "dark")
	s.Close()
	s, e = Open(p)
	if e != nil {
		t.Fatal(e)
	}
	defer s.Close()
	rows, e := s.History(10)
	if e != nil || len(rows) != 1 || rows[0].Text != "'; DROP TABLE events; --" {
		t.Fatal(rows, e)
	}
	if v, e := s.GetSetting("theme"); e != nil || v != "dark" {
		t.Fatal(v, e)
	}
	if e = s.Benchmark("run", "udp", "127.0.0.1", map[string]int{"sent": 2}, []any{1, 2}); e != nil {
		t.Fatal(e)
	}
	if e = s.Benchmark("run", "udp", "127.0.0.1", nil, nil); e == nil {
		t.Fatal("duplicate session accepted")
	}
}
