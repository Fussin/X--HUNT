package tls

import (
	"crypto/tls"
	"testing"
)

func TestManager(t *testing.T) {
	manager, err := NewManager()
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	hello := &tls.ClientHelloInfo{
		ServerName: "example.com",
	}

	cert, err := manager.GetCertificate(hello)
	if err != nil {
		t.Fatalf("Failed to get certificate: %v", err)
	}

	if cert == nil {
		t.Fatal("Certificate is nil")
	}

	if len(cert.Certificate) == 0 {
		t.Fatal("Certificate chain is empty")
	}
}
