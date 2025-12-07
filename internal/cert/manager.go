package cert

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"sync"
	"time"
)

// CertManager manages TLS certificates for MITM proxy
type CertManager struct {
	ca      *x509.Certificate
	caKey   *rsa.PrivateKey
	certMap sync.Map // hostname -> *tls.Certificate
}

// NewCertManager creates a new certificate manager with a CA certificate
func NewCertManager() (*CertManager, error) {
	// Generate CA private key
	caKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, fmt.Errorf("failed to generate CA key: %w", err)
	}

	// Create CA certificate template
	caTemplate := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization: []string{"4ebur-net MITM Proxy"},
			CommonName:   "4ebur-net CA",
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(10, 0, 0), // 10 years
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		IsCA:                  true,
	}

	// Create self-signed CA certificate
	caCertDER, err := x509.CreateCertificate(rand.Reader, caTemplate, caTemplate, &caKey.PublicKey, caKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create CA certificate: %w", err)
	}

	caCert, err := x509.ParseCertificate(caCertDER)
	if err != nil {
		return nil, fmt.Errorf("failed to parse CA certificate: %w", err)
	}

	return &CertManager{
		ca:    caCert,
		caKey: caKey,
	}, nil
}

// GetCertificate returns a certificate for the given hostname
func (m *CertManager) GetCertificate(hostname string) (*tls.Certificate, error) {
	// Check cache
	if cert, ok := m.certMap.Load(hostname); ok {
		return cert.(*tls.Certificate), nil
	}

	// Generate new certificate
	cert, err := m.generateCertificate(hostname)
	if err != nil {
		return nil, err
	}

	// Store in cache
	m.certMap.Store(hostname, cert)
	return cert, nil
}

// generateCertificate creates a new certificate for the hostname
func (m *CertManager) generateCertificate(hostname string) (*tls.Certificate, error) {
	// Generate private key for host
	hostKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, fmt.Errorf("failed to generate host key: %w", err)
	}

	// Create certificate template
	template := &x509.Certificate{
		SerialNumber: big.NewInt(time.Now().Unix()),
		Subject: pkix.Name{
			Organization: []string{"4ebur-net MITM"},
			CommonName:   hostname,
		},
		DNSNames:    []string{hostname},
		NotBefore:   time.Now(),
		NotAfter:    time.Now().AddDate(1, 0, 0), // 1 year
		KeyUsage:    x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
	}

	// Sign certificate with CA
	certDER, err := x509.CreateCertificate(rand.Reader, template, m.ca, &hostKey.PublicKey, m.caKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create certificate: %w", err)
	}

	// Create TLS certificate
	tlsCert := &tls.Certificate{
		Certificate: [][]byte{certDER, m.ca.Raw},
		PrivateKey:  hostKey,
	}

	return tlsCert, nil
}

// GetCACertPEM returns the CA certificate in PEM format
func (m *CertManager) GetCACertPEM() []byte {
	return pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: m.ca.Raw,
	})
}
