package config

import (
	"flag"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseConfig(t *testing.T) {
	tests := []struct {
		envVars     map[string]string
		name        string
		args        []string
		expected    Config
		expectError bool
	}{
		{
			name:    "Default values",
			envVars: map[string]string{},
			args:    []string{},
			expected: Config{
				ServerAddress:   defaultServerAddress,
				FileStoragePath: defaultFileStoragePath,
				DatabaseDSN:     defaultDatabaseDSN,
				SigningKey:      defaultSigningKey,
				StoreInterval:   defaultStoreInterval,
				Restore:         defaultRestoreFlag,
			},
			expectError: false,
		},
		{
			name: "Environment variables",
			envVars: map[string]string{
				"ADDRESS":           "envserver:9000",
				"FILE_STORAGE_PATH": "envfilestoragepath",
				"DATABASE_DSN":      "envdatabasedsn",
				"KEY":               "envkey",
				"STORE_INTERVAL":    "300",
				"RESTORE":           "true",
			},
			args: []string{},
			expected: Config{
				ServerAddress:   "envserver:9000",
				FileStoragePath: "envfilestoragepath",
				DatabaseDSN:     "envdatabasedsn",
				SigningKey:      "envkey",
				StoreInterval:   300,
				Restore:         true,
			},
			expectError: false,
		},
		{
			name:    "Command-line flags",
			envVars: map[string]string{},
			args: []string{
				"-a", "flagserver:8000",
				"-f", "flagfilestoragepath",
				"-d", "flagdatabasedsn",
				"-k", "flagkey",
				"-i", "500",
				"-r", "true",
			},
			expected: Config{
				ServerAddress:   "flagserver:8000",
				FileStoragePath: "flagfilestoragepath",
				DatabaseDSN:     "flagdatabasedsn",
				SigningKey:      "flagkey",
				StoreInterval:   500,
				Restore:         true,
			},
			expectError: false,
		},
		{
			name: "Flags override by environment",
			envVars: map[string]string{
				"ADDRESS":           "envserver:9000",
				"FILE_STORAGE_PATH": "envfilestoragepath",
				"DATABASE_DSN":      "envdatabasedsn",
				"KEY":               "envkey",
				"STORE_INTERVAL":    "300",
				"RESTORE":           "true",
			},
			args: []string{
				"-a", "flagserver:8000",
				"-f", "flagfilestoragepath",
				"-d", "flagdatabasedsn",
				"-k", "flagkey",
				"-i", "500",
				"-r", "true",
			},
			expected: Config{
				ServerAddress:   "envserver:9000",
				FileStoragePath: "envfilestoragepath",
				DatabaseDSN:     "envdatabasedsn",
				SigningKey:      "envkey",
				StoreInterval:   300,
				Restore:         true,
			},
			expectError: false,
		},
		{
			name: "Invalid environment variable",
			envVars: map[string]string{
				"STORE_INTERVAL": "invalid",
			},
			args:        []string{},
			expected:    Config{},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variables
			for key, value := range tt.envVars {
				require.NoError(t, os.Setenv(key, value))
				t.Cleanup(func() {
					require.NoError(t, os.Unsetenv(key)) // Ensure to unset after test
				})
			}

			// Reset command-line flags
			flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError) //nolint:reassign // for tests
			if len(tt.args) > 0 {
				os.Args = append([]string{"cmd"}, tt.args...) //nolint:reassign // for tests
			} else {
				os.Args = []string{"cmd"} //nolint:reassign // for tests
			}

			cfg, err := ParseConfig()
			if tt.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expected, *cfg)
			}
		})
	}
}
