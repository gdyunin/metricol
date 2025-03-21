package middleware

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"net/http"

	"github.com/gdyunin/metricol.git/pkg/sign"
	"github.com/labstack/echo/v4"
)

// Sign creates an Echo middleware that signs the HTTP response body using HMAC-SHA256.
// The middleware wraps the response writer so that after the response body is written,
// a "HashSHA256" header is added containing the signature computed using the provided key.
//
// Parameters:
//   - key: The secret key used to sign the response body.
//
// Returns:
//   - echo.MiddlewareFunc: The middleware function that applies the signing logic.
func Sign(key string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) (err error) {
			if key == "" {
				return next(c)
			}

			w := &signerWriter{
				ResponseWriter: c.Response().Writer,
				key:            key,
				body:           &bytes.Buffer{},
			}
			c.Response().Writer = w

			err = next(c)
			if err != nil {
				c.Error(err)
			}

			// Ensure WriteHeader is called after the response body is populated.
			if w.body.Len() > 0 {
				w.WriteHeader(http.StatusOK)
			}

			return nil
		}
	}
}

// signerWriter wraps an http.ResponseWriter to capture the response body
// and to add a signature header upon writing the header.
type signerWriter struct {
	http.ResponseWriter
	body *bytes.Buffer // body buffers the response data.
	key  string        // key is used to compute the HMAC signature.
}

// Write writes data to the buffer and to the underlying ResponseWriter.
//
// Parameters:
//   - data: The data to write.
//
// Returns:
//   - int: The number of bytes written.
//   - error: An error if the write fails.
func (w *signerWriter) Write(data []byte) (int, error) {
	w.body.Write(data)
	i, err := w.ResponseWriter.Write(data)
	if err != nil {
		err = fmt.Errorf("error writing data in signer writer: %w", err)
	}
	return i, err
}

// WriteHeader sets the "HashSHA256" header with the computed signature and writes the HTTP header.
//
// Parameters:
//   - statusCode: The HTTP status code to write.
func (w *signerWriter) WriteHeader(statusCode int) {
	w.Header().Set(
		"HashSHA256",
		hex.EncodeToString(sign.MakeSign(w.body.Bytes(), w.key)),
	)
	w.ResponseWriter.WriteHeader(statusCode)
}
