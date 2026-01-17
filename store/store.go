package store

import "time"

// Bucket represents the state of a token bucket for a specific key
type Bucket struct {
	// Tokens is the current number of tokens available
	Tokens int64
	// LastRefillAt is the last time tokens were added to the bucket
	LastRefillAt time.Time
}

// Store defines the interface for storing rate limiter state
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
