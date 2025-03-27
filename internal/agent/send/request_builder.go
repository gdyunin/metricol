package send

import (
	"encoding/base64"
	"fmt"

	"github.com/gdyunin/metricol.git/internal/agent/send/compress"
	"github.com/gdyunin/metricol.git/pkg/sign"
	"github.com/go-resty/resty/v2"
)

// RequestBuilder is responsible for creating HTTP requests with optional gzip compression.
// It holds an HTTP client and a gzip compressor to prepare requests for sending.
type RequestBuilder struct {
	httpClient *resty.Client        // httpClient is the HTTP client used for sending requests.
	compressor *compress.Compressor // compressor provides gzip compression for request bodies.
}

// NewRequestBuilder initializes and returns a new RequestBuilder instance.
//
// Parameters:
//   - httpClient: An instance of resty.Client used to send HTTP requests.
//
// Returns:
//   - *RequestBuilder: A pointer to the initialized RequestBuilder.
func NewRequestBuilder(httpClient *resty.Client) *RequestBuilder {
	return &RequestBuilder{
		httpClient: httpClient,
		compressor: compress.NewCompressor(),
	}
}

// Build creates a basic HTTP request with the specified method, endpoint, and body.
//
// Parameters:
//   - method: The HTTP method (e.g., GET, POST, etc.).
//   - endpoint: The URL endpoint for the request.
//   - body: The body content for the request as a byte slice.
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

// BuildWithGzip creates an HTTP request with a gzip-compressed body.
// It compresses the provided body data and, if a signing key is provided,
// computes and encodes a signature that is added as a header.
//
// Parameters:
//   - method: The HTTP method (e.g., POST, PUT, etc.).
//   - endpoint: The URL endpoint for the request.
//   - body: The body content to be compressed and included in the request.
//   - signingKey: A key used for signing the request payload; if empty, no signature is added.
//
// Returns:
//   - *resty.Request: The constructed HTTP request with a gzip-compressed body.
//   - error: An error if the compression process fails.
func (b *RequestBuilder) BuildWithGzip(method string, endpoint string, body []byte, signingKey string) (
	*resty.Request,
	error,
) {
	var s string
	if signingKey != "" {
		s = base64.StdEncoding.EncodeToString(sign.MakeSign(body, signingKey))
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
