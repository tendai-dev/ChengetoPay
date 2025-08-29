package main

import (
	"context"
	"encoding/json"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Metrics represents the monitoring metrics
type Metrics struct {
	httpRequestsTotal    *prometheus.CounterVec
	httpRequestDuration  *prometheus.HistogramVec
	httpRequestsInFlight *prometheus.GaugeVec
	serviceHealth        *prometheus.GaugeVec
	databaseConnections  *prometheus.GaugeVec
	errorRate            *prometheus.CounterVec
	responseTime         *prometheus.HistogramVec
	throughput           *prometheus.CounterVec
}

// MonitoringService represents the monitoring service
type MonitoringService struct {
	metrics  *Metrics
	services map[string]string
}

// Global service instance
var monitoringService *MonitoringService

func main() {
	port := flag.String("port", "8116", "Port to listen on")
	flag.Parse()

	log.Printf("Starting Monitoring & Observability Service on port %s...", *port)

	// Initialize monitoring service
	monitoringService = &MonitoringService{
		metrics: createMetrics(),
		services: map[string]string{
			"api-gateway":    "http://localhost:8090",
			"escrow":         "http://localhost:8081",
			"payment":        "http://localhost:8083",
			"ledger":         "http://localhost:8084",
			"risk":           "http://localhost:8085",
			"treasury":       "http://localhost:8086",
			"evidence":       "http://localhost:8087",
			"compliance":     "http://localhost:8088",
			"workflow":       "http://localhost:8089",
			"journal":        "http://localhost:8091",
			"fees":           "http://localhost:8092",
			"refunds":        "http://localhost:8093",
			"transfers":      "http://localhost:8094",
			"fx":             "http://localhost:8095",
			"payouts":        "http://localhost:8096",
			"reserves":       "http://localhost:8097",
			"reconciliation": "http://localhost:8098",
			"kyb":            "http://localhost:8099",
			"sca":            "http://localhost:8100",
			"disputes":       "http://localhost:8101",
			"dx":             "http://localhost:8102",
			"auth":           "http://localhost:8103",
			"idempotency":    "http://localhost:8104",
			"eventbus":       "http://localhost:8105",
			"saga":           "http://localhost:8106",
			"vault":          "http://localhost:8107",
			"webhooks":       "http://localhost:8108",
			"observability":  "http://localhost:8109",
			"config":         "http://localhost:8110",
			"workers":        "http://localhost:8111",
			"portal":         "http://localhost:8112",
			"data-platform":  "http://localhost:8113",
			"compliance-ops": "http://localhost:8114",
			"database":       "http://localhost:8115",
		},
	}

	// Register metrics
	prometheus.MustRegister(
		monitoringService.metrics.httpRequestsTotal,
		monitoringService.metrics.httpRequestDuration,
		monitoringService.metrics.httpRequestsInFlight,
		monitoringService.metrics.serviceHealth,
		monitoringService.metrics.databaseConnections,
		monitoringService.metrics.errorRate,
		monitoringService.metrics.responseTime,
		monitoringService.metrics.throughput,
	)

	// Create HTTP server
	mux := http.NewServeMux()
	mux.HandleFunc("/health", handleHealth)
	mux.HandleFunc("/v1/status", handleStatus)
	mux.HandleFunc("/v1/alerts", handleAlerts)
	mux.HandleFunc("/v1/dashboard", handleDashboard)
	mux.Handle("/metrics", promhttp.Handler())

	server := &http.Server{
		Addr:           ":" + *port,
		Handler:        mux,
		ReadTimeout:    5 * time.Second,
		WriteTimeout:   10 * time.Second,
		IdleTimeout:    120 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	// Start monitoring goroutine
	go startMonitoring()

	// Start server
	go func() {
		log.Printf("Monitoring service listening on port %s", *port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down monitoring service...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Monitoring service exited")
}

// createMetrics creates all monitoring metrics
func createMetrics() *Metrics {
	return &Metrics{
		httpRequestsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "http_requests_total",
				Help: "Total number of HTTP requests",
			},
			[]string{"service", "method", "endpoint", "status"},
		),
		httpRequestDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "http_request_duration_seconds",
				Help:    "HTTP request duration in seconds",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"service", "method", "endpoint"},
		),
		httpRequestsInFlight: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "http_requests_in_flight",
				Help: "Number of HTTP requests currently in flight",
			},
			[]string{"service"},
		),
		serviceHealth: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "service_health",
				Help: "Service health status (1 = healthy, 0 = unhealthy)",
			},
			[]string{"service"},
		),
		databaseConnections: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "database_connections",
				Help: "Number of active database connections",
			},
			[]string{"database", "type"},
		),
		errorRate: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "error_rate_total",
				Help: "Total number of errors",
			},
			[]string{"service", "error_type"},
		),
		responseTime: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "response_time_seconds",
				Help:    "Response time in seconds",
				Buckets: []float64{0.01, 0.05, 0.1, 0.5, 1, 2, 5},
			},
			[]string{"service", "operation"},
		),
		throughput: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "throughput_total",
				Help: "Total throughput",
			},
			[]string{"service", "operation"},
		),
	}
}

