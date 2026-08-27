package domain

import (
	"errors"
	"fmt"
)

var (
	ErrDuplicateID  = errors.New("duplicate user ID")
	ErrUserNotFound = errors.New("user not found")
	ErrLoggerClosed = errors.New("logger is closed")
)

// ValidationError represents a validation error with field-specific details.
type ValidationError struct {
	Field  string
	Reason string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("invalid %s: %s", e.Field, e.Reason)
}
