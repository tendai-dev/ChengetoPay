package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	port := flag.String("port", "8116", "Port to listen on")
	flag.Parse()

	log.Printf("Starting Monitoring Service on port %s...", *port)

	// Create HTTP server
	mux := http.NewServeMux()
	mux.HandleFunc("/health", handleHealth)
	mux.HandleFunc("/metrics", handleMetrics)
	mux.HandleFunc("/v1/status", handleStatus)

	server := &http.Server{
		Addr:           ":" + *port,
		Handler:        mux,
		ReadTimeout:    5 * time.Second,
		WriteTimeout:   10 * time.Second,
		IdleTimeout:    120 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

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
	log.Println("Monitoring service exited")
}

// handleHealth handles health check requests
func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `{"status":"healthy","service":"monitoring","timestamp":"%s"}`, time.Now().Format(time.RFC3339))
}

// handleMetrics handles metrics requests
func handleMetrics(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "# HELP service_uptime Service uptime in seconds\n# TYPE service_uptime counter\nservice_uptime %d\n", 100)
}

// handleStatus handles status requests
func handleStatus(w http.ResponseWriter, r *http.Request) {
	status := map[string]interface{}{
		"monitoring": map[string]interface{}{
			"status": "active",
			"mode":   "simplified",
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}
