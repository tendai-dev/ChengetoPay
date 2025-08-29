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
	port := flag.String("port", "8109", "Port to listen on")
	flag.Parse()

	log.Printf("Starting Observability & Audit Trail Microservice on port %s...", *port)

	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"status":"healthy","service":"observability","timestamp":"%s"}`, time.Now().Format(time.RFC3339))
	})

	mux.HandleFunc("/v1/logs", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]map[string]interface{}{
			{
				"id": "log_123",
				"level": "info",
				"service": "payment",
				"message": "Payment processed successfully",
				"timestamp": time.Now().Format(time.RFC3339),
			},
		})
	})

	mux.HandleFunc("/v1/audit", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]map[string]interface{}{
			{
				"id": "audit_123",
				"action": "payment.create",
				"user_id": "user_123",
				"resource": "payment_456",
				"timestamp": time.Now().Format(time.RFC3339),
			},
		})
	})

	server := &http.Server{
		Addr: ":" + *port,
		Handler: mux,
	}

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Observability & Audit Trail service exited")
}
