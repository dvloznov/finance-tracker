package logger

import (
	"context"
	"io"
	"os"
	"time"

	"github.com/rs/zerolog"
)

// ContextKey is the type for context keys used by the logger
type ContextKey string

const (
	// LoggerKey is the context key for the logger instance
	LoggerKey ContextKey = "logger"
)

// New creates a new structured logger with default configuration
func New() zerolog.Logger {
	output := zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: time.RFC3339,
	}
	return zerolog.New(output).With().Timestamp().Caller().Logger()
}

// NewWithWriter creates a new structured logger with a custom writer
func NewWithWriter(w io.Writer) zerolog.Logger {
	return zerolog.New(w).With().Timestamp().Caller().Logger()
}

// WithContext adds the logger to the context
func WithContext(ctx context.Context, logger zerolog.Logger) context.Context {
	return context.WithValue(ctx, LoggerKey, logger)
}

// FromContext retrieves the logger from the context or returns a default logger
func FromContext(ctx context.Context) zerolog.Logger {
	if logger, ok := ctx.Value(LoggerKey).(zerolog.Logger); ok {
		return logger
	}
	return New()
}

// WithFields adds structured fields to a logger
func WithFields(logger zerolog.Logger, fields map[string]interface{}) zerolog.Logger {
	ctx := logger.With()
	for k, v := range fields {
		ctx = ctx.Interface(k, v)
	}
	return ctx.Logger()
}
