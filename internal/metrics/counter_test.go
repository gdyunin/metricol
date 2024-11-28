package metrics

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCounter_Name(t *testing.T) {
	tests := []struct {
		name     string
		counter  *Counter
		expected string
	}{
		{
			name: "simple name test",
			counter: &Counter{
				name: "test_counter",
			},
			expected: "test_counter",
		},
		{
			name: "empty name",
			counter: &Counter{
				name: "",
			},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.expected, tt.counter.Name())
		})
	}
}

func TestCounter_SetFetcher(t *testing.T) {
	tests := []struct {
		name          string
		initialValue  int64
		fetcher       func() int64
		expectedValue int64
		expectError   bool
	}{
		{
			name:         "success set simple fetcher and update",
			initialValue: 0,
			fetcher: func() int64 {
				return 42
			},
			expectedValue: 42,
			expectError:   false,
		},
		{
			name:         "success set fetcher that return negative number and update",
			initialValue: 10,
			fetcher: func() int64 {
				return -7
			},
			expectedValue: 3,
			expectError:   false,
		},
		{
			name:          "try set empty fetcher and update",
			initialValue:  25,
			fetcher:       nil,
			expectedValue: 25,
			expectError:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Counter{
				name:  "test_counter",
				value: tt.initialValue,
			}

			if tt.fetcher != nil {
				c.SetFetcher(tt.fetcher)
			}

			err := c.Update()
			if err != nil {
				require.True(t, tt.expectError)
				require.Equal(t, ErrorFetcherNotSet, err.Error())
			}
			require.Equal(t, tt.expectedValue, c.value)
		})
	}
}

func TestCounter_SetFetcherAndReturn(t *testing.T) {
	tests := []struct {
		name          string
		initialValue  int64
		fetcher       func() int64
		expectedValue int64
		expectError   bool
	}{
		{
			name:         "success set simple fetcher and update",
			initialValue: 0,
			fetcher: func() int64 {
				return 42
			},
			expectedValue: 42,
			expectError:   false,
		},
		{
			name:         "success set fetcher that return negative number and update",
			initialValue: 10,
			fetcher: func() int64 {
				return -7
			},
			expectedValue: 3,
			expectError:   false,
		},
		{
			name:          "try set empty fetcher and update",
			initialValue:  25,
			fetcher:       nil,
			expectedValue: 25,
			expectError:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Counter{
				name:  "test_counter",
				value: tt.initialValue,
			}

			if tt.fetcher != nil {
				returnedG := c.SetFetcherAndReturn(tt.fetcher)
				require.Equal(t, c, returnedG)
			}

			err := c.Update()
			if err != nil {
				require.True(t, tt.expectError)
				require.Equal(t, ErrorFetcherNotSet, err.Error())
			}
			require.Equal(t, tt.expectedValue, c.value)
		})
	}
}

func TestCounter_StringValue(t *testing.T) {
	tests := []struct {
		name     string
		counter  *Counter
		expected string
	}{
		{
			name: "test with integer",
			counter: &Counter{
				value: 42,
			},
			expected: "42",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.expected, tt.counter.StringValue())
		})
	}
}

func TestCounter_Type(t *testing.T) {
	tests := []struct {
		name     string
		counter  *Counter
		expected string
	}{
		{
			name:     "simple test return type",
			counter:  &Counter{},
			expected: MetricTypeCounter,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.expected, tt.counter.Type())
		})
	}
}

func TestCounter_Update(t *testing.T) {
	tests := []struct {
		name          string
		initialValue  int64
		fetcher       func() int64
		expectedValue int64
		expectError   bool
	}{
		{
			name:          "success with fetcher that return 10",
			initialValue:  0,
			fetcher:       func() int64 { return 10 },
			expectedValue: 10,
			expectError:   false,
		},
		{
			name:          "success with fetcher that return 20",
			initialValue:  5,
			fetcher:       func() int64 { return 20 },
			expectedValue: 25,
			expectError:   false,
		},
		{
			name:          "no set fetcher",
			initialValue:  15,
			fetcher:       nil,
			expectedValue: 15,
			expectError:   true,
		},
		{
			name:          "success with fetcher that return negative number",
			initialValue:  30,
			fetcher:       func() int64 { return -5 },
			expectedValue: 25,
			expectError:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Counter{
				name:  "test_counter",
				value: tt.initialValue,
			}

			if tt.fetcher != nil {
				c.SetFetcher(tt.fetcher)
			}

			err := c.Update()
			if err != nil {
				require.True(t, tt.expectError)
				require.Equal(t, ErrorFetcherNotSet, err.Error())
			}
			require.Equal(t, tt.expectedValue, c.value)
		})
	}
}

func TestCounter_Value(t *testing.T) {
	tests := []struct {
		name     string
		counter  *Counter
		expected int64
	}{
		{
			name: "test with integer",
			counter: &Counter{
				value: 42,
			},
			expected: 42,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.expected, tt.counter.Value())
		})
	}
}

func TestNewCounterFromStrings(t *testing.T) {
	tests := []struct {
		name       string
		inputName  string
		inputValue string
		expectErr  bool
	}{
		{
			name:       "create counter with integer",
			inputName:  "test_counter_1",
			inputValue: "42",
			expectErr:  false,
		},
		{
			name:       "try create invalid counter",
			inputName:  "test_counter_invalid",
			inputValue: "invalid",
			expectErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			counter, err := newCounterFromStrings(tt.inputName, tt.inputValue)

			if err != nil {
				require.True(t, tt.expectErr)
				return
			}
			require.NotEmpty(t, counter)
			require.Equal(t, tt.inputName, counter.Name())
			require.Equal(t, tt.inputValue, counter.StringValue())
		})
	}
}
