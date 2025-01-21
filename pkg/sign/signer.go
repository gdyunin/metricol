package sign

import (
	"crypto/hmac"
	"crypto/sha256"
)

// MakeSign generates an HMAC-SHA256 signature for a given body and key.
//
// Parameters:
//   - body: The message to be signed.
//   - key: The secret key used to generate the signature.
//
// Returns:
//   - []byte: The generated HMAC-SHA256 signature.
func MakeSign(data []byte, key string) []byte {
	h := hmac.New(sha256.New, []byte(key))
	h.Write(data)
	return h.Sum(nil)
}
