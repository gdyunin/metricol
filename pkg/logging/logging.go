// Package logging provides functionality for creating and retrieving logger instances
// configured at different log levels using Uber's zap logging library. It manages a set
// of logger instances in a thread-safe manner and offers a fallback logger if logger
// creation fails.
package logging

import (
	"fmt"
	"strings"
	"sync"

	"go.uber.org/zap"
)

const (
	// LevelDEBUG represents the DEBUG log level.
	LevelDEBUG = "DEBUG"
	// LevelINFO represents the INFO log level.
	LevelINFO = "INFO"
	// LevelWARN represents the WARN log level.
	LevelWARN = "WARN"
	// LevelERROR represents the ERROR log level.
	LevelERROR = "ERROR"
	// LevelDPANIC represents the DPANIC log level.
	LevelDPANIC = "DPANIC"
	// LevelPANIC represents the PANIC log level.
	LevelPANIC = "PANIC"
	// LevelFATAL represents the FATAL log level.
	LevelFATAL = "FATAL"
)

var (
	// Var loggers stores created logger instances keyed by their log level.
	loggers = make(map[string]*zap.SugaredLogger)
	// Var defaultLogger is used as a fallback if logger creation fails.
	defaultLogger *zap.SugaredLogger
	// Var mu protects access to the loggers map.
	mu sync.Mutex
)

// init initializes the default logger used as a fallback.
// It attempts to create a production logger; if that fails, it falls back to an example logger.
func init() {
	var zl *zap.Logger
	var err error

	zl, err = zap.NewProduction()
	if err != nil {
		zl = zap.NewExample()
		zl.Error(
			"Error initializing default logger. Falling back to example logger.",
			zap.Error(err),
		)
	}

	defaultLogger = zl.Sugar()
}

// Logger retrieves or creates a logger for the specified log level.
// The log level string is converted to uppercase. If a logger for that level
// already exists, it is returned; otherwise, a new logger is created using createLogger.
// If creation fails, a warning is logged and the default logger is returned.
//
// Parameters:
//   - level: The desired log level (e.g., "INFO", "DEBUG").
//
// Returns:
//   - *zap.SugaredLogger: A logger instance configured for the specified level.
func Logger(level string) *zap.SugaredLogger {
	level = strings.ToUpper(level)

	mu.Lock()
	defer mu.Unlock()

	if logger, exist := loggers[level]; exist {
		return logger
	}

	newLogger, err := createLogger(level)
	if err != nil {
		defaultLogger.Warnf(
			"Fallback to default logger due to error creating logger for level '%s': %v",
			level,
			err,
		)
		return defaultLogger
	}

	loggers[level] = newLogger
	return newLogger
}

// createLogger creates a new logger configured for the specified log level.
// It parses the provided level into an atomic level and applies the production configuration.
// If building the logger fails, an error is returned.
//
// Parameters:
//   - level: The desired log level (e.g., "INFO", "DEBUG").
//
// Returns:
//   - *zap.SugaredLogger: A new logger instance configured for the specified level.
//   - error: An error if the logger could not be created.
func createLogger(level string) (*zap.SugaredLogger, error) {
	atomicLevel, err := zap.ParseAtomicLevel(level)
	if err != nil {
		return nil, fmt.Errorf("invalid log level '%s': failed to parse: %w", level, err)
	}

	cfg := zap.NewProductionConfig()
	cfg.Level = atomicLevel

	zl, err := cfg.Build()
	if err != nil {
		return nil, fmt.Errorf("configuration error: failed to build logger for level '%s': %w", level, err)
	}
	return zl.Sugar(), nil
}
