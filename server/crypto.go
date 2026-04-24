package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"io"

	"github.com/pkg/errors"
)

// encrypt encrypts plaintext using AES-256-GCM with the provided key.
// Returns base64-encoded ciphertext or an error.
func encrypt(key []byte, plaintext string) (string, error) {
	if len(key) == 0 {
		return "", errors.New("encryption key cannot be empty")
	}

	gcm, err := createGCM(key)
	if err != nil {
		return "", err
	}

	// Generate random nonce
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", errors.Wrap(err, "failed to generate nonce")
	}

	// Encrypt plaintext
	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)

	// Encode to base64 for storage
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// decrypt decrypts base64-encoded ciphertext using AES-256-GCM with the provided key.
// Returns plaintext or an error.
func decrypt(key []byte, ciphertext string) (string, error) {
	if len(key) == 0 {
		return "", errors.New("encryption key cannot be empty")
	}

	gcm, err := createGCM(key)
	if err != nil {
		return "", err
	}

	// Decode from base64
	data, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", errors.Wrap(err, "failed to decode base64")
	}

	// Verify we have enough data
	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return "", errors.New("ciphertext too short")
	}

	// Extract nonce and encrypted data
	nonce, encryptedData := data[:nonceSize], data[nonceSize:]

	// Decrypt
	plaintext, err := gcm.Open(nil, nonce, encryptedData, nil)
	if err != nil {
		return "", errors.Wrap(err, "failed to decrypt")
	}

	return string(plaintext), nil
}

// createGCM creates a GCM cipher from the provided key.
func createGCM(key []byte) (cipher.AEAD, error) {
	// Derive a 32-byte key from the provided key
	derivedKey := deriveKey(key)

	// Create AES cipher block
	block, err := aes.NewCipher(derivedKey)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create cipher block")
	}

	// Create GCM mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create GCM")
	}

	return gcm, nil
}

// deriveKey derives a 32-byte AES key from a password using SHA-256.
// This ensures any password length can be used as an encryption key.
func deriveKey(password []byte) []byte {
	hash := sha256.Sum256(password)
	return hash[:]
}
