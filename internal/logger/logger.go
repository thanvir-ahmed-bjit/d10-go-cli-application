package logger

import "context"

// Logger defines the interface for logging operations.
type Logger interface {
	Log(ctx context.Context, message string) error
	Flush() error
	Close() error
}
