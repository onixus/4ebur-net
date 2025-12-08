package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
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

	// Создаем обработчик с проверкой специальных путей
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Специальные endpoints (только для GET запросов, не CONNECT)
		if r.Method == http.MethodGet {
			switch r.URL.Path {
			case "/ca.crt":
				caCert := proxyServer.GetCACertificate()
				w.Header().Set("Content-Type", "application/x-x509-ca-cert")
				w.Header().Set("Content-Disposition", `attachment; filename="4ebur-net-ca.crt"`)
				_, _ = w.Write(caCert)
				log.Printf("📥 CA certificate downloaded from %s", r.RemoteAddr)
				return

			case "/stats":
				hits, misses, size, entries, hitRate := proxyServer.GetCacheStats()
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(fmt.Sprintf(
					`{"cache_hits":%d,"cache_misses":%d,"cache_size_bytes":%d,"cache_entries":%d,"hit_rate":%.2f}`,
					hits, misses, size, entries, hitRate,
				)))
				return

			case "/health":
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(`{"status":"ok","service":"4ebur-net"}`))
				return

			case "/":
				// Корневой путь - показываем информацию о прокси
				if !strings.Contains(r.Host, ":") || r.Host == "localhost:" + port {
					w.Header().Set("Content-Type", "text/html; charset=utf-8")
					_, _ = w.Write([]byte(fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
	<title>4ebur-net MITM Proxy</title>
	<style>
		body { font-family: monospace; margin: 40px; background: #1a1a1a; color: #00ff00; }
		h1 { color: #00ff00; }
		a { color: #00aaff; }
		pre { background: #0a0a0a; padding: 10px; border: 1px solid #00ff00; }
	</style>
</head>
<body>
	<h1>🚀 4ebur-net MITM Proxy Server</h1>
	<p><strong>Status:</strong> Running on port %s</p>
	
	<h2>📥 Downloads:</h2>
	<ul>
		<li><a href="/ca.crt">Download CA Certificate</a></li>
	</ul>
	
	<h2>📊 Endpoints:</h2>
	<ul>
		<li><a href="/stats">/stats</a> - Cache statistics (JSON)</li>
		<li><a href="/health">/health</a> - Health check (JSON)</li>
	</ul>
	
	<h2>🔧 Configuration:</h2>
	<pre>export HTTP_PROXY=http://localhost:%s
export HTTPS_PROXY=http://localhost:%s

# Or with curl:
curl -x http://localhost:%s https://example.com</pre>
	
	<h2>📖 Installation:</h2>
	<pre># 1. Download CA certificate
curl http://localhost:%s/ca.crt -o 4ebur-net-ca.crt

# 2. Install (Arch Linux)
sudo cp 4ebur-net-ca.crt /etc/ca-certificates/trust-source/anchors/
sudo trust extract-compat

# 3. Verify
trust list | grep -i 4ebur</pre>
</body>
</html>`, port, port, port, port, port)))
					return
				}
			}
		}

		// Все остальные запросы (включая CONNECT) идут в прокси
		proxyServer.ServeHTTP(w, r)
	})

	// Настраиваем HTTP сервер с оптимальными параметрами
	server := &http.Server{
		Addr:              ":" + port,
		Handler:           handler,
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
	log.Printf("🌐 Web interface: http://localhost:%s/", port)
	log.Printf("📥 Download CA certificate: http://localhost:%s/ca.crt", port)
	log.Printf("📊 Cache stats: http://localhost:%s/stats", port)
	log.Printf("💚 Health check: http://localhost:%s/health", port)
	log.Printf("🔧 Configure proxy: localhost:%s", port)
	log.Println("⚠️  Remember to install CA certificate in your trust store!")
	log.Println("───────────────────────────────────────────────────────────")

	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
