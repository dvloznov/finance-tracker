package logger

import (
	"bytes"
	"context"
	"testing"

	"github.com/rs/zerolog"
)

func TestNew(t *testing.T) {
	log := New()
	if log.GetLevel() == zerolog.Disabled {
		t.Error("Expected logger to be enabled")
	}
}

func TestNewWithWriter(t *testing.T) {
	buf := &bytes.Buffer{}
	log := NewWithWriter(buf)
	
	log.Info().Msg("test message")
	
	output := buf.String()
	if output == "" {
		t.Error("Expected log output, got empty string")
	}
	if !containsString(output, "test message") {
		t.Errorf("Expected output to contain 'test message', got: %s", output)
	}
}

func TestWithContext(t *testing.T) {
	log := New()
	ctx := context.Background()
	
	ctxWithLogger := WithContext(ctx, log)
	
	if ctxWithLogger.Value(LoggerKey) == nil {
		t.Error("Expected logger in context, got nil")
	}
}

func TestFromContext(t *testing.T) {
	buf := &bytes.Buffer{}
	testLog := NewWithWriter(buf)
	ctx := WithContext(context.Background(), testLog)
	
	retrievedLog := FromContext(ctx)
	retrievedLog.Info().Msg("test")
	
	if buf.Len() == 0 {
		t.Error("Expected log output from retrieved logger")
	}
}

func TestFromContext_DefaultLogger(t *testing.T) {
	ctx := context.Background()
	
	// Should return a default logger when none is in context
	log := FromContext(ctx)
	
	if log.GetLevel() == zerolog.Disabled {
		t.Error("Expected default logger to be enabled")
	}
}

func TestWithFields(t *testing.T) {
	buf := &bytes.Buffer{}
	log := NewWithWriter(buf)
	
	fields := map[string]interface{}{
		"user_id": "123",
		"action":  "test",
	}
	
	logWithFields := WithFields(log, fields)
	logWithFields.Info().Msg("test message")
	
	output := buf.String()
	if !containsString(output, "user_id") || !containsString(output, "123") {
		t.Errorf("Expected output to contain user_id field, got: %s", output)
	}
	if !containsString(output, "action") || !containsString(output, "test") {
		t.Errorf("Expected output to contain action field, got: %s", output)
	}
}

// Helper function to check if a string contains a substring
func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 || 
		(len(s) > 0 && (s[:len(substr)] == substr || containsString(s[1:], substr))))
}
