// Package utils provides utility functions for common operations
package utils

import (
	"fmt"
)

// Logger represents a logging interface
type Logger interface {
	Debug(args ...interface{})
	Error(args ...interface{})
	Info(args ...interface{})
	Warn(args ...interface{})
}

// DefaultLogger is a simple log writer implementation
type DefaultLogger struct {
	prefix string
}

// NewDefaultLogger creates a new default logger with optional prefix
func NewDefaultLogger(prefix string) *DefaultLogger {
	return &DefaultLogger{prefix: prefix}
}

// Debug logs debug message
func (l *DefaultLogger) Debug(args ...interface{}) {
	if l.prefix != "" {
		fmt.Printf("[%s] [DEBUG] %v\n", l.prefix, args)
	} else {
		fmt.Printf("[DEBUG] %v\n", args)
	}
}

// Error logs error message
func (l *DefaultLogger) Error(args ...interface{}) {
	if l.prefix != "" {
		fmt.Printf("[%s] [ERROR] %v\n", l.prefix, args)
	} else {
		fmt.Printf("[ERROR] %v\n", args)
	}
}

// Info logs info message
func (l *DefaultLogger) Info(args ...interface{}) {
	if l.prefix != "" {
		fmt.Printf("[%s] [INFO] %v\n", l.prefix, args)
	} else {
		fmt.Printf("[INFO] %v\n", args)
	}
}

// Warn logs warning message
func (l *DefaultLogger) Warn(args ...interface{}) {
	if l.prefix != "" {
		fmt.Printf("[%s] [WARN] %v\n", l.prefix, args)
	} else {
		fmt.Printf("[WARN] %v\n", args)
	}
}

// RedactAPIKey redacts API key values in strings for logging
func RedactAPIKey(input string) string {
	if len(input) > 5 {
		return input[:3] + "***" + input[len(input)-2:]
	}
	return "****"
}
