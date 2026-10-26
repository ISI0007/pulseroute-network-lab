package network

import (
	"context"
	"testing"
)

func TestDiscovery(t *testing.T) {
	d, e := Advertise("127.0.0.1:0", "test-peer", 42191)
	if e != nil {
		t.Fatal(e)
	}
	defer d.Close()
	p, e := Discover(context.Background(), d.Addr())
	if e != nil || len(p) != 1 || p[0].Name != "test-peer" || p[0].Address != "127.0.0.1:42191" {
		t.Fatal(p, e)
	}
}
