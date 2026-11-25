package limiter

import (
	"math"
	"sync"
	"time"

	"github.com/yasserelgammal/rate-limiter/store"
)

// TokenBucket implements the Token Bucket algorithm for rate limiting
type TokenBucket struct {
	config Config
	store  store.Store
	mu     sync.Mutex
	now    func() time.Time
}

var _ RateLimiter = (*TokenBucket)(nil)

// NewTokenBucket creates a new TokenBucket rate limiter
func NewTokenBucket(config Config, backingStore store.Store) (*TokenBucket, error) {
	if err := config.Validate(); err != nil {
		return nil, err
	}
	if backingStore == nil {
		return nil, ErrNilStore
	}

	return &TokenBucket{
		config: config,
		store:  backingStore,
		now:    time.Now,
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
			ResetAt:   tb.now(),
			Error:     ErrInvalidTokens,
		}
	}

	if atomicStore, ok := tb.store.(store.AtomicStore); ok {
		return tb.allowNAtomic(atomicStore, key, n)
	}

	// Preserve compatibility with Store implementations that predate AtomicStore.
	// The fallback coordinates calls through this TokenBucket instance only.
	tb.mu.Lock()
	defer tb.mu.Unlock()

	now := tb.now()
	bucket, allowed := tb.updatedBucket(tb.store.Get(key), now, n)
	tb.store.Set(key, bucket)
	return tb.result(bucket, now, n, allowed)
}

// Status returns the current state for key without consuming tokens.
func (tb *TokenBucket) Status(key string) Status {
	if atomicStore, ok := tb.store.(store.AtomicStore); ok {
		return tb.statusAtomic(atomicStore, key)
	}

	tb.mu.Lock()
	defer tb.mu.Unlock()

	now := tb.now()
	bucket := tb.refillBucket(tb.store.Get(key), now)
	tb.store.Set(key, bucket)
	return tb.status(bucket)
}

func (tb *TokenBucket) statusAtomic(atomicStore store.AtomicStore, key string) Status {
	now := tb.now()
	var status Status

	atomicStore.Update(key, func(bucket *store.Bucket) *store.Bucket {
		bucket = tb.refillBucket(bucket, now)
		status = tb.status(bucket)
		return bucket
	})

	return status
}

func (tb *TokenBucket) allowNAtomic(atomicStore store.AtomicStore, key string, n int64) Result {
	now := tb.now()
	var result Result

	atomicStore.Update(key, func(bucket *store.Bucket) *store.Bucket {
		var allowed bool
		bucket, allowed = tb.updatedBucket(bucket, now, n)
		result = tb.result(bucket, now, n, allowed)
		return bucket
	})

	return result
}

func (tb *TokenBucket) updatedBucket(bucket *store.Bucket, now time.Time, n int64) (*store.Bucket, bool) {
	bucket = tb.refillBucket(bucket, now)

	if bucket.Tokens >= n {
		bucket.Tokens -= n
		return bucket, true
	}

	return bucket, false
}

func (tb *TokenBucket) refillBucket(bucket *store.Bucket, now time.Time) *store.Bucket {
	if bucket == nil {
		bucket = &store.Bucket{
			Tokens:       tb.config.Burst,
			LastRefillAt: now,
		}
	}

	elapsed := now.Sub(bucket.LastRefillAt)
	if elapsed <= 0 {
		return bucket
	}
	if bucket.Tokens >= tb.config.Burst {
		// A full bucket cannot bank additional refill credit for later.
		bucket.LastRefillAt = now
		return bucket
	}

	tokensToAdd := int64(elapsed.Seconds() * float64(tb.config.Rate) / tb.config.Duration.Seconds())
	if tokensToAdd > 0 {
		space := tb.config.Burst - bucket.Tokens
		if tokensToAdd >= space {
			bucket.Tokens = tb.config.Burst
			// Once full, elapsed fractional time cannot earn additional credit.
			bucket.LastRefillAt = now
		} else {
			bucket.Tokens += tokensToAdd
			// Advance only by the time represented by whole tokens. Keeping the
			// remainder prevents frequent checks from lowering the effective rate.
			bucket.LastRefillAt = bucket.LastRefillAt.Add(tb.durationForTokens(tokensToAdd))
		}
	}

	return bucket
}

func (tb *TokenBucket) result(bucket *store.Bucket, now time.Time, n int64, allowed bool) Result {
	if allowed {
		return Result{
			Allowed:   true,
			Remaining: bucket.Tokens,
			ResetAt:   tb.calculateResetTime(bucket),
		}
	}

	tokensNeeded := n - bucket.Tokens
	retryAt := bucket.LastRefillAt.Add(tb.durationForTokens(tokensNeeded))
	retryAfter := retryAt.Sub(now)
	if retryAfter < 0 {
		retryAfter = 0
	}

	return Result{
		Allowed:    false,
		Remaining:  bucket.Tokens,
		ResetAt:    tb.calculateResetTime(bucket),
		RetryAfter: retryAfter,
	}
}

func (tb *TokenBucket) status(bucket *store.Bucket) Status {
	return Status{
		Remaining: bucket.Tokens,
		ResetAt:   tb.calculateResetTime(bucket),
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
	timeToFill := tb.durationForTokens(tokensNeeded)

	return bucket.LastRefillAt.Add(timeToFill)
}

// durationForTokens returns the minimum whole duration needed to earn n tokens.
// Rounding up prevents RetryAfter and ResetAt from advertising an early time.
func (tb *TokenBucket) durationForTokens(n int64) time.Duration {
	if n <= 0 {
		return 0
	}
	return time.Duration(math.Ceil(float64(n) * float64(tb.config.Duration) / float64(tb.config.Rate)))
}
