package websocket

import (
	"io"
	"log"
	"net"
)

// Processor handles a WebSocket connection.
type Processor struct {
	clientConn net.Conn
	targetConn net.Conn
}

// NewProcessor creates a new WebSocket processor.
func NewProcessor(clientConn, targetConn net.Conn) *Processor {
	return &Processor{
		clientConn: clientConn,
		targetConn: targetConn,
	}
}

// Process handles the WebSocket connection.
func (p *Processor) Process() {
	log.Printf("WebSocket processor started")

	done := make(chan struct{})

	go func() {
		defer p.clientConn.Close()
		defer p.targetConn.Close()
		io.Copy(p.targetConn, p.clientConn)
		done <- struct{}{}
	}()

	go func() {
		defer p.clientConn.Close()
		defer p.targetConn.Close()
		io.Copy(p.clientConn, p.targetConn)
		done <- struct{}{}
	}()

	<-done
	log.Printf("WebSocket connection closed")
}
