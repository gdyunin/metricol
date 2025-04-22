package send

import (
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

// BuildWithParams creates an HTTP request with a gzip-compressed body.
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
func (b *RequestBuilder) BuildWithParams(
	method string,
	endpoint string,
	body []byte,
	signingKey string,
	publicKeyPEM string,
) (
	*resty.Request,
	error,
) {
	var s string
	if signingKey != "" {
		s = base64.StdEncoding.EncodeToString(sign.MakeSign(body, signingKey))
	}

	// Encrypt the body using the provided public key.
	// Using hybrid crypto method for correct work with large body.
	// [ДЛЯ РЕВЬЮ] Использую тут гибридное шифрование, т.к. тело большое и не шифруется "маленькими" (4096) rsa.
	// Поэтому в качестве выхода из ситуации подобрал такой подход. Будет работать даже если мы откажемся от сжатия.
	var e string
	if publicKeyPEM != "" {
		encryptedBody, encryptedKey, err := encryptWithPublicKeyHybrid(body, publicKeyPEM)
		if err != nil {
			return nil, fmt.Errorf("encryption failed for request body: %w", err)
		}

		body = encryptedBody
		e = base64.StdEncoding.EncodeToString(encryptedKey)
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
	if e != "" {
		req.SetHeader("X-Encrypted-Key", e)
	}

	return req, nil
}

// encryptWithPublicKeyHybrid encrypts the given data using a hybrid encryption scheme
// that combines AES-GCM for data encryption and RSA for encrypting the AES key.
//
// Parameters:
//   - data: The plaintext data to be encrypted.
//   - publicKeyPEM: The RSA public key in PEM format used to encrypt the AES key.
//
// Returns:
//   - encryptedData: The encrypted data, including the AES-GCM nonce and ciphertext.
//   - encryptedKey: The AES key encrypted with the RSA public key.
//   - err: An error if any step of the encryption process fails.
//
// The function performs the following steps:
//  1. Parses the provided RSA public key in PEM format.
//  2. Generates a random 256-bit AES key.
//  3. Encrypts the data using AES-GCM with the generated AES key.
//  4. Encrypts the AES key using the RSA public key.
//
// Errors are returned if the public key is invalid, the AES key generation fails,
// or any encryption step encounters an issue.
func encryptWithPublicKeyHybrid(
	data []byte,
	publicKeyPEM string,
) (encryptedData []byte, encryptedKey []byte, err error) {
	block, _ := pem.Decode([]byte(publicKeyPEM))
	if block == nil || block.Type != "PUBLIC KEY" {
		return nil, nil, errors.New("invalid public key PEM format")
	}

	pubKeyInterface, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse public key: %w", err)
	}

	rsaPubKey, ok := pubKeyInterface.(*rsa.PublicKey)
	if !ok {
		return nil, nil, errors.New("provided key is not an RSA public key")
	}

	const aesKeySize = 32
	aesKey := make([]byte, aesKeySize)
	if _, err = rand.Read(aesKey); err != nil {
		return nil, nil, fmt.Errorf("failed to generate AES key: %w", err)
	}

	blockCipher, err := aes.NewCipher(aesKey)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create AES cipher: %w", err)
	}

	aesGCM, err := cipher.NewGCM(blockCipher)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create AES-GCM: %w", err)
	}

	nonce := make([]byte, aesGCM.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, nil, fmt.Errorf("failed to generate nonce: %w", err)
	}

	ciphertext := aesGCM.Seal(nonce, nonce, data, nil)

	encryptedAESKey, err := rsa.EncryptPKCS1v15(rand.Reader, rsaPubKey, aesKey)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to encrypt AES key with RSA: %w", err)
	}

	return ciphertext, encryptedAESKey, nil
}
