package ingestion

import (
	"sync"
	"testing"
	"time"
)

func TestIdempotencyStore_MarkSeen(t *testing.T) {
	store := NewIdempotencyStore(time.Hour)
	defer store.Close()

	if store.MarkSeen("event-1") {
		t.Error("first call to MarkSeen should return false")
	}

	if !store.MarkSeen("event-1") {
		t.Error("second call to MarkSeen should return true")
	}

	if store.MarkSeen("event-2") {
		t.Error("different event ID should return false")
	}
}

func TestIdempotencyStore_Cleanup(t *testing.T) {
	store := NewIdempotencyStore(50 * time.Millisecond)
	defer store.Close()

	store.MarkSeen("event-1")

	// Should still be seen immediately
	if !store.MarkSeen("event-1") {
		t.Error("event should still be seen before TTL")
	}

	// Wait for TTL + cleanup interval
	time.Sleep(150 * time.Millisecond)

	// Should be cleaned up now
	if store.MarkSeen("event-1") {
		t.Error("event should be cleaned up after TTL")
	}
}

func TestIdempotencyStore_Concurrent(t *testing.T) {
	store := NewIdempotencyStore(time.Hour)
	defer store.Close()

	var wg sync.WaitGroup
	seen := make([]bool, 100)

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			seen[idx] = store.MarkSeen("concurrent-event")
		}(i)
	}

	wg.Wait()

	// Exactly one goroutine should have seen it for the first time
	firstCount := 0
	for _, s := range seen {
		if !s {
			firstCount++
		}
	}
	if firstCount != 1 {
		t.Errorf("expected exactly 1 first-seen, got %d", firstCount)
	}
}
