package logger

import (
	"errors"
	"fmt"

	"go.uber.org/zap"
)

const (
	LevelINFO = "INFO" // Log level for informational messages.
)

var loggers = make(map[string]*zap.SugaredLogger) // Map to store cached loggers.

// Logger returns a SugaredLogger instance for the specified log level.
// It initializes the logger if it doesn't already exist, using the zap logging library.
// If an error occurs during logger initialization, it returns an error.
func Logger(level string) (*zap.SugaredLogger, error) {
	if level == "" {
		return nil, errors.New("error: expected non-empty level string but got empty string")
	}

	// If the logger for the given level already exists, return it.
	if loggers[level] != nil {
		return loggers[level], nil
	}

	// Parse the log level string.
	lvl, err := zap.ParseAtomicLevel(level)
	if err != nil {
		return nil, fmt.Errorf("error: failed to parse log level '%s': %w", level, err)
	}

	// Set up the production configuration for the logger.
	cfg := zap.NewProductionConfig()
	cfg.Level = lvl

	// Build the logger with the configuration.
	zl, err := cfg.Build()
	if err != nil {
		return nil, fmt.Errorf("error: failed to build logger with the specified configuration: %w", err)
	}

	// Create a SugaredLogger instance for easier logging.
	sl := zl.Sugar()

	// Ensure the logger is flushed properly when done.
	defer func() { _ = sl.Sync() }()

	// Cache the logger for future reuse.
	loggers[level] = sl
	return sl, nil
}
