package send

import (
	"encoding/base64"
	"testing"

	"github.com/gdyunin/metricol.git/internal/agent/send/compress"
	"github.com/gdyunin/metricol.git/pkg/sign"
	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
)

func TestBuild(t *testing.T) {
	client := resty.New()
	builder := NewRequestBuilder(client)

	tests := []struct {
		name     string
		method   string
		endpoint string
		body     []byte
	}{
		{
			name:     "Simple GET request",
			method:   "GET",
			endpoint: "https://example.com",
			body:     []byte("test"),
		},
		{
			name:     "Simple POST request",
			method:   "POST",
			endpoint: "https://example.com/data",
			body:     []byte(`{"key": "value"}`),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := builder.Build(tt.method, tt.endpoint, tt.body)
			assert.Equal(t, tt.method, req.Method)
			assert.Equal(t, tt.endpoint, req.URL)
			assert.Equal(t, tt.body, req.Body)
		})
	}
}

func TestBuildWithGzip(t *testing.T) {
	client := resty.New()
	builder := NewRequestBuilder(client)
	compressor := compress.NewCompressor()

	tests := []struct {
		name       string
		method     string
		endpoint   string
		signingKey string
		body       []byte
		expectErr  bool
	}{
		{
			name:       "Gzip compressed POST request with signing",
			method:     "POST",
			endpoint:   "https://example.com/data",
			body:       []byte(`{"key": "value"}`),
			signingKey: "secretKey",
		},
		{
			name:     "Gzip compressed POST request without signing",
			method:   "POST",
			endpoint: "https://example.com/data",
			body:     []byte(`{"key": "value"}`),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := builder.BuildWithParams(tt.method, tt.endpoint, tt.body, tt.signingKey, "")
			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.method, req.Method)
				assert.Equal(t, tt.endpoint, req.URL)
				compressedBody, _ := compressor.Compress(tt.body)
				assert.Equal(t, compressedBody, req.Body)

				if tt.signingKey != "" {
					expectedHash := base64.StdEncoding.EncodeToString(sign.MakeSign(tt.body, tt.signingKey))
					assert.Equal(t, expectedHash, req.Header.Get("HashSHA256"))
				}
			}
		})
	}
}
