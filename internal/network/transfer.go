package network

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/ISI0007/pulseroute-network-lab/internal/wire"
	"io"
	"net"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

var safeName = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9._ -]{0,119}$`)

func ValidName(name string) bool {
	return safeName.MatchString(name) && !strings.Contains(name, "..") && !strings.HasSuffix(name, ".") && !strings.HasSuffix(name, " ")
}
func (s *Server) receiveFile(c net.Conn, h wire.Header) error {
	id, e := hex.DecodeString(h.ID)
	if e != nil || len(id) != 16 {
		return errors.New("invalid transfer ID")
	}
	hash, e := hex.DecodeString(h.SHA256)
	if e != nil || len(hash) != 32 {
		return errors.New("invalid SHA-256")
	}
	if !ValidName(h.Name) || h.Size < 0 || h.Size > MaxFile {
		return errors.New("invalid file name or size (maximum 8 GiB)")
	}
	// Generated prefix avoids Windows device names, collisions and destination overwrites.
	f, e := os.CreateTemp(s.dir, ".partial-")
	if e != nil {
		return e
	}
	tmp := f.Name()
	defer os.Remove(tmp)
	defer f.Close()
	if e = wire.Send(c, wire.Header{Op: "ready", ID: h.ID}); e != nil {
		return e
	}
	hasher := sha256.New()
	var total int64
	last := time.Now()
	for {
		c.SetDeadline(time.Now().Add(30 * time.Second))
		k, b, e := wire.Read(c)
		if e != nil {
			return e
		}
		if k == wire.End {
			if len(b) != 0 {
				return errors.New("end frame must be empty")
			}
			break
		}
		if k != wire.Chunk || len(b) == 0 || len(b) > wire.ChunkSize || total+int64(len(b)) > h.Size {
			return errors.New("invalid file chunk")
		}
		if _, e = f.Write(b); e != nil {
			return e
		}
		hasher.Write(b)
		total += int64(len(b))
		if time.Since(last) > 100*time.Millisecond {
			s.emit(Event{Kind: "progress", ID: h.ID, Name: h.Name, Bytes: total, Total: h.Size, Status: "receiving"})
			last = time.Now()
		}
	}
	actual := hex.EncodeToString(hasher.Sum(nil))
	if total != h.Size || actual != h.SHA256 {
		return errors.New("integrity mismatch")
	}
	if e = f.Sync(); e != nil {
		return e
	}
	if e = f.Close(); e != nil {
		return e
	}
	dest := filepath.Join(s.dir, ID()+"-"+h.Name)
	if e = os.Rename(tmp, dest); e != nil {
		return e
	}
	if e = s.emit(Event{Kind: "transfer", ID: h.ID, Name: h.Name, Path: dest, SHA256: actual, Bytes: total, Total: h.Size, Status: "verified"}); e != nil {
		os.Remove(dest)
		return errors.New("history unavailable")
	}
	return wire.Send(c, wire.Header{Op: "complete", ID: h.ID, SHA256: actual, Size: total})
}

type Progress func(Event)

func SendFile(ctx context.Context, addr, path string, progress Progress) (wire.Header, error) {
	f, e := os.Open(path)
	if e != nil {
		return wire.Header{}, e
	}
	defer f.Close()
	st, e := f.Stat()
	if e != nil {
		return wire.Header{}, e
	}
	if !st.Mode().IsRegular() || st.Size() > MaxFile || !ValidName(filepath.Base(path)) {
		return wire.Header{}, errors.New("unsupported file size or name")
	}
	id := ID()
	emit := func(n int64, state string) {
		if progress != nil {
			progress(Event{Kind: "progress", ID: id, Name: st.Name(), Bytes: n, Total: st.Size(), Status: state, At: time.Now().UTC()})
		}
	}
	hash := sha256.New()
	buf := make([]byte, wire.ChunkSize)
	emit(0, "hashing")
	for {
		if e = ctx.Err(); e != nil {
			return wire.Header{}, e
		}
		n, err := f.Read(buf)
		if n > 0 {
			hash.Write(buf[:n])
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return wire.Header{}, err
		}
	}
	if _, e = f.Seek(0, 0); e != nil {
		return wire.Header{}, e
	}
	c, e := (&net.Dialer{Timeout: 5 * time.Second}).DialContext(ctx, "tcp", addr)
	if e != nil {
		return wire.Header{}, e
	}
	defer c.Close()
	stop := context.AfterFunc(ctx, func() { c.Close() })
	defer stop()
	h := wire.Header{Op: "put", ID: id, Name: st.Name(), Size: st.Size(), SHA256: hex.EncodeToString(hash.Sum(nil))}
	c.SetDeadline(time.Now().Add(30 * time.Second))
	if e = wire.Send(c, h); e != nil {
		return h, e
	}
	r, e := wire.Receive(c)
	if e != nil {
		return h, e
	}
	if r.Op != "ready" {
		return h, fmt.Errorf("receiver: %s", r.Error)
	}
	var total int64
	last := time.Now()
	emit(0, "sending")
	for {
		n, err := f.Read(buf)
		if n > 0 {
			c.SetDeadline(time.Now().Add(30 * time.Second))
			if e = wire.Write(c, wire.Chunk, buf[:n]); e != nil {
				return h, e
			}
			total += int64(n)
			if time.Since(last) > 100*time.Millisecond {
				emit(total, "sending")
				last = time.Now()
			}
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return h, err
		}
	}
	if e = wire.Write(c, wire.End, nil); e != nil {
		return h, e
	}
	r, e = wire.Receive(c)
	if e != nil {
		return h, e
	}
	if r.Op != "complete" || r.ID != id || r.Size != st.Size() || r.SHA256 != h.SHA256 {
		return h, fmt.Errorf("verification failed: %s", r.Error)
	}
	emit(total, "verified")
	return h, nil
}
