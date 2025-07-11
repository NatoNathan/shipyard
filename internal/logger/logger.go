package logger

import (
	"io"
	"os"
	"path/filepath"

	"github.com/charmbracelet/log"
)

var (
	// Logger is the global logger instance
	Logger *log.Logger
)

// LogLevel represents the available log levels
type LogLevel int

const (
	DebugLevel LogLevel = iota
	InfoLevel
	WarnLevel
	ErrorLevel
	FatalLevel
)

// Config holds the logger configuration
type Config struct {
	Level      LogLevel
	Output     io.Writer
	TimeFormat string
	Prefix     string
	LogFile    string
}

// DefaultConfig returns a default logger configuration
func DefaultConfig() *Config {
	return &Config{
		Level:      InfoLevel,
		Output:     os.Stderr,
		TimeFormat: "15:04:05",
		Prefix:     "shipyard",
		LogFile:    ".shipyard/logs/shipyard.log", // Always log to file by default
	}
}

// Init initializes the global logger with the provided configuration
func Init(config *Config) error {
	Logger = log.NewWithOptions(config.Output, log.Options{
		ReportCaller:    false,
		ReportTimestamp: true,
		TimeFormat:      config.TimeFormat,
		Prefix:          config.Prefix,
	})

	// Set log level
	switch config.Level {
	case DebugLevel:
		Logger.SetLevel(log.DebugLevel)
	case InfoLevel:
		Logger.SetLevel(log.InfoLevel)
	case WarnLevel:
		Logger.SetLevel(log.WarnLevel)
	case ErrorLevel:
		Logger.SetLevel(log.ErrorLevel)
	case FatalLevel:
		Logger.SetLevel(log.FatalLevel)
	}

	// Set output - always write to file
	logFile := config.LogFile
	if logFile == "" {
		logFile = ".shipyard/logs/shipyard.log" // Default log file location
	}

	// Create log directory if it doesn't exist
	logDir := filepath.Dir(logFile)
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return err
	}

	// Open log file
	file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return err
	}

	// Check if we're in an interactive shell by looking for PS1 environment variable
	// If PS1 is set, we're in an interactive shell - only log to file
	// If PS1 is not set (scripts, CI/CD, pipes), also log to stderr
	if os.Getenv("PS1") != "" {
		// Interactive shell - only log to file
		Logger.SetOutput(file)
	} else {
		// Non-interactive (scripts, CI/CD, pipes) - log to both file and stderr
		Logger.SetOutput(io.MultiWriter(config.Output, file))
	}

	return nil
}

// Convenience functions for common log operations
func Debug(msg string, keyvals ...interface{}) {
	if Logger != nil {
		Logger.Debug(msg, keyvals...)
	}
}

func Info(msg string, keyvals ...interface{}) {
	if Logger != nil {
		Logger.Info(msg, keyvals...)
	}
}

func Warn(msg string, keyvals ...interface{}) {
	if Logger != nil {
		Logger.Warn(msg, keyvals...)
	}
}

func Error(msg string, keyvals ...interface{}) {
	if Logger != nil {
		Logger.Error(msg, keyvals...)
	}
}

func Fatal(msg string, keyvals ...interface{}) {
	if Logger != nil {
		Logger.Fatal(msg, keyvals...)
	}
}

// With returns a new logger with the given key-value pairs
func With(keyvals ...interface{}) *log.Logger {
	if Logger != nil {
		return Logger.With(keyvals...)
	}
	return log.NewWithOptions(os.Stderr, log.Options{})
}
