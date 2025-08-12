package http2

import (
	"reflect"
	"testing"

	"golang.org/x/net/http2/hpack"
)

func TestDecodeHeaders(t *testing.T) {
	conn := NewConnection()

	// A sample HEADERS frame payload with a single header: :method: GET
	payload := []byte{0x82}

	headers, err := conn.DecodeHeaders(payload)
	if err != nil {
		t.Fatalf("DecodeHeaders failed: %v", err)
	}

	expected := []hpack.HeaderField{
		{Name: ":method", Value: "GET"},
	}

	if !reflect.DeepEqual(headers, expected) {
		t.Errorf("Expected headers %v, got %v", expected, headers)
	}
}
