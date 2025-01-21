package send

import (
	"encoding/hex"
	"fmt"

	"github.com/gdyunin/metricol.git/internal/agent/send/compress"
	"github.com/gdyunin/metricol.git/pkg/sign"
	"github.com/go-resty/resty/v2"
)

// RequestBuilder is responsible for creating HTTP requests, with support for gzip compression.
type RequestBuilder struct {
	httpClient *resty.Client        // HTTP client for sending requests.
	compressor *compress.Compressor // Compressor for gzip encoding.
}

// NewRequestBuilder initializes and returns a new RequestBuilder instance.
//
// Parameters:
//   - httpClient: An instance of resty.Client to send HTTP requests.
//
// Returns:
//   - *RequestBuilder: A pointer to the initialized RequestBuilder.
func NewRequestBuilder(httpClient *resty.Client) *RequestBuilder {
	return &RequestBuilder{
		httpClient: httpClient,
		compressor: compress.NewCompressor(),
	}
}

// Build creates a basic HTTP request with the given method, endpoint, and body.
//
// Parameters:
//   - method: The HTTP method (e.g., GET, POST, etc.).
//   - endpoint: The URL endpoint for the request.
//   - body: The body content for the request.
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

// BuildWithGzip creates an HTTP request with gzip-compressed body.
//
// Parameters:
//   - method: The HTTP method (e.g., POST, PUT, etc.).
//   - endpoint: The URL endpoint for the request.
//   - body: The body content to be compressed and included in the request.
//
// Returns:
//   - *resty.Request: The constructed HTTP request with gzip-compressed body.
//   - error: An error if compression fails.
func (b *RequestBuilder) BuildWithGzip(method string, endpoint string, body []byte, signingKey string) (
	*resty.Request,
	error,
) {
	var s string
	if signingKey != "" {
		s = hex.EncodeToString(sign.MakeSign(body, signingKey))
	}

	body, err := b.compressor.Compress(body)
	if err != nil {
		return nil, fmt.Errorf("gzip compression failed for request body: %w", err)
	}

	req := b.Build(method, endpoint, body)
	req.SetHeader("Content-Encoding", "gzip")
	if s != "" {
		req.SetHeader("HashSHA256", s)
	}

	return req, nil
}
