package send

import (
	"NewNewMetricol/internal/agent/send/compress"
	"fmt"

	"github.com/go-resty/resty/v2"
)

// RequestBuilder provides functionality for building HTTP requests with optional gzip compression.
type RequestBuilder struct {
	httpClient *resty.Client        // HTTP client for sending requests.
	compressor *compress.Compressor // Compressor for handling gzip compression.
}

// NewRequestBuilder initializes and returns a new RequestBuilder instance.
//
// Parameters:
//   - httpClient: A pointer to an instance of resty.Client.
//
// Returns:
//   - *RequestBuilder: A new RequestBuilder instance.
func NewRequestBuilder(httpClient *resty.Client) *RequestBuilder {
	return &RequestBuilder{
		httpClient: httpClient,
		compressor: compress.NewCompressor(),
	}
}

// Build constructs a new HTTP request without compression.
//
// Parameters:
//   - method: The HTTP method (e.g., "GET", "POST").
//   - endpoint: The target URL for the request.
//   - body: The request body as a byte slice.
//
// Returns:
//   - *resty.Request: The constructed HTTP request.
func (b *RequestBuilder) Build(method string, endpoint string, body []byte) *resty.Request {
	req := b.httpClient.R()
	req.Body = body
	req.Method = method
	req.URL = endpoint
	return req
}

// BuildWithGzip constructs a new HTTP request with gzip compression.
//
// Parameters:
//   - method: The HTTP method (e.g., "GET", "POST").
//   - endpoint: The target URL for the request.
//   - body: The request body as a byte slice.
//
// Returns:
//   - *resty.Request: The constructed HTTP request with gzip-compressed body.
//   - error: An error if compression fails.
func (b *RequestBuilder) BuildWithGzip(method string, endpoint string, body []byte) (*resty.Request, error) {
	body, err := b.compressor.Compress(body)
	if err != nil {
		return nil, fmt.Errorf("failed to compress body: %w", err)
	}

	req := b.Build(method, endpoint, body)
	req.SetHeader("Content-Encoding", "gzip")

	return req, nil
}
