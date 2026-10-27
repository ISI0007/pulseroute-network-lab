package store

import (
	"encoding/json"
	"fmt"
	"net"
	"strconv"
	"strings"

	"github.com/ISI0007/pulseroute-network-lab/internal/bench"
	"github.com/ISI0007/pulseroute-network-lab/internal/network"
)

type Settings struct {
	Name      string `json:"name"`
	TCP       string `json:"tcp"`
	UDP       string `json:"udp"`
	Target    string `json:"target"`
	Discovery string `json:"discovery"`
}

func DefaultSettings() Settings {
	return Settings{Name: "Operator", TCP: "127.0.0.1:42191", UDP: "127.0.0.1:42192", Target: "127.0.0.1:42191", Discovery: "255.255.255.255:42190"}
}

func (v Settings) Validate() error {
	if len(strings.TrimSpace(v.Name)) == 0 || len(v.Name) > 64 {
		return fmt.Errorf("peer name must contain 1..64 bytes")
	}
	for _, field := range []struct{ name, value string }{{"TCP", v.TCP}, {"UDP", v.UDP}, {"target", v.Target}, {"discovery", v.Discovery}} {
		host, port, err := net.SplitHostPort(field.value)
		n, parseErr := strconv.Atoi(port)
		if err != nil || parseErr != nil || n < 1 || n > 65535 || host == "" || strings.ContainsAny(host, " \t\r\n/\\") {
			return fmt.Errorf("%s must be host:port with a port from 1 to 65535", field.name)
		}
	}
	return nil
}

func (s *Store) LoadSettings() (Settings, error) {
	raw, err := s.GetSetting("node")
	if err != nil {
		return Settings{}, err
	}
	if raw == "" {
		return DefaultSettings(), nil
	}
	var v Settings
	if err = json.Unmarshal([]byte(raw), &v); err != nil {
		return v, fmt.Errorf("stored node settings: %w", err)
	}
	return v, v.Validate()
}

// The complete configuration is replaced in one SQL statement.
func (s *Store) SaveSettings(v Settings) error {
	if err := v.Validate(); err != nil {
		return err
	}
	raw, err := json.Marshal(v)
	if err != nil {
		return err
	}
	return s.Setting("node", string(raw))
}

func (s *Store) KnownPeers() ([]network.Peer, error) {
	rows, err := s.db.Query("SELECT name,address FROM peers ORDER BY last_seen DESC,address LIMIT 128")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	peers := []network.Peer{}
	for rows.Next() {
		var p network.Peer
		if err = rows.Scan(&p.Name, &p.Address); err != nil {
			return nil, err
		}
		peers = append(peers, p)
	}
	return peers, rows.Err()
}

// Return the newest 20 saved runs in oldest-to-newest display order.
func (s *Store) Results() ([]bench.Result, error) {
	rows, err := s.db.Query("SELECT result FROM benchmark_sessions ORDER BY rowid DESC LIMIT 20")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	results := []bench.Result{}
	for rows.Next() {
		var raw string
		var result bench.Result
		if err = rows.Scan(&raw); err != nil {
			return nil, err
		}
		if err = json.Unmarshal([]byte(raw), &result); err != nil {
			return nil, fmt.Errorf("stored benchmark: %w", err)
		}
		results = append(results, result)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	for i, j := 0, len(results)-1; i < j; i, j = i+1, j-1 {
		results[i], results[j] = results[j], results[i]
	}
	return results, nil
}
