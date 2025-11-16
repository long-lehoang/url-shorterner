// Package log provides structured logging utilities.
package log

import (
	"fmt"
	"log"
	"os"
)

// Logger provides structured logging functionality.
type Logger struct {
	*log.Logger
}

var (
	// Default logger instance
	defaultLogger *Logger
)

func init() {
	defaultLogger = NewLogger(os.Stdout, "", log.LstdFlags)
}

// NewLogger creates a new logger instance.
func NewLogger(output *os.File, prefix string, flags int) *Logger {
	return &Logger{
		Logger: log.New(output, prefix, flags),
	}
}

// Debug logs a debug message.
func Debug(format string, v ...interface{}) {
	defaultLogger.Debug(format, v...)
}

// Info logs an info message.
func Info(format string, v ...interface{}) {
	defaultLogger.Info(format, v...)
}

// Warn logs a warning message.
func Warn(format string, v ...interface{}) {
	defaultLogger.Warn(format, v...)
}

// Error logs an error message.
func Error(format string, v ...interface{}) {
	defaultLogger.Error(format, v...)
}

// Debug logs a debug message.
func (l *Logger) Debug(format string, v ...interface{}) {
	l.Printf("[DEBUG] %s", fmt.Sprintf(format, v...))
}

// Info logs an info message.
func (l *Logger) Info(format string, v ...interface{}) {
	l.Printf("[INFO] %s", fmt.Sprintf(format, v...))
}

// Warn logs a warning message.
func (l *Logger) Warn(format string, v ...interface{}) {
	l.Printf("[WARN] %s", fmt.Sprintf(format, v...))
}

// Error logs an error message.
func (l *Logger) Error(format string, v ...interface{}) {
	l.Printf("[ERROR] %s", fmt.Sprintf(format, v...))
}
