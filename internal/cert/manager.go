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

// Manager управляет генерацией и кэшированием сертификатов для MITM
type Manager struct {
	ca    *x509.Certificate
	caKey *rsa.PrivateKey
	cache sync.Map // host -> *tls.Certificate
	mu    sync.Mutex
}

// NewManager создает новый менеджер сертификатов с корневым CA
func NewManager() (*Manager, error) {
	// Создаем корневой CA сертификат
	ca := &x509.Certificate{
		SerialNumber: big.NewInt(time.Now().Unix()),
		Subject: pkix.Name{
			Organization: []string{"4ebur-net MITM Proxy"},
			CommonName:   "4ebur-net Root CA",
		},
		NotBefore:             time.Now().Add(-time.Hour),
		NotAfter:              time.Now().AddDate(10, 0, 0),
		IsCA:                  true,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		BasicConstraintsValid: true,
	}

	// Генерируем приватный ключ для CA (2048-bit для баланса безопасности/производительности)
	caKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, fmt.Errorf("failed to generate CA key: %w", err)
	}

	// Самоподписываем CA сертификат
	caBytes, err := x509.CreateCertificate(rand.Reader, ca, ca, &caKey.PublicKey, caKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create CA certificate: %w", err)
	}

	ca, err = x509.ParseCertificate(caBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse CA certificate: %w", err)
	}

	return &Manager{
		ca:    ca,
		caKey: caKey,
	}, nil
}

// GetCertForHost возвращает сертификат для указанного хоста (из кэша или генерирует новый)
func (m *Manager) GetCertForHost(host string) (*tls.Certificate, error) {
	// Быстрая проверка кэша без блокировки
	if cached, ok := m.cache.Load(host); ok {
		return cached.(*tls.Certificate), nil
	}

	// Генерируем новый сертификат с блокировкой
	m.mu.Lock()
	defer m.mu.Unlock()

	// Double-check после получения блокировки (может уже создать другая горутина)
	if cached, ok := m.cache.Load(host); ok {
		return cached.(*tls.Certificate), nil
	}

	// Создаем новый сертификат
	cert := &x509.Certificate{
		SerialNumber: big.NewInt(time.Now().UnixNano()),
		Subject: pkix.Name{
			Organization: []string{"4ebur-net"},
			CommonName:   host,
		},
		DNSNames:    []string{host},
		NotBefore:   time.Now().Add(-time.Hour),
		NotAfter:    time.Now().AddDate(1, 0, 0),
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		KeyUsage:    x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
	}

	// Генерируем ключ для сертификата
	certKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, fmt.Errorf("failed to generate cert key: %w", err)
	}

	// Подписываем сертификат нашим CA
	certBytes, err := x509.CreateCertificate(rand.Reader, cert, m.ca, &certKey.PublicKey, m.caKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create certificate: %w", err)
	}

	// Кодируем в PEM формат
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certBytes})
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(certKey)})

	tlsCert, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		return nil, fmt.Errorf("failed to create TLS certificate: %w", err)
	}

	// Сохраняем в кэш для последующего использования
	m.cache.Store(host, &tlsCert)
	return &tlsCert, nil
}

// GetCAPEM возвращает CA сертификат в формате PEM для установки в trust store
func (m *Manager) GetCAPEM() []byte {
	return pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: m.ca.Raw,
	})
}

// GetCacheSize возвращает количество закэшированных сертификатов
func (m *Manager) GetCacheSize() int {
	count := 0
	m.cache.Range(func(_, _ interface{}) bool {
		count++
		return true
	})
	return count
}
