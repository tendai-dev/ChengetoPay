package main

import (
	"context"
	"encoding/json"
	"flag"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

// Enhanced API Gateway with middleware and service discovery
func main() {
	port := flag.String("port", "8090", "Port to listen on")
	flag.Parse()

	log.Printf("Starting Enhanced API Gateway on port %s...", *port)

	// Initialize service registry
	registry := NewServiceRegistry()
	registry.InitializeDefaultServices()

	// Initialize rate limiter (100 requests per minute per IP)
	rateLimiter := NewRateLimiter(100, time.Minute)

	// Create context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start health checking in background
	go registry.StartHealthChecking(ctx, 30*time.Second)

	// Create enhanced router
	mux := http.NewServeMux()

	// Health check endpoint
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		handleEnhancedHealth(w, r, registry)
	})

	// Service discovery endpoint
	mux.HandleFunc("/services", func(w http.ResponseWriter, r *http.Request) {
		handleServiceDiscovery(w, r, registry)
	})

	// API Gateway endpoints with enhanced routing
	mux.HandleFunc("/api/v1/", func(w http.ResponseWriter, r *http.Request) {
		handleEnhancedAPIRequest(w, r, registry)
	})

	// Root endpoint
	mux.HandleFunc("/", handleEnhancedRoot)

	// Apply middleware chain
	handler := ChainMiddleware(
		RecoveryMiddleware(),
		LoggingMiddleware(),
		SecurityMiddleware(),
		CORSMiddleware(),
		RateLimitMiddleware(rateLimiter),
		MetricsMiddleware(),
	)(mux)

	// Create server with optimized settings
	server := &http.Server{
		Addr:           ":" + *port,
		Handler:        handler,
		ReadTimeout:    15 * time.Second,
		WriteTimeout:   15 * time.Second,
		IdleTimeout:    120 * time.Second,
		MaxHeaderBytes: 1 << 20, // 1MB
	}

	// Start server in goroutine
	go func() {
		log.Printf("Enhanced API Gateway listening on port %s", *port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	log.Println("Shutting down API Gateway...")
	cancel()

	// Graceful shutdown with timeout
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("Server shutdown error: %v", err)
	} else {
		log.Println("API Gateway shutdown complete")
	}
}

