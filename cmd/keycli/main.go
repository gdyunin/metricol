// Package main provides a CLI tool for generating RSA key pairs and saving them to files.
package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"flag"
	"fmt"
	"os"
)

const defaultKeySize = 2048

func main() {
	// Command-line flags
	var (
		privateKeyPath string
		publicKeyPath  string
		keySize        int
	)

	flag.StringVar(&privateKeyPath, "private", "private_key.pem", "Path to save the private key")
	flag.StringVar(&publicKeyPath, "public", "public_key.pem", "Path to save the public key")
	flag.IntVar(&keySize, "size", defaultKeySize, "Key size in bits (2048 or 4096 is recommended)")
	flag.Parse()

	fmt.Println("Generating RSA key pair...")

	// Generate the RSA private key
	privateKey, err := rsa.GenerateKey(rand.Reader, keySize)
	if err != nil {
		panic(fmt.Errorf("failed to generate private key: %w", err))
	}

	// Save the private key to a file
	if err := savePrivateKey(privateKey, privateKeyPath); err != nil {
		panic(err)
	}

	// Save the public key to a file
	if err := savePublicKey(&privateKey.PublicKey, publicKeyPath); err != nil {
		panic(err)
	}

	fmt.Println("Keys successfully generated!")
	fmt.Println("Private key saved to:", privateKeyPath)
	fmt.Println("Public key saved to:", publicKeyPath)
}

// savePrivateKey saves the RSA private key to the specified file in PEM format.
// The file will be created or overwritten if it already exists.
func savePrivateKey(key *rsa.PrivateKey, filepath string) error {
	f, err := os.Create(filepath)
	if err != nil {
		return fmt.Errorf("failed to create private key file: %w", err)
	}
	defer func() { _ = f.Close() }()

	// Marshal the private key to PKCS#1 ASN.1 DER format
	privateKeyBytes := x509.MarshalPKCS1PrivateKey(key)

	// Create a PEM block for the private key
	pemBlock := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: privateKeyBytes,
	}

	// Write the PEM block to the file
	if err = pem.Encode(f, pemBlock); err != nil {
		return fmt.Errorf("failed to encode private key to PEM format")
	}
	return nil
}

// savePublicKey saves the RSA public key to the specified file in PEM format.
// The file will be created or overwritten if it already exists.
func savePublicKey(key *rsa.PublicKey, filepath string) error {
	f, err := os.Create(filepath)
	if err != nil {
		return fmt.Errorf("failed to create public key file: %w", err)
	}
	defer func() { _ = f.Close() }()

	// Marshal the public key to PKIX ASN.1 DER format
	publicKeyBytes, err := x509.MarshalPKIXPublicKey(key)
	if err != nil {
		return fmt.Errorf("failed to serialize public key: %w", err)
	}

	// Create a PEM block for the public key
	pemBlock := &pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: publicKeyBytes,
	}

	// Write the PEM block to the file
	if err := pem.Encode(f, pemBlock); err != nil {
		return fmt.Errorf("failed to encode public key to PEM format: %w", err)
	}
	return nil
}
