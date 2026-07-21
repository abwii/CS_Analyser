package veto

import (
	"sync"

	"github.com/google/uuid"
)

type Manager struct {
	mu       sync.RWMutex
	sessions map[uuid.UUID]*Session
}

func NewManager() *Manager {
	return &Manager{sessions: make(map[uuid.UUID]*Session)}
}

func (m *Manager) CreateSession(format Format, teamA, teamB string, maps []string) (*Session, error) {
	session, err := newSession(format, teamA, teamB, maps)
	if err != nil {
		return nil, err
	}

	m.mu.Lock()
	m.sessions[session.ID()] = session
	m.mu.Unlock()

	return session, nil
}

func (m *Manager) Get(id uuid.UUID) (*Session, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	session, ok := m.sessions[id]
	return session, ok
}
