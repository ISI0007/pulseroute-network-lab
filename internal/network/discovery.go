package network

import (
	"context"
	"encoding/json"
	"errors"
	"net"
	"strconv"
	"time"
)

type Peer struct {
	Name    string `json:"name"`
	Address string `json:"address"`
}
type discoveryPacket struct {
	Magic string `json:"magic"`
	Nonce string `json:"nonce"`
	Name  string `json:"name"`
	Port  int    `json:"port"`
}
type Discovery struct {
	c    *net.UDPConn
	done chan struct{}
}

func Advertise(bind, name string, tcpPort int) (*Discovery, error) {
	if len(name) > 64 || tcpPort < 1 || tcpPort > 65535 {
		return nil, errors.New("invalid advertisement")
	}
	a, e := net.ResolveUDPAddr("udp4", bind)
	if e != nil {
		return nil, e
	}
	c, e := net.ListenUDP("udp4", a)
	if e != nil {
		return nil, e
	}
	d := &Discovery{c, make(chan struct{})}
	go func() {
		defer close(d.done)
		b := make([]byte, 1024)
		for {
			n, a, e := c.ReadFromUDP(b)
			if e != nil {
				return
			}
			var p discoveryPacket
			if json.Unmarshal(b[:n], &p) != nil || p.Magic != "PULSEROUTE_DISCOVER_1" || len(p.Nonce) != 32 {
				continue
			}
			reply, _ := json.Marshal(discoveryPacket{Magic: "PULSEROUTE_PEER_1", Nonce: p.Nonce, Name: name, Port: tcpPort})
			c.SetWriteDeadline(time.Now().Add(time.Second))
			c.WriteToUDP(reply, a)
		}
	}()
	return d, nil
}
func (d *Discovery) Addr() string { return d.c.LocalAddr().String() }
func (d *Discovery) Close()       { d.c.Close(); <-d.done }
func Discover(ctx context.Context, target string) ([]Peer, error) {
	a, e := net.ResolveUDPAddr("udp4", target)
	if e != nil {
		return nil, e
	}
	c, e := net.ListenUDP("udp4", &net.UDPAddr{IP: net.IPv4zero, Port: 0})
	if e != nil {
		return nil, e
	}
	defer c.Close()
	stop := context.AfterFunc(ctx, func() { c.Close() })
	defer stop()
	nonce := ID()
	b, _ := json.Marshal(discoveryPacket{Magic: "PULSEROUTE_DISCOVER_1", Nonce: nonce})
	c.SetDeadline(time.Now().Add(1200 * time.Millisecond))
	if _, e = c.WriteToUDP(b, a); e != nil {
		return nil, e
	}
	out := []Peer{}
	seen := map[string]bool{}
	buf := make([]byte, 1024)
	for {
		n, a, e := c.ReadFromUDP(buf)
		if e != nil {
			if ctx.Err() != nil {
				return nil, ctx.Err()
			}
			if ne, ok := e.(net.Error); ok && ne.Timeout() {
				return out, nil
			}
			return nil, e
		}
		var p discoveryPacket
		if json.Unmarshal(buf[:n], &p) != nil || p.Magic != "PULSEROUTE_PEER_1" || p.Nonce != nonce || len(p.Name) > 64 || p.Port < 1 || p.Port > 65535 {
			continue
		}
		addr := net.JoinHostPort(a.IP.String(), strconv.Itoa(p.Port))
		if !seen[addr] && len(out) < 128 {
			seen[addr] = true
			out = append(out, Peer{Name: p.Name, Address: addr})
		}
	}
}
