package storage

import (
	"os"
	"reflect"
	"testing"
)

func TestStore(t *testing.T) {
	// Create a temporary database for the test.
	f, err := os.CreateTemp("", "sentinelx-test.db")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(f.Name())

	store, err := NewStore(f.Name())
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}
	defer store.Close()

	// Create a sample session.
	session := &Session{
		ID: []byte("test-session"),
		Entries: []Entry{
			{
				Request:  Request{Raw: []byte("GET / HTTP/1.1")},
				Response: Response{Raw: []byte("HTTP/1.1 200 OK")},
			},
		},
	}

	// Save the session.
	if err := store.SaveSession(session); err != nil {
		t.Fatalf("Failed to save session: %v", err)
	}

	// Get the session.
	retrieved, err := store.GetSession(session.ID)
	if err != nil {
		t.Fatalf("Failed to get session: %v", err)
	}

	// Compare the sessions.
	if !reflect.DeepEqual(session, retrieved) {
		t.Errorf("Expected session %v, got %v", session, retrieved)
	}
}
