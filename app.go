//go:build desktop

package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/ISI0007/pulseroute-network-lab/internal/bench"
	"github.com/ISI0007/pulseroute-network-lab/internal/nearby"
	"github.com/ISI0007/pulseroute-network-lab/internal/network"
	"github.com/ISI0007/pulseroute-network-lab/internal/store"
	"github.com/ISI0007/pulseroute-network-lab/internal/wire"
	wr "github.com/wailsapp/wails/v2/pkg/runtime"
	"io"
	"mime"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Snapshot struct {
	SelfID      string          `json:"selfId"`
	CallPeers   []wire.Peer     `json:"callPeers"`
	CallSignals []CallEnvelope  `json:"callSignals"`
	Settings    store.Settings  `json:"settings"`
	Listening   string          `json:"listening"`
	UDP         string          `json:"udp"`
	Connected   string          `json:"connected"`
	Error       string          `json:"error"`
	History     []network.Event `json:"history"`
	Progress    []network.Event `json:"progress"`
	Peers       []network.Peer  `json:"peers"`
	Results     []bench.Result  `json:"results"`
	Dropped     []string        `json:"dropped"`
	Theme       string          `json:"theme"`
	DataDir     string          `json:"dataDir"`
}
type App struct {
	callSeq      uint64
	callRevision uint64
	nearbyServer *nearby.Server
	life         context.Context
	cancel       context.CancelFunc
	closing      bool
	ctx          context.Context
	mu           sync.Mutex
	ops          sync.Mutex
	wg           sync.WaitGroup
	db           *store.Store
	dir          string
	server       *network.Server
	echo         *bench.Echo
	discovery    *network.Discovery
	client       *network.Client
	reconnect    context.CancelFunc
	transfer     context.CancelFunc
	state        Snapshot
	progress     map[string]network.Event
	allowed      map[string]string
}

