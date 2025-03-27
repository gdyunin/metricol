package middleware

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
)

// Auth creates an authentication middleware for Echo that verifies HMAC-SHA256 signatures.
// If a key is provided, the middleware checks for the "HashSHA256" header and verifies the
// signature against the raw request body using the secret key.
// If the key is empty or no signature is provided, the request proceeds without verification.
//
// Parameters:
//   - key: The secret key used to verify the HMAC signature.
//
// Returns:
//   - echo.MiddlewareFunc: The configured middleware function.
func Auth(key string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) (err error) {
			if key == "" {
				return next(c)
			}

			sign := c.Request().Header.Get("HashSHA256")
			if sign == "" {
				return next(c) // https://app.pachca.com/chats?thread_message_id=419958556
			}

			rawBody, err := getRawBody(c.Request())
			if err != nil {
				return c.String(
					http.StatusInternalServerError,
					http.StatusText(http.StatusInternalServerError),
				)
			}

			if !checkSign(rawBody, sign, key) {
				return c.String(http.StatusBadRequest, http.StatusText(http.StatusBadRequest))
			}

			return next(c)
		}
	}
}

// getRawBody retrieves the raw body from an http.Request.
//
// Parameters:
//   - req: The HTTP request from which to read the body.
//
// Returns:
//   - []byte: The raw request body.
//   - error: An error if the body cannot be read.
func getRawBody(req *http.Request) ([]byte, error) {
	rawBody, err := io.ReadAll(req.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read body: %w", err)
	}
	defer func() {
		if err = req.Body.Close(); err != nil {
			log.Errorf("failed to close body: %v", err)
		}
	}()

	// Restore the Body so it can be read later in the chain.
	req.Body = io.NopCloser(bytes.NewBuffer(rawBody))

	return rawBody, nil
}

// checkSign verifies the HMAC-SHA256 signature for the provided body using the secret key.
//
// Parameters:
//   - body: The original request body.
//   - sign: The base64-encoded signature provided in the request header.
//   - key: The secret key used to compute the signature.
//
// Returns:
//   - bool: True if the computed signature matches the provided one; otherwise, false.
func checkSign(body []byte, sign string, key string) bool {
	h := hmac.New(sha256.New, []byte(key))
	h.Write(body)

	gotSign, err := base64.StdEncoding.DecodeString(sign)
	if err != nil {
		return false
	}
	validSign := h.Sum(nil)
	return hmac.Equal(gotSign, validSign)
}
