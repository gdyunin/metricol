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
	loggers       = make(map[string]*zap.SugaredLogger)
	defaultLogger *zap.SugaredLogger
	mu            sync.Mutex
)

// init initializes the default logger used as a fallback.
func init() {
	zl, err := zap.NewProduction()
	if err != nil {
		panic(fmt.Sprintf("Initialization error: failed to initialize default logger: %v", err))
	}
	defaultLogger = zl.Sugar()
}

// Logger retrieves or creates a logger for the specified log level.
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
		defaultLogger.Warnf("Fallback to default logger due to error creating logger for level '%s': %v", level, err)
		return defaultLogger
	}

	loggers[level] = newLogger
	return newLogger
}

// createLogger creates a new logger configured for the specified log level.
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
