package auth

import (
	"context"
	"fmt"
	"time"

	"github.com/onichange/pos-system/pkg/cache"
)

// TokenStore manages JWT token whitelist in Redis
type TokenStore struct {
	cache cache.Cache
}

// NewTokenStore creates a new token store
func NewTokenStore(cache cache.Cache) *TokenStore {
	return &TokenStore{
		cache: cache,
	}
}

// StoreAccessToken stores an access token in whitelist
func (ts *TokenStore) StoreAccessToken(ctx context.Context, tokenID, userID string, expiry time.Duration) error {
	key := fmt.Sprintf("token:access:%s", tokenID)
	value := fmt.Sprintf("user:%s", userID)
	return ts.cache.Set(ctx, key, value, expiry)
}

// StoreRefreshToken stores a refresh token in whitelist
func (ts *TokenStore) StoreRefreshToken(ctx context.Context, tokenID, userID string, expiry time.Duration) error {
	key := fmt.Sprintf("token:refresh:%s", tokenID)
	value := fmt.Sprintf("user:%s", userID)
	return ts.cache.Set(ctx, key, value, expiry)
}

// ValidateToken checks if token exists in whitelist
func (ts *TokenStore) ValidateToken(ctx context.Context, tokenID string, isRefresh bool) (bool, error) {
	tokenType := "access"
	if isRefresh {
		tokenType = "refresh"
	}
	key := fmt.Sprintf("token:%s:%s", tokenType, tokenID)
	return ts.cache.Exists(ctx, key)
}

// RevokeToken revokes a token
func (ts *TokenStore) RevokeToken(ctx context.Context, tokenID string, isRefresh bool) error {
	tokenType := "access"
	if isRefresh {
		tokenType = "refresh"
	}
	key := fmt.Sprintf("token:%s:%s", tokenType, tokenID)
	return ts.cache.Delete(ctx, key)
}

// RotateToken rotates a token (revokes old, stores new)
func (ts *TokenStore) RotateToken(ctx context.Context, oldTokenID, newTokenID, userID string, isRefresh bool, expiry time.Duration) error {
	// Revoke old token
	if err := ts.RevokeToken(ctx, oldTokenID, isRefresh); err != nil {
		return err
	}

	// Store new token
	if isRefresh {
		return ts.StoreRefreshToken(ctx, newTokenID, userID, expiry)
	}
	return ts.StoreAccessToken(ctx, newTokenID, userID, expiry)
}

