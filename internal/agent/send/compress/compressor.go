// Package compress provides functionality for compressing data using the gzip algorithm.
// It encapsulates a gzip.Writer along with a buffer and mutex to safely compress data concurrently.
package compress

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"sync"
)

// Compressor provides methods for compressing data using gzip.
// It maintains an internal buffer and a gzip.Writer to perform compression.
// A mutex is used to ensure thread-safe operations.
type Compressor struct {
	mu     *sync.Mutex   // mu protects the buffer and writer during compression.
	buf    *bytes.Buffer // buf holds the compressed data.
	writer *gzip.Writer  // writer is used to compress data using gzip.
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

// Compress compresses the provided data using gzip and returns the compressed bytes.
// It resets the internal buffer and writer after the compression is complete.
//
// Parameters:
//   - data: The byte slice containing the data to be compressed.
//
// Returns:
//   - []byte: The compressed data.
//   - error: An error if compression fails; otherwise, nil.
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
