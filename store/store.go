package store

import "time"

// Bucket represents the state of a token bucket for a specific key
type Bucket struct {
	// Tokens is the current number of tokens available
	Tokens int64
	// LastRefillAt is the last time tokens were added to the bucket
	LastRefillAt time.Time
}

// Store defines the basic interface for storing rate limiter state.
// Implement AtomicStore when instances may share this store concurrently.
type Store interface {
	// Get retrieves the bucket for the given key
	// Returns nil if the key doesn't exist
	Get(key string) *Bucket

	// Set stores the bucket for the given key
	Set(key string, bucket *Bucket)

	// Delete removes the bucket for the given key
	Delete(key string)

	// Clear removes all buckets
	Clear()
}

// AtomicStore extends Store with an atomic read-modify-write operation.
// The update function runs while the key is locked and must not call back into
// the same store. Returning nil deletes the key.
type AtomicStore interface {
	Store

	Update(key string, update func(bucket *Bucket) *Bucket) *Bucket
}
