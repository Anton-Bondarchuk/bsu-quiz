package repository

import (
	"context"
	"fmt"
	"bsu-quiz/telegram/internal/domain/models"
	"sync"
	"time"
)

// MemoryStorage implements Storage using in-memory maps
type MemoryStorage struct {
	states          map[string]models.State
	data            map[string]map[string]interface{}
	mu              sync.RWMutex
	cleanupInterval time.Duration
	lastAccess      map[string]time.Time
}

// NewMemoryStorage creates a new memory-based storage
func NewMemoryStorage(cleanupInterval time.Duration) *MemoryStorage {
	storage := &MemoryStorage{
		states:          make(map[string]models.State),
		data:            make(map[string]map[string]interface{}),
		lastAccess:      make(map[string]time.Time),
		cleanupInterval: cleanupInterval,
	}

	if cleanupInterval > 0 {
		go storage.cleanup()
	}

	return storage
}

// cleanup periodically removes old states and data
func (s *MemoryStorage) cleanup() {
	ticker := time.NewTicker(s.cleanupInterval)
	defer ticker.Stop()

	for range ticker.C {
		now := time.Now()
		s.mu.Lock()

		for key, lastAccess := range s.lastAccess {
			if now.Sub(lastAccess) > s.cleanupInterval*2 {
				delete(s.states, key)
				delete(s.data, key)
				delete(s.lastAccess, key)
			}
		}

		s.mu.Unlock()
	}
}

// generateKey generates a unique key for a chat+user combination
func generateKey(chatID, userID int64) string {
	return fmt.Sprintf("%d:%d", chatID, userID)
}

// Get implements Storage.Get
func (s *MemoryStorage) Get(ctx context.Context, chatID int64, userID int64) (models.State, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	k := generateKey(chatID, userID)
	s.lastAccess[k] = time.Now()
	return s.states[k], nil
}

// Set implements Storage.Set
func (s *MemoryStorage) Set(ctx context.Context, chatID int64, userID int64, state models.State) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	k := generateKey(chatID, userID)
	s.states[k] = state
	s.lastAccess[k] = time.Now()
	return nil
}

// Delete implements Storage.Delete
func (s *MemoryStorage) Delete(ctx context.Context, chatID int64, userID int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	k := generateKey(chatID, userID)
	delete(s.states, k)
	s.lastAccess[k] = time.Now()
	return nil
}

// GetData implements Storage.GetData
func (s *MemoryStorage) GetData(ctx context.Context, chatID int64, userID int64, key string) (interface{}, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	k := generateKey(chatID, userID)
	s.lastAccess[k] = time.Now()

	userData, exists := s.data[k]
	if !exists {
		return nil, nil
	}

	return userData[key], nil
}

// SetData implements Storage.SetData
func (s *MemoryStorage) SetData(ctx context.Context, chatID int64, userID int64, key string, value interface{}) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	k := generateKey(chatID, userID)
	s.lastAccess[k] = time.Now()

	if s.data[k] == nil {
		s.data[k] = make(map[string]interface{})
	}

	s.data[k][key] = value
	return nil
}

// ClearData implements Storage.ClearData
func (s *MemoryStorage) ClearData(ctx context.Context, chatID int64, userID int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	k := generateKey(chatID, userID)
	delete(s.data, k)
	s.lastAccess[k] = time.Now()
	return nil
}
