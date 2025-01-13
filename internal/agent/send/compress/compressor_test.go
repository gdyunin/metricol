package compress

import (
	"bytes"
	"compress/gzip"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewCompressor(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "Initialize new Compressor"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			compressor := NewCompressor()

			assert.NotNil(t, compressor)
			assert.NotNil(t, compressor.buf)
			assert.NotNil(t, compressor.writer)
		})
	}
}

func TestCompressor_Compress(t *testing.T) {
	tests := []struct {
		name        string
		input       []byte
		expectError bool
	}{
		{
			name:        "Compress empty data",
			input:       []byte{},
			expectError: false,
		},
		{
			name:        "Compress small data",
			input:       []byte("test"),
			expectError: false,
		},
		{
			name:        "Compress large data",
			input:       bytes.Repeat([]byte("a"), 10000),
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			compressor := NewCompressor()
			compressedData, err := compressor.Compress(tt.input)

			if tt.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, compressedData)
				assert.True(t, len(compressedData) > 0)

				reader, err := gzip.NewReader(bytes.NewReader(compressedData))
				require.NoError(t, err)

				decompressedData, err := io.ReadAll(reader)
				require.NoError(t, err)
				assert.Equal(t, tt.input, decompressedData)

				assert.NoError(t, reader.Close())
			}
		})
	}
}
