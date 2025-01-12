package logging

import (
	"fmt"
	"strings"
	"sync"

	"go.uber.org/zap"
)

const (
	LevelINFO = "INFO"
)

var (
	loggers       = make(map[string]*zap.SugaredLogger)
	defaultLogger *zap.SugaredLogger
	mu            sync.Mutex
)

func init() {
	zl, err := zap.NewProduction()
	if err != nil {
		panic(fmt.Sprintf("Failed init default logger: %v", err))
	}
	defaultLogger = zl.Sugar()
}

func Logger(level string) *zap.SugaredLogger {
	level = strings.ToUpper(level)

	mu.Lock()
	defer mu.Unlock()

	if logger, exist := loggers[level]; exist {
		return logger
	}

	if newLogger, err := createLogger(level); err != nil {
		return defaultLogger
	} else {
		loggers[level] = newLogger
		return newLogger
	}
}

func createLogger(level string) (*zap.SugaredLogger, error) {
	atomicLevel, err := zap.ParseAtomicLevel(level)
	if err != nil {
		return nil, fmt.Errorf("error: failed to parse log level '%s': %w", level, err)
	}

	cfg := zap.NewProductionConfig()
	cfg.Level = atomicLevel

	zl, err := cfg.Build()
	if err != nil {
		return nil, fmt.Errorf("error: failed to build logger with the specified configuration: %w", err)
	}
	return zl.Sugar(), nil
}
