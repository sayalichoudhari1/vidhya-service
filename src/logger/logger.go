// Package logger provides a minimal structured logger for Vidhya Service.
//
// It intentionally avoids any external logging framework so the service has
// zero dependency risk here; it writes leveled, timestamped lines to stdout
// which is what both local `docker compose logs` and AWS CloudWatch expect.
package logger

import (
	"fmt"
	"log"
	"os"
	"strings"
)

// Level represents a log severity level.
type Level int

// Log levels, ordered by severity.
const (
	LevelDebug Level = iota
	LevelInfo
	LevelWarn
	LevelError
)

// Logger is a small leveled logger writing to stdout.
type Logger struct {
	minLevel Level
	std      *log.Logger
}

// Log is the package-level logger instance used across the service.
var Log = New("INFO")

// New creates a Logger configured with the given minimum level
// (DEBUG, INFO, WARN, ERROR - case-insensitive; defaults to INFO).
func New(levelName string) *Logger {
	return &Logger{
		minLevel: parseLevel(levelName),
		std:      log.New(os.Stdout, "", log.LstdFlags),
	}
}

func parseLevel(name string) Level {
	switch strings.ToUpper(strings.TrimSpace(name)) {
	case "DEBUG":
		return LevelDebug
	case "WARN", "WARNING":
		return LevelWarn
	case "ERROR":
		return LevelError
	default:
		return LevelInfo
	}
}

func (l *Logger) log(level Level, levelName, requestID, format string, args ...interface{}) {
	if level < l.minLevel {
		return
	}
	msg := fmt.Sprintf(format, args...)
	if requestID != "" {
		l.std.Printf("[%s] [req:%s] %s", levelName, requestID, msg)
		return
	}
	l.std.Printf("[%s] %s", levelName, msg)
}

// Debug logs a debug-level message, optionally tagged with a request ID.
func (l *Logger) Debug(requestID, format string, args ...interface{}) {
	l.log(LevelDebug, "DEBUG", requestID, format, args...)
}

// Info logs an info-level message, optionally tagged with a request ID.
func (l *Logger) Info(requestID, format string, args ...interface{}) {
	l.log(LevelInfo, "INFO", requestID, format, args...)
}

// Warn logs a warning-level message, optionally tagged with a request ID.
func (l *Logger) Warn(requestID, format string, args ...interface{}) {
	l.log(LevelWarn, "WARN", requestID, format, args...)
}

// Error logs an error-level message. code is a short machine-readable tag
// used to grep logs; it may be left empty.
func (l *Logger) Error(requestID, code, format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	if code != "" {
		msg = fmt.Sprintf("[%s] %s", code, msg)
	}
	l.log(LevelError, "ERROR", requestID, "%s", msg)
}

// Configure re-initializes the package-level Log with the given level.
func Configure(levelName string) {
	Log = New(levelName)
}
