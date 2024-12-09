package metrics

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCounter_SetFetcher(t *testing.T) {
	tests := []struct {
		name          string
		initialValue  int64
		fetcher       func() int64
		expectedValue int64
		expectError   bool
	}{
		{
			name:         "Success set simple fetcher and update",
			initialValue: 0,
			fetcher: func() int64 {
				return 42
			},
			expectedValue: 42,
			expectError:   false,
		},
		{
			name:         "Success set fetcher that return negative number and update",
			initialValue: 10,
			fetcher: func() int64 {
				return -7
			},
			expectedValue: -7,
			expectError:   false,
		},
		{
			name:          "Try set empty fetcher and update",
			initialValue:  25,
			fetcher:       nil,
			expectedValue: 25,
			expectError:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Counter{
				Name:  "test_counter",
				Value: tt.initialValue,
			}

			if tt.fetcher != nil {
				c.SetFetcher(tt.fetcher)
			}

			err := c.Update()
			if err != nil {
				require.True(t, tt.expectError)
				require.EqualError(t, fmt.Errorf("error updating metric %s: fetcher not set", c.Name), err.Error())
				return
			}
			require.Equal(t, tt.expectedValue, c.Value)
			require.False(t, tt.expectError)
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
			name:         "Success set simple fetcher and update",
			initialValue: 0,
			fetcher: func() int64 {
				return 42
			},
			expectedValue: 42,
			expectError:   false,
		},
		{
			name:         "Success set fetcher that return negative number and update",
			initialValue: 10,
			fetcher: func() int64 {
				return -7
			},
			expectedValue: -7,
			expectError:   false,
		},
		{
			name:          "Try set empty fetcher and update",
			initialValue:  25,
			fetcher:       nil,
			expectedValue: 25,
			expectError:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Counter{
				Name:  "test_counter",
				Value: tt.initialValue,
			}

			if tt.fetcher != nil {
				returnedG := c.SetFetcherAndReturn(tt.fetcher)
				require.Equal(t, c, returnedG)
			}

			err := c.Update()
			if err != nil {
				require.True(t, tt.expectError)
				require.EqualError(t, fmt.Errorf("error updating metric %s: fetcher not set", c.Name), err.Error())
				return
			}
			require.Equal(t, tt.expectedValue, c.Value)
			require.False(t, tt.expectError)
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
			name: "Test with integer",
			counter: &Counter{
				Value: 42,
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

func TestCounter_Update(t *testing.T) {
	tests := []struct {
		name          string
		initialValue  int64
		fetcher       func() int64
		expectedValue int64
		expectError   bool
	}{
		{
			name:          "Success with fetcher that return 10",
			initialValue:  0,
			fetcher:       func() int64 { return 10 },
			expectedValue: 10,
			expectError:   false,
		},
		{
			name:          "Success with fetcher that return 20",
			initialValue:  5,
			fetcher:       func() int64 { return 20 },
			expectedValue: 20,
			expectError:   false,
		},
		{
			name:          "No set fetcher",
			initialValue:  15,
			fetcher:       nil,
			expectedValue: 15,
			expectError:   true,
		},
		{
			name:          "Success with fetcher that return negative number",
			initialValue:  30,
			fetcher:       func() int64 { return -5 },
			expectedValue: -5,
			expectError:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Counter{
				Name:  "test_counter",
				Value: tt.initialValue,
			}

			if tt.fetcher != nil {
				c.SetFetcher(tt.fetcher)
			}

			err := c.Update()
			if err != nil {
				require.True(t, tt.expectError)
				require.EqualError(t, fmt.Errorf("error updating metric %s: fetcher not set", c.Name), err.Error())
				return
			}
			require.Equal(t, tt.expectedValue, c.Value)
			require.False(t, tt.expectError)
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
			name:       "Create counter with integer",
			inputName:  "test_counter_1",
			inputValue: "42",
			expectErr:  false,
		},
		{
			name:       "Try create invalid counter",
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
			switch c := counter.(type) {
			case *Counter:
				require.Equal(t, tt.inputName, c.Name)
				require.Equal(t, tt.inputValue, c.StringValue())
			default:
				require.Fail(t, "Metric isn`t counter!")
			}
		})
	}
}
