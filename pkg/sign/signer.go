// Package sign provides functionality for generating HMAC-SHA256 signatures.
// It includes functions to sign data using a secret key, ensuring data integrity and authenticity.
package sign

import (
	"crypto/hmac"
	"crypto/sha256"
)

// MakeSign generates an HMAC-SHA256 signature for the provided data using the given key.
// The function creates a new HMAC using SHA256, writes the data into it, and returns the computed signature.
//
// Parameters:
//   - data: The message to be signed as a byte slice.
//   - key: The secret key used for signing the data.
//
// Returns:
//   - []byte: The generated HMAC-SHA256 signature.
func MakeSign(data []byte, key string) []byte {
	h := hmac.New(sha256.New, []byte(key))
	h.Write(data)
	return h.Sum(nil)
}
