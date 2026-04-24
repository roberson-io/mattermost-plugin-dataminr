package main

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEncryptDecrypt(t *testing.T) {
	t.Run("successfully encrypts and decrypts data", func(t *testing.T) {
		key := []byte("test-encryption-key-32-bytes!!!")
		plaintext := "sensitive-data-to-encrypt"

		encrypted, err := encrypt(key, plaintext)
		require.NoError(t, err)
		require.NotEmpty(t, encrypted)

		// Encrypted text should be different from plaintext
		assert.NotEqual(t, plaintext, encrypted)

		decrypted, err := decrypt(key, encrypted)
		require.NoError(t, err)
		assert.Equal(t, plaintext, decrypted)
	})

	t.Run("encrypted output is different each time (IV randomization)", func(t *testing.T) {
		key := []byte("test-encryption-key-32-bytes!!!")
		plaintext := "same-data"

		encrypted1, err := encrypt(key, plaintext)
		require.NoError(t, err)

		encrypted2, err := encrypt(key, plaintext)
		require.NoError(t, err)

		// Due to random IV, encrypted outputs should differ
		assert.NotEqual(t, encrypted1, encrypted2)

		// But both should decrypt to the same plaintext
		decrypted1, err := decrypt(key, encrypted1)
		require.NoError(t, err)
		assert.Equal(t, plaintext, decrypted1)

		decrypted2, err := decrypt(key, encrypted2)
		require.NoError(t, err)
		assert.Equal(t, plaintext, decrypted2)
	})

	t.Run("fails to encrypt with empty key", func(t *testing.T) {
		key := []byte("")
		plaintext := "test-data"

		_, err := encrypt(key, plaintext)
		assert.Error(t, err)
	})

	t.Run("fails to decrypt with wrong key", func(t *testing.T) {
		key1 := []byte("test-encryption-key-32-bytes!!!")
		key2 := []byte("different-key-32-bytes-here!!!")
		plaintext := "secret-data"

		encrypted, err := encrypt(key1, plaintext)
		require.NoError(t, err)

		_, err = decrypt(key2, encrypted)
		assert.Error(t, err)
	})

	t.Run("fails to decrypt corrupted data", func(t *testing.T) {
		key := []byte("test-encryption-key-32-bytes!!!")
		corruptedData := "corrupted-random-string"

		_, err := decrypt(key, corruptedData)
		assert.Error(t, err)
	})

	t.Run("handles empty plaintext", func(t *testing.T) {
		key := []byte("test-encryption-key-32-bytes!!!")
		plaintext := ""

		encrypted, err := encrypt(key, plaintext)
		require.NoError(t, err)

		decrypted, err := decrypt(key, encrypted)
		require.NoError(t, err)
		assert.Equal(t, plaintext, decrypted)
	})

	t.Run("handles large plaintext", func(t *testing.T) {
		key := []byte("test-encryption-key-32-bytes!!!")
		// Create a large JSON-like payload simulating credentials
		plaintext := strings.Repeat("large-data-block-", 1000)

		encrypted, err := encrypt(key, plaintext)
		require.NoError(t, err)

		decrypted, err := decrypt(key, encrypted)
		require.NoError(t, err)
		assert.Equal(t, plaintext, decrypted)
	})
}

func TestKeyDerivation(t *testing.T) {
	t.Run("derives consistent key from password", func(t *testing.T) {
		password := "my-secure-password"

		key1 := deriveKey([]byte(password))
		key2 := deriveKey([]byte(password))

		assert.Equal(t, key1, key2)
		assert.Len(t, key1, 32) // AES-256 requires 32-byte key
	})

	t.Run("different passwords produce different keys", func(t *testing.T) {
		password1 := "password-one"
		password2 := "password-two"

		key1 := deriveKey([]byte(password1))
		key2 := deriveKey([]byte(password2))

		assert.NotEqual(t, key1, key2)
	})

	t.Run("handles empty password", func(t *testing.T) {
		password := ""
		key := deriveKey([]byte(password))

		assert.NotNil(t, key)
		assert.Len(t, key, 32)
	})
}
