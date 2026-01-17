package limiter

import (
	"time"

	"github.com/yasserelgammal/rate-limiter/store"
)

// TokenBucket implements the Token Bucket algorithm for rate limiting
type TokenBucket struct {
	config Config
	store  store.Store
}

// NewTokenBucket creates a new TokenBucket rate limiter
func NewTokenBucket(config Config, store store.Store) (*TokenBucket, error) {
	if err := config.Validate(); err != nil {
		return nil, err
	}

	return &TokenBucket{
		config: config,
		store:  store,
	}, nil
}

// Allow checks if a request for the given key should be allowed
func (tb *TokenBucket) Allow(key string) bool {
	result := tb.AllowN(key, 1)
	return result.Allowed
}

// AllowN checks if n requests for the given key should be allowed
func (tb *TokenBucket) AllowN(key string, n int64) Result {
	if n <= 0 {
		return Result{
			Allowed:   false,
			Remaining: 0,
			ResetAt:   time.Now(),
		}
	}

	now := time.Now()

	// Get or create bucket state
	bucket := tb.store.Get(key)
	if bucket == nil {
		// Initialize new bucket with full capacity
		bucket = &store.Bucket{
			Tokens:       tb.config.Burst,
			LastRefillAt: now,
		}
	}

	// Calculate tokens to add based on time elapsed
	elapsed := now.Sub(bucket.LastRefillAt)
	tokensToAdd := int64(elapsed.Seconds() * float64(tb.config.Rate) / tb.config.Duration.Seconds())

	if tokensToAdd > 0 {
		bucket.Tokens += tokensToAdd
		if bucket.Tokens > tb.config.Burst {
			bucket.Tokens = tb.config.Burst
		}
		bucket.LastRefillAt = now
	}

	// Check if we have enough tokens
	if bucket.Tokens >= n {
		bucket.Tokens -= n
		tb.store.Set(key, bucket)

		return Result{
			Allowed:   true,
			Remaining: bucket.Tokens,
			ResetAt:   tb.calculateResetTime(bucket),
		}
	}

	// Not enough tokens
	tb.store.Set(key, bucket)

	// Calculate retry after duration
	tokensNeeded := n - bucket.Tokens
	retryAfter := time.Duration(float64(tokensNeeded) * tb.config.Duration.Seconds() / float64(tb.config.Rate) * float64(time.Second))

	return Result{
		Allowed:    false,
		Remaining:  bucket.Tokens,
		ResetAt:    tb.calculateResetTime(bucket),
		RetryAfter: retryAfter,
	}
}

// Reset clears the rate limit state for the given key
func (tb *TokenBucket) Reset(key string) {
	tb.store.Delete(key)
}

// ResetAll clears all rate limit states
func (tb *TokenBucket) ResetAll() {
	tb.store.Clear()
}

// calculateResetTime calculates when the bucket will be fully refilled
func (tb *TokenBucket) calculateResetTime(bucket *store.Bucket) time.Time {
	if bucket.Tokens >= tb.config.Burst {
		return bucket.LastRefillAt
	}

	tokensNeeded := tb.config.Burst - bucket.Tokens
	timeToFill := time.Duration(float64(tokensNeeded) * tb.config.Duration.Seconds() / float64(tb.config.Rate) * float64(time.Second))

	return bucket.LastRefillAt.Add(timeToFill)
}
