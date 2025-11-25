package encryption

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"io"

	"golang.org/x/crypto/argon2"
)

const (
	// Argon2id parameters for password hashing
	argon2Time    = 2
	argon2Memory  = 64 * 1024 // 64MB
	argon2Threads = 4
	argon2KeyLen  = 32
)

// HashPassword hashes a password using Argon2id
func HashPassword(password string) (string, error) {
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}

	hash := argon2.IDKey([]byte(password), salt, argon2Time, argon2Memory, argon2Threads, argon2KeyLen)

	// Encode: salt + hash
	encoded := base64.RawStdEncoding.EncodeToString(append(salt, hash...))
	return encoded, nil
}

// VerifyPassword verifies a password against a hash
func VerifyPassword(password, hash string) (bool, error) {
	decoded, err := base64.RawStdEncoding.DecodeString(hash)
	if err != nil {
		return false, err
	}

	if len(decoded) < 16 {
		return false, errors.New("invalid hash format")
	}

	salt := decoded[:16]
	expectedHash := decoded[16:]

	computedHash := argon2.IDKey([]byte(password), salt, argon2Time, argon2Memory, argon2Threads, argon2KeyLen)

	// Time-constant comparison
	return constantTimeCompare(computedHash, expectedHash), nil
}

// constantTimeCompare performs a time-constant comparison
func constantTimeCompare(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}

	var result byte
	for i := 0; i < len(a); i++ {
		result |= a[i] ^ b[i]
	}

	return result == 0
}

// AESGCMEncrypt encrypts data using AES-256-GCM
func AESGCMEncrypt(plaintext []byte, key []byte) ([]byte, error) {
	// Derive key from input (in production, use proper key derivation)
	hasher := sha256.New()
	hasher.Write(key)
	derivedKey := hasher.Sum(nil)

	block, err := aes.NewCipher(derivedKey)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)
	return ciphertext, nil
}

// AESGCMDecrypt decrypts data using AES-256-GCM
func AESGCMDecrypt(ciphertext []byte, key []byte) ([]byte, error) {
	// Derive key from input (in production, use proper key derivation)
	hasher := sha256.New()
	hasher.Write(key)
	derivedKey := hasher.Sum(nil)

	block, err := aes.NewCipher(derivedKey)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, errors.New("ciphertext too short")
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, err
	}

	return plaintext, nil
}
