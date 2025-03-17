package logging

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLogger(t *testing.T) {
	tests := []struct {
		name      string
		level     string
		expectErr bool
	}{
		{name: "Valid DEBUG level", level: LevelDEBUG},
		{name: "Valid INFO level", level: LevelINFO},
		{name: "Valid WARN level", level: LevelWARN},
		{name: "Valid ERROR level", level: LevelERROR},
		{name: "Valid DPANIC level", level: LevelDPANIC},
		{name: "Valid PANIC level", level: LevelPANIC},
		{name: "Valid FATAL level", level: LevelFATAL},
		{name: "Invalid log level", level: "INVALID", expectErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := Logger(tt.level)
			assert.NotNil(t, logger, "Logger should not be nil for level %s", tt.level)

			if tt.expectErr {
				assert.Equal(t, defaultLogger, logger, "Invalid log level should return default logger")
			} else {
				assert.NotEqual(t, defaultLogger, logger, "Valid log level should return a new logger")
			}
		})
	}

	// Test logger caching
	t.Run("Logger caching", func(t *testing.T) {
		logger1 := Logger(LevelINFO)
		logger2 := Logger(LevelINFO)
		assert.Equal(
			t,
			logger1,
			logger2,
			"Logger should be cached and return the same instance for the same level",
		)
	})
}

func TestCreateLogger(t *testing.T) {
	tests := []struct {
		name      string
		level     string
		expectErr bool
	}{
		{name: "Valid DEBUG level", level: LevelDEBUG},
		{name: "Valid INFO level", level: LevelINFO},
		{name: "Invalid log level", level: "INVALID", expectErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger, err := createLogger(tt.level)

			if tt.expectErr {
				assert.Nil(t, logger, "Logger should be nil for invalid level %s", tt.level)
				assert.Error(t, err, "Error should be returned for invalid level %s", tt.level)
			} else {
				assert.NotNil(t, logger, "Logger should not be nil for valid level %s", tt.level)
				assert.NoError(t, err, "No error should be returned for valid level %s", tt.level)
			}
		})
	}
}
