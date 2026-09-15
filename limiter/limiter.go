package limiter

import "time"

// Result contains the result of a rate limit check
type Result struct {
	// Allowed indicates whether the request should be allowed
	Allowed bool
	// Remaining is the number of requests remaining in the current window
	Remaining int64
	// ResetAt is the time when the rate limit will reset
	ResetAt time.Time
	// RetryAfter is the duration to wait before retrying (only set when Allowed is false)
	RetryAfter time.Duration
}

// Status describes the current state of a rate limit without consuming tokens.
type Status struct {
	// Remaining is the number of tokens currently available.
	Remaining int64
	// ResetAt is the time when the bucket will be fully refilled.
	ResetAt time.Time
}

// RateLimiter defines the interface for rate limiting implementations
type RateLimiter interface {
	// Allow checks if a request for the given key should be allowed
	// Returns true if the request is allowed, false otherwise
	Allow(key string) bool

	// AllowN checks if n requests for the given key should be allowed
	// Returns a Result with detailed information about the rate limit status
	AllowN(key string, n int64) Result

	// Reset clears the rate limit state for the given key
	Reset(key string)

	// ResetAll clears all rate limit states
	ResetAll()
}

// Config holds the configuration for a rate limiter
type Config struct {
	// Rate is the number of requests allowed per Duration
	Rate int64
	// Duration is the time window for the rate limit
	Duration time.Duration
	// Burst is the maximum number of requests that can be made in a single burst
	// For Token Bucket, this is the bucket capacity
	Burst int64
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	if c.Rate <= 0 {
		return ErrInvalidRate
	}
	if c.Duration <= 0 {
		return ErrInvalidDuration
	}
	if c.Burst <= 0 {
		return ErrInvalidBurst
	}
	return nil
}
