package logger

import (
	"bytes"
	"encoding/json"
	"os"
	"testing"
	"time"
)

func TestNewLogger(t *testing.T) {
	logger := NewLogger()
	if logger == nil {
		t.Fatal("NewLogger returned nil")
	}
}

func TestLogHTTPRequest(t *testing.T) {
	// Capture log output
	var buf bytes.Buffer
	os.Setenv("LOG_FORMAT", "json")
	os.Setenv("LOG_LEVEL", "info")
	defer os.Unsetenv("LOG_FORMAT")
	defer os.Unsetenv("LOG_LEVEL")

	logger := NewLogger()

	// Redirect output to buffer
	// Note: This is simplified - in real tests you'd use zerolog.TestWriter
	logger.LogHTTPRequest(
		"GET",
		"https://example.com/api",
		"192.168.1.1:12345",
		200,
		50*time.Millisecond,
		true,
		"Mozilla/5.0",
	)

	// Test passes if no panic occurs
}

func TestLogCacheOperation(t *testing.T) {
	os.Setenv("LOG_LEVEL", "debug")
	defer os.Unsetenv("LOG_LEVEL")

	logger := NewLogger()
	logger.LogCacheOperation("get", "test-key", true, 1024)

	// Test passes if no panic
}

func TestLoggerWithFields(t *testing.T) {
	logger := NewLogger()

	fields := map[string]interface{}{
		"user_id":    123,
		"session_id": "abc-def",
	}

	loggerWithFields := logger.WithFields(fields)
	if loggerWithFields == nil {
		t.Fatal("WithFields returned nil")
	}

	loggerWithFields.Info("test message")
}

func TestLogLevels(t *testing.T) {
	logger := NewLogger()

	tests := []struct {
		name string
		fn   func()
	}{
		{"Info", func() { logger.Info("info message") }},
		{"Debug", func() { logger.Debug("debug message") }},
		{"Warn", func() { logger.Warn("warn message") }},
		{"Error", func() { logger.Error(nil, "error message") }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Should not panic
			tt.fn()
		})
	}
}

func BenchmarkLogHTTPRequest(b *testing.B) {
	logger := NewLogger()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.LogHTTPRequest(
			"GET",
			"https://example.com/api",
			"192.168.1.1:12345",
			200,
			50*time.Millisecond,
			true,
			"Mozilla/5.0",
		)
	}
}
