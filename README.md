# Rate Limiter

[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)](https://go.dev/)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/yasserelgammal/rate-limiter)](https://goreportcard.com/report/github.com/yasserelgammal/rate-limiter)

A production-ready, thread-safe rate limiter library for Go applications. Control the number of requests per key (user/IP/token) in a given time frame using the Token Bucket algorithm.

## 🎯 Problem Being Solved

In modern applications, you often need to:
- **Prevent abuse** by limiting API requests per user/IP
- **Protect resources** from being overwhelmed by too many requests
- **Ensure fair usage** across multiple clients
- **Comply with rate limits** when calling external APIs

This library provides a simple, efficient, and thread-safe solution for implementing rate limiting in your Go applications.

## ✨ Features

- 🚀 **Production-ready** - Thread-safe with comprehensive test coverage
- 🎯 **Simple API** - Easy to integrate with just a few lines of code
- 🔒 **Thread-safe** - Uses `sync.RWMutex` for concurrent access
- 💾 **In-memory storage** - Fast, zero-dependency storage (Redis support planned)
- 🪣 **Token Bucket algorithm** - Smooth rate limiting with burst support
- 📊 **Detailed results** - Get remaining tokens, reset time, and retry-after duration
- 🧹 **Automatic cleanup** - Optional cleanup of stale entries to prevent memory leaks
- 🧪 **Well-tested** - Extensive unit tests including concurrency tests

## 📦 Installation

```bash
go get github.com/yasserelgammal/rate-limiter
```

## 🚀 Quick Start

### Basic Usage

```go
package main

import (
    "fmt"
    "time"

    "github.com/yasserelgammal/rate-limiter/limiter"
    "github.com/yasserelgammal/rate-limiter/store"
)

func main() {
    // Configure: 10 requests per second with burst of 20
    config := limiter.Config{
        Rate:     10,
        Duration: time.Second,
        Burst:    20,
    }

    // Create in-memory store
    memStore := store.NewMemoryStore(5 * time.Minute)
    defer memStore.Close()

    // Create rate limiter
    rateLimiter, err := limiter.NewTokenBucket(config, memStore)
    if err != nil {
        panic(err)
    }

    // Check if request is allowed
    userID := "user123"
    if rateLimiter.Allow(userID) {
        fmt.Println("✅ Request allowed")
    } else {
        fmt.Println("❌ Rate limit exceeded")
    }
}
```

### Advanced Usage with Detailed Results

```go
// Get detailed rate limit information
result := rateLimiter.AllowN("user123", 1)

if result.Allowed {
    fmt.Printf("✅ Request allowed. Remaining: %d\n", result.Remaining)
} else {
    fmt.Printf("❌ Rate limited. Retry after: %v\n", result.RetryAfter)
    fmt.Printf("   Reset at: %v\n", result.ResetAt)
}
```

### HTTP Middleware Example

```go
func rateLimitMiddleware(next http.HandlerFunc) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        // Use IP address as rate limit key
        clientIP := r.RemoteAddr
        
        result := rateLimiter.AllowN(clientIP, 1)
        
        // Set rate limit headers
        w.Header().Set("X-RateLimit-Remaining", fmt.Sprintf("%d", result.Remaining))
        w.Header().Set("X-RateLimit-Reset", result.ResetAt.Format(time.RFC3339))
        
        if !result.Allowed {
            w.Header().Set("Retry-After", fmt.Sprintf("%.0f", result.RetryAfter.Seconds()))
            w.WriteHeader(http.StatusTooManyRequests)
            json.NewEncoder(w).Encode(map[string]string{
                "error": "Rate limit exceeded",
                "retry_after": result.RetryAfter.String(),
            })
            return
        }
        
        next(w, r)
    }
}
```

## 📚 API Reference

### Configuration

```go
type Config struct {
    Rate     int64         // Number of requests allowed per Duration
    Duration time.Duration // Time window for the rate limit
    Burst    int64         // Maximum burst capacity
}
```

**Example configurations:**
- `Rate: 100, Duration: time.Minute, Burst: 100` - 100 requests per minute
- `Rate: 10, Duration: time.Second, Burst: 20` - 10 req/s with burst of 20
- `Rate: 1000, Duration: time.Hour, Burst: 1000` - 1000 requests per hour

### RateLimiter Interface

```go
type RateLimiter interface {
    // Allow checks if a single request should be allowed
    Allow(key string) bool
    
    // AllowN checks if n requests should be allowed
    AllowN(key string, n int64) Result
    
    // Reset clears rate limit for a specific key
    Reset(key string)
    
    // ResetAll clears all rate limits
    ResetAll()
}
```

### Result Structure

```go
type Result struct {
    Allowed    bool          // Whether the request is allowed
    Remaining  int64         // Tokens remaining in current window
    ResetAt    time.Time     // When the rate limit resets
    RetryAfter time.Duration // How long to wait before retrying (if denied)
}
```

## 🏗️ Project Structure

```
rate-limiter/
├── limiter/                      # Core rate limiting logic
│   ├── limiter.go                # Interface and configuration
│   ├── token_bucket.go           # Token Bucket implementation
│   ├── token_bucket_test.go      # Comprehensive tests
│   └── errors.go                 # Error definitions
│
├── store/                        # Storage implementations
│   ├── store.go                  # Storage interface
│   ├── memory.go                 # Thread-safe in-memory store
│   └── memory_test.go            # Store tests
│
├── examples/                     # Usage examples
│   ├── basic/                    # Basic usage example
│   │   └── main.go
│   └── http_server/              # HTTP server with middleware
│       ├── templates/            # HTML templates
│       │   └── index.html
│       └── main.go
│
├── go.mod                        # Go module definition
├── Makefile                      # Development tasks
├── README.md                     # Main documentation
├── LICENSE                       # MIT License
├── CONTRIBUTING.md               # Contribution guidelines
└── .gitignore                    # Git ignore rules
```

## 🧪 Running Tests

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run tests with race detection
go test -race ./...

# Run benchmarks
go test -bench=. ./...
```

### Test Coverage

The library includes comprehensive tests covering:
- ✅ Basic allow/deny functionality
- ✅ Token refill behavior
- ✅ Multiple concurrent keys
- ✅ Thread-safety with race detection
- ✅ Edge cases and error handling
- ✅ Benchmark tests for performance

## 🎮 Running the Example

```bash
# Navigate to the example directory
cd examples/http_server

# Run the server
go run main.go
```

Then open your browser to `http://localhost:8080` for an interactive demo, or use curl:

```bash
# Make a request
curl http://localhost:8080/api/data

# Check rate limit status
curl http://localhost:8080/api/status

# Test rate limiting
for i in {1..15}; do
  curl -i http://localhost:8080/api/data
  echo ""
done
```

## 🔧 Configuration Examples

### Strict Rate Limiting (API Protection)
```go
config := limiter.Config{
    Rate:     100,              // 100 requests
    Duration: time.Minute,      // per minute
    Burst:    100,              // no burst allowance
}
```

### Lenient with Burst (User-facing API)
```go
config := limiter.Config{
    Rate:     10,               // 10 requests
    Duration: time.Second,      // per second
    Burst:    50,               // allow bursts up to 50
}
```

### Hourly Limit (Background Jobs)
```go
config := limiter.Config{
    Rate:     1000,             // 1000 requests
    Duration: time.Hour,        // per hour
    Burst:    1000,             // full capacity available
}
```

## 🤝 Contributing

Contributions are welcome! Please feel free to submit a Pull Request. For major changes, please open an issue first to discuss what you would like to change.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 📞 Support & Contact

If you need help, have a question, or want to suggest an improvement:

- 🐛 **Issues**: Report bugs or request features  
  https://github.com/yasserelgammal/rate-limiter/issues

- 💬 **Discussions**: Ask questions or share ideas  
  https://github.com/yasserelgammal/rate-limiter/discussions

- 💼 **LinkedIn**: Connect with me  
  https://www.linkedin.com/in/elgammal


---

**Made with ❤️ by [Yasser Elgammal](https://github.com/yasserelgammal)**

If you find this project useful, please consider giving it a ⭐️ on GitHub!
