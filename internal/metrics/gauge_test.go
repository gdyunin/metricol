package metrics

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGauge_Name(t *testing.T) {
	tests := []struct {
		name     string
		gauge    *Gauge
		expected string
	}{
		{
			name: "simple name test",
			gauge: &Gauge{
				name: "test_gauge",
			},
			expected: "test_gauge",
		},
		{
			name: "empty name",
			gauge: &Gauge{
				name: "",
			},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.expected, tt.gauge.Name())
		})
	}
}

func TestGauge_SetFetcher(t *testing.T) {
	tests := []struct {
		name          string
		initialValue  float64
		fetcher       func() float64
		expectedValue float64
		expectError   bool
	}{
		{
			name:         "success set simple fetcher and update",
			initialValue: 0.0,
			fetcher: func() float64 {
				return 42.0
			},
			expectedValue: 42.0,
			expectError:   false,
		},
		{
			name:         "success set fetcher that return negative number and update",
			initialValue: 10.0,
			fetcher: func() float64 {
				return -7.5
			},
			expectedValue: -7.5,
			expectError:   false,
		},
		{
			name:         "success set fetcher that return big number and update",
			initialValue: 100.0,
			fetcher: func() float64 {
				return 1e9
			},
			expectedValue: 1e9,
			expectError:   false,
		},
		{
			name:          "try set empty fetcher and update",
			initialValue:  25.0,
			fetcher:       nil,
			expectedValue: 25.0,
			expectError:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := &Gauge{
				name:  "test_gauge",
				value: tt.initialValue,
			}

			if tt.fetcher != nil {
				g.SetFetcher(tt.fetcher)
			}

			err := g.Update()
			if err != nil {
				require.True(t, tt.expectError)
				require.Equal(t, ErrorFetcherNotSet, err.Error())
			}
			require.Equal(t, tt.expectedValue, g.value)
		})
	}
}

func TestGauge_SetFetcherAndReturn(t *testing.T) {
	tests := []struct {
		name          string
		initialValue  float64
		fetcher       func() float64
		expectedValue float64
		expectError   bool
	}{
		{
			name:         "success set simple fetcher and update",
			initialValue: 0.0,
			fetcher: func() float64 {
				return 42.0
			},
			expectedValue: 42.0,
			expectError:   false,
		},
		{
			name:         "success set fetcher that return negative number and update",
			initialValue: 10.0,
			fetcher: func() float64 {
				return -7.5
			},
			expectedValue: -7.5,
			expectError:   false,
		},
		{
			name:         "success set fetcher that return big number and update",
			initialValue: 100.0,
			fetcher: func() float64 {
				return 1e9
			},
			expectedValue: 1e9,
			expectError:   false,
		},
		{
			name:          "try set empty fetcher and update",
			initialValue:  25.0,
			fetcher:       nil,
			expectedValue: 25.0,
			expectError:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := &Gauge{
				name:  "test_gauge",
				value: tt.initialValue,
			}

			if tt.fetcher != nil {
				returnedG := g.SetFetcherAndReturn(tt.fetcher)
				require.Equal(t, g, returnedG)
			}

			err := g.Update()
			if err != nil {
				require.True(t, tt.expectError)
				require.Equal(t, ErrorFetcherNotSet, err.Error())
			}
			require.Equal(t, tt.expectedValue, g.value)
		})
	}
}

func TestGauge_StringValue(t *testing.T) {
	tests := []struct {
		name     string
		gauge    *Gauge
		expected string
	}{
		{
			name: "test with integer",
			gauge: &Gauge{
				value: 42.0,
			},
			expected: "42",
		},
		{
			name: "test with float",
			gauge: &Gauge{
				value: 3.1415,
			},
			expected: "3.1415",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.expected, tt.gauge.StringValue())
		})
	}
}

func TestGauge_Type(t *testing.T) {
	tests := []struct {
		name     string
		gauge    *Gauge
		expected string
	}{
		{
			name:     "simple test return type",
			gauge:    &Gauge{},
			expected: MetricTypeGauge,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.expected, tt.gauge.Type())
		})
	}
}

func TestGauge_Update(t *testing.T) {
	tests := []struct {
		name          string
		initialValue  float64
		fetcher       func() float64
		expectedValue float64
		expectError   bool
	}{
		{
			name:          "success with fetcher that return 10.0",
			initialValue:  0.0,
			fetcher:       func() float64 { return 10.0 },
			expectedValue: 10.0,
			expectError:   false,
		},
		{
			name:          "success with fetcher that return 20.5",
			initialValue:  5.0,
			fetcher:       func() float64 { return 20.5 },
			expectedValue: 20.5,
			expectError:   false,
		},
		{
			name:          "no set fetcher",
			initialValue:  15.0,
			fetcher:       nil,
			expectedValue: 15.0,
			expectError:   true,
		},
		{
			name:          "success with fetcher that return negative number",
			initialValue:  30.0,
			fetcher:       func() float64 { return -5.5 },
			expectedValue: -5.5,
			expectError:   false,
		},
		{
			name:          "success with fetcher that return big number",
			initialValue:  100.0,
			fetcher:       func() float64 { return 1e6 },
			expectedValue: 1e6,
			expectError:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := &Gauge{
				name:  "test_gauge",
				value: tt.initialValue,
			}

			if tt.fetcher != nil {
				g.SetFetcher(tt.fetcher)
			}

			err := g.Update()
			if err != nil {
				require.True(t, tt.expectError)
				require.Equal(t, ErrorFetcherNotSet, err.Error())
			}
			require.Equal(t, tt.expectedValue, g.value)
		})
	}
}

func TestGauge_Value(t *testing.T) {
	tests := []struct {
		name     string
		gauge    *Gauge
		expected float64
	}{
		{
			name: "test with integer",
			gauge: &Gauge{
				value: 42.0,
			},
			expected: 42.0,
		},
		{
			name: "test with float",
			gauge: &Gauge{
				value: 3.1415,
			},
			expected: 3.1415,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.expected, tt.gauge.Value())
		})
	}
}

func TestNewGaugeFromStrings(t *testing.T) {
	tests := []struct {
		name       string
		inputName  string
		inputValue string
		expectErr  bool
	}{
		{
			name:       "create gauge with integer",
			inputName:  "test_gauge_1",
			inputValue: "42",
			expectErr:  false,
		},
		{
			name:       "create gauge with float",
			inputName:  "test_gauge_2",
			inputValue: "3.14",
			expectErr:  false,
		},
		{
			name:       "try create invalid gauge",
			inputName:  "test_gauge_invalid",
			inputValue: "invalid",
			expectErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gauge, err := newGaugeFromStrings(tt.inputName, tt.inputValue)

			if err != nil {
				require.True(t, tt.expectErr)
				return
			}
			require.NotEmpty(t, gauge)
			require.Equal(t, tt.inputName, gauge.Name())
			require.Equal(t, tt.inputValue, gauge.StringValue())
		})
	}
}
