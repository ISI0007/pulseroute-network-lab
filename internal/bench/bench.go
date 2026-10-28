// Package bench measures sequential application echo probes, not link capacity.
package bench

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/binary"
	"errors"
	"github.com/ISI0007/pulseroute-network-lab/internal/wire"
	"math"
	"net"
	"time"
)

type Sample struct {
	Sequence int     `json:"sequence"`
	RTTMS    float64 `json:"rttMs"`
	Received bool    `json:"received"`
	Arrival  int     `json:"arrival"`
	Bytes    int     `json:"bytes"`
}
type Result struct {
	Protocol    string   `json:"protocol"`
	Address     string   `json:"address"`
	Sent        int      `json:"sent"`
	Received    int      `json:"received"`
	Duplicates  int      `json:"duplicates"`
	OutOfOrder  int      `json:"outOfOrder"`
	LossPercent float64  `json:"lossPercent"`
	RTTMS       float64  `json:"rttMs"`
	JitterMS    float64  `json:"jitterMs"`
	Mbps        float64  `json:"mbps"`
	DurationMS  float64  `json:"durationMs"`
	Samples     []Sample `json:"samples"`
}

func Summarize(protocol, addr string, sent int, elapsed time.Duration, samples []Sample) Result {
	r := Result{Protocol: protocol, Address: addr, Sent: sent, DurationMS: float64(elapsed) / float64(time.Millisecond), Samples: samples}
	seen := map[int]bool{}
	last := -1
	var prev float64
	var totalBytes int
	for _, s := range samples {
		if !s.Received || s.Sequence < 0 || s.Sequence >= sent {
			continue
		}
		if seen[s.Sequence] {
			r.Duplicates++
			continue
		}
		seen[s.Sequence] = true
		r.Received++
		r.RTTMS += s.RTTMS
		totalBytes += s.Bytes
		if s.Sequence < last {
			r.OutOfOrder++
		}
		if s.Sequence > last {
			last = s.Sequence
		}
		if r.Received > 1 {
			r.JitterMS += (math.Abs(s.RTTMS-prev) - r.JitterMS) / 16
		}
		prev = s.RTTMS
	}
	if sent > 0 {
		r.LossPercent = 100 * float64(sent-r.Received) / float64(sent)
	}
	if r.Received > 0 {
		r.RTTMS /= float64(r.Received)
	}
	if elapsed > 0 {
		r.Mbps = float64(totalBytes) * 8 / elapsed.Seconds() / 1e6
	}
	return r
}
func Run(ctx context.Context, protocol, addr string, count, size int) (Result, error) {
	if (protocol != "tcp" && protocol != "udp") || count < 1 || count > 1000 || size < 16 || size > 1200 {
		return Result{}, errors.New("protocol tcp/udp; count 1..1000; bytes 16..1200")
	}
	c, e := (&net.Dialer{Timeout: 3 * time.Second}).DialContext(ctx, protocol, addr)
	if e != nil {
		return Result{}, e
	}
	defer c.Close()
	stop := context.AfterFunc(ctx, func() { c.Close() })
	defer stop()
	if protocol == "tcp" {
		c.SetDeadline(time.Now().Add(time.Second))
		if e = wire.Send(c, wire.Header{Op: "echo"}); e != nil {
			return Result{}, e
		}
	}
	payload := make([]byte, size)
	copy(payload, []byte("PLS1"))
	// A per-run cookie rejects replies from older or unrelated experiments.
	if _, e = rand.Read(payload[8:16]); e != nil {
		return Result{}, e
	}
	buf := make([]byte, 2048)
	samples := make([]Sample, 0, count)
	start := time.Now()
	stamps := make([]time.Time, count)
	for i := 0; i < count; i++ {
		if ctx.Err() != nil {
			return Result{}, ctx.Err()
		}
		binary.BigEndian.PutUint32(payload[4:8], uint32(i))
		sent := time.Now()
		stamps[i] = sent
		c.SetDeadline(sent.Add(250 * time.Millisecond))
		if protocol == "tcp" {
			e = wire.Write(c, wire.Chunk, payload)
		} else {
			_, e = c.Write(payload)
		}
		if e != nil {
			return Result{}, e
		}
		sample := Sample{Sequence: i, Bytes: size, Arrival: -1}
		for {
			var b []byte
			var err error
			if protocol == "tcp" {
				var k byte
				k, b, err = wire.Read(c)
				if err == nil && k != wire.Chunk {
					err = errors.New("unexpected echo frame")
				}
			} else {
				var n int
				n, err = c.Read(buf)
				b = buf[:n]
			}
			if err != nil {
				if ctx.Err() != nil {
					return Result{}, ctx.Err()
				}
				if protocol == "tcp" {
					return Result{}, err
				}
				if ne, ok := err.(net.Error); !ok || !ne.Timeout() {
					return Result{}, err
				}
				break
			}
			if len(b) != size || string(b[:4]) != "PLS1" || !bytes.Equal(b[8:], payload[8:]) {
				if protocol == "tcp" {
					return Result{}, errors.New("echo payload mismatch")
				}
				continue
			}
			seq := int(binary.BigEndian.Uint32(b[4:8]))
			if seq != i {
				if seq >= 0 && seq < i && len(samples) < count*4 {
					samples = append(samples, Sample{Sequence: seq, RTTMS: float64(time.Since(stamps[seq])) / float64(time.Millisecond), Received: true, Arrival: len(samples), Bytes: size})
				}
				continue
			}
			sample.Received = true
			sample.RTTMS = float64(time.Since(sent)) / float64(time.Millisecond)
			sample.Arrival = len(samples)
			break
		}
		samples = append(samples, sample)
	}
	return Summarize(protocol, addr, count, time.Since(start), samples), nil
}

type Echo struct {
	conn net.PacketConn
	done chan struct{}
}

func StartEcho(addr string) (*Echo, error) {
	c, e := net.ListenPacket("udp", addr)
	if e != nil {
		return nil, e
	}
	s := &Echo{c, make(chan struct{})}
	go func() {
		defer close(s.done)
		b := make([]byte, 1201)
		for {
			n, a, e := c.ReadFrom(b)
			if e != nil {
				return
			}
			if n >= 16 && n <= 1200 && string(b[:4]) == "PLS1" {
				c.SetWriteDeadline(time.Now().Add(time.Second))
				c.WriteTo(b[:n], a)
			}
		}
	}()
	return s, nil
}
func (e *Echo) Addr() string { return e.conn.LocalAddr().String() }
func (e *Echo) Close()       { e.conn.Close(); <-e.done }
