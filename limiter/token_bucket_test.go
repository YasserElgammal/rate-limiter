package limiter

import (
	"sync"
	"testing"
	"time"

	"github.com/yasserelgammal/rate-limiter/store"
)

func TestTokenBucket_Allow(t *testing.T) {
	config := Config{
		Rate:     10,
		Duration: time.Second,
		Burst:    10,
	}

	s := store.NewMemoryStore(0)
	defer s.Close()

	tb, err := NewTokenBucket(config, s)
	if err != nil {
		t.Fatalf("Failed to create token bucket: %v", err)
	}

	// First 10 requests should be allowed (burst capacity)
	for i := 0; i < 10; i++ {
		if !tb.Allow("test-key") {
			t.Errorf("Request %d should be allowed", i+1)
		}
	}

	// 11th request should be denied (bucket empty)
	if tb.Allow("test-key") {
		t.Error("Request 11 should be denied")
	}
}

func TestTokenBucket_AllowN(t *testing.T) {
	config := Config{
		Rate:     10,
		Duration: time.Second,
		Burst:    10,
	}

	s := store.NewMemoryStore(0)
	defer s.Close()

	tb, err := NewTokenBucket(config, s)
	if err != nil {
		t.Fatalf("Failed to create token bucket: %v", err)
	}

	// Request 5 tokens
	result := tb.AllowN("test-key", 5)
	if !result.Allowed {
		t.Error("Request for 5 tokens should be allowed")
	}
	if result.Remaining != 5 {
		t.Errorf("Expected 5 remaining tokens, got %d", result.Remaining)
	}

	// Request 6 more tokens (should fail, only 5 remaining)
	result = tb.AllowN("test-key", 6)
	if result.Allowed {
		t.Error("Request for 6 tokens should be denied")
	}
	if result.Remaining != 5 {
		t.Errorf("Expected 5 remaining tokens, got %d", result.Remaining)
	}
}

func TestTokenBucket_StatusDoesNotConsumeTokens(t *testing.T) {
	config := Config{Rate: 1, Duration: time.Hour, Burst: 10}
	s := store.NewMemoryStore(0)
	defer s.Close()

	tb, err := NewTokenBucket(config, s)
	if err != nil {
		t.Fatalf("create token bucket: %v", err)
	}

	if result := tb.AllowN("test-key", 4); !result.Allowed {
		t.Fatal("expected initial request to be allowed")
	}

	first := tb.Status("test-key")
	second := tb.Status("test-key")
	if first.Remaining != 6 || second.Remaining != 6 {
		t.Fatalf("status consumed tokens: first=%d second=%d", first.Remaining, second.Remaining)
	}

	if result := tb.Allow("test-key"); !result {
		t.Fatal("expected token to remain available after status checks")
	}
	if remaining := tb.Status("test-key").Remaining; remaining != 5 {
		t.Fatalf("expected 5 remaining tokens, got %d", remaining)
	}
}

func TestTokenBucket_Refill(t *testing.T) {
	config := Config{
		Rate:     10, // 10 tokens per second
		Duration: time.Second,
		Burst:    10,
	}

	s := store.NewMemoryStore(0)
	defer s.Close()

	tb, err := NewTokenBucket(config, s)
	if err != nil {
		t.Fatalf("Failed to create token bucket: %v", err)
	}

	// Consume all tokens
	for i := 0; i < 10; i++ {
		tb.Allow("test-key")
	}

	// Should be denied
	if tb.Allow("test-key") {
		t.Error("Request should be denied (bucket empty)")
	}

	// Wait for refill (100ms = 1 token at 10 tokens/second)
	time.Sleep(150 * time.Millisecond)

	// Should be allowed now
	if !tb.Allow("test-key") {
		t.Error("Request should be allowed after refill")
	}
}

func TestTokenBucket_Reset(t *testing.T) {
	config := Config{
		Rate:     10,
		Duration: time.Second,
		Burst:    10,
	}

	s := store.NewMemoryStore(0)
	defer s.Close()

	tb, err := NewTokenBucket(config, s)
	if err != nil {
		t.Fatalf("Failed to create token bucket: %v", err)
	}

	// Consume all tokens
	for i := 0; i < 10; i++ {
		tb.Allow("test-key")
	}

	// Should be denied
	if tb.Allow("test-key") {
		t.Error("Request should be denied")
	}

	// Reset the key
	tb.Reset("test-key")

	// Should be allowed now (bucket reset to full capacity)
	if !tb.Allow("test-key") {
		t.Error("Request should be allowed after reset")
	}
}

func TestTokenBucket_MultipleKeys(t *testing.T) {
	config := Config{
		Rate:     5,
		Duration: time.Second,
		Burst:    5,
	}

	s := store.NewMemoryStore(0)
	defer s.Close()

	tb, err := NewTokenBucket(config, s)
	if err != nil {
		t.Fatalf("Failed to create token bucket: %v", err)
	}

	// Consume all tokens for key1
	for i := 0; i < 5; i++ {
		tb.Allow("key1")
	}

	// key1 should be denied
	if tb.Allow("key1") {
		t.Error("key1 should be denied")
	}

	// key2 should still be allowed (different bucket)
	if !tb.Allow("key2") {
		t.Error("key2 should be allowed")
	}
}

