package bench

import (
	"context"
	"github.com/ISI0007/pulseroute-network-lab/internal/network"
	"testing"
	"time"
)

func TestMetrics(t *testing.T) {
	r := Summarize("udp", "synthetic", 4, time.Second, []Sample{{Sequence: 0, Received: true, RTTMS: 10, Bytes: 100}, {Sequence: 2, Received: true, RTTMS: 14, Bytes: 100}, {Sequence: 2, Received: true}, {Sequence: 1, Received: true, RTTMS: 12, Bytes: 100}})
	if r.Received != 3 || r.Duplicates != 1 || r.OutOfOrder != 1 || r.LossPercent != 25 || r.RTTMS != 12 || r.Mbps != 0.0024 {
		t.Fatal(r)
	}
}
func TestEchoLoopback(t *testing.T) {
	s, e := network.Start("127.0.0.1:0", t.TempDir(), nil)
	if e != nil {
		t.Fatal(e)
	}
	defer s.Close()
	u, e := StartEcho("127.0.0.1:0")
	if e != nil {
		t.Fatal(e)
	}
	defer u.Close()
	for protocol, addr := range map[string]string{"tcp": s.Addr(), "udp": u.Addr()} {
		r, e := Run(context.Background(), protocol, addr, 20, 256)
		if e != nil || r.Received != 20 || r.LossPercent != 0 {
			t.Fatal(r, e)
		}
	}
}
