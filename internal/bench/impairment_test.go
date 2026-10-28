package bench

import (
	"context"
	"encoding/binary"
	"math"
	"net"
	"testing"
	"time"
)

// This local fixture changes actual datagram delivery; it does not substitute
// synthetic Result values for network measurements.
func udpFixture(t *testing.T, reply func(net.PacketConn, []byte, net.Addr)) string {
	t.Helper()
	c, err := net.ListenPacket("udp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	done := make(chan struct{})
	go func() {
		defer close(done)
		buf := make([]byte, 2048)
		for {
			n, addr, err := c.ReadFrom(buf)
			if err != nil {
				return
			}
			reply(c, append([]byte(nil), buf[:n]...), addr)
		}
	}()
	t.Cleanup(func() { c.Close(); <-done })
	return c.LocalAddr().String()
}

func TestUDPLossDuplicateAndLateObservations(t *testing.T) {
	var first, delayed []byte
	addr := udpFixture(t, func(c net.PacketConn, b []byte, a net.Addr) {
		switch binary.BigEndian.Uint32(b[4:8]) {
		case 0:
			first = b
			c.WriteTo(b, a)
		case 1:
			delayed = b // No reply until after probe 2 has completed.
		case 2:
			c.WriteTo(first, a) // Duplicate of already observed probe 0.
			c.WriteTo(b, a)
		case 3:
			c.WriteTo(delayed, a) // Probe 1 arrives out of order.
			c.WriteTo(b, a)
		}
	})
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	r, err := Run(ctx, "udp", addr, 4, 64)
	if err != nil {
		t.Fatal(err)
	}
	if r.Received != 4 || r.LossPercent != 0 || r.Duplicates != 1 || r.OutOfOrder != 1 {
		t.Fatalf("incorrect observations: %+v", r)
	}
	if len(r.Samples) != 6 || r.Samples[1].Received {
		t.Fatalf("missing timed-out sample or late observations: %+v", r.Samples)
	}
	if r.RTTMS <= 0 || r.Mbps <= 0 {
		t.Fatal("missing measured metrics", r)
	}
}

func TestUDPRejectsAlteredEchoAndMeasuresLoss(t *testing.T) {
	addr := udpFixture(t, func(c net.PacketConn, b []byte, a net.Addr) {
		if binary.BigEndian.Uint32(b[4:8]) == 0 {
			b[8] ^= 0xff // Correct sequence/magic, incorrect experiment cookie.
			c.WriteTo(b, a)
		} else {
			c.WriteTo(b, a)
		}
	})
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	r, err := Run(ctx, "udp", addr, 3, 64)
	if err != nil {
		t.Fatal(err)
	}
	if r.Received != 2 || math.Abs(r.LossPercent-100.0/3) > 0.000001 || r.Samples[0].Received {
		t.Fatalf("altered reply counted as delivered: %+v", r)
	}
}

func TestUDPCancellationUnblocksRead(t *testing.T) {
	received := make(chan struct{}, 1)
	addr := udpFixture(t, func(_ net.PacketConn, _ []byte, _ net.Addr) {
		select {
		case received <- struct{}{}:
		default:
		}
	})
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan error, 1)
	go func() { _, err := Run(ctx, "udp", addr, 100, 64); done <- err }()
	select {
	case <-received:
	case <-time.After(3 * time.Second):
		t.Fatal("no probe sent")
	}
	cancel()
	select {
	case err := <-done:
		if err == nil {
			t.Fatal("canceled run succeeded")
		}
	case <-time.After(time.Second):
		t.Fatal("cancellation did not unblock socket read")
	}
}
