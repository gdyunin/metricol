package compress

import (
	"bytes"
	"compress/gzip"
	"fmt"
)

// Compressor provides functionality for compressing data using gzip.
type Compressor struct {
	buf    *bytes.Buffer
	writer *gzip.Writer
}

// NewCompressor initializes and returns a new Compressor instance.
func NewCompressor() *Compressor {
	buf := &bytes.Buffer{}
	writer := gzip.NewWriter(buf)

	return &Compressor{
		buf:    buf,
		writer: writer,
	}
}

// Compress compresses the given data using gzip and returns the compressed bytes.
// If an error occurs during compression, it is returned.
func (c *Compressor) Compress(data []byte) ([]byte, error) {
	defer c.buf.Reset()         // Reset the buffer to reuse.
	defer c.writer.Reset(c.buf) // Reset the gzip writer to reuse.

	// Write data to the gzip writer.
	if _, err := c.writer.Write(data); err != nil {
		return nil, fmt.Errorf("failed to write data to gzip writer: %w", err)
	}

	// Close the gzip writer to finalize the compression.
	if err := c.writer.Close(); err != nil {
		return nil, fmt.Errorf("failed to close gzip writer: %w", err)
	}

	// Return the compressed data from the buffer.
	return c.buf.Bytes(), nil
}
