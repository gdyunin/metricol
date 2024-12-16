package logger

import (
	"fmt"

	"go.uber.org/zap"
)

// SugarLogger is a global instance (singleton) of a SugaredLogger for structured logging.
var SugarLogger *zap.SugaredLogger

// InitializeSugarLogger sets up the SugaredLogger with the specified logging level.
// It parses the provided level string and configures the logger accordingly.
func InitializeSugarLogger(level string) error {
	lvl, err := zap.ParseAtomicLevel(level)
	if err != nil {
		return err
	}

	// Check if a SugaredLogger already exists and its level matches the specified level.
	if SugarLogger != nil && SugarLogger.Level() == lvl.Level() {
		return fmt.Errorf("a SugarLogger with this level (%s) already exists", lvl)
	}

	cfg := zap.NewProductionConfig()
	cfg.Level = lvl // Set the logging level in the configuration.

	zl, err := cfg.Build()
	if err != nil {
		return err
	}

	sugared := zl.Sugar()
	SugarLogger = sugared

	return nil
}
