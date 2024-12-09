package agent

import (
	"flag"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

// TestParseAgentConfig tests the ParseConfig function.
func TestParseAgentConfig(t *testing.T) {
	tests := []struct {
		name           string
		envVars        map[string]string
		flags          []string
		expectedConfig *Config
		wantErr        bool
	}{
		{
			name: "Default config",
			expectedConfig: &Config{
				ServerAddress:  "localhost:8080",
				PollInterval:   2,
				ReportInterval: 10,
			},
		},
		{
			name: "Override with environment variables",
			envVars: map[string]string{
				"ADDRESS":         "127.0.0.1:9090",
				"POLL_INTERVAL":   "5",
				"REPORT_INTERVAL": "15",
			},
			expectedConfig: &Config{
				ServerAddress:  "127.0.0.1:9090",
				PollInterval:   5,
				ReportInterval: 15,
			},
		},
		{
			name:  "Override with command-line flags",
			flags: []string{"-a", "192.168.1.1:8080", "-p", "3", "-r", "12"},
			expectedConfig: &Config{
				ServerAddress:  "192.168.1.1:8080",
				PollInterval:   3,
				ReportInterval: 12,
			},
		},
		{
			name: "Override with invalid environment variables",
			envVars: map[string]string{
				"ADDRESS":         "127.0.0.1:9090",
				"POLL_INTERVAL":   "5",
				"REPORT_INTERVAL": "invalid",
			},
			expectedConfig: nil,
			wantErr:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError) //nolint // need for test

			// Set environment variables for the test-case
			for key, value := range tt.envVars {
				require.NoError(t, os.Setenv(key, value))
				t.Cleanup(func() {
					require.NoError(t, os.Unsetenv(key)) // Ensure to unset after test
				})
			}

			// Set command-line flags for the test-case
			if len(tt.flags) > 0 {
				os.Args = append([]string{"cmd"}, tt.flags...) //nolint // need for test
			} else {
				os.Args = []string{"cmd"} //nolint // need for test
			}

			cfg, err := ParseConfig()
			if err != nil {
				require.True(t, tt.wantErr)
				return
			}

			require.Equal(t, tt.expectedConfig, cfg)
		})
	}
}
