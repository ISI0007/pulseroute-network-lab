package network

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/ISI0007/pulseroute-network-lab/internal/wire"
	"net"
	"os"
	"sync"
	"time"
)

const MaxFile = int64(8 * 1024 * 1024 * 1024)

type Event struct {
	Kind   string    `json:"kind"`
	ID     string    `json:"id"`
	Peer   string    `json:"peer"`
	Text   string    `json:"text"`
	Name   string    `json:"name"`
	Path   string    `json:"path"`
	SHA256 string    `json:"sha256"`
	Bytes  int64     `json:"bytes"`
	Total  int64     `json:"total"`
	Status string    `json:"status"`
	At     time.Time `json:"at"`
}
type Sink func(Event) error
type session struct {
	conn net.Conn
	mu   sync.Mutex
	name string
}

func (s *session) send(h wire.Header) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.conn.SetWriteDeadline(time.Now().Add(5 * time.Second))
	return wire.Send(s.conn, h)
}

type Server struct {
	ln      net.Listener
	dir     string
	sink    Sink
	mu      sync.Mutex
	clients map[*session]bool
	conns   map[net.Conn]bool
	closed  bool
	wg      sync.WaitGroup
	slots   chan struct{}
}

func ID() string {
	var b [16]byte
	if _, e := rand.Read(b[:]); e != nil {
		panic(e)
	}
	return hex.EncodeToString(b[:])
}
func Start(addr, dir string, sink Sink) (*Server, error) {
	if e := os.MkdirAll(dir, 0700); e != nil {
		return nil, e
	}
	ln, e := net.Listen("tcp", addr)
	if e != nil {
		return nil, e
	}
	s := &Server{ln: ln, dir: dir, sink: sink, clients: make(map[*session]bool), conns: make(map[net.Conn]bool), slots: make(chan struct{}, 64)}
	s.wg.Add(1)
	go s.accept()
	return s, nil
}
func (s *Server) Addr() string { return s.ln.Addr().String() }
func (s *Server) emit(e Event) error {
	e.At = time.Now().UTC()
	if s.sink != nil {
		return s.sink(e)
	}
	return nil
}
func (s *Server) Close() error {
	s.mu.Lock()
	if s.closed {
		s.mu.Unlock()
		return nil
	}
	s.closed = true
	s.ln.Close()
	for c := range s.conns {
		c.Close()
	}
	s.mu.Unlock()
	s.wg.Wait()
	return nil
}
func (s *Server) accept() {
	defer s.wg.Done()
	for {
		c, e := s.ln.Accept()
		if e != nil {
			return
		}
		select {
		case s.slots <- struct{}{}:
		default:
			c.Close()
			continue
		}
		s.mu.Lock()
		if s.closed {
			s.mu.Unlock()
			c.Close()
			<-s.slots
			return
		}
		s.conns[c] = true
		s.wg.Add(1)
		s.mu.Unlock()
		go func() {
			defer s.wg.Done()
			defer func() { c.Close(); s.mu.Lock(); delete(s.conns, c); s.mu.Unlock(); <-s.slots }()
			s.handle(c)
		}()
	}
}
func (s *Server) handle(c net.Conn) {
	c.SetDeadline(time.Now().Add(15 * time.Second))
	h, e := wire.Receive(c)
	if e != nil {
		return
	}
	switch h.Op {
	case "hello":
		s.chat(c, h)
	case "put":
		if e = s.receiveFile(c, h); e != nil {
			wire.Send(c, wire.Header{Op: "error", ID: h.ID, Error: e.Error()})
			s.emit(Event{Kind: "transfer", ID: h.ID, Name: h.Name, Total: h.Size, Status: "failed", Text: e.Error()})
		}
	case "ping":
		wire.Send(c, wire.Header{Op: "pong", ID: h.ID})
	case "echo":
		for i := 0; i < 10000; i++ {
			k, b, e := wire.Read(c)
			if e != nil || k != wire.Chunk {
				return
			}
			c.SetDeadline(time.Now().Add(5 * time.Second))
			if e = wire.Write(c, k, b); e != nil {
				return
			}
		}
	default:
		wire.Send(c, wire.Header{Op: "error", Error: "unknown operation"})
	}
}
func (s *Server) chat(c net.Conn, h wire.Header) {
	if len(h.Peer) < 1 || len(h.Peer) > 64 {
		wire.Send(c, wire.Header{Op: "error", Error: "peer name must be 1..64 bytes"})
		return
	}
	p := &session{conn: c, name: h.Peer}
	s.mu.Lock()
	s.clients[p] = true
	s.mu.Unlock()
	defer func() { s.mu.Lock(); delete(s.clients, p); s.mu.Unlock() }()
	if p.send(wire.Header{Op: "welcome", Peer: p.name}) != nil {
		return
	}
	for {
		c.SetReadDeadline(time.Now().Add(90 * time.Second))
		h, e := wire.Receive(c)
		if e != nil {
			return
		}
		if h.Op == "ping" {
			if p.send(wire.Header{Op: "pong", ID: h.ID}) != nil {
				return
			}
			continue
		}
		if h.Op != "message" || len(h.Text) == 0 || len(h.Text) > 8192 || len(h.ID) != 32 {
			p.send(wire.Header{Op: "error", ID: h.ID, Error: "invalid message"})
			continue
		}
		h.Peer = p.name
		if e = s.emit(Event{Kind: "message", ID: h.ID, Peer: p.name, Text: h.Text, Status: "accepted"}); e != nil {
			p.send(wire.Header{Op: "error", ID: h.ID, Error: "history unavailable"})
			continue
		}
		s.mu.Lock()
		peers := make([]*session, 0, len(s.clients))
		for q := range s.clients {
			peers = append(peers, q)
		}
		s.mu.Unlock()
		count := 0
		for _, q := range peers {
			if q.send(h) == nil {
				count++
			} else {
				q.conn.Close()
			}
		}
		if p.send(wire.Header{Op: "ack", ID: h.ID, Count: count}) != nil {
			return
		}
	}
}

