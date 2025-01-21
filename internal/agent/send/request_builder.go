package send

import (
	"crypto/hmac"
	"crypto/sha256"
	"fmt"
	"io"

	"github.com/gdyunin/metricol.git/internal/agent/send/compress"
	"github.com/labstack/gommon/log"

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
func (b *RequestBuilder) BuildWithGzip(method string, endpoint string, body []byte) (*resty.Request, error) {
	body, err := b.compressor.Compress(body)
	if err != nil {
		return nil, fmt.Errorf("gzip compression failed for request body: %w", err)
	}

	req := b.Build(method, endpoint, body)
	req.SetHeader("Content-Encoding", "gzip")

	return req, nil
}

// SignRequest signs a given resty.Request using a specified key.
//
// Parameters:
//   - req: The HTTP request to be signed.
//   - key: The secret key used to generate the HMAC-SHA256 signature.
//
// Returns:
//   - error: An error if the signing process fails; otherwise, nil.
func (b *RequestBuilder) SignRequest(req *resty.Request, key string) error {
	rawBody, err := b.getRawBody(req)
	if err != nil {
		return fmt.Errorf("failed get raw body: %w", err)
	}

	sign := b.makeSign(rawBody, key)

	req.SetBody(rawBody)
	req.SetHeader("HashSHA256", string(sign))

	return nil
}

// getRawBody retrieves the raw body of a resty.Request.
//
// Parameters:
//   - req: The HTTP request whose body needs to be read.
//
// Returns:
//   - []byte: The raw body of the request as a byte slice.
//   - error: An error if the body retrieval or reading fails; otherwise, nil.
func (b *RequestBuilder) getRawBody(req *resty.Request) ([]byte, error) {
	bodyReader, err := req.RawRequest.GetBody()
	if err != nil {
		return nil, fmt.Errorf("failed to get body: %w", err)
	}
	defer func() {
		if err = bodyReader.Close(); err != nil {
			log.Errorf("failed close body: %v", err)
		}
	}()

	bodyBytes, err := io.ReadAll(bodyReader)
	if err != nil {
		return nil, fmt.Errorf("failed to read body: %w", err)
	}
	return bodyBytes, nil
}

// makeSign generates an HMAC-SHA256 signature for a given body and key.
//
// Parameters:
//   - body: The message to be signed.
//   - key: The secret key used to generate the signature.
//
// Returns:
//   - []byte: The generated HMAC-SHA256 signature.
func (b *RequestBuilder) makeSign(body []byte, key string) []byte {
	h := hmac.New(sha256.New, []byte(key))
	h.Write(body)
	return h.Sum(nil)
}