// handleEnhancedHealth provides comprehensive health status
func handleEnhancedHealth(w http.ResponseWriter, r *http.Request, registry *ServiceRegistry) {
	services := registry.GetAllServices()
	servicesHealth := make(map[string]interface{})

	totalServices := 0
	healthyServices := 0

	for serviceName, instances := range services {
		serviceHealth := make(map[string]interface{})
		healthyInstances := 0
		totalInstances := len(instances)

		instancesStatus := make([]map[string]interface{}, 0)
		for _, instance := range instances {
			instanceStatus := map[string]interface{}{
				"id":      instance.ID,
				"address": instance.Address,
				"port":    instance.Port,
				"health":  instance.Health,
				"tags":    instance.Tags,
			}
			instancesStatus = append(instancesStatus, instanceStatus)

			if instance.Health == "passing" {
				healthyInstances++
			}
		}

		serviceHealth["instances"] = instancesStatus
		serviceHealth["total_instances"] = totalInstances
		serviceHealth["healthy_instances"] = healthyInstances
		serviceHealth["status"] = "unhealthy"

		if healthyInstances > 0 {
			serviceHealth["status"] = "healthy"
			healthyServices++
		}

		servicesHealth[serviceName] = serviceHealth
		totalServices++
	}

	overallStatus := "unhealthy"
	if healthyServices > 0 {
		overallStatus = "healthy"
	}

	response := GatewayResponse{
		Status:  overallStatus,
		Message: "Enhanced API Gateway Health Status",
		Data: map[string]interface{}{
			"gateway": map[string]interface{}{
				"status":    overallStatus,
				"uptime":    time.Since(startTime).String(),
				"timestamp": time.Now().Format(time.RFC3339),
			},
			"services": servicesHealth,
			"summary": map[string]interface{}{
				"total_services":   totalServices,
				"healthy_services": healthyServices,
			},
		},
		Timestamp: time.Now().Format(time.RFC3339),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleServiceDiscovery returns service registry information
func handleServiceDiscovery(w http.ResponseWriter, r *http.Request, registry *ServiceRegistry) {
	services := registry.GetAllServices()

	response := GatewayResponse{
		Status:  "success",
		Message: "Service Discovery Information",
		Data:    services,
		Timestamp: time.Now().Format(time.RFC3339),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleEnhancedAPIRequest routes API requests with enhanced features
func handleEnhancedAPIRequest(w http.ResponseWriter, r *http.Request, registry *ServiceRegistry) {
	// Extract service name from path: /api/v1/{service}/...
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 4 {
		http.Error(w, "Invalid API path", http.StatusBadRequest)
		return
	}

	serviceName := pathParts[3] + "-service" // Add -service suffix for consistency

	// Get healthy service endpoint from registry
	serviceURL, err := registry.GetServiceEndpoint(serviceName)
	if err != nil {
		log.Printf("Service discovery failed for %s: %v", serviceName, err)
		http.Error(w, "Service unavailable", http.StatusServiceUnavailable)
		return
	}

	// Create reverse proxy with enhanced error handling
	targetURL, err := url.Parse(serviceURL)
	if err != nil {
		log.Printf("Invalid service URL %s: %v", serviceURL, err)
		http.Error(w, "Service configuration error", http.StatusInternalServerError)
		return
	}

	proxy := httputil.NewSingleHostReverseProxy(targetURL)

	// Enhanced proxy error handling
	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		log.Printf("Proxy error for %s: %v", serviceName, err)
		
		response := GatewayResponse{
			Status:    "error",
			Message:   "Service temporarily unavailable",
			Timestamp: time.Now().Format(time.RFC3339),
		}
		
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadGateway)
		json.NewEncoder(w).Encode(response)
	}

	// Modify request path to remove /api/v1/{service} prefix
	originalPath := r.URL.Path
	r.URL.Path = "/v1/" + strings.Join(pathParts[4:], "/")

	// Add tracing headers
	if reqCtx, ok := r.Context().Value("requestContext").(*RequestContext); ok {
		r.Header.Set("X-Request-ID", reqCtx.RequestID)
		r.Header.Set("X-Gateway-Service", serviceName)
	}

	log.Printf("Proxying %s %s -> %s%s", r.Method, originalPath, serviceURL, r.URL.Path)

	// Proxy the request
	proxy.ServeHTTP(w, r)
}

// handleEnhancedRoot provides enhanced API documentation
func handleEnhancedRoot(w http.ResponseWriter, r *http.Request) {
	response := GatewayResponse{
		Status:  "success",
		Message: "Enhanced Escrow Platform API Gateway",
		Data: map[string]interface{}{
			"version": "4.0.0",
			"features": []string{
				"Service Discovery",
				"Health Checking",
				"Rate Limiting",
				"Request Tracing",
				"Circuit Breaking",
				"Load Balancing",
				"Security Headers",
				"CORS Support",
			},
			"endpoints": map[string]interface{}{
				"health":           "/health - Gateway and service health status",
				"services":         "/services - Service discovery information",
				"api":              "/api/v1/{service}/* - Proxied service endpoints",
				"escrow":           "/api/v1/escrow/* - Escrow service endpoints",
				"payment":          "/api/v1/payment/* - Payment service endpoints",
				"ledger":           "/api/v1/ledger/* - Ledger service endpoints",
				"risk":             "/api/v1/risk/* - Risk service endpoints",
				"treasury":         "/api/v1/treasury/* - Treasury service endpoints",
			},
			"rate_limits": map[string]interface{}{
				"requests_per_minute": 100,
				"burst_capacity":      100,
			},
			"uptime": time.Since(startTime).String(),
		},
		Timestamp: time.Now().Format(time.RFC3339),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

var startTime = time.Now()
