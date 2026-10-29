// Package wire implements bounded, binary-safe PulseRoute v1 frames.
package wire

import (
	"encoding/binary"
	"encoding/json"
	"errors"
	"io"
)

const MaxFrame = 128 * 1024
const ChunkSize = 64 * 1024
const (
	Control byte = 1
	Chunk   byte = 2
	End     byte = 3
)

type Peer struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type Header struct {
	Version  int    `json:"version"`
	Op       string `json:"op"`
	ID       string `json:"id,omitempty"`
	Peer     string `json:"peer,omitempty"`
	Text     string `json:"text,omitempty"`
	Name     string `json:"name,omitempty"`
	Size     int64  `json:"size,omitempty"`
	SHA256   string `json:"sha256,omitempty"`
	Error    string `json:"error,omitempty"`
	Count    int    `json:"count,omitempty"`
	From     string `json:"from,omitempty"`
	To       string `json:"to,omitempty"`
	Peers    []Peer `json:"peers,omitempty"`
	Revision uint64 `json:"revision,omitempty"`
}

func Write(w io.Writer, k byte, b []byte) error {
	if len(b)+1 > MaxFrame {
		return errors.New("frame exceeds limit")
	}
	var p [5]byte
	binary.BigEndian.PutUint32(p[:4], uint32(len(b)+1))
	p[4] = k
	if e := writeAll(w, p[:]); e != nil {
		return e
	}
	return writeAll(w, b)
}
func writeAll(w io.Writer, p []byte) error {
	for len(p) > 0 {
		n, e := w.Write(p)
		if e != nil {
			return e
		}
		if n <= 0 {
			return io.ErrShortWrite
		}
		p = p[n:]
	}
	return nil
}
func Read(r io.Reader) (byte, []byte, error) {
	var p [4]byte
	if _, e := io.ReadFull(r, p[:]); e != nil {
		return 0, nil, e
	}
	n := binary.BigEndian.Uint32(p[:])
	if n < 1 || n > MaxFrame {
		return 0, nil, errors.New("invalid frame length")
	}
	b := make([]byte, int(n))
	if _, e := io.ReadFull(r, b); e != nil {
		return 0, nil, e
	}
	return b[0], b[1:], nil
}
func Send(w io.Writer, h Header) error {
	h.Version = 1
	b, e := json.Marshal(h)
	if e != nil {
		return e
	}
	return Write(w, Control, b)
}
func Receive(r io.Reader) (Header, error) {
	var h Header
	k, b, e := Read(r)
	if e != nil {
		return h, e
	}
	if k != Control {
		return h, errors.New("expected control frame")
	}
	if e = json.Unmarshal(b, &h); e != nil {
		return h, e
	}
	if h.Version != 1 {
		return h, errors.New("unsupported version")
	}
	return h, nil
}
