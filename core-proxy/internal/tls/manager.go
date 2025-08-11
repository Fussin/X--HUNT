package tls

import (
	"crypto/tls"
	"sync"
)

// Manager is a TLS manager that can generate certificates on the fly.
type Manager struct {
	ca    *CA
	cache sync.Map // host -> *tls.Certificate
}

// NewManager creates a new TLS manager.
func NewManager() (*Manager, error) {
	ca, err := NewCA()
	if err != nil {
		return nil, err
	}

	return &Manager{
		ca: ca,
	}, nil
}

// GetCertificate returns a certificate for the given client hello info.
func (m *Manager) GetCertificate(hello *tls.ClientHelloInfo) (*tls.Certificate, error) {
	if cert, ok := m.cache.Load(hello.ServerName); ok {
		return cert.(*tls.Certificate), nil
	}

	cert, key, err := m.ca.SignHost(hello.ServerName)
	if err != nil {
		return nil, err
	}

	tlsCert := &tls.Certificate{
		Certificate: [][]byte{cert.Raw},
		PrivateKey:  key,
	}

	m.cache.Store(hello.ServerName, tlsCert)
	return tlsCert, nil
}
