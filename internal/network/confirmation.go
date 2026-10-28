package network

import (
	"context"
	"errors"
	"fmt"

	"github.com/ISI0007/pulseroute-network-lab/internal/wire"
)

// SendConfirmed waits for server acceptance and its socket-write count.
// Timeout means the outcome is unknown; this method never resends automatically.
func (c *Client) SendConfirmed(ctx context.Context, text string) (wire.Header, error) {
	if err := ctx.Err(); err != nil {
		return wire.Header{}, err
	}
	if len(text) == 0 || len(text) > 8192 {
		return wire.Header{}, fmt.Errorf("message must be 1..8192 bytes")
	}
	id := ID()
	result := make(chan wire.Header, 1)
	c.pendingMu.Lock()
	if len(c.pending) >= 128 {
		c.pendingMu.Unlock()
		return wire.Header{}, errors.New("too many messages awaiting acknowledgment")
	}
	c.pending[id] = result
	c.pendingMu.Unlock()
	defer func() { c.pendingMu.Lock(); delete(c.pending, id); c.pendingMu.Unlock() }()
	if err := c.p.send(wire.Header{Op: "message", ID: id, Text: text}); err != nil {
		return wire.Header{ID: id}, err
	}
	select {
	case h := <-result:
		if h.Op == "error" {
			return h, fmt.Errorf("server rejected message: %s", h.Error)
		}
		return h, nil
	case <-ctx.Done():
		return wire.Header{ID: id}, fmt.Errorf("acknowledgment unavailable; delivery outcome unknown: %w", ctx.Err())
	case <-c.done:
		return wire.Header{ID: id}, errors.New("connection closed before acknowledgment; delivery outcome unknown")
	}
}
