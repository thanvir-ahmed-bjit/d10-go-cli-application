package logger

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"sync"
	"testing"
	"time"

	"d10-go-cli-application/internal/domain"
)

func TestNewAsyncLoggerValidation(t *testing.T) {
	tests := []struct {
		name       string
		output     io.Writer
		bufferSize int
		wantErr    bool
	}{
		{
			name:       "valid logger",
			output:     &bytes.Buffer{},
			bufferSize: 10,
			wantErr:    false,
		},
		{
			name:       "nil output",
			output:     nil,
			bufferSize: 10,
			wantErr:    true,
		},
		{
			name:       "negative buffer size",
			output:     &bytes.Buffer{},
			bufferSize: -1,
			wantErr:    true,
		},
		{
			name:       "zero buffer size",
			output:     &bytes.Buffer{},
			bufferSize: 0,
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewAsyncLogger(tt.output, tt.bufferSize)
			if (err != nil) != tt.wantErr {
				t.Fatalf("got error %v, want error %v", err != nil, tt.wantErr)
			}
		})
	}
}

func TestAsyncLoggerMessageDelivery(t *testing.T) {
	output := &bytes.Buffer{}
	logger, err := NewAsyncLogger(output, 10)
	if err != nil {
		t.Fatal(err)
	}
	defer logger.Close()

	ctx := context.Background()
	if err := logger.Log(ctx, "test message"); err != nil {
		t.Fatal(err)
	}

	logger.Close()

	result := output.String()
	if !strings.Contains(result, "test message") {
		t.Fatalf("expected 'test message' in output, got %q", result)
	}
}

func TestAsyncLoggerMultipleMessages(t *testing.T) {
	output := &bytes.Buffer{}
	logger, err := NewAsyncLogger(output, 10)
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()
	for i := 1; i <= 5; i++ {
		if err := logger.Log(ctx, fmt.Sprintf("message %d", i)); err != nil {
			t.Fatal(err)
		}
	}

	logger.Close()

	result := output.String()
	for i := 1; i <= 5; i++ {
		if !strings.Contains(result, fmt.Sprintf("message %d", i)) {
			t.Fatalf("expected 'message %d' in output", i)
		}
	}
}

func TestAsyncLoggerContextCancellation(t *testing.T) {
	output := &bytes.Buffer{}
	logger, err := NewAsyncLogger(output, 1)
	if err != nil {
		t.Fatal(err)
	}
	defer logger.Close()

	// Fill the buffer
	ctx := context.Background()
	if err := logger.Log(ctx, "message 1"); err != nil {
		t.Fatal(err)
	}

	// Create a cancelled context
	cancelledCtx, cancel := context.WithCancel(context.Background())
	cancel()

	// Try to send with cancelled context - should return error
	err = logger.Log(cancelledCtx, "message 2")
	if err == nil {
		t.Fatal("expected error from cancelled context")
	}
	if err != context.Canceled {
		t.Fatalf("expected context.Canceled, got %v", err)
	}
}

func TestAsyncLoggerShutdownDrains(t *testing.T) {
	output := &bytes.Buffer{}
	logger, err := NewAsyncLogger(output, 10)
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()
	messages := []string{"msg1", "msg2", "msg3"}
	for _, msg := range messages {
		if err := logger.Log(ctx, msg); err != nil {
			t.Fatal(err)
		}
	}

	logger.Close()

	result := output.String()
	for _, msg := range messages {
		if !strings.Contains(result, msg) {
			t.Fatalf("expected %q in output", msg)
		}
	}
}

func TestAsyncLoggerRepeatedClose(t *testing.T) {
	output := &bytes.Buffer{}
	logger, err := NewAsyncLogger(output, 10)
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()
	if err := logger.Log(ctx, "message"); err != nil {
		t.Fatal(err)
	}

	if err := logger.Close(); err != nil {
		t.Fatal(err)
	}

	// Second close should not panic
	if err := logger.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestAsyncLoggerLoggingAfterClose(t *testing.T) {
	output := &bytes.Buffer{}
	logger, err := NewAsyncLogger(output, 10)
	if err != nil {
		t.Fatal(err)
	}

	logger.Close()

	ctx := context.Background()
	err = logger.Log(ctx, "message after close")
	if !errors.Is(err, domain.ErrLoggerClosed) {
		t.Fatalf("expected ErrLoggerClosed, got %v", err)
	}
}

func TestAsyncLoggerConcurrentLogging(t *testing.T) {
	output := &bytes.Buffer{}
	logger, err := NewAsyncLogger(output, 100)
	if err != nil {
		t.Fatal(err)
	}
	defer logger.Close()

	var wg sync.WaitGroup
	ctx := context.Background()

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < 10; j++ {
				_ = logger.Log(ctx, fmt.Sprintf("goroutine %d message %d", id, j))
			}
		}(i)
	}

	wg.Wait()
	logger.Close()

	result := output.String()
	lines := strings.Split(result, "\n")
	// Should have 100 log lines (plus extra empty lines from split)
	if len(lines) < 100 {
		t.Fatalf("expected at least 100 log lines, got %d", len(lines))
	}
}

func TestAsyncLoggerConcurrentLoggingAndShutdown(t *testing.T) {
	output := &bytes.Buffer{}
	logger, err := NewAsyncLogger(output, 50)
	if err != nil {
		t.Fatal(err)
	}

	var wg sync.WaitGroup
	ctx := context.Background()

	// Start logging goroutines
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < 20; j++ {
				_ = logger.Log(ctx, fmt.Sprintf("goroutine %d message %d", id, j))
			}
		}(i)
	}

	// Wait a bit then close
	time.Sleep(10 * time.Millisecond)
	wg.Wait()
	logger.Close()

	result := output.String()
	lines := strings.Split(strings.TrimSpace(result), "\n")
	// All 100 messages should be logged
	if len(lines) != 100 {
		t.Fatalf("expected 100 log lines, got %d", len(lines))
	}
}

func TestAsyncLoggerTimestampFormat(t *testing.T) {
	output := &bytes.Buffer{}
	logger, err := NewAsyncLogger(output, 10)
	if err != nil {
		t.Fatal(err)
	}
	defer logger.Close()

	ctx := context.Background()
	if err := logger.Log(ctx, "test"); err != nil {
		t.Fatal(err)
	}

	logger.Close()

	result := output.String()
	// Check for timestamp pattern like [2006-01-02 15:04:05]
	if !strings.Contains(result, "[") || !strings.Contains(result, "]") {
		t.Fatalf("expected timestamp in output, got %q", result)
	}
}
