package http2

import (
	"encoding/binary"
	"errors"
	"io"
)

// Frame types
const (
	FrameData         FrameType = 0x0
	FrameHeaders      FrameType = 0x1
	FramePriority     FrameType = 0x2
	FrameRSTStream    FrameType = 0x3
	FrameSettings     FrameType = 0x4
	FramePushPromise  FrameType = 0x5
	FramePing         FrameType = 0x6
	FrameGoAway       FrameType = 0x7
	FrameWindowUpdate FrameType = 0x8
	FrameContinuation FrameType = 0x9
)

// FrameType is the type of an HTTP/2 frame.
type FrameType uint8

// FrameHeader is the 9-byte header of an HTTP/2 frame.
type FrameHeader struct {
	Length   uint32
	Type     FrameType
	Flags    uint8
	StreamID uint32
}

// readFrameHeader reads a frame header from the given reader.
func readFrameHeader(buf []byte) (FrameHeader, error) {
	fh := FrameHeader{
		Length:   (uint32(buf[0])<<16 | uint32(buf[1])<<8 | uint32(buf[2])),
		Type:     FrameType(buf[3]),
		Flags:    buf[4],
		StreamID: binary.BigEndian.Uint32(buf[5:]) & (1<<31 - 1),
	}
	return fh, nil
}

// Framer reads and writes HTTP/2 frames.
type Framer struct {
	r io.Reader
}

// NewFramer creates a new Framer.
func NewFramer(r io.Reader) *Framer {
	return &Framer{r: r}
}

// ReadFrame reads a single frame.
func (f *Framer) ReadFrame() (FrameHeader, []byte, error) {
	headerBuf := make([]byte, 9)
	if _, err := io.ReadFull(f.r, headerBuf); err != nil {
		return FrameHeader{}, nil, err
	}

	fh, err := readFrameHeader(headerBuf)
	if err != nil {
		return FrameHeader{}, nil, err
	}

	payload := make([]byte, fh.Length)
	if _, err := io.ReadFull(f.r, payload); err != nil {
		return FrameHeader{}, nil, err
	}

	return fh, payload, nil
}

// Setting is a single setting in a SETTINGS frame.
type Setting struct {
	ID    uint16
	Value uint32
}

// SettingsFrame is the payload of a SETTINGS frame.
type SettingsFrame []Setting

// ParseSettingsFrame parses the payload of a SETTINGS frame.
func ParseSettingsFrame(payload []byte) (SettingsFrame, error) {
	if len(payload)%6 != 0 {
		return nil, errors.New("invalid settings frame payload length")
	}
	var settings SettingsFrame
	for i := 0; i < len(payload); i += 6 {
		id := binary.BigEndian.Uint16(payload[i:])
		value := binary.BigEndian.Uint32(payload[i+2:])
		settings = append(settings, Setting{ID: id, Value: value})
	}
	return settings, nil
}
