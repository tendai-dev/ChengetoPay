#!/bin/bash

# Script to create simplified main.go for all services that need it

services=(
  "risk-service:8085"
  "treasury-service:8086"
  "evidence-service:8087"
  "compliance-service:8088"
  "workflow-service:8089"
  "journal-service:8091"
  "fees-service:8092"
  "refunds-service:8093"
  "transfers-service:8094"
  "fx-service:8095"
  "payouts-service:8096"
  "reserves-service:8097"
  "reconciliation-service:8098"
  "kyb-service:8099"
  "sca-service:8100"
  "disputes-service:8101"
  "dx-service:8102"
  "auth-service:8103"
  "idempotency-service:8104"
  "eventbus-service:8105"
  "saga-service:8106"
  "webhooks-service:8108"
  "observability-service:8109"
  "config-service:8110"
  "workers-service:8111"
  "portal-service:8112"
  "data-platform-service:8113"
  "compliance-ops-service:8114"
)

for service_port in "${services[@]}"; do
  IFS=':' read -r service port <<< "$service_port"
  
  if [ ! -f "$service/main.go" ] || grep -q "package main" "$service/main.go" 2>/dev/null && ! grep -q "func main()" "$service/main.go" 2>/dev/null; then
    echo "Creating simplified main.go for $service on port $port..."
    
    cat > "$service/main.go" << EOF
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
	port := flag.String("port", "$port", "Port to listen on")
	flag.Parse()

	serviceName := "${service%-service}"
	log.Printf("Starting %s Service on port %s...", serviceName, *port)

	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, \`{"status":"healthy","service":"%s","timestamp":"%s"}\`, serviceName, time.Now().Format(time.RFC3339))
	})
	
	mux.HandleFunc("/v1/status", func(w http.ResponseWriter, r *http.Request) {
		status := map[string]interface{}{
			"service": serviceName,
			"status": "active",
			"port": *port,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(status)
	})

	server := &http.Server{
		Addr:           ":" + *port,
		Handler:        mux,
		ReadTimeout:    5 * time.Second,
		WriteTimeout:   10 * time.Second,
		IdleTimeout:    120 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	go func() {
		log.Printf("%s service listening on port %s", serviceName, *port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Printf("Shutting down %s service...", serviceName)
	log.Printf("%s service exited", serviceName)
}
EOF
  fi
  
  # Ensure go.mod exists and is simple
  if [ ! -f "$service/go.mod" ]; then
    echo "Creating go.mod for $service..."
    cat > "$service/go.mod" << EOF
module $service

go 1.21
EOF
  fi
  
  # Create empty go.sum if it doesn't exist
  if [ ! -f "$service/go.sum" ]; then
    touch "$service/go.sum"
  fi
done

echo "All services simplified!"
