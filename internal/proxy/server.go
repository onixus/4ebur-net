package proxy

import (
	"bufio"
	"crypto/tls"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/onixus/4ebur-net/internal/cert"
	"github.com/onixus/4ebur-net/pkg/pool"
)

// Server представляет MITM прокси-сервер с оптимизацией для высоких нагрузок
type Server struct {
	certManager *cert.Manager
	transport   *http.Transport
	bufferPool  *pool.BufferPool
}

// NewProxyServer создает новый прокси-сервер с оптимальными настройками
func NewProxyServer() (*Server, error) {
	certManager, err := cert.NewManager()
	if err != nil {
		return nil, fmt.Errorf("failed to create cert manager: %w", err)
	}

	// Читаем настройки из переменных окружения
	maxIdleConns := getEnvInt("MAX_IDLE_CONNS", 1000)
	maxIdleConnsPerHost := getEnvInt("MAX_IDLE_CONNS_PER_HOST", 100)
	maxConnsPerHost := getEnvInt("MAX_CONNS_PER_HOST", 100)

	log.Printf("🔧 Transport config: MaxIdleConns=%d, MaxIdleConnsPerHost=%d, MaxConnsPerHost=%d",
		maxIdleConns, maxIdleConnsPerHost, maxConnsPerHost)

	// Оптимизированный транспорт для высоких нагрузок
	transport := &http.Transport{
		MaxIdleConns:        maxIdleConns,
		MaxIdleConnsPerHost: maxIdleConnsPerHost,
		MaxConnsPerHost:     maxConnsPerHost,
		IdleConnTimeout:     90 * time.Second,
		DisableCompression:  true, // Отключаем для максимальной производительности
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		ForceAttemptHTTP2:     true,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}

	return &Server{
		certManager: certManager,
		transport:   transport,
		bufferPool:  pool.NewBufferPool(32 * 1024), // 32KB буферы для оптимального I/O
	}, nil
}

// ServeHTTP обрабатывает входящие HTTP/HTTPS запросы
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodConnect {
		s.handleConnect(w, r)
	} else {
		s.handleHTTP(w, r)
	}
}

