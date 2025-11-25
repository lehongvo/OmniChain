package encryption

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHashPassword(t *testing.T) {
	password := "test-password-123"
	hash, err := HashPassword(password)
	require.NoError(t, err)
	assert.NotEmpty(t, hash)
	assert.NotEqual(t, password, hash)
}

func TestVerifyPassword(t *testing.T) {
	password := "test-password-123"
	hash, err := HashPassword(password)
	require.NoError(t, err)

	valid, err := VerifyPassword(password, hash)
	require.NoError(t, err)
	assert.True(t, valid)
}

func TestVerifyPassword_Invalid(t *testing.T) {
	password := "test-password-123"
	wrongPassword := "wrong-password"
	hash, err := HashPassword(password)
	require.NoError(t, err)

	valid, err := VerifyPassword(wrongPassword, hash)
	require.NoError(t, err)
	assert.False(t, valid)
}

func TestAESGCMEncryptDecrypt(t *testing.T) {
	key := []byte("test-key-32-characters-long!!")
	plaintext := []byte("sensitive data to encrypt")

	ciphertext, err := AESGCMEncrypt(plaintext, key)
	require.NoError(t, err)
	assert.NotEmpty(t, ciphertext)
	assert.NotEqual(t, plaintext, ciphertext)

	decrypted, err := AESGCMDecrypt(ciphertext, key)
	require.NoError(t, err)
	assert.Equal(t, plaintext, decrypted)
}

func TestAESGCMDecrypt_InvalidKey(t *testing.T) {
	key := []byte("test-key-32-characters-long!!")
	wrongKey := []byte("wrong-key-32-characters-long!!")
	plaintext := []byte("sensitive data")

	ciphertext, err := AESGCMEncrypt(plaintext, key)
	require.NoError(t, err)

	_, err = AESGCMDecrypt(ciphertext, wrongKey)
	assert.Error(t, err)
}

