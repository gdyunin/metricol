package sign

import (
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMakeSign(t *testing.T) {
	tests := []struct {
		name     string
		key      string
		expected string
		data     []byte
	}{
		{
			name:     "Empty data and key",
			data:     []byte{},
			key:      "",
			expected: "b613679a0814d9ec772f95d778c35fc5ff1697c493715653c6c712144292c5ad",
		},
		{
			name:     "Basic message",
			data:     []byte("hello"),
			key:      "secret",
			expected: "88aab3ede8d3adf94d26ab90d3bafd4a2083070c3bcce9c014ee04a443847c0b",
		},
		{
			name:     "Different key",
			data:     []byte("hello"),
			key:      "anotherkey",
			expected: "9407da994bfd10e1e90c70251b53383e7ea2be6f5d0cb4d4c1a4f6d5d14e6e0b",
		},
		{
			name:     "Longer message",
			data:     []byte("this is a much longer test message"),
			key:      "key",
			expected: "516dcbd5666433eb6e5bd5f1c4469e680a581c8413a26d99a40eead30431f703",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			signature := MakeSign(tt.data, tt.key)
			actual := hex.EncodeToString(signature)
			assert.Equal(
				t,
				tt.expected,
				actual,
				"Test failed for case: %s. Expected: %s, Got: %s, Data: %s, Key: %s",
				tt.name,
				tt.expected,
				actual,
				string(tt.data),
				tt.key,
			)
		})
	}
}
