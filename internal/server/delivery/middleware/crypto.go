package middleware

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

// [ДЛЯ РЕВЬЮ] Этот волшебный 🩼 -- плата за экономию на переделывании `internal/server/delivery/http_server.go`...
var cryptoIgnoredPath = map[string]bool{
	"/":     true,
	"/ping": true,
}

func Crypto(cryptoKey string, logger *zap.SugaredLogger) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) (err error) {
			if cryptoKey == "" || cryptoIgnoredPath[c.Request().URL.Path] {
				return next(c)
			}

			encryptedKeyB64 := c.Request().Header.Get("X-Encrypted-Key")
			encryptedKey, err := base64.StdEncoding.DecodeString(encryptedKeyB64)
			if err != nil {
				logger.Errorf("failed to decode base64 encrypted key: %v", err)
				return c.String(
					http.StatusInternalServerError,
					http.StatusText(http.StatusInternalServerError),
				)
			}

			encryptedBody, err := io.ReadAll(c.Request().Body)
			if err != nil {
				logger.Errorf("failed to read request body: %v", err)
				return c.String(
					http.StatusInternalServerError,
					http.StatusText(http.StatusInternalServerError),
				)
			}

			decryptedBody, err := decryptWithPrivateKeyHybrid(encryptedBody, encryptedKey, cryptoKey)
			if err != nil {
				return c.String(http.StatusBadRequest, http.StatusText(http.StatusBadRequest))
			}

			c.Request().Body = io.NopCloser(bytes.NewReader(decryptedBody))

			return next(c)
		}
	}
}

func decryptWithPrivateKeyHybrid(
	encryptedData []byte,
	encryptedKey []byte,
	privateKeyPEM string,
) ([]byte, error) {
	block, _ := pem.Decode([]byte(privateKeyPEM))
	if block == nil || block.Type != "RSA PRIVATE KEY" {
		return nil, errors.New("invalid private key PEM format")
	}

	privKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key: %w", err)
	}

	aesKey, err := rsa.DecryptPKCS1v15(rand.Reader, privKey, encryptedKey)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt AES key: %w", err)
	}

	blockCipher, err := aes.NewCipher(aesKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create AES cipher: %w", err)
	}

	aesGCM, err := cipher.NewGCM(blockCipher)
	if err != nil {
		return nil, fmt.Errorf("failed to create AES-GCM: %w", err)
	}

	if len(encryptedData) < aesGCM.NonceSize() {
		return nil, errors.New("encrypted data too short")
	}

	nonce := encryptedData[:aesGCM.NonceSize()]
	ciphertext := encryptedData[aesGCM.NonceSize():]

	plaintext, err := aesGCM.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt data: %w", err)
	}

	return plaintext, nil
}
