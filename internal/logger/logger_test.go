package logger

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/rs/zerolog"
)

func TestNewConsoleOnly(t *testing.T) {
	log := NewConsoleOnly()
	if log.GetLevel() == zerolog.Disabled {
		t.Error("Expected logger to be enabled")
	}
}

func TestNew(t *testing.T) {
	tmp := t.TempDir() + "/test.log"
	log, closer, err := New(tmp)
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}
	defer closer.Close()
	if log.GetLevel() == zerolog.Disabled {
		t.Error("Expected logger to be enabled")
	}
}

func TestNewWithWriter(t *testing.T) {
	buf := &bytes.Buffer{}
	log := NewWithWriter(buf)
	
	log.Info().Msg("test message")
	
	output := buf.String()
	if !strings.Contains(output, "test message") {
		t.Errorf("Expected output to contain 'test message', got: %s", output)
	}
}

func TestWithContext(t *testing.T) {
	log := NewConsoleOnly()
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
	if !strings.Contains(output, "user_id") || !strings.Contains(output, "123") {
		t.Errorf("Expected output to contain user_id field, got: %s", output)
	}
	if !strings.Contains(output, "action") || !strings.Contains(output, "test") {
		t.Errorf("Expected output to contain action field, got: %s", output)
	}
}
