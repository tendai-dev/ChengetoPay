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
	port := flag.String("port", "8118", "Port to listen on")
	flag.Parse()

	log.Printf("Starting Service Discovery on port %s...", *port)

	// Create HTTP server
	mux := http.NewServeMux()
	mux.HandleFunc("/health", handleHealth)
	mux.HandleFunc("/v1/status", handleStatus)
	mux.HandleFunc("/v1/services", handleServices)

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
	log.Println("Service Discovery exited")
}

// handleHealth handles health check requests
func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `{"status":"healthy","service":"service-discovery","timestamp":"%s"}`, time.Now().Format(time.RFC3339))
}

// handleStatus handles status requests
func handleStatus(w http.ResponseWriter, r *http.Request) {
	status := map[string]interface{}{
		"service_discovery": map[string]interface{}{
			"status": "active",
			"mode":   "simplified",
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}

// handleServices returns list of services
func handleServices(w http.ResponseWriter, r *http.Request) {
	services := map[string]interface{}{
		"services": []string{
			"escrow-service",
			"payment-service",
			"ledger-service",
			"database-service",
			"message-queue-service",
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(services)
}
