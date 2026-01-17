package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
	"time"

	"github.com/yasserelgammal/rate-limiter/limiter"
	"github.com/yasserelgammal/rate-limiter/store"
)

var (
	rateLimiter limiter.RateLimiter
	config      limiter.Config
)

// TemplateData holds data passed to HTML templates
type TemplateData struct {
	Rate     int64
	Duration time.Duration
	Burst    int64
}

func main() {
	// Configure rate limiter: 5 requests per 10 seconds, burst of 10
	config = limiter.Config{
		Rate:     5,
		Duration: 10 * time.Second,
		Burst:    10,
	}

	// Create in-memory store with cleanup every 5 minutes
	memStore := store.NewMemoryStore(5 * time.Minute)
	defer memStore.Close()

	// Create token bucket rate limiter
	var err error
	rateLimiter, err = limiter.NewTokenBucket(config, memStore)
	if err != nil {
		log.Fatalf("Failed to create rate limiter: %v", err)
	}

	// Setup HTTP routes
	http.HandleFunc("/api/data", rateLimitMiddleware(dataHandler))
	http.HandleFunc("/api/status", statusHandler)
	http.HandleFunc("/", homeHandler)

	// Start server
	addr := ":8080"
	fmt.Printf("🚀 Server starting on http://localhost%s\n", addr)
	fmt.Println("📊 Rate limit: 5 requests per 10 seconds (burst: 10)")
	fmt.Println("\nEndpoints:")
	fmt.Println("  GET  /              - Home page with instructions")
	fmt.Println("  GET  /api/data      - Rate-limited API endpoint")
	fmt.Println("  GET  /api/status    - Check rate limit status")
	fmt.Println("\nTry: curl http://localhost:8080/api/data")

	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

// rateLimitMiddleware wraps handlers with rate limiting based on IP address
func rateLimitMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Use IP address as the rate limit key
		// In production, you might want to use user ID, API key, etc.
		key := getClientIP(r)

		// Check rate limit with detailed result
		result := rateLimiter.AllowN(key, 1)

		// Set rate limit headers
		w.Header().Set("X-RateLimit-Limit", "5")
		w.Header().Set("X-RateLimit-Remaining", fmt.Sprintf("%d", result.Remaining))
		w.Header().Set("X-RateLimit-Reset", result.ResetAt.Format(time.RFC3339))

		if !result.Allowed {
			// Request denied - rate limit exceeded
			w.Header().Set("Retry-After", fmt.Sprintf("%.0f", result.RetryAfter.Seconds()))
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusTooManyRequests)

			response := map[string]interface{}{
				"error":               "Rate limit exceeded",
				"message":             fmt.Sprintf("Too many requests. Please retry after %v", result.RetryAfter.Round(time.Second)),
				"retry_after_seconds": result.RetryAfter.Seconds(),
				"reset_at":            result.ResetAt.Format(time.RFC3339),
			}

			json.NewEncoder(w).Encode(response)
			log.Printf("❌ Rate limit exceeded for %s - Retry after: %v", key, result.RetryAfter.Round(time.Millisecond))
			return
		}

		// Request allowed
		log.Printf("✅ Request allowed for %s - Remaining: %d", key, result.Remaining)
		next(w, r)
	}
}

// dataHandler handles the main API endpoint
func dataHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	response := map[string]interface{}{
		"message":   "Success! Your request was processed.",
		"timestamp": time.Now().Format(time.RFC3339),
		"data": map[string]string{
			"id":   "12345",
			"name": "Sample Data",
		},
	}

	json.NewEncoder(w).Encode(response)
}

// statusHandler returns the current rate limit status
func statusHandler(w http.ResponseWriter, r *http.Request) {
	key := getClientIP(r)

	// Check status without consuming a token
	result := rateLimiter.AllowN(key, 0)

	w.Header().Set("Content-Type", "application/json")

	response := map[string]interface{}{
		"client_ip": key,
		"rate_limit": map[string]interface{}{
			"limit":     5,
			"remaining": result.Remaining,
			"reset_at":  result.ResetAt.Format(time.RFC3339),
		},
	}

	json.NewEncoder(w).Encode(response)
}

// homeHandler serves a simple HTML page with instructions
func homeHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	// Try to find the template file in a few common locations
	possiblePaths := []string{
		filepath.Join("templates", "index.html"),
		filepath.Join("examples", "http_server", "templates", "index.html"),
		filepath.Join("..", "examples", "http_server", "templates", "index.html"), // In case running from root/limiter
	}

	var tmpl *template.Template
	var err error

	for _, path := range possiblePaths {
		tmpl, err = template.ParseFiles(path)
		if err == nil {
			break
		}
	}

	if err != nil {
		log.Printf("Could not find template in paths: %v", possiblePaths)
		http.Error(w, "Could not load template file. Please ensure running from project root or examples/http_server directory.", http.StatusInternalServerError)
		return
	}

	data := TemplateData{
		Rate:     config.Rate,
		Duration: config.Duration,
		Burst:    config.Burst,
	}

	w.Header().Set("Content-Type", "text/html")
	if err := tmpl.Execute(w, data); err != nil {
		log.Printf("Error executing template: %v", err)
	}
}

// getClientIP extracts the client IP address from the request
func getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header (for proxies/load balancers)
	forwarded := r.Header.Get("X-Forwarded-For")
	if forwarded != "" {
		return forwarded
	}

	// Check X-Real-IP header
	realIP := r.Header.Get("X-Real-IP")
	if realIP != "" {
		return realIP
	}

	// Fall back to RemoteAddr
	return r.RemoteAddr
}
