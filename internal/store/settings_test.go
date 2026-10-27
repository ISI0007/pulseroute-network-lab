package store

import (
	"fmt"
	"github.com/ISI0007/pulseroute-network-lab/internal/bench"
	"path/filepath"
	"testing"
)

func TestRestoreSettingsPeersAndMeasurements(t *testing.T) {
	p := filepath.Join(t.TempDir(), "restore.db")
	s, err := Open(p)
	if err != nil {
		t.Fatal(err)
	}
	original := DefaultSettings()
	original.Name = "peer-two"
	original.Target = "[::1]:42191"
	if err = s.SaveSettings(original); err != nil {
		t.Fatal(err)
	}
	invalid := original
	invalid.Target = "127.0.0.1:70000"
	if err = s.SaveSettings(invalid); err == nil {
		t.Fatal("invalid endpoint replaced valid settings")
	}
	if err = s.Peer("peer-two", original.Target); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 25; i++ {
		r := bench.Result{Protocol: "udp", Sent: i, Samples: []bench.Sample{{Sequence: 0, RTTMS: float64(i), Received: true}}}
		if err = s.Benchmark(fmt.Sprint(i), "udp", original.Target, r, []any{r.Samples[0]}); err != nil {
			t.Fatal(err)
		}
	}
	if err = s.Close(); err != nil {
		t.Fatal(err)
	}
	s, err = Open(p)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	restored, err := s.LoadSettings()
	if err != nil || restored != original {
		t.Fatalf("settings: %+v, %v", restored, err)
	}
	peers, err := s.KnownPeers()
	if err != nil || len(peers) != 1 || peers[0].Address != original.Target {
		t.Fatalf("peers: %+v, %v", peers, err)
	}
	results, err := s.Results()
	if err != nil || len(results) != 20 || results[0].Sent != 5 || results[19].Sent != 24 || results[19].Samples[0].RTTMS != 24 {
		t.Fatalf("results: %+v, %v", results, err)
	}
}
