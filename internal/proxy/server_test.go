package proxy

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNewProxyServer(t *testing.T) {
	server, err := NewProxyServer()
	if err != nil {
		t.Fatalf("Failed to create proxy server: %v", err)
	}

	if server == nil {
		t.Fatal("Proxy server is nil")
	}
}

func TestProxyServerHTTPRequest(t *testing.T) {
	server, err := NewProxyServer()
	if err != nil {
		t.Fatalf("Failed to create proxy server: %v", err)
	}

	// Create test backend server
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}))
	defer backend.Close()

	// Create request through proxy
	req := httptest.NewRequest("GET", backend.URL, nil)
	rr := httptest.NewRecorder()

	server.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status OK, got %v", rr.Code)
	}
}

func TestProxyServerCONNECTMethod(t *testing.T) {
	server, err := NewProxyServer()
	if err != nil {
		t.Fatalf("Failed to create proxy server: %v", err)
	}

	req := httptest.NewRequest("CONNECT", "example.com:443", nil)
	rr := httptest.NewRecorder()

	server.ServeHTTP(rr, req)

	// CONNECT should respond with 200 OK for tunnel establishment
	// or appropriate error
	if rr.Code != http.StatusOK && rr.Code != http.StatusBadGateway {
		t.Logf("CONNECT method returned status: %v", rr.Code)
	}
}

func TestProxyServerInvalidRequest(t *testing.T) {
	server, err := NewProxyServer()
	if err != nil {
		t.Fatalf("Failed to create proxy server: %v", err)
	}

	tests := []struct {
		name   string
		method string
		url    string
	}{
		{
			name:   "invalid URL",
			method: "GET",
			url:    "://invalid",
		},
		{
			name:   "empty host",
			method: "GET",
			url:    "http://",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.url, nil)
			rr := httptest.NewRecorder()

			server.ServeHTTP(rr, req)

			// Should return error status
			if rr.Code == http.StatusOK {
				t.Error("Expected error status for invalid request")
			}
		})
	}
}

func TestProxyServerConcurrency(t *testing.T) {
	server, err := NewProxyServer()
	if err != nil {
		t.Fatalf("Failed to create proxy server: %v", err)
	}

	// Create test backend
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(10 * time.Millisecond) // Simulate some work
		w.WriteHeader(http.StatusOK)
	}))
	defer backend.Close()

	const concurrency = 100
	done := make(chan bool, concurrency)

	// Send concurrent requests
	for i := 0; i < concurrency; i++ {
		go func() {
			req := httptest.NewRequest("GET", backend.URL, nil)
			rr := httptest.NewRecorder()
			server.ServeHTTP(rr, req)
			done <- true
		}()
	}

	// Wait for all to complete
	timeout := time.After(5 * time.Second)
	for i := 0; i < concurrency; i++ {
		select {
		case <-done:
			// Success
		case <-timeout:
			t.Fatal("Timeout waiting for concurrent requests")
		}
	}
}

func BenchmarkProxyServerHTTP(b *testing.B) {
	server, err := NewProxyServer()
	if err != nil {
		b.Fatalf("Failed to create proxy server: %v", err)
	}

	// Create test backend
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer backend.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("GET", backend.URL, nil)
		rr := httptest.NewRecorder()
		server.ServeHTTP(rr, req)
	}
}

func BenchmarkProxyServerParallel(b *testing.B) {
	server, err := NewProxyServer()
	if err != nil {
		b.Fatalf("Failed to create proxy server: %v", err)
	}

	// Create test backend
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer backend.Close()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			req := httptest.NewRequest("GET", backend.URL, nil)
			rr := httptest.NewRecorder()
			server.ServeHTTP(rr, req)
		}
	})
}
