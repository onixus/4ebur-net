package logger

import (
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// Logger wraps zerolog with proxy-specific functionality
type Logger struct {
	logger zerolog.Logger
}

// NewLogger creates a new production-ready logger
func NewLogger() *Logger {
	// Configure time format
	zerolog.TimeFieldFormat = time.RFC3339Nano

	// Determine output format
	var output = os.Stdout
	if os.Getenv("LOG_FORMAT") == "pretty" {
		output = zerolog.ConsoleWriter{
			Out:        os.Stdout,
			TimeFormat: "15:04:05",
		}
	}

	// Parse log level
	level := zerolog.InfoLevel
	if lvl := os.Getenv("LOG_LEVEL"); lvl != "" {
		parsedLevel, err := zerolog.ParseLevel(lvl)
		if err == nil {
			level = parsedLevel
		}
	}

	// Create logger with base fields
	logger := zerolog.New(output).
		Level(level).
		With().
		Timestamp().
		Str("service", "4ebur-net").
		Str("version", getVersion()).
		Logger()

	return &Logger{logger: logger}
}

// Info logs an info message
func (l *Logger) Info(msg string) {
	l.logger.Info().Msg(msg)
}

// Debug logs a debug message
func (l *Logger) Debug(msg string) {
	l.logger.Debug().Msg(msg)
}

// Warn logs a warning message
func (l *Logger) Warn(msg string) {
	l.logger.Warn().Msg(msg)
}

// Error logs an error
func (l *Logger) Error(err error, msg string) {
	l.logger.Error().Err(err).Msg(msg)
}

// ErrorWithFields logs an error with additional fields
func (l *Logger) ErrorWithFields(err error, msg string, fields map[string]interface{}) {
	event := l.logger.Error().Err(err)

	for k, v := range fields {
		event = event.Interface(k, v)
	}

	event.Msg(msg)
}

// LogHTTPRequest logs an HTTP request with structured fields
func (l *Logger) LogHTTPRequest(method, url, remoteAddr string, statusCode int, latency time.Duration, cacheHit bool, userAgent string) {
	event := l.logger.Info()

	// Adjust log level based on status code
	if statusCode >= 400 && statusCode < 500 {
		event = l.logger.Warn()
	} else if statusCode >= 500 {
		event = l.logger.Error()
	}

	event.
		Str("method", method).
		Str("url", url).
		Str("remote_addr", remoteAddr).
		Int("status", statusCode).
		Dur("latency_ms", latency).
		Bool("cache_hit", cacheHit).
		Str("user_agent", userAgent).
		Msg("http_request")
}

// LogCacheOperation logs cache operations
func (l *Logger) LogCacheOperation(operation, key string, hit bool, sizeBytes int64) {
	l.logger.Debug().
		Str("operation", operation).
		Str("key", key).
		Bool("hit", hit).
		Int64("size_bytes", sizeBytes).
		Msg("cache_operation")
}

// LogCacheStats logs cache statistics
func (l *Logger) LogCacheStats(hits, misses uint64, hitRate float64, sizeBytes int64, entries int) {
	l.logger.Info().
		Uint64("hits", hits).
		Uint64("misses", misses).
		Float64("hit_rate", hitRate).
		Int64("size_bytes", sizeBytes).
		Int("entries", entries).
		Msg("cache_stats")
}

// WithField returns a new logger with an additional field
func (l *Logger) WithField(key string, value interface{}) *Logger {
	return &Logger{
		logger: l.logger.With().Interface(key, value).Logger(),
	}
}

// WithFields returns a new logger with multiple additional fields
func (l *Logger) WithFields(fields map[string]interface{}) *Logger {
	ctx := l.logger.With()
	for k, v := range fields {
		ctx = ctx.Interface(k, v)
	}
	return &Logger{
		logger: ctx.Logger(),
	}
}

// getVersion returns application version from env or default
func getVersion() string {
	if v := os.Getenv("VERSION"); v != "" {
		return v
	}
	return "dev"
}
