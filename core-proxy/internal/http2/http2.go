package http2

import (
	"bytes"
	"io"
	"log"
	"net"
)

// The client connection preface is a 24-byte sequence that the client sends
// at the beginning of an HTTP/2 connection.
const clientPreface = "PRI * HTTP/2.0\r\n\r\nSM\r\n\r\n"

// Processor handles a single HTTP/2 connection.
type Processor struct {
	clientConn net.Conn
}

// NewProcessor creates a new HTTP/2 processor.
func NewProcessor(clientConn net.Conn) *Processor {
	return &Processor{
		clientConn: clientConn,
	}
}

// Process handles the HTTP/2 connection.
func (p *Processor) Process() {
	log.Printf("HTTP/2 processor started for %s", p.clientConn.RemoteAddr())

	// Send an initial empty SETTINGS frame.
	// TODO: Send our actual settings.
	if _, err := p.clientConn.Write([]byte{0x00, 0x00, 0x00, 0x04, 0x00, 0x00, 0x00, 0x00, 0x00}); err != nil {
		log.Printf("Failed to write initial SETTINGS frame: %v", err)
		return
	}

	// Read the connection preface.
	preface := make([]byte, len(clientPreface))
	if _, err := io.ReadFull(p.clientConn, preface); err != nil {
		log.Printf("Failed to read preface: %v", err)
		return
	}

	if !bytes.Equal(preface, []byte(clientPreface)) {
		log.Printf("Invalid preface received")
		return
	}

	log.Printf("Preface received successfully")

	framer := NewFramer(p.clientConn)
	for {
		fh, payload, err := framer.ReadFrame()
		if err != nil {
			log.Printf("Failed to read frame: %v", err)
			return
		}
		log.Printf("Read frame: type=%v, length=%d", fh.Type, fh.Length)

		switch fh.Type {
		case FrameSettings:
			settings, err := ParseSettingsFrame(payload)
			if err != nil {
				log.Printf("Failed to parse SETTINGS frame: %v", err)
				return
			}
			log.Printf("Received SETTINGS frame: %v", settings)
		default:
			// TODO: Handle other frame types.
		}
	}
}
