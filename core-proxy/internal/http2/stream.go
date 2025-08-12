package http2

// StreamState is the state of an HTTP/2 stream.
type StreamState int

const (
	StreamStateIdle StreamState = iota
	StreamStateOpen
	StreamStateHalfClosedLocal
	StreamStateHalfClosedRemote
	StreamStateClosed
)

// Stream represents an HTTP/2 stream.
type Stream struct {
	ID    uint32
	State StreamState
	// TODO: Add flow control window and other fields.
}
