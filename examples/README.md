# Examples

This directory contains practical examples demonstrating how to use the rate limiter library.

Requires Go 1.27 or later. Run the commands below from the repository root.

## Available Examples

### 1. Basic Example (`basic/`)

A simple command-line example demonstrating core features:

- ✅ Simple `Allow()` checks
- ✅ Detailed `AllowN()` with result information
- ✅ Multiple users with independent limits
- ✅ Burst request handling
- ✅ Token refill over time

**Run it:**
```bash
go run ./examples/basic
```

**What you'll see:**
- 12 sequential requests showing allow/deny
- Detailed rate limit information
- Multiple users being rate limited independently
- Burst capacity demonstration
- Real-time token refill

### 2. HTTP Server Example (`http_server/`)

An HTTP server demonstrating rate limiting middleware:

**Features:**
- 🌐 Interactive web UI at `http://localhost:8080`
- 📊 Rate limiting by IP address
- 🎯 Common rate-limit headers (`X-RateLimit-*`)
- 📝 JSON error responses
- ✅ Status endpoint to check limits
- 🎨 Beautiful, responsive interface

**Endpoints:**
- `GET /` - Interactive demo page
- `GET /api/data` - Rate-limited API endpoint
- `GET /api/status` - Check rate limit status

**Run it:**
```bash
go run ./examples/http_server
```

Then visit: `http://localhost:8080`

**Test with curl:**
```bash
# Make a single request
curl http://localhost:8080/api/data

# Check status
curl http://localhost:8080/api/status

# Test rate limiting (15 rapid requests)
for i in {1..15}; do
  echo "Request $i:"
  curl -i http://localhost:8080/api/data
  echo ""
done
```

## Configuration Examples

### Strict API Protection
```go
config := limiter.Config{
    Rate:     100,              // 100 requests
    Duration: time.Minute,      // per minute
    Burst:    100,              // no extra burst
}
```

### Lenient User-Facing API
```go
config := limiter.Config{
    Rate:     10,               // 10 requests
    Duration: time.Second,      // per second
    Burst:    50,               // allow bursts up to 50
}
```

### Background Job Limiting
```go
config := limiter.Config{
    Rate:     1000,             // 1000 requests
    Duration: time.Hour,        // per hour
    Burst:    1000,
}
```

## Common Use Cases

### 1. Rate Limit by User ID
```go
userID := getUserID(r)
if !rateLimiter.Allow(userID) {
    return errors.New("rate limit exceeded")
}
```

### 2. Rate Limit by IP Address
```go
host, _, err := net.SplitHostPort(r.RemoteAddr)
if err != nil {
    http.Error(w, "invalid remote address", http.StatusBadRequest)
    return
}

clientIP := host
result := rateLimiter.AllowN(clientIP, 1)
if !result.Allowed {
    http.Error(w, "Too Many Requests", 429)
    return
}
```

Do not trust `X-Forwarded-For` or `X-Real-IP` directly. When deploying behind
a reverse proxy, accept forwarded headers only when the connection peer is a
configured trusted proxy, then parse the expected forwarded-header format.

### 3. Rate Limit by API Token
```go
apiToken := r.Header.Get("Authorization")
if !rateLimiter.Allow(apiToken) {
    return errors.New("API rate limit exceeded")
}
```

### 4. Multiple Rate Limits
```go
// Per-user limit
if !userLimiter.Allow(userID) {
    return errors.New("user rate limit exceeded")
}

// Global limit
if !globalLimiter.Allow("global") {
    return errors.New("global rate limit exceeded")
}
```

## HTTP Middleware Pattern

```go
func rateLimitMiddleware(next http.HandlerFunc) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        // Extract key (IP, user ID, token, etc.)
        key := getClientIP(r)

        // Check rate limit
        result := rateLimiter.AllowN(key, 1)

        // Set standard headers
        w.Header().Set("X-RateLimit-Limit", "10")
        w.Header().Set("X-RateLimit-Remaining",
            fmt.Sprintf("%d", result.Remaining))
        w.Header().Set("X-RateLimit-Reset",
            result.ResetAt.Format(time.RFC3339))

        if !result.Allowed {
            // Rate limit exceeded
            w.Header().Set("Retry-After",
                fmt.Sprintf("%.0f", result.RetryAfter.Seconds()))
            w.WriteHeader(http.StatusTooManyRequests)

            json.NewEncoder(w).Encode(map[string]interface{}{
                "error": "Rate limit exceeded",
                "retry_after": result.RetryAfter.String(),
                "reset_at": result.ResetAt.Format(time.RFC3339),
            })
            return
        }

        // Request allowed
        next(w, r)
    }
}

// Usage
http.HandleFunc("/api/data", rateLimitMiddleware(dataHandler))
```

## Testing Your Integration

```bash
# Run the example
go run ./examples/http_server

# In another terminal, test it
curl http://localhost:8080/api/data

# Load test with Apache Bench
ab -n 100 -c 10 http://localhost:8080/api/data

# Or with hey
hey -n 100 -c 10 http://localhost:8080/api/data
```

## Next Steps

1. **Customize the configuration** for your use case
2. **Integrate the middleware** into your application
3. **Add monitoring** to track rate limit hits
4. **Consider distributed limiting** with Redis (coming soon)

`MemoryStore` coordinates concurrent decisions across limiter instances through
`store.AtomicStore`. A custom shared store should implement that optional
interface to provide the same guarantee.

## Need Help?

- 📖 [Main Documentation](../README.md)
- 🚀 [Quick Start Guide](../QUICKSTART.md)
- 🐛 [Report Issues](https://github.com/yasserelgammal/rate-limiter/issues)

---

**Happy rate limiting! 🚀**
