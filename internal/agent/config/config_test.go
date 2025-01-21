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
		expected    Config
		name        string
		args        []string
		expectError bool
	}{
		{
			name:    "Default values",
			envVars: map[string]string{},
			args:    []string{},
			expected: Config{
				ServerAddress:  defaultServerAddress,
				PollInterval:   defaultPollInterval,
				ReportInterval: defaultReportInterval,
			},
			expectError: false,
		},
		{
			name: "Environment variables",
			envVars: map[string]string{
				"ADDRESS":         "envserver:9000",
				"POLL_INTERVAL":   "5",
				"REPORT_INTERVAL": "15",
			},
			args: []string{},
			expected: Config{
				ServerAddress:  "envserver:9000",
				PollInterval:   5,
				ReportInterval: 15,
			},
			expectError: false,
		},
		{
			name:    "Command-line flags",
			envVars: map[string]string{},
			args: []string{
				"-a", "flagserver:8000",
				"-p", "6",
				"-r", "12",
			},
			expected: Config{
				ServerAddress:  "flagserver:8000",
				PollInterval:   6,
				ReportInterval: 12,
			},
			expectError: false,
		},
		{
			name: "Flags override by environment",
			envVars: map[string]string{
				"ADDRESS":         "envserver:9000",
				"POLL_INTERVAL":   "5",
				"REPORT_INTERVAL": "15",
			},
			args: []string{
				"-a", "flagserver:8000",
				"-p", "6",
				"-r", "12",
			},
			expected: Config{
				ServerAddress:  "envserver:9000",
				PollInterval:   5,
				ReportInterval: 15,
			},
			expectError: false,
		},
		{
			name: "Invalid environment variable",
			envVars: map[string]string{
				"POLL_INTERVAL": "invalid",
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
