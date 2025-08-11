package http2

import (
	"bytes"
	"testing"
)

func TestFramer_ReadFrame(t *testing.T) {
	// Create a buffer with a sample frame.
	// This is a SETTINGS frame with no payload.
	// Header: length=0, type=SETTINGS, flags=0, streamID=0
	var buf bytes.Buffer
	buf.Write([]byte{0x00, 0x00, 0x00, 0x04, 0x00, 0x00, 0x00, 0x00, 0x00})

	framer := NewFramer(&buf)
	fh, payload, err := framer.ReadFrame()
	if err != nil {
		t.Fatalf("ReadFrame failed: %v", err)
	}

	if fh.Length != 0 {
		t.Errorf("Expected length 0, got %d", fh.Length)
	}
	if fh.Type != FrameSettings {
		t.Errorf("Expected type SETTINGS, got %v", fh.Type)
	}
	if fh.Flags != 0 {
		t.Errorf("Expected flags 0, got %d", fh.Flags)
	}
	if fh.StreamID != 0 {
		t.Errorf("Expected stream ID 0, got %d", fh.StreamID)
	}
	if len(payload) != 0 {
		t.Errorf("Expected empty payload, got %d bytes", len(payload))
	}
}

func TestParseSettingsFrame(t *testing.T) {
	// A sample SETTINGS frame payload with two settings.
	// Setting 1: id=1, value=100
	// Setting 2: id=2, value=200
	payload := []byte{
		0x00, 0x01, 0x00, 0x00, 0x00, 0x64, // id=1, value=100
		0x00, 0x02, 0x00, 0x00, 0x00, 0xc8, // id=2, value=200
	}

	settings, err := ParseSettingsFrame(payload)
	if err != nil {
		t.Fatalf("ParseSettingsFrame failed: %v", err)
	}

	if len(settings) != 2 {
		t.Fatalf("Expected 2 settings, got %d", len(settings))
	}

	if settings[0].ID != 1 || settings[0].Value != 100 {
		t.Errorf("Expected setting {1, 100}, got {%d, %d}", settings[0].ID, settings[0].Value)
	}
	if settings[1].ID != 2 || settings[1].Value != 200 {
		t.Errorf("Expected setting {2, 200}, got {%d, %d}", settings[1].ID, settings[1].Value)
	}
}
