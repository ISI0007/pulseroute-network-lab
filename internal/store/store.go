package store

import (
	"database/sql"
	_ "embed"
	"encoding/json"
	"fmt"
	"github.com/ISI0007/pulseroute-network-lab/internal/network"
	_ "modernc.org/sqlite"
	"time"
)

//go:embed migrations/001_initial.sql
var migration string

type Store struct{ db *sql.DB }

func Open(path string) (*Store, error) {
	db, e := sql.Open("sqlite", path)
	if e != nil {
		return nil, e
	}
	db.SetMaxOpenConns(1)
	fail := func(e error) (*Store, error) { db.Close(); return nil, e }
	if _, e = db.Exec("PRAGMA busy_timeout=5000; PRAGMA journal_mode=WAL; PRAGMA foreign_keys=ON;"); e != nil {
		return fail(e)
	}
	tx, e := db.Begin()
	if e != nil {
		return fail(e)
	}
	var version int
	if e = tx.QueryRow("PRAGMA user_version").Scan(&version); e != nil {
		tx.Rollback()
		return fail(e)
	}
	if version > 1 {
		tx.Rollback()
		return fail(fmt.Errorf("database version %d is newer than supported", version))
	}
	if version == 0 {
		if _, e = tx.Exec(migration); e != nil {
			tx.Rollback()
			return fail(e)
		}
		if _, e = tx.Exec("PRAGMA user_version=1"); e != nil {
			tx.Rollback()
			return fail(e)
		}
	}
	if e = tx.Commit(); e != nil {
		return fail(e)
	}
	return &Store{db}, nil
}
func (s *Store) Close() error { return s.db.Close() }
func (s *Store) Record(e network.Event) error {
	if e.Kind == "progress" {
		return nil
	}
	if e.At.IsZero() {
		e.At = time.Now().UTC()
	}
	b, err := json.Marshal(e)
	if err != nil {
		return err
	}
	_, err = s.db.Exec("INSERT INTO events(kind,reference_id,created_at,payload) VALUES(?,?,?,?)", e.Kind, e.ID, e.At.Format(time.RFC3339Nano), string(b))
	return err
}
func (s *Store) History(limit int) ([]network.Event, error) {
	if limit < 1 || limit > 500 {
		limit = 200
	}
	rows, e := s.db.Query("SELECT payload FROM events ORDER BY seq DESC LIMIT ?", limit)
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	out := make([]network.Event, 0)
	for rows.Next() {
		var b string
		if e = rows.Scan(&b); e != nil {
			return nil, e
		}
		var v network.Event
		if e = json.Unmarshal([]byte(b), &v); e != nil {
			return nil, e
		}
		out = append(out, v)
	}
	return out, rows.Err()
}
func (s *Store) Setting(key, value string) error {
	_, e := s.db.Exec("INSERT INTO settings(key,value) VALUES(?,?) ON CONFLICT(key) DO UPDATE SET value=excluded.value", key, value)
	return e
}
func (s *Store) GetSetting(key string) (string, error) {
	var v string
	e := s.db.QueryRow("SELECT value FROM settings WHERE key=?", key).Scan(&v)
	if e == sql.ErrNoRows {
		return "", nil
	}
	return v, e
}
func (s *Store) Peer(name, addr string) error {
	_, e := s.db.Exec("INSERT INTO peers(address,name,last_seen) VALUES(?,?,?) ON CONFLICT(address) DO UPDATE SET name=excluded.name,last_seen=excluded.last_seen", addr, name, time.Now().UTC().Format(time.RFC3339Nano))
	return e
}
func (s *Store) Benchmark(id, protocol, addr string, result any, samples []any) error {
	tx, e := s.db.Begin()
	if e != nil {
		return e
	}
	defer tx.Rollback()
	b, e := json.Marshal(result)
	if e != nil {
		return e
	}
	if _, e = tx.Exec("INSERT INTO benchmark_sessions(id,protocol,address,created_at,result) VALUES(?,?,?,?,?)", id, protocol, addr, time.Now().UTC().Format(time.RFC3339Nano), string(b)); e != nil {
		return e
	}
	for i, v := range samples {
		b, e = json.Marshal(v)
		if e != nil {
			return e
		}
		if _, e = tx.Exec("INSERT INTO benchmark_measurements(session_id,sequence,payload) VALUES(?,?,?)", id, i, string(b)); e != nil {
			return e
		}
	}
	return tx.Commit()
}