func TestTokenBucket_Concurrency(t *testing.T) {
	config := Config{
		Rate:     100,
		Duration: time.Second,
		Burst:    100,
	}

	s := store.NewMemoryStore(0)
	defer s.Close()

	tb, err := NewTokenBucket(config, s)
	if err != nil {
		t.Fatalf("Failed to create token bucket: %v", err)
	}

	const numGoroutines = 50
	const requestsPerGoroutine = 10

	var wg sync.WaitGroup
	allowedCount := make(chan int, numGoroutines)

	// Launch concurrent requests
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			allowed := 0
			for j := 0; j < requestsPerGoroutine; j++ {
				if tb.Allow("concurrent-key") {
					allowed++
				}
			}
			allowedCount <- allowed
		}(i)
	}

	wg.Wait()
	close(allowedCount)

	// Count total allowed requests
	total := 0
	for count := range allowedCount {
		total += count
	}

	// Should allow exactly burst capacity (100)
	if total != 100 {
		t.Errorf("Expected exactly 100 allowed requests, got %d", total)
	}
}

func TestTokenBucket_ConcurrencyAcrossInstances(t *testing.T) {
	config := Config{
		Rate:     1,
		Duration: time.Hour,
		Burst:    100,
	}

	s := store.NewMemoryStore(0)
	defer s.Close()

	first, err := NewTokenBucket(config, s)
	if err != nil {
		t.Fatalf("create first token bucket: %v", err)
	}
	second, err := NewTokenBucket(config, s)
	if err != nil {
		t.Fatalf("create second token bucket: %v", err)
	}

	const workers = 500
	start := make(chan struct{})
	allowed := make(chan bool, workers)
	var wg sync.WaitGroup

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func(limiter *TokenBucket) {
			defer wg.Done()
			<-start
			allowed <- limiter.Allow("shared-key")
		}([]*TokenBucket{first, second}[i%2])
	}

	close(start)
	wg.Wait()
	close(allowed)

	var total int64
	for result := range allowed {
		if result {
			total++
		}
	}

	if total != config.Burst {
		t.Fatalf("expected exactly %d allowed requests, got %d", config.Burst, total)
	}
}

func TestTokenBucket_ConcurrentKeys(t *testing.T) {
	config := Config{
		Rate:     10,
		Duration: time.Second,
		Burst:    10,
	}

	s := store.NewMemoryStore(0)
	defer s.Close()

	tb, err := NewTokenBucket(config, s)
	if err != nil {
		t.Fatalf("Failed to create token bucket: %v", err)
	}

	const numKeys = 100
	const requestsPerKey = 5

	var wg sync.WaitGroup

	// Launch concurrent requests for different keys
	for i := 0; i < numKeys; i++ {
		wg.Add(1)
		go func(keyID int) {
			defer wg.Done()
			key := string(rune('A'+(keyID%26))) + string(rune('0'+(keyID/26)))
			for j := 0; j < requestsPerKey; j++ {
				tb.Allow(key)
			}
		}(i)
	}

	wg.Wait()

	// Verify store has correct number of keys
	if s.Size() != numKeys {
		t.Errorf("Expected %d keys in store, got %d", numKeys, s.Size())
	}
}

func TestTokenBucket_InvalidConfig(t *testing.T) {
	s := store.NewMemoryStore(0)
	defer s.Close()

	tests := []struct {
		name   string
		config Config
		errMsg error
	}{
		{
			name: "invalid rate",
			config: Config{
				Rate:     0,
				Duration: time.Second,
				Burst:    10,
			},
			errMsg: ErrInvalidRate,
		},
		{
			name: "invalid duration",
			config: Config{
				Rate:     10,
				Duration: 0,
				Burst:    10,
			},
			errMsg: ErrInvalidDuration,
		},
		{
			name: "invalid burst",
			config: Config{
				Rate:     10,
				Duration: time.Second,
				Burst:    0,
			},
			errMsg: ErrInvalidBurst,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewTokenBucket(tt.config, s)
			if err != tt.errMsg {
				t.Errorf("Expected error %v, got %v", tt.errMsg, err)
			}
		})
	}
}

func TestTokenBucket_RetryAfter(t *testing.T) {
	config := Config{
		Rate:     10,
		Duration: time.Second,
		Burst:    10,
	}

	s := store.NewMemoryStore(0)
	defer s.Close()

	tb, err := NewTokenBucket(config, s)
	if err != nil {
		t.Fatalf("Failed to create token bucket: %v", err)
	}

	// Consume all tokens
	tb.AllowN("test-key", 10)

	// Try to get 1 more token
	result := tb.AllowN("test-key", 1)
	if result.Allowed {
		t.Error("Request should be denied")
	}

	// RetryAfter should be approximately 100ms (1 token at 10 tokens/second)
	expectedRetryAfter := 100 * time.Millisecond
	tolerance := 10 * time.Millisecond

	if result.RetryAfter < expectedRetryAfter-tolerance || result.RetryAfter > expectedRetryAfter+tolerance {
		t.Errorf("Expected RetryAfter around %v, got %v", expectedRetryAfter, result.RetryAfter)
	}
}

func BenchmarkTokenBucket_Allow(b *testing.B) {
	config := Config{
		Rate:     1000,
		Duration: time.Second,
		Burst:    1000,
	}

	s := store.NewMemoryStore(0)
	defer s.Close()

	tb, _ := NewTokenBucket(config, s)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tb.Allow("bench-key")
	}
}

func BenchmarkTokenBucket_AllowParallel(b *testing.B) {
	config := Config{
		Rate:     10000,
		Duration: time.Second,
		Burst:    10000,
	}

	s := store.NewMemoryStore(0)
	defer s.Close()

	tb, _ := NewTokenBucket(config, s)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			tb.Allow("bench-key")
		}
	})
}
