package main

import (
	"encoding/json"
	"flag"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"
)

// ServiceRegistry maps service names to their URLs
var ServiceRegistry = map[string]string{
	"escrow":            "http://localhost:8081",
	"payment":           "http://localhost:8083",
	"ledger":            "http://localhost:8084",
	"risk":              "http://localhost:8085",
	"treasury":          "http://localhost:8086",
	"evidence":          "http://localhost:8087",
	"compliance":        "http://localhost:8088",
	"workflow":          "http://localhost:8089",
	"journal":           "http://localhost:8091",
	"fees":              "http://localhost:8092",
	"refunds":           "http://localhost:8093",
	"transfers":         "http://localhost:8094",
	"fx":                "http://localhost:8095",
	"payouts":           "http://localhost:8096",
	"reserves":          "http://localhost:8097",
	"reconciliation":    "http://localhost:8098",
	"kyb":               "http://localhost:8099",
	"sca":               "http://localhost:8100",
	"disputes":          "http://localhost:8101",
	"dx":                "http://localhost:8102",
	"auth":              "http://localhost:8103",
	"idempotency":       "http://localhost:8104",
	"eventbus":          "http://localhost:8105",
	"saga":              "http://localhost:8106",
	"vault-old":         "http://localhost:8107",
	"webhooks":          "http://localhost:8108",
	"observability":     "http://localhost:8109",
	"config":            "http://localhost:8110",
	"workers":           "http://localhost:8111",
	"portal":            "http://localhost:8112",
	"data-platform":     "http://localhost:8113",
	"compliance-ops":    "http://localhost:8114",
	"database":          "http://localhost:8115",
	"monitoring":        "http://localhost:8116",
	"message-queue":     "http://localhost:8117",
	"service-discovery": "http://localhost:8118",
	"vault":             "http://localhost:8119",
}

// GatewayResponse represents the API Gateway response
type GatewayResponse struct {
	Status    string      `json:"status"`
	Message   string      `json:"message"`
	Data      interface{} `json:"data,omitempty"`
	Timestamp string      `json:"timestamp"`
}

func main() {
	port := flag.String("port", "8080", "Port to listen on")
	flag.Parse()

	log.Printf("Starting API Gateway on port %s...", *port)

	// Create router
	mux := http.NewServeMux()

	// Health check endpoint
	mux.HandleFunc("/health", handleHealth)

	// API Gateway endpoints
	mux.HandleFunc("/api/v1/", handleAPIRequest)

	// Root endpoint
	mux.HandleFunc("/", handleRoot)

	// Create server with optimized settings
	server := &http.Server{
		Addr:           ":" + *port,
		Handler:        mux,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		IdleTimeout:    120 * time.Second,
		MaxHeaderBytes: 1 << 20, // 1MB
	}

	log.Printf("API Gateway listening on port %s", *port)
	log.Printf("Available services: %v", getServiceList())

	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

// handleHealth handles health check requests
func handleHealth(w http.ResponseWriter, r *http.Request) {
	// Check all microservices health
	servicesHealth := make(map[string]string)

	for serviceName, serviceURL := range ServiceRegistry {
		resp, err := http.Get(serviceURL + "/health")
		if err != nil {
			servicesHealth[serviceName] = "unhealthy"
		} else {
			resp.Body.Close()
			if resp.StatusCode == 200 {
				servicesHealth[serviceName] = "healthy"
			} else {
				servicesHealth[serviceName] = "unhealthy"
			}
		}
	}

	response := GatewayResponse{
		Status:    "healthy",
		Message:   "API Gateway is running",
		Data:      servicesHealth,
		Timestamp: time.Now().Format(time.RFC3339),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleAPIRequest routes API requests to appropriate microservices
func handleAPIRequest(w http.ResponseWriter, r *http.Request) {
	// Extract service name from path: /api/v1/{service}/...
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 4 {
		http.Error(w, "Invalid API path", http.StatusBadRequest)
		return
	}

	serviceName := pathParts[3]
	serviceURL, exists := ServiceRegistry[serviceName]
	if !exists {
		http.Error(w, "Service not found", http.StatusNotFound)
		return
	}

	// Create reverse proxy
	targetURL, _ := url.Parse(serviceURL)
	proxy := httputil.NewSingleHostReverseProxy(targetURL)

	// Modify request path to remove /api/v1/{service} prefix
	r.URL.Path = "/" + strings.Join(pathParts[4:], "/")

	// Add CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	// Handle preflight requests
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Proxy the request
	proxy.ServeHTTP(w, r)
}

// handleRoot handles root endpoint
func handleRoot(w http.ResponseWriter, r *http.Request) {
	response := GatewayResponse{
		Status:  "success",
		Message: "Escrow Platform API Gateway",
		Data: map[string]interface{}{
			"version":  "3.0.0",
			"services": getServiceList(),
			"endpoints": map[string]string{
				"health":            "/health",
				"escrow":            "/api/v1/escrow",
				"payment":           "/api/v1/payment",
				"ledger":            "/api/v1/ledger",
				"risk":              "/api/v1/risk",
				"treasury":          "/api/v1/treasury",
				"evidence":          "/api/v1/evidence",
				"compliance":        "/api/v1/compliance",
				"workflow":          "/api/v1/workflow",
				"journal":           "/api/v1/journal",
				"fees":              "/api/v1/fees",
				"refunds":           "/api/v1/refunds",
				"transfers":         "/api/v1/transfers",
				"fx":                "/api/v1/fx",
				"payouts":           "/api/v1/payouts",
				"reserves":          "/api/v1/reserves",
				"reconciliation":    "/api/v1/reconciliation",
				"kyb":               "/api/v1/kyb",
				"sca":               "/api/v1/sca",
				"disputes":          "/api/v1/disputes",
				"dx":                "/api/v1/dx",
				"auth":              "/api/v1/auth",
				"idempotency":       "/api/v1/idempotency",
				"eventbus":          "/api/v1/eventbus",
				"saga":              "/api/v1/saga",
				"vault":             "/api/v1/vault",
				"webhooks":          "/api/v1/webhooks",
				"observability":     "/api/v1/observability",
				"config":            "/api/v1/config",
				"workers":           "/api/v1/workers",
				"portal":            "/api/v1/portal",
				"data-platform":     "/api/v1/data-platform",
				"compliance-ops":    "/api/v1/compliance-ops",
				"database":          "/api/v1/database",
				"monitoring":        "/api/v1/monitoring",
				"message-queue":     "/api/v1/message-queue",
				"service-discovery": "/api/v1/service-discovery",
			},
		},
		Timestamp: time.Now().Format(time.RFC3339),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// getServiceList returns list of available services
func getServiceList() []string {
	services := make([]string, 0, len(ServiceRegistry))
	for service := range ServiceRegistry {
		services = append(services, service)
	}
	return services
}