func newApp(dir string) (*App, error) {
	if e := os.MkdirAll(dir, 0700); e != nil {
		return nil, e
	}
	db, e := store.Open(filepath.Join(dir, "history.db"))
	if e != nil {
		return nil, e
	}
	theme, e := db.GetSetting("theme")
	if e != nil {
		db.Close()
		return nil, e
	}
	if !validTheme(theme) {
		theme = "forest"
	}
	a := &App{db: db, dir: dir, progress: map[string]network.Event{}, allowed: map[string]string{}, state: Snapshot{Theme: theme, DataDir: dir, Peers: []network.Peer{}, Results: []bench.Result{}, Dropped: []string{}}}
	a.life, a.cancel = context.WithCancel(context.Background())
	a.state.Settings, e = db.LoadSettings()
	if e == nil {
		a.state.Peers, e = db.KnownPeers()
	}
	if e == nil {
		a.state.Results, e = db.Results()
	}
	if e != nil {
		a.cancel()
		db.Close()
		return nil, e
	}
	return a, nil
}
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	wr.OnFileDrop(ctx, func(x, y int, paths []string) {
		a.mu.Lock()
		defer a.mu.Unlock()
		for _, p := range paths {
			if len(a.state.Dropped) < 100 {
				a.state.Dropped = append(a.state.Dropped, p)
				a.allowed[p] = p
			}
		}
	})
}
func (a *App) shutdown(ctx context.Context) {
	a.mu.Lock()
	if a.closing {
		a.mu.Unlock()
		return
	}
	a.closing = true
	a.cancel()
	a.mu.Unlock()
	a.Disconnect()
	a.CancelTransfer()
	a.StopHost()
	a.StopNearby()
	a.wg.Wait()
	a.db.Close()
}
func (a *App) setError(e error) {
	if e == nil {
		return
	}
	a.mu.Lock()
	a.state.Error = e.Error()
	a.mu.Unlock()
}
func (a *App) record(e network.Event) error {
	if e.At.IsZero() {
		e.At = time.Now().UTC()
	}
	if e.Kind == "progress" || e.Kind == "transfer" {
		a.mu.Lock()
		if _, exists := a.progress[e.ID]; !exists && len(a.progress) >= 128 {
			oldest := ""
			var at time.Time
			for id, v := range a.progress {
				terminal := v.Status == "verified" || v.Status == "sent-verified" || v.Status == "failed" || v.Status == "cancelled" || v.Status == "received"
				if terminal && (oldest == "" || v.At.Before(at)) {
					oldest, at = id, v.At
				}
			}
			if oldest != "" {
				delete(a.progress, oldest)
			}
		}
		if _, exists := a.progress[e.ID]; exists || len(a.progress) < 128 {
			a.progress[e.ID] = e
		}
		a.mu.Unlock()
		if e.Kind == "progress" {
			return nil
		}
	}
	err := a.db.Record(e)
	if err != nil {
		a.setError(err)
	}
	return err
}
func (a *App) Snapshot() (Snapshot, error) {
	if err := a.beginOperation(); err != nil {
		return Snapshot{}, err
	}
	defer a.wg.Done()
	history, e := a.db.History(200)
	if e != nil {
		return Snapshot{}, e
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	s := a.state
	s.History = history
	s.Progress = []network.Event{}
	for _, v := range a.progress {
		s.Progress = append(s.Progress, v)
	}
	s.Peers = append([]network.Peer{}, a.state.Peers...)
	s.CallPeers = append([]wire.Peer{}, a.state.CallPeers...)
	s.CallSignals = append([]CallEnvelope{}, a.state.CallSignals...)
	s.Results = append([]bench.Result{}, a.state.Results...)
	s.Dropped = append([]string{}, a.state.Dropped...)
	a.state.Dropped = nil
	return s, nil
}
func (a *App) StartHost(tcp, udp, name string) error {
	if err := a.beginOperation(); err != nil {
		return err
	}
	defer a.wg.Done()
	a.ops.Lock()
	defer a.ops.Unlock()
	a.mu.Lock()
	exists := a.server != nil
	a.mu.Unlock()
	if exists {
		return errors.New("listener is already running")
	}
	s, e := network.Start(tcp, filepath.Join(a.dir, "received"), a.record)
	if e != nil {
		return e
	}
	u, e := bench.StartEcho(udp)
	if e != nil {
		s.Close()
		return e
	}
	host, p, _ := net.SplitHostPort(s.Addr())
	port, _ := strconv.Atoi(p)
	d, de := network.Advertise(net.JoinHostPort(host, "42190"), name, port)
	a.mu.Lock()
	a.server = s
	a.echo = u
	a.discovery = d
	a.state.Listening = s.Addr()
	a.state.UDP = u.Addr()
	if de != nil {
		a.state.Error = "TCP/UDP started; discovery unavailable: " + de.Error()
	}
	a.mu.Unlock()
	return nil
}
func (a *App) StopHost() {
	a.ops.Lock()
	defer a.ops.Unlock()
	a.mu.Lock()
	s, u, d := a.server, a.echo, a.discovery
	a.server = nil
	a.echo = nil
	a.discovery = nil
	a.state.Listening = ""
	a.state.UDP = ""
	a.mu.Unlock()
	if d != nil {
		d.Close()
	}
	if u != nil {
		u.Close()
	}
	if s != nil {
		s.Close()
	}
}
func (a *App) Connect(addr, name string) error {
	if err := a.beginOperation(); err != nil {
		return err
	}
	defer a.wg.Done()
	a.ops.Lock()
	defer a.ops.Unlock()
	a.disconnect()
	ctx, cancel := context.WithCancel(a.life)
	c, e := network.Connect(ctx, addr, name)
	if e != nil {
		cancel()
		return e
	}
	if e = a.db.Peer(name, addr); e != nil {
		c.Close()
		cancel()
		return e
	}
	a.mu.Lock()
	a.client = c
	a.reconnect = cancel
	a.state.Connected = addr
	a.state.SelfID = c.ID
	a.state.CallPeers = []wire.Peer{}
	a.state.CallSignals = []CallEnvelope{}
	a.callRevision = 0
	a.mu.Unlock()
	a.wg.Add(1)
	go func() {
		defer a.wg.Done()
		current := c
		delay := 500 * time.Millisecond
		for {
			for h := range current.Events {
				if ctx.Err() != nil {
					return
				}
				if h.Op == "peers" {
					a.mu.Lock()
					if a.client == current && ctx.Err() == nil && h.Revision >= a.callRevision {
						a.state.CallPeers = append([]wire.Peer{}, h.Peers...)
						a.callRevision = h.Revision
					}
					a.mu.Unlock()
				}
				if h.Op == "signal" {
					a.pushCallFrom(current, CallEnvelope{From: h.From, Peer: h.Peer, ID: h.ID, Text: h.Text})
				}
				if h.Op == "message" {
					a.record(network.Event{Kind: "message", ID: h.ID, Peer: h.Peer, Text: h.Text, Status: "received"})
				}
				if h.Op == "error" {
					a.setError(errors.New(h.Error))
					a.pushCallFrom(current, CallEnvelope{ID: h.ID, Error: h.Error})
				}
			}
			if ctx.Err() != nil {
				return
			}
			a.mu.Lock()
			if a.client == current {
				a.client = nil
				a.state.Connected = ""
				a.state.SelfID = ""
				a.state.CallPeers = []wire.Peer{}
				a.callRevision = 0
			}
			a.mu.Unlock()
			a.pushCall(CallEnvelope{Error: "Signaling connection closed. Call ended."})
			for {
				select {
				case <-ctx.Done():
					return
				case <-time.After(delay):
				}
				next, e := network.Connect(ctx, addr, name)
				if e != nil {
					a.setError(fmt.Errorf("reconnecting: %w", e))
					if delay < 8*time.Second {
						delay *= 2
					}
					continue
				}
				if ctx.Err() != nil {
					next.Close()
					return
				}
				a.mu.Lock()
				if ctx.Err() != nil {
					a.mu.Unlock()
					next.Close()
					return
				}
				a.client = next
				a.state.Connected = addr
				a.state.SelfID = next.ID
				a.state.CallPeers = []wire.Peer{}
				a.state.CallSignals = []CallEnvelope{}
				a.callRevision = 0
				a.mu.Unlock()
				current = next
				delay = 500 * time.Millisecond
				break
			}
		}
	}()
	return nil
}
func (a *App) Disconnect() {
	a.ops.Lock()
	defer a.ops.Unlock()
	a.disconnect()
}
func (a *App) disconnect() {
	a.mu.Lock()
	cancel, c := a.reconnect, a.client
	if cancel != nil {
		cancel()
	}
	a.reconnect = nil
	a.client = nil
	a.state.Connected = ""
	a.state.SelfID = ""
	a.state.CallPeers = []wire.Peer{}
	a.state.CallSignals = []CallEnvelope{}
	a.callRevision = 0
	a.mu.Unlock()
	if c != nil {
		c.Close()
	}
	if c != nil || cancel != nil {
		a.pushCall(CallEnvelope{Error: "Signaling connection closed. Call ended."})
	}
}
func (a *App) SendMessage(text string) (string, error) {
	if err := a.beginOperation(); err != nil {
		return "", err
	}
	defer a.wg.Done()
	a.mu.Lock()
	c := a.client
	a.mu.Unlock()
	if c == nil {
		return "", errors.New("connect before sending")
	}
	ctx, cancel := context.WithTimeout(a.life, 10*time.Second)
	defer cancel()
	ack, err := c.SendConfirmed(ctx, text)
	if err != nil {
		return ack.ID, err
	}
	return ack.ID, a.record(network.Event{Kind: "delivery", ID: ack.ID, Status: "accepted", Count: ack.Count})
}
func (a *App) ChooseFiles() ([]string, error) {
	paths, e := wr.OpenMultipleFilesDialog(a.ctx, wr.OpenDialogOptions{Title: "Queue files for transfer"})
	if e == nil {
		a.mu.Lock()
		for _, p := range paths {
			a.allowed[p] = p
		}
		a.mu.Unlock()
	}
	return paths, e
}
func (a *App) SendFile(addr, path string) error {
	if err := a.beginOperation(); err != nil {
		return err
	}
	defer a.wg.Done()
	a.mu.Lock()
	if a.transfer != nil {
		a.mu.Unlock()
		return errors.New("a transfer is already active")
	}
	if a.allowed[path] == "" {
		a.mu.Unlock()
		return errors.New("select the file through the file picker or drop zone")
	}
	ctx, cancel := context.WithCancel(a.life)
	a.transfer = cancel
	a.mu.Unlock()
	a.wg.Add(1)
	defer a.wg.Done()
	defer func() { cancel(); a.mu.Lock(); a.transfer = nil; a.mu.Unlock() }()
	var last network.Event
	h, e := network.SendFile(ctx, addr, path, func(ev network.Event) { last = ev; a.record(ev) })
	if e != nil {
		last.Status = "failed"
		if ctx.Err() != nil {
			last.Status = "cancelled"
		}
		last.Text = e.Error()
		last.Kind = "transfer"
		if last.ID == "" {
			last.ID = network.ID()
			last.Name = filepath.Base(path)
		}
		a.record(last)
		return e
	}
	return a.record(network.Event{Kind: "transfer", ID: h.ID, Peer: addr, Name: h.Name, SHA256: h.SHA256, Bytes: h.Size, Total: h.Size, Status: "sent-verified"})
}
func (a *App) CancelTransfer() {
	a.mu.Lock()
	cancel := a.transfer
	a.mu.Unlock()
	if cancel != nil {
		cancel()
	}
}
func (a *App) Discover(target string) ([]network.Peer, error) {
	if err := a.beginOperation(); err != nil {
		return nil, err
	}
	defer a.wg.Done()
	ctx, cancel := context.WithTimeout(a.life, 2*time.Second)
	defer cancel()
	peers, e := network.Discover(ctx, target)
	if e != nil {
		return nil, e
	}
	for _, p := range peers {
		if e = a.db.Peer(p.Name, p.Address); e != nil {
			return nil, e
		}
	}
	a.mu.Lock()
	a.state.Peers = peers
	a.mu.Unlock()
	return peers, nil
}
func (a *App) Benchmark(protocol, addr string, count, size int) (bench.Result, error) {
	if err := a.beginOperation(); err != nil {
		return bench.Result{}, err
	}
	defer a.wg.Done()
	ctx, cancel := context.WithTimeout(a.life, 30*time.Second)
	defer cancel()
	r, e := bench.Run(ctx, protocol, addr, count, size)
	if e != nil {
		return r, e
	}
	samples := make([]any, len(r.Samples))
	for i, v := range r.Samples {
		samples[i] = v
	}
	if e = a.db.Benchmark(network.ID(), protocol, addr, r, samples); e != nil {
		return r, e
	}
	a.mu.Lock()
	a.state.Results = append(a.state.Results, r)
	if len(a.state.Results) > 20 {
		a.state.Results = a.state.Results[len(a.state.Results)-20:]
	}
	a.mu.Unlock()
	return r, nil
}
func (a *App) SetTheme(theme string) error {
	if err := a.beginOperation(); err != nil {
		return err
	}
	defer a.wg.Done()
	if !validTheme(theme) {
		return errors.New("invalid theme")
	}
	if e := a.db.Setting("theme", theme); e != nil {
		return e
	}
	a.mu.Lock()
	a.state.Theme = theme
	a.mu.Unlock()
	return nil
}
func (a *App) ClearError() { a.mu.Lock(); a.state.Error = ""; a.mu.Unlock() }
func (a *App) ExportResults() (string, error) {
	a.mu.Lock()
	b, e := json.MarshalIndent(a.state.Results, "", "  ")
	a.mu.Unlock()
	if e != nil {
		return "", e
	}
	p, e := wr.SaveFileDialog(a.ctx, wr.SaveDialogOptions{Title: "Export raw benchmark measurements", DefaultFilename: "pulseroute-results.json"})
	if e != nil || p == "" {
		return p, e
	}
	return p, os.WriteFile(p, b, 0600)
}
func (a *App) Preview(path string) (string, error) {
	if err := a.beginOperation(); err != nil {
		return "", err
	}
	defer a.wg.Done()
	a.mu.Lock()
	allowed := a.allowed[path] != ""
	a.mu.Unlock()
	if !allowed {
		history, e := a.db.History(500)
		if e != nil {
			return "", e
		}
		for _, v := range history {
			if v.Path == path && (v.Status == "verified" || v.Status == "received") {
				allowed = true
				break
			}
		}
	}
	if !allowed {
		return "", errors.New("file was not selected or received")
	}
	f, e := os.Open(path)
	if e != nil {
		return "", e
	}
	defer f.Close()
	st, e := f.Stat()
	if e != nil {
		return "", e
	}
	if !st.Mode().IsRegular() {
		return "", errors.New("preview requires a regular file")
	}
	if st.Size() > 16*1024*1024 {
		return "", errors.New("inline preview limit is 16 MiB; transferred files remain available in the data folder")
	}
	ext := strings.ToLower(filepath.Ext(path))
	typ := mime.TypeByExtension(ext)
	if ext != ".png" && ext != ".jpg" && ext != ".jpeg" && ext != ".gif" && ext != ".webp" && ext != ".mp4" && ext != ".webm" {
		return "", errors.New("preview supports PNG, JPEG, GIF, WebP, MP4 and WebM")
	}
	b, e := io.ReadAll(io.LimitReader(f, 16*1024*1024+1))
	if e != nil {
		return "", e
	}
	if len(b) > 16*1024*1024 {
		return "", errors.New("file grew beyond the inline preview limit")
	}
	if typ == "" {
		typ = http.DetectContentType(b)
	}
	return "data:" + typ + ";base64," + base64.StdEncoding.EncodeToString(b), nil
}

// beginOperation prevents new database/network work once shutdown starts.
func (a *App) beginOperation() error {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.closing {
		return errors.New("application is shutting down")
	}
	a.wg.Add(1)
	return nil
}
func (a *App) SaveSettings(settings store.Settings) error {
	if err := a.beginOperation(); err != nil {
		return err
	}
	defer a.wg.Done()
	if err := a.db.SaveSettings(settings); err != nil {
		return err
	}
	a.mu.Lock()
	a.state.Settings = settings
	a.mu.Unlock()
	return nil
}

func validTheme(theme string) bool {
	switch theme {
	case "dark", "light", "forest", "ocean":
		return true
	default:
		return false
	}
}
