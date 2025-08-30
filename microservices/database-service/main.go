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

// HealthStatus represents health check status
type HealthStatus struct {
	Status    string `json:"status"`
	Service   string `json:"service"`
	Timestamp string `json:"timestamp"`
}

func main() {
	port := flag.String("port", "8115", "Port to listen on")
	flag.Parse()

	log.Printf("Starting Database Service on port %s...", *port)

	// Setup HTTP handlers
	http.HandleFunc("/health", healthHandler)
	http.HandleFunc("/", rootHandler)

	// Start server
	server := &http.Server{
		Addr:    ":" + *port,
		Handler: nil,
	}

	// Graceful shutdown
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan

		log.Println("Shutting down database service...")
		server.Close()
	}()

	log.Printf("Database service listening on port %s", *port)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Server failed to start: %v", err)
	}
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	status := HealthStatus{
		Status:    "healthy",
		Service:   "database-service",
		Timestamp: time.Now().Format(time.RFC3339),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(status)
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Database Service - Project X\nStatus: Running\nTime: %s", time.Now().Format(time.RFC3339))
}
