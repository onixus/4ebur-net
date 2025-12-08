package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/onixus/4ebur-net/internal/proxy"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	// Создаем прокси-сервер
	proxyServer, err := proxy.NewProxyServer()
	if err != nil {
		log.Fatalf("Failed to create proxy server: %v", err)
	}

	// Получаем порт из переменной окружения или используем по умолчанию
	port := os.Getenv("PROXY_PORT")
	if port == "" {
		port = "1488"
	}

	// Создаем HTTP mux для специальных endpoints
	mux := http.NewServeMux()

	// Endpoint для скачивания CA сертификата
	mux.HandleFunc("/ca.crt", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		caCert := proxyServer.GetCACertificate()
		w.Header().Set("Content-Type", "application/x-x509-ca-cert")
		w.Header().Set("Content-Disposition", "attachment; filename=\"4ebur-net-ca.crt\"")
		_, _ = w.Write(caCert)

		log.Printf("📥 CA certificate downloaded from %s", r.RemoteAddr)
	})

	// Endpoint для статистики кеша
	mux.HandleFunc("/stats", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		hits, misses, size, entries, hitRate := proxyServer.GetCacheStats()
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(fmt.Sprintf(
			`{"cache_hits":%d,"cache_misses":%d,"cache_size_bytes":%d,"cache_entries":%d,"hit_rate":%.2f}`,
			hits, misses, size, entries, hitRate,
		)))
	})

	// Health check endpoint
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"status":"ok","service":"4ebur-net"}`))
	})

	// Все остальные запросы идут в прокси
	mux.HandleFunc("/", proxyServer.ServeHTTP)

	// Настраиваем HTTP сервер с оптимальными параметрами
	server := &http.Server{
		Addr:              ":" + port,
		Handler:           mux,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       120 * time.Second,
		ReadHeaderTimeout: 10 * time.Second,
		MaxHeaderBytes:    1 << 20, // 1MB
	}

	log.Println("╔═══════════════════════════════════════════════════════════╗")
	log.Println("║         4ebur-net MITM Proxy Server Started              ║")
	log.Println("╚═══════════════════════════════════════════════════════════╝")
	log.Printf("🚀 Listening on port: %s", port)
	log.Printf("🔧 Configure proxy: localhost:%s", port)
	log.Printf("📥 Download CA certificate: http://localhost:%s/ca.crt", port)
	log.Printf("📊 Cache stats: http://localhost:%s/stats", port)
	log.Printf("💚 Health check: http://localhost:%s/health", port)
	log.Println("⚠️  Remember to install CA certificate in your trust store!")
	log.Println("───────────────────────────────────────────────────────────")

	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