type Client struct {
	p      *session
	Events chan wire.Header
	done   chan struct{}
	once   sync.Once
}

func Connect(ctx context.Context, addr, name string) (*Client, error) {
	c, e := (&net.Dialer{Timeout: 5 * time.Second}).DialContext(ctx, "tcp", addr)
	if e != nil {
		return nil, e
	}
	c.SetDeadline(time.Now().Add(5 * time.Second))
	if e = wire.Send(c, wire.Header{Op: "hello", Peer: name}); e != nil {
		c.Close()
		return nil, e
	}
	h, e := wire.Receive(c)
	if e != nil || h.Op != "welcome" {
		c.Close()
		if e == nil {
			e = errors.New(h.Error)
		}
		return nil, e
	}
	c.SetDeadline(time.Time{})
	cl := &Client{p: &session{conn: c, name: name}, Events: make(chan wire.Header, 256), done: make(chan struct{})}
	go cl.read()
	go cl.heartbeat()
	return cl, nil
}
func (c *Client) read() {
	defer close(c.Events)
	defer c.Close()
	for {
		h, e := wire.Receive(c.p.conn)
		if e != nil {
			return
		}
		select {
		case c.Events <- h:
		case <-c.done:
			return
		}
	}
}
func (c *Client) heartbeat() {
	t := time.NewTicker(20 * time.Second)
	defer t.Stop()
	for {
		select {
		case <-c.done:
			return
		case <-t.C:
			if e := c.p.send(wire.Header{Op: "ping", ID: ID()}); e != nil {
				c.Close()
				return
			}
		}
	}
}
func (c *Client) Send(text string) (string, error) {
	if len(text) == 0 || len(text) > 8192 {
		return "", fmt.Errorf("message must be 1..8192 bytes")
	}
	id := ID()
	return id, c.p.send(wire.Header{Op: "message", ID: id, Text: text})
}
func (c *Client) Close() { c.once.Do(func() { close(c.done); c.p.conn.Close() }) }
