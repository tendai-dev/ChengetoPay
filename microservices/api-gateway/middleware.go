package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"
)

// Middleware represents HTTP middleware
type Middleware func(http.Handler) http.Handler

// RequestContext holds request-scoped data
type RequestContext struct {
	RequestID string
	UserID    string
	StartTime time.Time
}

// RateLimiter implements token bucket rate limiting
type RateLimiter struct {
	mu       sync.RWMutex
	buckets  map[string]*TokenBucket
	capacity int
	refill   time.Duration
}

// TokenBucket represents a token bucket for rate limiting
type TokenBucket struct {
	tokens    int
	capacity  int
	lastRefill time.Time
	mu        sync.Mutex
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(capacity int, refillInterval time.Duration) *RateLimiter {
	return &RateLimiter{
		buckets:  make(map[string]*TokenBucket),
		capacity: capacity,
		refill:   refillInterval,
	}
}

// Allow checks if a request should be allowed
func (rl *RateLimiter) Allow(key string) bool {
	rl.mu.Lock()
	bucket, exists := rl.buckets[key]
	if !exists {
		bucket = &TokenBucket{
			tokens:     rl.capacity,
			capacity:   rl.capacity,
			lastRefill: time.Now(),
		}
		rl.buckets[key] = bucket
	}
	rl.mu.Unlock()

	return bucket.consume()
}

// consume attempts to consume a token from the bucket
func (tb *TokenBucket) consume() bool {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(tb.lastRefill)
	
	// Refill tokens based on elapsed time
	tokensToAdd := int(elapsed.Seconds())
	if tokensToAdd > 0 {
		tb.tokens = min(tb.capacity, tb.tokens+tokensToAdd)
		tb.lastRefill = now
	}

	if tb.tokens > 0 {
		tb.tokens--
		return true
	}

	return false
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// LoggingMiddleware logs HTTP requests
func LoggingMiddleware() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			
			// Create request context
			ctx := context.WithValue(r.Context(), "requestContext", &RequestContext{
				RequestID: generateRequestID(),
				StartTime: start,
			})
			r = r.WithContext(ctx)

			// Wrap response writer to capture status code
			wrapped := &ResponseWriter{ResponseWriter: w, statusCode: http.StatusOK}
			
			next.ServeHTTP(wrapped, r)
			
			duration := time.Since(start)
			log.Printf("[%s] %s %s - %d - %v", 
				getRequestID(r), 
				r.Method, 
				r.URL.Path, 
				wrapped.statusCode, 
				duration)
		})
	}
}

// ResponseWriter wraps http.ResponseWriter to capture status code
type ResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *ResponseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// RateLimitMiddleware implements rate limiting
func RateLimitMiddleware(limiter *RateLimiter) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Use IP address as rate limit key
			clientIP := getClientIP(r)
			
			if !limiter.Allow(clientIP) {
				http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
				return
			}
			
			next.ServeHTTP(w, r)
		})
	}
}

// AuthMiddleware handles authentication
func AuthMiddleware() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Skip auth for health checks and public endpoints
			if r.URL.Path == "/health" || r.URL.Path == "/" {
				next.ServeHTTP(w, r)
				return
			}

			// Extract token from Authorization header
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, "Authorization header required", http.StatusUnauthorized)
				return
			}

			// Simple token validation (in production, use proper JWT validation)
			if !strings.HasPrefix(authHeader, "Bearer ") {
				http.Error(w, "Invalid authorization format", http.StatusUnauthorized)
				return
			}

			token := strings.TrimPrefix(authHeader, "Bearer ")
			if !isValidToken(token) {
				http.Error(w, "Invalid token", http.StatusUnauthorized)
				return
			}

			// Add user context
			ctx := r.Context()
			if reqCtx, ok := ctx.Value("requestContext").(*RequestContext); ok {
				reqCtx.UserID = extractUserIDFromToken(token)
			}

			next.ServeHTTP(w, r)
		})
	}
}

// CORSMiddleware handles CORS headers
func CORSMiddleware() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Request-ID")
			w.Header().Set("Access-Control-Expose-Headers", "X-Request-ID")

			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// SecurityMiddleware adds security headers
func SecurityMiddleware() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-Content-Type-Options", "nosniff")
			w.Header().Set("X-Frame-Options", "DENY")
			w.Header().Set("X-XSS-Protection", "1; mode=block")
			w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
			
			// Add request ID header
			if reqCtx, ok := r.Context().Value("requestContext").(*RequestContext); ok {
				w.Header().Set("X-Request-ID", reqCtx.RequestID)
			}

			next.ServeHTTP(w, r)
		})
	}
}

// RecoveryMiddleware recovers from panics
func RecoveryMiddleware() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					log.Printf("Panic recovered: %v", err)
					
					response := GatewayResponse{
						Status:    "error",
						Message:   "Internal server error",
						Timestamp: time.Now().Format(time.RFC3339),
					}
					
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusInternalServerError)
					json.NewEncoder(w).Encode(response)
				}
			}()
			
			next.ServeHTTP(w, r)
		})
	}
}

// MetricsMiddleware collects request metrics
func MetricsMiddleware() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			wrapped := &ResponseWriter{ResponseWriter: w, statusCode: http.StatusOK}
			
			next.ServeHTTP(wrapped, r)
			
			duration := time.Since(start)
			
			// In production, send metrics to monitoring system
			log.Printf("METRICS: method=%s path=%s status=%d duration=%v", 
				r.Method, r.URL.Path, wrapped.statusCode, duration)
		})
	}
}

// ChainMiddleware chains multiple middleware functions
func ChainMiddleware(middlewares ...Middleware) Middleware {
	return func(next http.Handler) http.Handler {
		for i := len(middlewares) - 1; i >= 0; i-- {
			next = middlewares[i](next)
		}
		return next
	}
}

// Helper functions

func generateRequestID() string {
	return fmt.Sprintf("req_%d_%d", time.Now().UnixNano(), time.Now().Nanosecond()%1000)
}

func getRequestID(r *http.Request) string {
	if reqCtx, ok := r.Context().Value("requestContext").(*RequestContext); ok {
		return reqCtx.RequestID
	}
	return "unknown"
}

func getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header first
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		ips := strings.Split(xff, ",")
		return strings.TrimSpace(ips[0])
	}
	
	// Check X-Real-IP header
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}
	
	// Fall back to remote address
	ip := r.RemoteAddr
	if colon := strings.LastIndex(ip, ":"); colon != -1 {
		ip = ip[:colon]
	}
	return ip
}

func isValidToken(token string) bool {
	// Simple token validation - in production use proper JWT validation
	// For now, accept any token that's at least 10 characters
	return len(token) >= 10
}

func extractUserIDFromToken(token string) string {
	// Simple user ID extraction - in production parse JWT claims
	// For now, use a hash of the token
	return fmt.Sprintf("user_%d", len(token))
}
