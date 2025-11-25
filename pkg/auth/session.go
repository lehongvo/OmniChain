package auth

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"

	"github.com/onichange/pos-system/pkg/cache"
)

// Session represents a user session
type Session struct {
	ID        string
	UserID    string
	DeviceID  string
	IPAddress string
	UserAgent string
	CreatedAt time.Time
	ExpiresAt time.Time
	LastUsed  time.Time
}

// SessionManager manages user sessions
type SessionManager struct {
	cache           cache.Cache
	maxSessions     int
	sessionDuration time.Duration
	mu              sync.RWMutex
	sessions        map[string][]string // userID -> sessionIDs
}

// NewSessionManager creates a new session manager
func NewSessionManager(cache cache.Cache, maxSessions int, sessionDuration time.Duration) *SessionManager {
	return &SessionManager{
		cache:           cache,
		maxSessions:     maxSessions,
		sessionDuration: sessionDuration,
		sessions:        make(map[string][]string),
	}
}

// CreateSession creates a new session
func (sm *SessionManager) CreateSession(ctx context.Context, userID, deviceID, ipAddress, userAgent string) (*Session, error) {
	sessionID := uuid.New().String()
	now := time.Now()

	session := &Session{
		ID:        sessionID,
		UserID:    userID,
		DeviceID:  deviceID,
		IPAddress: ipAddress,
		UserAgent: userAgent,
		CreatedAt: now,
		ExpiresAt: now.Add(sm.sessionDuration),
		LastUsed:  now,
	}

	// Check concurrent session limit
	sm.mu.Lock()
	userSessions := sm.sessions[userID]
	if len(userSessions) >= sm.maxSessions {
		// Remove oldest session
		if len(userSessions) > 0 {
			oldestSessionID := userSessions[0]
			sm.cache.Delete(ctx, fmt.Sprintf("session:%s", oldestSessionID))
			sm.sessions[userID] = userSessions[1:]
		}
	}
	sm.sessions[userID] = append(sm.sessions[userID], sessionID)
	sm.mu.Unlock()

	// Store session in cache
	key := fmt.Sprintf("session:%s", sessionID)
	// In production, serialize session to JSON
	if err := sm.cache.Set(ctx, key, sessionID, sm.sessionDuration); err != nil {
		return nil, err
	}

	return session, nil
}

// GetSession retrieves a session
func (sm *SessionManager) GetSession(ctx context.Context, sessionID string) (*Session, error) {
	key := fmt.Sprintf("session:%s", sessionID)
	_, err := sm.cache.Get(ctx, key)
	if err != nil {
		return nil, fmt.Errorf("session not found")
	}

	// In production, deserialize from cache
	// For now, return a basic session
	return &Session{
		ID: sessionID,
	}, nil
}

// InvalidateSession invalidates a session
func (sm *SessionManager) InvalidateSession(ctx context.Context, sessionID string) error {
	key := fmt.Sprintf("session:%s", sessionID)
	return sm.cache.Delete(ctx, key)
}

// InvalidateUserSessions invalidates all sessions for a user
func (sm *SessionManager) InvalidateUserSessions(ctx context.Context, userID string) error {
	sm.mu.RLock()
	sessionIDs := sm.sessions[userID]
	sm.mu.RUnlock()

	for _, sessionID := range sessionIDs {
		sm.InvalidateSession(ctx, sessionID)
	}

	sm.mu.Lock()
	delete(sm.sessions, userID)
	sm.mu.Unlock()

	return nil
}

// DeviceFingerprint generates a device fingerprint
func DeviceFingerprint(userAgent, ipAddress string) string {
	// Simple fingerprint - in production, use more sophisticated method
	data := fmt.Sprintf("%s:%s", userAgent, ipAddress)
	hash := fmt.Sprintf("%x", []byte(data))
	return hash[:16] // Return first 16 chars
}
