package tls

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"net"
	"os"
	"sync"
	"time"
)

// RootCA holds root CA cert and key
type RootCA struct {
	Cert *x509.Certificate
	Key  crypto.PrivateKey
	PEM  []byte // cert PEM
	KeyPEM []byte
}

// Manager signs leaf certs using root CA (in-memory or loaded)
type Manager struct {
	root   *RootCA
	cache  map[string]tlsCertEntry
	mtx    sync.RWMutex
	ttl    time.Duration
}

type tlsCertEntry struct {
	certPEM []byte
	keyPEM  []byte
	notAfter time.Time
}

type Config struct {
	// If provided, manager will try to load root CA from these files. If missing, generate new root in-memory.
	RootCertPath string
	RootKeyPath  string

	// Defaults for leaf certificates
	DefaultTTL time.Duration

	// On-disk cache directory for generated leaf certs (optional)
	CacheDir string
}

// NewManager creates a manager. If root files exist they will be loaded.
func NewManager(cfg Config) (*Manager, error) {
	m := &Manager{
		cache: make(map[string]tlsCertEntry),
		ttl:   24 * time.Hour,
	}
	if cfg.DefaultTTL > 0 {
		m.ttl = cfg.DefaultTTL
	}

	if cfg.RootCertPath != "" && cfg.RootKeyPath != "" {
		if _, err := os.Stat(cfg.RootCertPath); err == nil {
			root, err := loadRootFromFiles(cfg.RootCertPath, cfg.RootKeyPath)
			if err != nil {
				return nil, fmt.Errorf("loading root CA: %w", err)
			}
			m.root = root
			return m, nil
		}
	}

	// generate ephemeral root CA
	root, err := generateRootCA()
	if err != nil {
		return nil, fmt.Errorf("generate root CA: %w", err)
	}
	m.root = root
	return m, nil
}

// loadRootFromFiles loads PEM files
func loadRootFromFiles(certPath, keyPath string) (*RootCA, error) {
	certPEM, err := os.ReadFile(certPath)
	if err != nil { return nil, err }
	keyPEM, err := os.ReadFile(keyPath)
	if err != nil { return nil, err }
	block, _ := pem.Decode(certPEM)
	if block == nil { return nil, fmt.Errorf("invalid cert PEM") }
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil { return nil, err }

	blockk, _ := pem.Decode(keyPEM)
	if blockk == nil { return nil, fmt.Errorf("invalid key PEM") }
	var key crypto.PrivateKey
	if pk, err := x509.ParsePKCS1PrivateKey(blockk.Bytes); err == nil {
		key = pk
	} else if pk2, err := x509.ParseECPrivateKey(blockk.Bytes); err == nil {
		key = pk2
	} else {
		return nil, fmt.Errorf("unsupported private key type")
	}
	return &RootCA{Cert: cert, Key: key, PEM: certPEM, KeyPEM: keyPEM}, nil
}

// generateRootCA makes a new self-signed root CA (ECDSA)
func generateRootCA() (*RootCA, error) {
	// generate ECDSA key
	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil { return nil, err }

	serial, _ := rand.Int(rand.Reader, big.NewInt(1<<62))
	tmpl := &x509.Certificate{
		SerialNumber: serial,
		Subject: pkix.Name{
			Organization: []string{"SentinelX Root CA"},
			CommonName:   "SentinelX Root CA",
		},
		NotBefore: time.Now().Add(-1 * time.Hour),
		NotAfter:  time.Now().Add(10 * 365 * 24 * time.Hour),
		KeyUsage:  x509.KeyUsageCertSign | x509.KeyUsageCRLSign | x509.KeyUsageDigitalSignature,
		IsCA:      true,
		BasicConstraintsValid: true,
		MaxPathLen: 2,
	}
	certDER, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &priv.PublicKey, priv)
	if err != nil { return nil, err }

	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})
	keyBytes, err := x509.MarshalECPrivateKey(priv)
	if err != nil { return nil, err }
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: keyBytes})
	cert, err := x509.ParseCertificate(certDER)
	if err != nil { return nil, err }
	return &RootCA{Cert: cert, Key: priv, PEM: certPEM, KeyPEM: keyPEM}, nil
}

// GenerateLeafForHost signs a leaf certificate for the provided hostnames (sni)
func (m *Manager) GenerateLeafForHost(hosts []string) (certPEM, keyPEM []byte, err error) {
	if len(hosts) == 0 {
		return nil, nil, fmt.Errorf("no hostnames provided")
	}
	m.mtx.RLock()
	// naive cache key: first host
	key := hosts[0]
	if e, ok := m.cache[key]; ok && time.Now().Before(e.notAfter) {
		c := e
		m.mtx.RUnlock()
		return c.certPEM, c.keyPEM, nil
	}
	m.mtx.RUnlock()

	// generate RSA key for leaf (fast)
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil { return nil, nil, err }

	serial, _ := rand.Int(rand.Reader, big.NewInt(1<<62))
	notBefore := time.Now().Add(-1 * time.Hour)
	notAfter := time.Now().Add(m.ttl)

	tmpl := &x509.Certificate{
		SerialNumber: serial,
		Subject: pkix.Name{
			CommonName: hosts[0],
			Organization: []string{"SentinelX MITM"},
		},
		NotBefore: notBefore,
		NotAfter:  notAfter,
		KeyUsage:  x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
	}

	// add SANs
	var ips []net.IP
	for _, h := range hosts {
		if ip := net.ParseIP(h); ip != nil {
			ips = append(ips, ip)
		} else {
			tmpl.DNSNames = append(tmpl.DNSNames, h)
		}
	}
	if len(ips) > 0 {
		tmpl.IPAddresses = ips
	}

	leafDER, err := x509.CreateCertificate(rand.Reader, tmpl, m.root.Cert, &priv.PublicKey, m.root.Key)
	if err != nil { return nil, nil, err }

	certPEM = pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: leafDER})
	keyBytes := x509.MarshalPKCS1PrivateKey(priv)
	keyPEM = pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: keyBytes})

	// cache it
	m.mtx.Lock()
	m.cache[key] = tlsCertEntry{certPEM: certPEM, keyPEM: keyPEM, notAfter: notAfter}
	m.mtx.Unlock()
	return certPEM, keyPEM, nil
}

// RootPEM returns the root CA certificate PEM (for trusting)
func (m *Manager) RootPEM() []byte {
	if m.root == nil { return nil }
	return m.root.PEM
}

// SaveRootTo writes root cert and key to files (used for manual trust)
func (m *Manager) SaveRootTo(certPath, keyPath string) error {
	if m.root == nil { return fmt.Errorf("no root loaded") }
	if err := os.WriteFile(certPath, m.root.PEM, 0644); err != nil { return err }
	if err := os.WriteFile(keyPath, m.root.KeyPEM, 0600); err != nil { return err }
	return nil
}
