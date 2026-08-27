package logger

import (
	"context"
	"fmt"
	"io"
	"sync"
	"time"

	"d10-go-cli-application/internal/domain"
)

// AsyncLogger is a Logger implementation that writes log messages asynchronously
// through a background goroutine.
type AsyncLogger struct {
	messages chan string
	done     chan struct{}
	output   io.Writer

	mu     sync.RWMutex
	closed bool
	once   sync.Once
}

// NewAsyncLogger creates a new AsyncLogger with the specified output and buffer size.
// It returns an error if output is nil or bufferSize is negative.
func NewAsyncLogger(output io.Writer, bufferSize int) (*AsyncLogger, error) {
	if output == nil {
		return nil, fmt.Errorf("output is nil")
	}
	if bufferSize < 0 {
		return nil, fmt.Errorf("buffer size must be non-negative")
	}

	logger := &AsyncLogger{
		messages: make(chan string, bufferSize),
		done:     make(chan struct{}),
		output:   output,
	}

	go logger.run()
	return logger, nil
}

// Log submits a log message to the logger.
// It returns an error if the logger is closed or if the context is cancelled.
func (l *AsyncLogger) Log(ctx context.Context, message string) error {
	l.mu.RLock()
	if l.closed {
		l.mu.RUnlock()
		return domain.ErrLoggerClosed
	}
	l.mu.RUnlock()

	select {
	case l.messages <- message:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// Close gracefully shuts down the logger, allowing any accepted messages to be written.
// It is safe to call Close multiple times.
func (l *AsyncLogger) Close() error {
	l.once.Do(func() {
		l.mu.Lock()
		l.closed = true
		l.mu.Unlock()

		close(l.messages)
		<-l.done
	})
	return nil
}

// run is the background goroutine that processes log messages.
func (l *AsyncLogger) run() {
	defer close(l.done)

	for message := range l.messages {
		timestamp := time.Now().Format("2006-01-02 15:04:05")
		fmt.Fprintf(l.output, "[%s] %s\n", timestamp, message)
	}
}
