package main

import (
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
		port = "8080"
	}

	// Настраиваем HTTP сервер с оптимальными параметрами
	server := &http.Server{
		Addr:              ":" + port,
		Handler:           proxyServer,
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
	log.Println("⚠️  Remember to install CA certificate in your trust store!")
	log.Println("───────────────────────────────────────────────────────────")

	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