// handleHTTP обрабатывает обычные HTTP запросы
func (s *Server) handleHTTP(w http.ResponseWriter, r *http.Request) {
	// Удаляем hop-by-hop заголовки
	removeHopHeaders(r.Header)

	// Копируем запрос для пересылки
	outReq := r.Clone(r.Context())
	if outReq.URL.Scheme == "" {
		outReq.URL.Scheme = "http"
	}
	if outReq.URL.Host == "" {
		outReq.URL.Host = r.Host
	}

	// Логирование запроса
	log.Printf("→ HTTP %s %s", r.Method, outReq.URL)

	// Отправляем запрос к целевому серверу
	resp, err := s.transport.RoundTrip(outReq)
	if err != nil {
		log.Printf("✗ Error forwarding request: %v", err)
		http.Error(w, fmt.Sprintf("Proxy error: %v", err), http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	// Копируем заголовки ответа
	copyHeaders(w.Header(), resp.Header)
	w.WriteHeader(resp.StatusCode)

	// Используем буферный пул для эффективного копирования (zero-copy оптимизация)
	buf := s.bufferPool.Get()
	defer s.bufferPool.Put(buf)

	if _, err := io.CopyBuffer(w, resp.Body, buf); err != nil {
		log.Printf("✗ Error copying response body: %v", err)
	}

	log.Printf("← HTTP %d %s", resp.StatusCode, resp.Status)
}

// handleConnect обрабатывает HTTPS CONNECT запросы с MITM
func (s *Server) handleConnect(w http.ResponseWriter, r *http.Request) {
	host := r.URL.Host
	if host == "" {
		host = r.Host
	}

	log.Printf("→ CONNECT %s", host)

	// Hijack TCP соединение для raw доступа
	hijacker, ok := w.(http.Hijacker)
	if !ok {
		log.Println("✗ Hijacking not supported")
		http.Error(w, "Hijacking not supported", http.StatusInternalServerError)
		return
	}

	clientConn, _, err := hijacker.Hijack()
	if err != nil {
		log.Printf("✗ Hijack error: %v", err)
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}
	defer clientConn.Close()

	// Отправляем клиенту успешный ответ
	if _, err := clientConn.Write([]byte("HTTP/1.1 200 Connection Established\r\n\r\n")); err != nil {
		log.Printf("✗ Failed to write 200 response: %v", err)
		return
	}

	// Извлекаем hostname из host:port
	hostname, _, _ := net.SplitHostPort(host)
	if hostname == "" {
		hostname = host
	}

	// Получаем или генерируем сертификат для хоста
	tlsCert, err := s.certManager.GetCertForHost(hostname)
	if err != nil {
		log.Printf("✗ Failed to get certificate for %s: %v", hostname, err)
		return
	}

	// Устанавливаем TLS соединение с клиентом
	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{*tlsCert},
		MinVersion:   tls.VersionTLS12,
	}
	tlsClientConn := tls.Server(clientConn, tlsConfig)
	defer tlsClientConn.Close()

	// Выполняем TLS handshake
	if err := tlsClientConn.Handshake(); err != nil {
		log.Printf("✗ TLS handshake failed: %v", err)
		return
	}

	// Читаем HTTP запрос от клиента через TLS
	reader := bufio.NewReader(tlsClientConn)
	req, err := http.ReadRequest(reader)
	if err != nil {
		if err != io.EOF {
			log.Printf("✗ Failed to read request: %v", err)
		}
		return
	}

	// Устанавливаем TLS соединение с целевым сервером
	targetConn, err := tls.Dial("tcp", host, &tls.Config{
		ServerName:         hostname,
		InsecureSkipVerify: false,
	})
	if err != nil {
		log.Printf("✗ Failed to connect to %s: %v", host, err)
		return
	}
	defer targetConn.Close()

	// Подготавливаем запрос для отправки
	req.RequestURI = ""
	req.URL.Scheme = "https"
	req.URL.Host = host

	// Точка инспектирования запроса (можно добавить логику анализа)
	log.Printf("🔍 MITM Request: %s %s", req.Method, req.URL)

	// Отправляем запрос к целевому серверу
	if err := req.Write(targetConn); err != nil {
		log.Printf("✗ Failed to write request to target: %v", err)
		return
	}

	// Читаем ответ от целевого сервера
	resp, err := http.ReadResponse(bufio.NewReader(targetConn), req)
	if err != nil {
		log.Printf("✗ Failed to read response: %v", err)
		return
	}
	defer resp.Body.Close()

	// Точка инспектирования ответа
	log.Printf("🔍 MITM Response: %d %s", resp.StatusCode, resp.Status)

	// Отправляем ответ клиенту
	if err := resp.Write(tlsClientConn); err != nil {
		log.Printf("✗ Failed to write response to client: %v", err)
	}
}

// Вспомогательные функции

// removeHopHeaders удаляет hop-by-hop заголовки согласно RFC 2616
func removeHopHeaders(h http.Header) {
	hopHeaders := []string{
		"Connection",
		"Proxy-Connection",
		"Keep-Alive",
		"Proxy-Authenticate",
		"Proxy-Authorization",
		"Te",
		"Trailer",
		"Transfer-Encoding",
		"Upgrade",
	}
	for _, header := range hopHeaders {
		h.Del(header)
	}
}

// copyHeaders копирует HTTP заголовки из источника в назначение
func copyHeaders(dst, src http.Header) {
	for k, vv := range src {
		for _, v := range vv {
			dst.Add(k, v)
		}
	}
}

// getEnvInt читает целочисленное значение из переменной окружения
func getEnvInt(key string, defaultValue int) int {
	if val := os.Getenv(key); val != "" {
		if i, err := strconv.Atoi(val); err == nil {
			return i
		}
	}
	return defaultValue
}