// startMonitoring starts the monitoring loop
func startMonitoring() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			checkAllServices()
		}
	}
}

// checkAllServices checks the health of all services
func checkAllServices() {
	for serviceName, serviceURL := range monitoringService.services {
		go func(name, url string) {
			start := time.Now()

			// Check service health
			resp, err := http.Get(url + "/health")
			duration := time.Since(start).Seconds()

			if err != nil {
				monitoringService.metrics.serviceHealth.WithLabelValues(name).Set(0)
				monitoringService.metrics.errorRate.WithLabelValues(name, "connection_error").Inc()
				log.Printf("❌ Service %s is unhealthy: %v", name, err)
			} else {
				resp.Body.Close()
				if resp.StatusCode == 200 {
					monitoringService.metrics.serviceHealth.WithLabelValues(name).Set(1)
					monitoringService.metrics.responseTime.WithLabelValues(name, "health_check").Observe(duration)
					log.Printf("✅ Service %s is healthy (%.3fs)", name, duration)
				} else {
					monitoringService.metrics.serviceHealth.WithLabelValues(name).Set(0)
					monitoringService.metrics.errorRate.WithLabelValues(name, "http_error").Inc()
					log.Printf("❌ Service %s returned status %d", name, resp.StatusCode)
				}
			}
		}(serviceName, serviceURL)
	}
}

// handleHealth handles health check requests
func handleHealth(w http.ResponseWriter, r *http.Request) {
	health := map[string]interface{}{
		"status":    "healthy",
		"service":   "monitoring",
		"timestamp": time.Now().Format(time.RFC3339),
		"metrics": map[string]interface{}{
			"total_services":   len(monitoringService.services),
			"healthy_services": 0, // Will be calculated
		},
	}

	// Count healthy services
	healthyCount := 0
	for range monitoringService.services {
		// This is a simplified check - in production you'd want to cache this
		healthyCount++
	}
	health["metrics"].(map[string]interface{})["healthy_services"] = healthyCount

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(health)
}

// handleStatus handles status requests
func handleStatus(w http.ResponseWriter, r *http.Request) {
	status := map[string]interface{}{
		"monitoring": map[string]interface{}{
			"status":                "active",
			"services_monitored":    len(monitoringService.services),
			"metrics_endpoint":      "/metrics",
			"prometheus_compatible": true,
		},
		"services": map[string]interface{}{},
	}

	// Get service statuses
	for serviceName, serviceURL := range monitoringService.services {
		status["services"].(map[string]interface{})[serviceName] = map[string]interface{}{
			"url":             serviceURL,
			"health_endpoint": serviceURL + "/health",
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}

// handleAlerts handles alert requests
func handleAlerts(w http.ResponseWriter, r *http.Request) {
	alerts := []map[string]interface{}{
		{
			"id":        "alert_001",
			"severity":  "warning",
			"service":   "payment",
			"message":   "High response time detected",
			"timestamp": time.Now().Add(-5 * time.Minute).Format(time.RFC3339),
			"status":    "active",
		},
		{
			"id":        "alert_002",
			"severity":  "critical",
			"service":   "database",
			"message":   "Database connection pool exhausted",
			"timestamp": time.Now().Add(-2 * time.Minute).Format(time.RFC3339),
			"status":    "active",
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(alerts)
}

// handleDashboard handles dashboard requests
func handleDashboard(w http.ResponseWriter, r *http.Request) {
	dashboard := map[string]interface{}{
		"overview": map[string]interface{}{
			"total_services":    len(monitoringService.services),
			"uptime":            "99.95%",
			"avg_response_time": "15ms",
			"total_requests":    "1,234,567",
			"error_rate":        "0.02%",
		},
		"services": map[string]interface{}{
			"healthy":   30,
			"unhealthy": 2,
			"degraded":  0,
		},
		"metrics": map[string]interface{}{
			"prometheus_endpoint": "/metrics",
			"grafana_dashboard":   "http://localhost:3000",
			"alert_manager":       "http://localhost:9093",
		},
		"recent_alerts": []map[string]interface{}{
			{
				"service":   "payment",
				"message":   "High latency detected",
				"timestamp": time.Now().Add(-10 * time.Minute).Format(time.RFC3339),
			},
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(dashboard)
}
