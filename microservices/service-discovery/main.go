package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/hashicorp/consul/api"
)

// ServiceDiscovery represents the service discovery service
type ServiceDiscovery struct {
	consulClient *api.Client
	services     map[string]*ServiceInfo
}

// ServiceInfo represents service information
type ServiceInfo struct {
	ID       string            `json:"id"`
	Name     string            `json:"name"`
	Address  string            `json:"address"`
	Port     int               `json:"port"`
	Tags     []string          `json:"tags"`
	Metadata map[string]string `json:"metadata"`
	Status   string            `json:"status"`
	LastSeen time.Time         `json:"last_seen"`
}

// Global service instance
var sdService *ServiceDiscovery

func main() {
	port := flag.String("port", "8118", "Port to listen on")
	consulAddr := flag.String("consul", "localhost:8500", "Consul address")
	flag.Parse()

	log.Printf("Starting Service Discovery Service on port %s...", *port)

	// Initialize service discovery
	if err := initializeServiceDiscovery(*consulAddr); err != nil {
		log.Fatalf("Failed to initialize service discovery: %v", err)
	}

	// Create HTTP server
	mux := http.NewServeMux()
	mux.HandleFunc("/health", handleHealth)
	mux.HandleFunc("/v1/services", handleServices)
	mux.HandleFunc("/v1/register", handleRegister)
	mux.HandleFunc("/v1/deregister", handleDeregister)
	mux.HandleFunc("/v1/discover", handleDiscover)
	mux.HandleFunc("/v1/health", handleHealthCheck)

	server := &http.Server{
		Addr:           ":" + *port,
		Handler:        mux,
		ReadTimeout:    5 * time.Second,
		WriteTimeout:   10 * time.Second,
		IdleTimeout:    120 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	// Start health monitoring
	go startHealthMonitoring()

	// Start server
	go func() {
		log.Printf("Service Discovery listening on port %s", *port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down service discovery...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Service Discovery exited")
}

// initializeServiceDiscovery initializes Consul client and service discovery
func initializeServiceDiscovery(consulAddr string) error {
	config := api.DefaultConfig()
	config.Address = consulAddr

	client, err := api.NewClient(config)
	if err != nil {
		return fmt.Errorf("failed to create Consul client: %w", err)
	}

	sdService = &ServiceDiscovery{
		consulClient: client,
		services:     make(map[string]*ServiceInfo),
	}

	// Register default services
	if err := registerDefaultServices(); err != nil {
		return fmt.Errorf("failed to register default services: %w", err)
	}

	log.Println("✅ Service Discovery initialized successfully")
	return nil
}

// registerDefaultServices registers all microservices with Consul
func registerDefaultServices() error {
	defaultServices := []ServiceInfo{
		{ID: "api-gateway", Name: "api-gateway", Address: "localhost", Port: 8090, Tags: []string{"api", "gateway"}},
		{ID: "escrow", Name: "escrow", Address: "localhost", Port: 8081, Tags: []string{"financial", "escrow"}},
		{ID: "payment", Name: "payment", Address: "localhost", Port: 8083, Tags: []string{"financial", "payment"}},
		{ID: "ledger", Name: "ledger", Address: "localhost", Port: 8084, Tags: []string{"financial", "ledger"}},
		{ID: "risk", Name: "risk", Address: "localhost", Port: 8085, Tags: []string{"security", "risk"}},
		{ID: "treasury", Name: "treasury", Address: "localhost", Port: 8086, Tags: []string{"financial", "treasury"}},
		{ID: "evidence", Name: "evidence", Address: "localhost", Port: 8087, Tags: []string{"compliance", "evidence"}},
		{ID: "compliance", Name: "compliance", Address: "localhost", Port: 8088, Tags: []string{"compliance", "security"}},
		{ID: "workflow", Name: "workflow", Address: "localhost", Port: 8089, Tags: []string{"workflow", "automation"}},
		{ID: "journal", Name: "journal", Address: "localhost", Port: 8091, Tags: []string{"financial", "journal"}},
		{ID: "fees", Name: "fees", Address: "localhost", Port: 8092, Tags: []string{"financial", "fees"}},
		{ID: "refunds", Name: "refunds", Address: "localhost", Port: 8093, Tags: []string{"financial", "refunds"}},
		{ID: "transfers", Name: "transfers", Address: "localhost", Port: 8094, Tags: []string{"financial", "transfers"}},
		{ID: "fx", Name: "fx", Address: "localhost", Port: 8095, Tags: []string{"financial", "fx"}},
		{ID: "payouts", Name: "payouts", Address: "localhost", Port: 8096, Tags: []string{"financial", "payouts"}},
		{ID: "reserves", Name: "reserves", Address: "localhost", Port: 8097, Tags: []string{"financial", "reserves"}},
		{ID: "reconciliation", Name: "reconciliation", Address: "localhost", Port: 8098, Tags: []string{"financial", "reconciliation"}},
		{ID: "kyb", Name: "kyb", Address: "localhost", Port: 8099, Tags: []string{"compliance", "kyb"}},
		{ID: "sca", Name: "sca", Address: "localhost", Port: 8100, Tags: []string{"security", "sca"}},
		{ID: "disputes", Name: "disputes", Address: "localhost", Port: 8101, Tags: []string{"compliance", "disputes"}},
		{ID: "dx", Name: "dx", Address: "localhost", Port: 8102, Tags: []string{"developer", "dx"}},
		{ID: "auth", Name: "auth", Address: "localhost", Port: 8103, Tags: []string{"security", "auth"}},
		{ID: "idempotency", Name: "idempotency", Address: "localhost", Port: 8104, Tags: []string{"infrastructure", "idempotency"}},
		{ID: "eventbus", Name: "eventbus", Address: "localhost", Port: 8105, Tags: []string{"infrastructure", "events"}},
		{ID: "saga", Name: "saga", Address: "localhost", Port: 8106, Tags: []string{"infrastructure", "saga"}},
		{ID: "vault", Name: "vault", Address: "localhost", Port: 8107, Tags: []string{"security", "vault"}},
		{ID: "webhooks", Name: "webhooks", Address: "localhost", Port: 8108, Tags: []string{"infrastructure", "webhooks"}},
		{ID: "observability", Name: "observability", Address: "localhost", Port: 8109, Tags: []string{"monitoring", "observability"}},
		{ID: "config", Name: "config", Address: "localhost", Port: 8110, Tags: []string{"infrastructure", "config"}},
		{ID: "workers", Name: "workers", Address: "localhost", Port: 8111, Tags: []string{"infrastructure", "workers"}},
		{ID: "portal", Name: "portal", Address: "localhost", Port: 8112, Tags: []string{"developer", "portal"}},
		{ID: "data-platform", Name: "data-platform", Address: "localhost", Port: 8113, Tags: []string{"infrastructure", "data"}},
		{ID: "compliance-ops", Name: "compliance-ops", Address: "localhost", Port: 8114, Tags: []string{"compliance", "ops"}},
		{ID: "database", Name: "database", Address: "localhost", Port: 8115, Tags: []string{"infrastructure", "database"}},
		{ID: "monitoring", Name: "monitoring", Address: "localhost", Port: 8116, Tags: []string{"monitoring", "metrics"}},
		{ID: "message-queue", Name: "message-queue", Address: "localhost", Port: 8117, Tags: []string{"infrastructure", "message-queue"}},
	}

	for _, service := range defaultServices {
		if err := registerService(service); err != nil {
			log.Printf("Warning: Failed to register service %s: %v", service.ID, err)
		}
	}

	return nil
}

// registerService registers a service with Consul
func registerService(service ServiceInfo) error {
	registration := &api.AgentServiceRegistration{
		ID:      service.ID,
		Name:    service.Name,
		Address: service.Address,
		Port:    service.Port,
		Tags:    service.Tags,
		Check: &api.AgentServiceCheck{
			HTTP:                           fmt.Sprintf("http://%s:%d/health", service.Address, service.Port),
			Interval:                       "10s",
			Timeout:                        "5s",
			DeregisterCriticalServiceAfter: "30s",
		},
	}

	err := sdService.consulClient.Agent().ServiceRegister(registration)
	if err != nil {
		return fmt.Errorf("failed to register service %s: %w", service.ID, err)
	}

	sdService.services[service.ID] = &service
	log.Printf("✅ Registered service: %s (%s:%d)", service.Name, service.Address, service.Port)

	return nil
}

// startHealthMonitoring starts monitoring service health
func startHealthMonitoring() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			checkAllServicesHealth()
		}
	}
}

