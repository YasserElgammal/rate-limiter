package main

import (
	"fmt"
	"time"

	"github.com/yasserelgammal/rate-limiter/limiter"
	"github.com/yasserelgammal/rate-limiter/store"
)

func main() {
	fmt.Println("🚀 Rate Limiter - Basic Example")
	fmt.Println("================================")

	// Configure rate limiter: 5 requests per second with burst of 10
	config := limiter.Config{
		Rate:     5,
		Duration: time.Second,
		Burst:    10,
	}

	// Create in-memory store with cleanup every 5 minutes
	memStore := store.NewMemoryStore(5 * time.Minute)
	defer memStore.Close()

	// Create token bucket rate limiter
	rateLimiter, err := limiter.NewTokenBucket(config, memStore)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Configuration: %d requests per %v (burst: %d)\n\n",
		config.Rate, config.Duration, config.Burst)

	// Example 1: Simple Allow check
	fmt.Println("Example 1: Simple Allow Check")
	fmt.Println("------------------------------")
	userID := "user123"

	for i := 1; i <= 12; i++ {
		if rateLimiter.Allow(userID) {
			fmt.Printf("Request %2d: ✅ Allowed\n", i)
		} else {
			fmt.Printf("Request %2d: ❌ Denied (rate limit exceeded)\n", i)
		}
	}

	// Reset for next example
	rateLimiter.Reset(userID)
	fmt.Println()

	// Example 2: Detailed AllowN with result information
	fmt.Println("Example 2: Detailed Result Information")
	fmt.Println("---------------------------------------")

	for i := 1; i <= 12; i++ {
		result := rateLimiter.AllowN(userID, 1)

		if result.Allowed {
			fmt.Printf("Request %2d: ✅ Allowed - Remaining: %d tokens\n",
				i, result.Remaining)
		} else {
			fmt.Printf("Request %2d: ❌ Denied - Retry after: %v\n",
				i, result.RetryAfter.Round(time.Millisecond))
		}
	}

	// Reset for next example
	rateLimiter.Reset(userID)
	fmt.Println()

	// Example 3: Multiple users
	fmt.Println("Example 3: Multiple Users (Independent Limits)")
	fmt.Println("----------------------------------------------")

	users := []string{"alice", "bob", "charlie"}

	for _, user := range users {
		// Each user makes 3 requests
		for i := 1; i <= 3; i++ {
			result := rateLimiter.AllowN(user, 1)
			status := "✅"
			if !result.Allowed {
				status = "❌"
			}
			fmt.Printf("User %-8s Request %d: %s (Remaining: %d)\n",
				user, i, status, result.Remaining)
		}
	}

	fmt.Println()

	// Example 4: Burst requests
	fmt.Println("Example 4: Burst Requests")
	fmt.Println("-------------------------")

	rateLimiter.Reset("burst-user")

	// Make 10 requests instantly (should all succeed due to burst)
	fmt.Println("Making 10 instant requests (burst capacity):")
	for i := 1; i <= 10; i++ {
		result := rateLimiter.AllowN("burst-user", 1)
		status := "✅"
		if !result.Allowed {
			status = "❌"
		}
		fmt.Printf("  Request %2d: %s (Remaining: %d)\n", i, status, result.Remaining)
	}

	// 11th request should fail
	result := rateLimiter.AllowN("burst-user", 1)
	fmt.Printf("  Request 11: ❌ Denied - Retry after: %v\n",
		result.RetryAfter.Round(time.Millisecond))

	fmt.Println()

	// Example 5: Token refill demonstration
	fmt.Println("Example 5: Token Refill Over Time")
	fmt.Println("----------------------------------")

	rateLimiter.Reset("refill-user")

	// Consume all tokens
	for i := 0; i < 10; i++ {
		rateLimiter.Allow("refill-user")
	}

	fmt.Println("All tokens consumed. Waiting for refill...")

	// Check every 200ms for 1 second
	for i := 0; i < 5; i++ {
		time.Sleep(200 * time.Millisecond)
		result := rateLimiter.AllowN("refill-user", 1)

		if result.Allowed {
			fmt.Printf("After %dms: ✅ Token available (Remaining: %d)\n",
				(i+1)*200, result.Remaining)
		} else {
			fmt.Printf("After %dms: ❌ Still rate limited (Retry after: %v)\n",
				(i+1)*200, result.RetryAfter.Round(time.Millisecond))
		}
	}

	fmt.Println("\n✨ Examples completed!")
}
