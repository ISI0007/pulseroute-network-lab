package wire

import (
	"bytes"
	"encoding/binary"
	"io"
	"testing"
)

type shortWriter struct{ bytes.Buffer }

func (w *shortWriter) Write(p []byte) (int, error) {
	if len(p) > 2 {
		p = p[:2]
	}
	return w.Buffer.Write(p)
}
func TestBinaryAndShortWrites(t *testing.T) {
	w := new(shortWriter)
	p := []byte{0, 255, 3, 0}
	if e := Write(w, Chunk, p); e != nil {
		t.Fatal(e)
	}
	k, b, e := Read(&w.Buffer)
	if e != nil || k != Chunk || !bytes.Equal(p, b) {
		t.Fatalf("%d %v %v", k, b, e)
	}
}
func TestBoundsAndTruncation(t *testing.T) {
	for _, n := range []uint32{0, MaxFrame + 1} {
		var b [4]byte
		binary.BigEndian.PutUint32(b[:], n)
		if _, _, e := Read(bytes.NewReader(b[:])); e == nil {
			t.Fatal("accepted invalid length")
		}
	}
	if _, _, e := Read(bytes.NewReader([]byte{0, 0, 0, 3, 2})); e != io.ErrUnexpectedEOF {
		t.Fatal(e)
	}
}
func FuzzRead(f *testing.F) {
	f.Add([]byte{0, 0, 0, 2, 1, 0})
	f.Fuzz(func(t *testing.T, b []byte) { Read(bytes.NewReader(b)) })
}