// checkAllServicesHealth checks health of all registered services
func checkAllServicesHealth() {
	for serviceID, service := range sdService.services {
		healthy, err := checkServiceHealth(service)
		if err != nil {
			log.Printf("❌ Health check failed for %s: %v", serviceID, err)
			service.Status = "unhealthy"
		} else if healthy {
			service.Status = "healthy"
			service.LastSeen = time.Now()
		} else {
			service.Status = "unhealthy"
		}
	}
}

// checkServiceHealth checks if a service is healthy
func checkServiceHealth(service *ServiceInfo) (bool, error) {
	url := fmt.Sprintf("http://%s:%d/health", service.Address, service.Port)

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	return resp.StatusCode == 200, nil
}

// handleHealth handles health check requests
func handleHealth(w http.ResponseWriter, r *http.Request) {
	health := map[string]interface{}{
		"status":    "healthy",
		"service":   "service-discovery",
		"timestamp": time.Now().Format(time.RFC3339),
		"consul":    "connected",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(health)
}

// handleServices handles service listing requests
func handleServices(w http.ResponseWriter, r *http.Request) {
	services := make([]*ServiceInfo, 0, len(sdService.services))
	for _, service := range sdService.services {
		services = append(services, service)
	}

	response := map[string]interface{}{
		"status":   "success",
		"services": services,
		"count":    len(services),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleRegister handles service registration requests
func handleRegister(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var service ServiceInfo
	if err := json.NewDecoder(r.Body).Decode(&service); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := registerService(service); err != nil {
		http.Error(w, fmt.Sprintf("Failed to register service: %v", err), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"status":  "success",
		"message": "Service registered successfully",
		"data": map[string]interface{}{
			"service_id": service.ID,
			"name":       service.Name,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleDeregister handles service deregistration requests
func handleDeregister(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var request struct {
		ServiceID string `json:"service_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	err := sdService.consulClient.Agent().ServiceDeregister(request.ServiceID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to deregister service: %v", err), http.StatusInternalServerError)
		return
	}

	delete(sdService.services, request.ServiceID)

	response := map[string]interface{}{
		"status":  "success",
		"message": "Service deregistered successfully",
		"data": map[string]interface{}{
			"service_id": request.ServiceID,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleDiscover handles service discovery requests
func handleDiscover(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	serviceName := query.Get("service")
	tag := query.Get("tag")

	if serviceName == "" {
		http.Error(w, "Service name is required", http.StatusBadRequest)
		return
	}

	services, _, err := sdService.consulClient.Health().Service(serviceName, tag, true, nil)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to discover services: %v", err), http.StatusInternalServerError)
		return
	}

	var discoveredServices []map[string]interface{}
	for _, service := range services {
		discoveredServices = append(discoveredServices, map[string]interface{}{
			"id":      service.Service.ID,
			"name":    service.Service.Service,
			"address": service.Service.Address,
			"port":    service.Service.Port,
			"tags":    service.Service.Tags,
			"status":  service.Checks.AggregatedStatus(),
		})
	}

	response := map[string]interface{}{
		"status":   "success",
		"services": discoveredServices,
		"count":    len(discoveredServices),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleHealthCheck handles health check aggregation requests
func handleHealthCheck(w http.ResponseWriter, r *http.Request) {
	healthStatus := make(map[string]interface{})
	totalServices := len(sdService.services)
	healthyServices := 0

	for serviceID, service := range sdService.services {
		healthy, _ := checkServiceHealth(service)
		status := "unhealthy"
		if healthy {
			status = "healthy"
			healthyServices++
		}

		healthStatus[serviceID] = map[string]interface{}{
			"name":      service.Name,
			"status":    status,
			"last_seen": service.LastSeen,
		}
	}

	response := map[string]interface{}{
		"status": "success",
		"data": map[string]interface{}{
			"total_services":     totalServices,
			"healthy_services":   healthyServices,
			"unhealthy_services": totalServices - healthyServices,
			"health_percentage":  float64(healthyServices) / float64(totalServices) * 100,
			"services":           healthStatus,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
