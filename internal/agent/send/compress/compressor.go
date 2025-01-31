package compress

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"sync"
)

// Compressor provides methods for compressing data using gzip.
type Compressor struct {
	mu     *sync.Mutex
	buf    *bytes.Buffer // Buffer to hold compressed data.
	writer *gzip.Writer  // Gzip writer for compression.
}

// NewCompressor initializes and returns a new Compressor instance.
//
// Returns:
//   - *Compressor: A pointer to the newly created Compressor instance.
func NewCompressor() *Compressor {
	buf := &bytes.Buffer{}
	writer := gzip.NewWriter(buf)

	return &Compressor{
		mu:     &sync.Mutex{},
		buf:    buf,
		writer: writer,
	}
}

// Compress compresses the given data using gzip and returns the compressed bytes.
//
// Parameters:
//   - data: The data to be compressed.
//
// Returns:
//   - []byte: The compressed data.
//   - error: An error if compression fails.
func (c *Compressor) Compress(data []byte) ([]byte, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	defer c.buf.Reset()
	defer c.writer.Reset(c.buf)

	if _, err := c.writer.Write(data); err != nil {
		return nil, fmt.Errorf("compression error: unable to write data to gzip writer: %w", err)
	}

	if err := c.writer.Close(); err != nil {
		return nil, fmt.Errorf("compression error: unable to close gzip writer: %w", err)
	}

	return c.buf.Bytes(), nil
}
