package registry

import (
	"sync"
	"time"
)

// ClientSession tracks a client's affinity to a specific backend server
type ClientSession struct {
	ID         string    // Unique session ID
	ServerID   string    // Which backend server this session is pinned to
	LastActive time.Time // Last time this session was used
	CreatedAt  time.Time // When this session was created
}

// SessionManager handles session affinity
type SessionManager struct {
	sessions map[string]*ClientSession
	mu       sync.RWMutex
	ttl      time.Duration
}

// NewSessionManager creates a new session manager
func NewSessionManager(ttl time.Duration) *SessionManager {
	if ttl == 0 {
		ttl = 30 * time.Minute
	}

	sm := &SessionManager{
		sessions: make(map[string]*ClientSession),
		ttl:      ttl,
	}

	// Start background cleanup
	go sm.cleanupLoop()
	return sm
}

// GetOrCreate retrieves an existing session or creates a new one
func (sm *SessionManager) GetOrCreate(sessionID string, preferredServerID string) *ClientSession {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	// Check if session exists
	if session, exists := sm.sessions[sessionID]; exists {
		session.LastActive = time.Now()
		return session
	}

	// Create new session
	session := &ClientSession{
		ID:         sessionID,
		ServerID:   preferredServerID,
		LastActive: time.Now(),
		CreatedAt:  time.Now(),
	}

	sm.sessions[sessionID] = session
	return session
}

// Get retrieves a session by ID
func (sm *SessionManager) Get(sessionID string) (*ClientSession, bool) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	session, exists := sm.sessions[sessionID]
	return session, exists
}

// UpdateServer changes which server a session is pinned to
func (sm *SessionManager) UpdateServer(sessionID string, newServerID string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if session, exists := sm.sessions[sessionID]; exists {
		session.ServerID = newServerID
		session.LastActive = time.Now()
	}
}

// Remove deletes a session
func (sm *SessionManager) Remove(sessionID string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	delete(sm.sessions, sessionID)
}

// cleanupLoop runs periodically to remove stale sessions
func (sm *SessionManager) cleanupLoop() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		sm.cleanup()
	}
}

// cleanup removes sessions that haven't been active
func (sm *SessionManager) cleanup() {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	now := time.Now()
	for id, session := range sm.sessions {
		if now.Sub(session.LastActive) > sm.ttl {
			delete(sm.sessions, id)
		}
	}
}

// GetStats returns session statistics
func (sm *SessionManager) GetStats() map[string]interface{} {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	return map[string]interface{}{
		"total_sessions": len(sm.sessions),
		"ttl_minutes":    sm.ttl.Minutes(),
	}
}
