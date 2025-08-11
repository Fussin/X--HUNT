package storage

import (
	"encoding/json"
	"go.etcd.io/bbolt"
	"time"
)

// Store is a BoltDB-based storage for sessions.
type Store struct {
	db *bbolt.DB
}

// NewStore creates a new Store.
func NewStore(path string) (*Store, error) {
	db, err := bbolt.Open(path, 0600, &bbolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		return nil, err
	}

	// Create the "sessions" bucket if it doesn't exist.
	err = db.Update(func(tx *bbolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte("sessions"))
		return err
	})
	if err != nil {
		return nil, err
	}

	return &Store{db: db}, nil
}

// Close closes the database.
func (s *Store) Close() error {
	return s.db.Close()
}

// Request represents a stored HTTP request.
type Request struct {
	Raw []byte
}

// Response represents a stored HTTP response.
type Response struct {
	Raw []byte
}

// Entry represents a request-response pair.
type Entry struct {
	Request  Request
	Response Response
}

// Session represents a collection of entries.
type Session struct {
	ID      []byte
	Entries []Entry
}

// SaveSession saves a session to the database.
func (s *Store) SaveSession(session *Session) error {
	return s.db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("sessions"))
		data, err := json.Marshal(session)
		if err != nil {
			return err
		}
		return b.Put(session.ID, data)
	})
}

// GetSession retrieves a session from the database.
func (s *Store) GetSession(id []byte) (*Session, error) {
	var session Session
	err := s.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("sessions"))
		data := b.Get(id)
		if data == nil {
			return nil // Not found
		}
		return json.Unmarshal(data, &session)
	})
	if err != nil {
		return nil, err
	}
	return &session, nil
}
