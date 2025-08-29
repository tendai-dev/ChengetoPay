package main

import (
	"net/http"
	"strings"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

// SecurityManager handles API Gateway security
type SecurityManager struct {
	rateLimiters    map[string]*rate.Limiter
	rateLimitersMux sync.RWMutex
	apiKeys         map[string]string
	blacklistedIPs  map[string]time.Time
	blacklistMux    sync.RWMutex
}

// NewSecurityManager creates a new security manager
func NewSecurityManager() *SecurityManager {
	return &SecurityManager{
		rateLimiters:   make(map[string]*rate.Limiter),
		apiKeys:        make(map[string]string),
		blacklistedIPs: make(map[string]time.Time),
	}
}

// RateLimitingMiddleware implements rate limiting
func (sm *SecurityManager) RateLimitingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		clientIP := sm.getClientIP(r)

		if sm.isIPBlacklisted(clientIP) {
			http.Error(w, "Access denied", http.StatusForbidden)
			return
		}

		limiter := sm.getRateLimiter(clientIP)
		if !limiter.Allow() {
			sm.recordFailedAttempt(clientIP)
			http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// AuthenticationMiddleware validates API keys
func (sm *SecurityManager) AuthenticationMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if sm.isPublicEndpoint(r.URL.Path) {
			next.ServeHTTP(w, r)
			return
		}

		apiKey := r.Header.Get("X-API-Key")
		if apiKey == "" || !sm.validateAPIKey(apiKey) {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// SecurityHeadersMiddleware adds security headers
func (sm *SecurityManager) SecurityHeadersMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		w.Header().Set("Strict-Transport-Security", "max-age=31536000")

		next.ServeHTTP(w, r)
	})
}

// Helper methods
func (sm *SecurityManager) getClientIP(r *http.Request) string {
	if ip := r.Header.Get("X-Forwarded-For"); ip != "" {
		return strings.Split(ip, ",")[0]
	}
	return r.RemoteAddr
}

func (sm *SecurityManager) getRateLimiter(clientIP string) *rate.Limiter {
	sm.rateLimitersMux.Lock()
	defer sm.rateLimitersMux.Unlock()

	limiter, exists := sm.rateLimiters[clientIP]
	if !exists {
		limiter = rate.NewLimiter(rate.Limit(10), 30) // 10 req/sec, burst of 30
		sm.rateLimiters[clientIP] = limiter
	}
	return limiter
}

func (sm *SecurityManager) validateAPIKey(apiKey string) bool {
	_, exists := sm.apiKeys[apiKey]
	return exists
}

func (sm *SecurityManager) isPublicEndpoint(path string) bool {
	publicPaths := []string{"/health", "/metrics", "/docs"}
	for _, publicPath := range publicPaths {
		if strings.HasPrefix(path, publicPath) {
			return true
		}
	}
	return false
}

func (sm *SecurityManager) isIPBlacklisted(clientIP string) bool {
	sm.blacklistMux.RLock()
	defer sm.blacklistMux.RUnlock()

	blacklistTime, exists := sm.blacklistedIPs[clientIP]
	if !exists {
		return false
	}

	if time.Since(blacklistTime) > time.Hour {
		sm.blacklistMux.RUnlock()
		sm.blacklistMux.Lock()
		delete(sm.blacklistedIPs, clientIP)
		sm.blacklistMux.Unlock()
		sm.blacklistMux.RLock()
		return false
	}

	return true
}

func (sm *SecurityManager) recordFailedAttempt(clientIP string) {
	sm.blacklistMux.Lock()
	defer sm.blacklistMux.Unlock()
	sm.blacklistedIPs[clientIP] = time.Now()
}
