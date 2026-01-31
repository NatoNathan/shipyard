package logger

import (
	"fmt"
	"io"
	"os"
	"strings"
	"time"
)

// Global logger instance
var globalLogger *Logger

func init() {
	// Initialize with default logger
	globalLogger = New(os.Stdout, LevelInfo, false)
}

// Level represents the logging level
type Level int

const (
	LevelDebug Level = iota
	LevelInfo
	LevelWarn
	LevelError
)

// String returns the string representation of the level
func (l Level) String() string {
	switch l {
	case LevelDebug:
		return "DEBUG"
	case LevelInfo:
		return "INFO"
	case LevelWarn:
		return "WARN"
	case LevelError:
		return "ERROR"
	default:
		return "UNKNOWN"
	}
}

// ParseLevel parses a string into a Level
func ParseLevel(s string) (Level, error) {
	switch strings.ToLower(s) {
	case "debug":
		return LevelDebug, nil
	case "info":
		return LevelInfo, nil
	case "warn", "warning":
		return LevelWarn, nil
	case "error":
		return LevelError, nil
	default:
		return LevelInfo, fmt.Errorf("invalid log level: %s", s)
	}
}

// Logger provides structured logging with verbosity levels
type Logger struct {
	writer io.Writer
	level  Level
	quiet  bool
}

// New creates a new Logger instance
func New(writer io.Writer, level Level, quiet bool) *Logger {
	return &Logger{
		writer: writer,
		level:  level,
		quiet:  quiet,
	}
}

// Debug logs a debug-level message
func (l *Logger) Debug(format string, args ...interface{}) {
	if l.level <= LevelDebug && !l.quiet {
		l.log(LevelDebug, format, args...)
	}
}

// Info logs an info-level message
func (l *Logger) Info(format string, args ...interface{}) {
	if l.level <= LevelInfo && !l.quiet {
		l.log(LevelInfo, format, args...)
	}
}

// Warn logs a warning-level message
func (l *Logger) Warn(format string, args ...interface{}) {
	if l.level <= LevelWarn && !l.quiet {
		l.log(LevelWarn, format, args...)
	}
}

// Error logs an error-level message (not suppressed by quiet mode)
func (l *Logger) Error(format string, args ...interface{}) {
	if l.level <= LevelError {
		l.log(LevelError, format, args...)
	}
}

// log is the internal logging method
func (l *Logger) log(level Level, format string, args ...interface{}) {
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	message := fmt.Sprintf(format, args...)
	
	fmt.Fprintf(l.writer, "[%s] %s %s\n", level.String(), timestamp, message)
}

// SetLevel changes the logging level
func (l *Logger) SetLevel(level Level) {
	l.level = level
}

// SetQuiet enables or disables quiet mode
func (l *Logger) SetQuiet(quiet bool) {
	l.quiet = quiet
}

// Get returns the global logger instance
func Get() *Logger {
	return globalLogger
}

// SetGlobal sets the global logger instance
func SetGlobal(l *Logger) {
	globalLogger = l
}

// Infof is a convenience method for logging info messages
func (l *Logger) Infof(format string, args ...interface{}) {
	l.Info(format, args...)
}

// Debugf is a convenience method for logging debug messages
func (l *Logger) Debugf(format string, args ...interface{}) {
	l.Debug(format, args...)
}

// Warnf is a convenience method for logging warning messages
func (l *Logger) Warnf(format string, args ...interface{}) {
	l.Warn(format, args...)
}

// Errorf is a convenience method for logging error messages
func (l *Logger) Errorf(format string, args ...interface{}) {
	l.Error(format, args...)
}
