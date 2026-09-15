package store

import (
	"sync"
	"time"
)

// MemoryStore is a thread-safe in-memory implementation of the Store interface
type MemoryStore struct {
	mu      sync.RWMutex
	buckets map[string]*Bucket
	// Optional: cleanup of expired entries
	cleanupInterval time.Duration
	stopCleanup     chan struct{}
}

// NewMemoryStore creates a new in-memory store
// If cleanupInterval is > 0, it will periodically remove stale entries
func NewMemoryStore(cleanupInterval time.Duration) *MemoryStore {
	store := &MemoryStore{
		buckets:         make(map[string]*Bucket),
		cleanupInterval: cleanupInterval,
		stopCleanup:     make(chan struct{}),
	}

	// Start cleanup goroutine if interval is set
	if cleanupInterval > 0 {
		go store.cleanupLoop()
	}

	return store
}

// Get retrieves the bucket for the given key
func (m *MemoryStore) Get(key string) *Bucket {
	m.mu.RLock()
	defer m.mu.RUnlock()

	bucket, exists := m.buckets[key]
	if !exists {
		return nil
	}

	// Return a copy to prevent external modifications
	return &Bucket{
		Tokens:       bucket.Tokens,
		LastRefillAt: bucket.LastRefillAt,
	}
}

// Set stores the bucket for the given key
func (m *MemoryStore) Set(key string, bucket *Bucket) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Store a copy to prevent external modifications
	m.buckets[key] = &Bucket{
		Tokens:       bucket.Tokens,
		LastRefillAt: bucket.LastRefillAt,
	}
}

// Update atomically reads, modifies, and stores a bucket.
func (m *MemoryStore) Update(key string, update func(bucket *Bucket) *Bucket) *Bucket {
	m.mu.Lock()
	defer m.mu.Unlock()

	var current *Bucket
	if bucket, exists := m.buckets[key]; exists {
		current = cloneBucket(bucket)
	}

	updated := update(current)
	if updated == nil {
		delete(m.buckets, key)
		return nil
	}

	m.buckets[key] = cloneBucket(updated)
	return cloneBucket(updated)
}

// Delete removes the bucket for the given key
func (m *MemoryStore) Delete(key string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.buckets, key)
}

// Clear removes all buckets
func (m *MemoryStore) Clear() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.buckets = make(map[string]*Bucket)
}

// Close stops the cleanup goroutine
func (m *MemoryStore) Close() {
	if m.cleanupInterval > 0 {
		close(m.stopCleanup)
	}
}

// cleanupLoop periodically removes stale entries
func (m *MemoryStore) cleanupLoop() {
	ticker := time.NewTicker(m.cleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			m.cleanup()
		case <-m.stopCleanup:
			return
		}
	}
}

// cleanup removes entries that haven't been accessed recently
// This helps prevent memory leaks in long-running applications
func (m *MemoryStore) cleanup() {
	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now()
	staleThreshold := 1 * time.Hour // Remove entries older than 1 hour

	for key, bucket := range m.buckets {
		if now.Sub(bucket.LastRefillAt) > staleThreshold {
			delete(m.buckets, key)
		}
	}
}

// Size returns the number of entries in the store
func (m *MemoryStore) Size() int {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return len(m.buckets)
}

func cloneBucket(bucket *Bucket) *Bucket {
	return &Bucket{
		Tokens:       bucket.Tokens,
		LastRefillAt: bucket.LastRefillAt,
	}
}
