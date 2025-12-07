package cert

import (
	"crypto/tls"
	"testing"
	"time"
)

func TestNewCertManager(t *testing.T) {
	manager, err := NewCertManager()
	if err != nil {
		t.Fatalf("Failed to create cert manager: %v", err)
	}

	if manager == nil {
		t.Fatal("Cert manager is nil")
	}

	if manager.ca == nil {
		t.Fatal("CA certificate is nil")
	}

	if manager.caKey == nil {
		t.Fatal("CA private key is nil")
	}
}

func TestGetCertificate(t *testing.T) {
	manager, err := NewCertManager()
	if err != nil {
		t.Fatalf("Failed to create cert manager: %v", err)
	}

	tests := []struct {
		name     string
		hostname string
		wantErr  bool
	}{
		{
			name:     "valid hostname",
			hostname: "example.com",
			wantErr:  false,
		},
		{
			name:     "subdomain",
			hostname: "api.example.com",
			wantErr:  false,
		},
		{
			name:     "localhost",
			hostname: "localhost",
			wantErr:  false,
		},
		{
			name:     "IP address",
			hostname: "192.168.1.1",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cert, err := manager.GetCertificate(tt.hostname)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetCertificate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if cert == nil {
					t.Error("Certificate is nil")
					return
				}

				if len(cert.Certificate) == 0 {
					t.Error("Certificate chain is empty")
				}

				if cert.PrivateKey == nil {
					t.Error("Private key is nil")
				}
			}
		})
	}
}

func TestCertificateCaching(t *testing.T) {
	manager, err := NewCertManager()
	if err != nil {
		t.Fatalf("Failed to create cert manager: %v", err)
	}

	hostname := "cache-test.example.com"

	// First call - should create new cert
	cert1, err := manager.GetCertificate(hostname)
	if err != nil {
		t.Fatalf("First GetCertificate() failed: %v", err)
	}

	// Second call - should return cached cert
	cert2, err := manager.GetCertificate(hostname)
	if err != nil {
		t.Fatalf("Second GetCertificate() failed: %v", err)
	}

	// Certificates should be identical (cached)
	if cert1 != cert2 {
		t.Error("Certificates are not cached - different instances returned")
	}
}

func TestConcurrentCertificateGeneration(t *testing.T) {
	manager, err := NewCertManager()
	if err != nil {
		t.Fatalf("Failed to create cert manager: %v", err)
	}

	const concurrency = 50
	hostname := "concurrent-test.example.com"

	done := make(chan *tls.Certificate, concurrency)
	errors := make(chan error, concurrency)

	// Start concurrent requests
	for i := 0; i < concurrency; i++ {
		go func() {
			cert, err := manager.GetCertificate(hostname)
			if err != nil {
				errors <- err
				return
			}
			done <- cert
		}()
	}

	// Collect results
	for i := 0; i < concurrency; i++ {
		select {
		case <-done:
			// Success
		case err := <-errors:
			t.Errorf("Concurrent certificate generation failed: %v", err)
		case <-time.After(5 * time.Second):
			t.Fatal("Timeout waiting for certificate generation")
		}
	}
}

func BenchmarkGetCertificate(b *testing.B) {
	manager, err := NewCertManager()
	if err != nil {
		b.Fatalf("Failed to create cert manager: %v", err)
	}

	hostname := "benchmark.example.com"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := manager.GetCertificate(hostname)
		if err != nil {
			b.Fatalf("GetCertificate() failed: %v", err)
		}
	}
}

func BenchmarkConcurrentGetCertificate(b *testing.B) {
	manager, err := NewCertManager()
	if err != nil {
		b.Fatalf("Failed to create cert manager: %v", err)
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			hostname := "benchmark-" + string(rune(i%10)) + ".example.com"
			_, err := manager.GetCertificate(hostname)
			if err != nil {
				b.Errorf("GetCertificate() failed: %v", err)
			}
			i++
		}
	})
}
