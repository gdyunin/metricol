package logger

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

/*
TestLogger tests the Logger function.
It ensures correct logger creation, caching, and error handling for invalid log levels.
*/
func TestLogger(t *testing.T) {
	tests := []struct {
		name          string
		logLevel      string
		expectError   bool
		errorContains string
	}{
		{
			name:        "Valid log level INFO",
			logLevel:    "INFO",
			expectError: false,
		},
		{
			name:        "Valid log level DEBUG",
			logLevel:    "DEBUG",
			expectError: false,
		},
		{
			name:        "Valid log level ERROR",
			logLevel:    "ERROR",
			expectError: false,
		},
		{
			name:          "Invalid log level",
			logLevel:      "INVALID",
			expectError:   true,
			errorContains: "failed to parse log level 'INVALID'",
		},
		{
			name:          "Empty log level",
			logLevel:      "",
			expectError:   true,
			errorContains: "expected non-empty level string but got empty string",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Call the Logger function with the specified log level.
			logger, err := Logger(tc.logLevel)

			if tc.expectError {
				// Validate that an error is returned.
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.errorContains)
				assert.Nil(t, logger)
			} else {
				// Validate that no error is returned.
				require.NoError(t, err)
				require.NotNil(t, logger)

				// Check if the logger is cached.
				cachedLogger, exists := loggers[tc.logLevel]
				assert.True(t, exists, "Logger should be cached after initialization.")
				assert.Equal(t, logger, cachedLogger, "Returned logger should match the cached logger.")
			}
		})
	}
}

/*
TestLogger_Caching tests that the Logger function reuses cached loggers for the same log level.
*/
func TestLogger_Caching(t *testing.T) {
	logLevel := "INFO"

	// Get the logger for the first time.
	firstLogger, err := Logger(logLevel)
	require.NoError(t, err)
	require.NotNil(t, firstLogger)

	// Get the logger for the same log level again.
	secondLogger, err := Logger(logLevel)
	require.NoError(t, err)
	require.NotNil(t, secondLogger)

	// Verify that the same instance is returned.
	assert.Equal(t, firstLogger, secondLogger, "Logger instances for the same level should be identical.")
}
