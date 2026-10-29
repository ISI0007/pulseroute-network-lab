//go:build desktop

package main

import (
	"context"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestNearbyPickerAuthorizationAndRevocation(t *testing.T) {
	a, err := newApp(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	defer a.shutdown(context.Background())
	file := filepath.Join(t.TempDir(), "selected.txt")
	if err = os.WriteFile(file, []byte("selected content"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err = a.StartNearby("127.0.0.1:0", []string{file}); err == nil {
		t.Fatal("unselected file was shared")
	}
	a.mu.Lock()
	a.allowed[file] = file
	a.mu.Unlock()
	info, err := a.StartNearby("127.0.0.1:0", []string{file})
	if err != nil {
		t.Fatal(err)
	}
	if !info.Active || len(info.Files) != 1 || info.LocalURL == "" {
		t.Fatal("incorrect sharing state", info)
	}
	if _, err = a.StartNearby("127.0.0.1:0", nil); err == nil {
		t.Fatal("active session silently replaced")
	}
	link, err := url.Parse(info.LocalURL)
	if err != nil {
		t.Fatal(err)
	}
	client := &http.Client{Timeout: 2 * time.Second}
	endpoint := "http://" + link.Host + "/api/files"
	response, err := client.Get(endpoint)
	if err != nil {
		t.Fatal(err)
	}
	response.Body.Close()
	if response.StatusCode != http.StatusUnauthorized {
		t.Fatal("unauthenticated portal API accepted", response.StatusCode)
	}
	request, _ := http.NewRequest(http.MethodGet, endpoint, nil)
	request.Header.Set("X-PulseRoute-Token", link.Fragment)
	response, err = client.Do(request)
	if err != nil {
		t.Fatal(err)
	}
	response.Body.Close()
	if response.StatusCode != http.StatusOK {
		t.Fatal("authorized listing failed", response.StatusCode)
	}
	a.StopNearby()
	status, err := a.NearbyStatus()
	if err != nil || status.Active {
		t.Fatal("stop did not clear state", err)
	}
	if response, err = client.Do(request); err == nil {
		response.Body.Close()
		t.Fatal("revoked listener remained reachable")
	}
	if _, err = a.StartNearby("127.0.0.1:0", nil); err != nil {
		t.Fatal(err)
	}
	a.mu.Lock()
	server := a.nearbyServer
	a.mu.Unlock()
	a.shutdown(context.Background())
	if !server.Closed() {
		t.Fatal("shutdown leaked nearby listener")
	}
}
