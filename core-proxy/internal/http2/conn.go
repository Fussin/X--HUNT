package http2

import (
	"bytes"
	"golang.org/x/net/http2/hpack"
	"sync"
)

// Connection manages the state of an HTTP/2 connection.
type Connection struct {
	mu            sync.Mutex
	streams       map[uint32]*Stream
	hpackDecoder  *hpack.Decoder
	hpackEncoder  *hpack.Encoder
	// TODO: Add flow control windows.
}

// NewConnection creates a new HTTP/2 connection.
func NewConnection() *Connection {
	var buf bytes.Buffer
	return &Connection{
		streams:      make(map[uint32]*Stream),
		hpackDecoder: hpack.NewDecoder(4096, nil),
		hpackEncoder: hpack.NewEncoder(&buf),
	}
}

// GetStream returns a stream by its ID.
func (c *Connection) GetStream(id uint32) *Stream {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.streams[id]
}

// AddStream adds a new stream to the connection.
func (c *Connection) AddStream(stream *Stream) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.streams[stream.ID] = stream
}

// DecodeHeaders decodes a HEADERS frame payload.
func (c *Connection) DecodeHeaders(payload []byte) ([]hpack.HeaderField, error) {
	return c.hpackDecoder.DecodeFull(payload)
}
