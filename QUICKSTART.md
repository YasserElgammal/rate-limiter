# Quick Start Guide

Get started with the Rate Limiter library in 5 minutes!

## Requirements

- Go 1.27 or later

## Installation

```bash
go get github.com/yasserelgammal/rate-limiter
```

## Step 1: Import the Package

```go
import (
    "time"

    "github.com/yasserelgammal/rate-limiter/limiter"
    "github.com/yasserelgammal/rate-limiter/store"
)
```

## Step 2: Configure the Rate Limiter

```go
// Allow 10 requests per second with burst of 20
config := limiter.Config{
    Rate:     10,              // 10 requests
    Duration: time.Second,     // per second
    Burst:    20,              // burst capacity
}
```

## Step 3: Create a Store

```go
// In-memory store with cleanup every 5 minutes
memStore := store.NewMemoryStore(5 * time.Minute)
defer memStore.Close()
```

`MemoryStore` performs rate-limit decisions atomically, including when several
`TokenBucket` instances share it. Custom stores should implement
`store.AtomicStore` when they need the same shared-instance guarantee.

## Step 4: Create the Rate Limiter

```go
rateLimiter, err := limiter.NewTokenBucket(config, memStore)
if err != nil {
    panic(err)
}
```

## Step 5: Use It!

### Simple Check

```go
if rateLimiter.Allow("user123") {
    // Request allowed - process it
    fmt.Println("✅ Request allowed")
} else {
    // Request denied - rate limit exceeded
    fmt.Println("❌ Rate limit exceeded")
}
```

### Detailed Information

```go
result := rateLimiter.AllowN("user123", 1)

if result.Allowed {
    fmt.Printf("✅ Allowed - %d tokens remaining\n", result.Remaining)
} else {
    fmt.Printf("❌ Denied - Retry after %v\n", result.RetryAfter)
}
```

## Complete Example

```go
package main

import (
    "fmt"
    "time"

    "github.com/yasserelgammal/rate-limiter/limiter"
    "github.com/yasserelgammal/rate-limiter/store"
)

func main() {
    // Configure
    config := limiter.Config{
        Rate:     5,
        Duration: time.Second,
        Burst:    10,
    }

    // Create store
    memStore := store.NewMemoryStore(5 * time.Minute)
    defer memStore.Close()

    // Create limiter
    rateLimiter, err := limiter.NewTokenBucket(config, memStore)
    if err != nil {
        panic(err)
    }

    // Use it
    for i := 1; i <= 12; i++ {
        result := rateLimiter.AllowN("user123", 1)

        if result.Allowed {
            fmt.Printf("Request %d: ✅ Allowed (Remaining: %d)\n",
                i, result.Remaining)
        } else {
            fmt.Printf("Request %d: ❌ Denied (Retry after: %v)\n",
                i, result.RetryAfter)
        }
    }
}
```

## HTTP Middleware Example

```go
func rateLimitMiddleware(next http.HandlerFunc) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        clientIP := r.RemoteAddr
        result := rateLimiter.AllowN(clientIP, 1)

        // Set headers
        w.Header().Set("X-RateLimit-Remaining",
            fmt.Sprintf("%d", result.Remaining))

        if !result.Allowed {
            w.WriteHeader(http.StatusTooManyRequests)
            json.NewEncoder(w).Encode(map[string]string{
                "error": "Rate limit exceeded",
            })
            return
        }

        next(w, r)
    }
}

// Use it
http.HandleFunc("/api/data", rateLimitMiddleware(dataHandler))
```

## Common Configurations

### Strict API Protection
```go
config := limiter.Config{
    Rate:     100,           // 100 requests
    Duration: time.Minute,   // per minute
    Burst:    100,           // no extra burst
}
```

### User-Facing API
```go
config := limiter.Config{
    Rate:     10,            // 10 requests
    Duration: time.Second,   // per second
    Burst:    50,            // allow bursts
}
```

### Background Jobs
```go
config := limiter.Config{
    Rate:     1000,          // 1000 requests
    Duration: time.Hour,     // per hour
    Burst:    1000,
}
```

## Running the Examples

```bash
# Basic example
go run ./examples/basic

# HTTP server example
go run ./examples/http_server
# Then visit http://localhost:8080
```

## Testing Your Integration

```bash
# Run tests
go test ./...

# With race detection
go test -race ./...
```

## Next Steps

- Read the [README.md](README.md) for detailed documentation
- Check [PROJECT_STATUS.md](PROJECT_STATUS.md) for the current roadmap
- Review [CHANGELOG.MD](CHANGELOG.MD) for release history
- See [examples/](examples/) for more usage patterns
- Read [CONTRIBUTING.md](CONTRIBUTING.md) to contribute

## Need Help?

- 📖 [Full Documentation](README.md)
- 🐛 [Report Issues](https://github.com/yasserelgammal/rate-limiter/issues)
- 💬 [Discussions](https://github.com/yasserelgammal/rate-limiter/discussions)

---

**You're ready to go! 🚀**
