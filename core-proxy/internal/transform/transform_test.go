package transform

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"testing"
)

func TestMatchReplaceInterceptor(t *testing.T) {
	interceptor := &MatchReplaceInterceptor{
		Match:   []byte("Hello"),
		Replace: []byte("Goodbye"),
	}

	res := &http.Response{
		Body: io.NopCloser(bytes.NewBufferString("Hello, world!")),
	}

	if err := interceptor.InterceptResponse(res); err != nil {
		t.Fatalf("InterceptResponse failed: %v", err)
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}

	expected := "Goodbye, world!"
	if string(body) != expected {
		t.Errorf("Expected body '%s', got '%s'", expected, string(body))
	}
}

func TestJSONInterceptor(t *testing.T) {
	interceptor := &JSONInterceptor{
		ModifyFunc: func(data map[string]interface{}) (map[string]interface{}, error) {
			data["modified"] = true
			return data, nil
		},
	}

	res := &http.Response{
		Body: io.NopCloser(bytes.NewBufferString(`{"hello": "world"}`)),
	}

	if err := interceptor.InterceptResponse(res); err != nil {
		t.Fatalf("InterceptResponse failed: %v", err)
	}

	var data map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&data); err != nil {
		t.Fatalf("Failed to decode response body: %v", err)
	}

	if modified, ok := data["modified"]; !ok || !modified.(bool) {
		t.Errorf("Expected 'modified' to be true, got %v", data["modified"])
	}
}
