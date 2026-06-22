package ingestion

import (
	"sync"
	"time"
)

// IdempotencyStore tracks recently processed event IDs to prevent duplicate processing.
type IdempotencyStore struct {
	mu      sync.Mutex
	seen    map[string]time.Time
	ttl     time.Duration
	stopCh  chan struct{}
}

// NewIdempotencyStore creates a store that deduplicates events for the given TTL window.
func NewIdempotencyStore(ttl time.Duration) *IdempotencyStore {
	s := &IdempotencyStore{
		seen:   make(map[string]time.Time),
		ttl:    ttl,
		stopCh: make(chan struct{}),
	}
	go s.cleanup()
	return s
}

// MarkSeen returns true if the event was already seen, otherwise marks it as seen and returns false.
func (s *IdempotencyStore) MarkSeen(eventID string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.seen[eventID]; exists {
		return true
	}
	s.seen[eventID] = time.Now()
	return false
}

func (s *IdempotencyStore) cleanup() {
	ticker := time.NewTicker(s.ttl / 2)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.mu.Lock()
			now := time.Now()
			for id, seenAt := range s.seen {
				if now.Sub(seenAt) > s.ttl {
					delete(s.seen, id)
				}
			}
			s.mu.Unlock()
		case <-s.stopCh:
			return
		}
	}
}

// Close stops the background cleanup goroutine.
func (s *IdempotencyStore) Close() {
	close(s.stopCh)
}
