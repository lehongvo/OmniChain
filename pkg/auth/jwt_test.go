package auth

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJWTManager_GenerateTokenPair(t *testing.T) {
	manager := NewJWTManager(
		"test-access-secret-key-minimum-32-characters-long",
		"test-refresh-secret-key-minimum-32-characters-long",
		15*time.Minute,
		7*24*time.Hour,
		"test-issuer",
	)

	userID := "user-123"
	email := "test@example.com"
	roles := []string{"user", "admin"}
	deviceID := "device-123"

	tokenPair, err := manager.GenerateTokenPair(userID, email, roles, deviceID)
	require.NoError(t, err)
	assert.NotEmpty(t, tokenPair.AccessToken)
	assert.NotEmpty(t, tokenPair.RefreshToken)
	assert.True(t, tokenPair.ExpiresAt.After(time.Now()))
}

func TestJWTManager_ValidateAccessToken(t *testing.T) {
	manager := NewJWTManager(
		"test-access-secret-key-minimum-32-characters-long",
		"test-refresh-secret-key-minimum-32-characters-long",
		15*time.Minute,
		7*24*time.Hour,
		"test-issuer",
	)

	userID := "user-123"
	email := "test@example.com"
	roles := []string{"user"}
	deviceID := "device-123"

	tokenPair, err := manager.GenerateTokenPair(userID, email, roles, deviceID)
	require.NoError(t, err)

	claims, err := manager.ValidateAccessToken(tokenPair.AccessToken)
	require.NoError(t, err)
	assert.Equal(t, userID, claims.UserID)
	assert.Equal(t, email, claims.Email)
	assert.Equal(t, roles, claims.Roles)
	assert.Equal(t, deviceID, claims.DeviceID)
}

func TestJWTManager_ValidateInvalidToken(t *testing.T) {
	manager := NewJWTManager(
		"test-access-secret-key-minimum-32-characters-long",
		"test-refresh-secret-key-minimum-32-characters-long",
		15*time.Minute,
		7*24*time.Hour,
		"test-issuer",
	)

	_, err := manager.ValidateAccessToken("invalid-token")
	assert.Error(t, err)
}

func TestJWTManager_ValidateRefreshToken(t *testing.T) {
	manager := NewJWTManager(
		"test-access-secret-key-minimum-32-characters-long",
		"test-refresh-secret-key-minimum-32-characters-long",
		15*time.Minute,
		7*24*time.Hour,
		"test-issuer",
	)

	userID := "user-123"
	email := "test@example.com"
	roles := []string{"user"}
	deviceID := "device-123"

	tokenPair, err := manager.GenerateTokenPair(userID, email, roles, deviceID)
	require.NoError(t, err)

	claims, err := manager.ValidateRefreshToken(tokenPair.RefreshToken)
	require.NoError(t, err)
	assert.Equal(t, userID, claims.UserID)
}

