//go:build desktop

package main

import (
	"errors"
	"fmt"
	"net"
	"net/url"
	"path/filepath"
	"sort"
	"strconv"
	"time"

	"github.com/ISI0007/pulseroute-network-lab/internal/nearby"
)

type NearbyBind struct {
	Label   string `json:"label"`
	Address string `json:"address"`
}
type NearbyInfo struct {
	Active         bool                `json:"active"`
	Listening      string              `json:"listening"`
	URLs           []string            `json:"urls"`
	LocalURL       string              `json:"localUrl"`
	ExpiresAt      time.Time           `json:"expiresAt"`
	Files          []nearby.SharedFile `json:"files"`
	BindChoices    []NearbyBind        `json:"bindChoices"`
	MaxUploadBytes int64               `json:"maxUploadBytes"`
}

func nearbyBinds() []NearbyBind {
	out := []NearbyBind{}
	interfaces, _ := net.Interfaces()
	for _, iface := range interfaces {
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			continue
		}
		addrs, _ := iface.Addrs()
		for _, addr := range addrs {
			ip, _, err := net.ParseCIDR(addr.String())
			if err != nil || ip.To4() == nil || !ip.IsPrivate() {
				continue
			}
			out = append(out, NearbyBind{Label: iface.Name + " / " + ip.String(), Address: net.JoinHostPort(ip.String(), "42193")})
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Address < out[j].Address })
	return append(out, NearbyBind{Label: "This PC only (local test)", Address: "127.0.0.1:42193"})
}
func nearbyInfo(s *nearby.Server) NearbyInfo {
	info := NearbyInfo{URLs: []string{}, Files: []nearby.SharedFile{}, BindChoices: nearbyBinds(), MaxUploadBytes: nearby.MaxUpload}
	if s == nil || s.Closed() {
		return info
	}
	info.Active = true
	info.Listening = s.Addr()
	info.ExpiresAt = s.ExpiresAt()
	info.Files = s.Files()
	host, port, _ := net.SplitHostPort(s.Addr())
	link := func(ip string) string {
		return (&url.URL{Scheme: "http", Host: net.JoinHostPort(ip, port), Path: "/", Fragment: s.Token()}).String()
	}
	if host == "0.0.0.0" || host == "::" || host == "" {
		for _, choice := range info.BindChoices {
			ip, _, _ := net.SplitHostPort(choice.Address)
			if ip != "127.0.0.1" {
				info.URLs = append(info.URLs, link(ip))
			}
		}
		info.LocalURL = link("127.0.0.1")
	} else if ip := net.ParseIP(host); ip != nil && ip.IsLoopback() {
		info.LocalURL = link(host)
	} else {
		info.URLs = append(info.URLs, link(host))
	}
	return info
}
func (a *App) NearbyStatus() (NearbyInfo, error) {
	if err := a.beginOperation(); err != nil {
		return NearbyInfo{}, err
	}
	defer a.wg.Done()
	a.mu.Lock()
	s := a.nearbyServer
	a.mu.Unlock()
	return nearbyInfo(s), nil
}
func (a *App) StartNearby(address string, paths []string) (NearbyInfo, error) {
	if err := a.beginOperation(); err != nil {
		return NearbyInfo{}, err
	}
	defer a.wg.Done()
	host, port, err := net.SplitHostPort(address)
	n, pe := strconv.Atoi(port)
	if err != nil || pe != nil || n < 0 || n > 65535 || net.ParseIP(host) == nil {
		return NearbyInfo{}, errors.New("nearby listener must use a local IP address and port")
	}
	if len(paths) > 128 {
		return NearbyInfo{}, errors.New("select at most 128 nearby files")
	}
	a.ops.Lock()
	defer a.ops.Unlock()
	a.mu.Lock()
	if a.closing {
		a.mu.Unlock()
		return NearbyInfo{}, errors.New("application is shutting down")
	}
	for _, path := range paths {
		if a.allowed[path] == "" {
			a.mu.Unlock()
			return NearbyInfo{}, errors.New("select nearby files through the native file picker")
		}
	}
	previous := a.nearbyServer
	a.mu.Unlock()
	if previous != nil && !previous.Closed() {
		return NearbyInfo{}, errors.New("stop the existing nearby session before changing files")
	}
	s, err := nearby.Start(address, filepath.Join(a.dir, "received", "nearby"), paths, a.record)
	if err != nil {
		return NearbyInfo{}, fmt.Errorf("nearby sharing: %w", err)
	}
	a.mu.Lock()
	a.nearbyServer = s
	a.mu.Unlock()
	return nearbyInfo(s), nil
}
func (a *App) StopNearby() {
	a.ops.Lock()
	defer a.ops.Unlock()
	a.mu.Lock()
	s := a.nearbyServer
	a.nearbyServer = nil
	a.mu.Unlock()
	if s != nil {
		_ = s.Close()
	}
}
