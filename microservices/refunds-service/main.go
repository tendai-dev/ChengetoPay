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
)

// Global services
var (
	refundsService *Service
)

func main() {
	// Parse command line flags
	port := flag.String("port", "8093", "Port to listen on")
	flag.Parse()

	log.Printf("Starting Refunds Microservice on port %s...", *port)

	// Initialize refunds service with mock repository
	refundsService = NewService(&MockRepository{}, nil)

	// Create router
	mux := http.NewServeMux()

	// Health check endpoint
	mux.HandleFunc("/health", handleHealth)

	// API endpoints
	mux.HandleFunc("/v1/refunds", handleRefunds)
	mux.HandleFunc("/v1/refunds/", handleRefundByID)
	mux.HandleFunc("/v1/reconcile", handleReconcile)

	// Create server with optimized settings for high-performance refund processing
	server := &http.Server{
		Addr:         ":" + *port,
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
		MaxHeaderBytes: 1 << 20, // 1MB
	}

	// Start server in goroutine
	go func() {
		log.Printf("Refunds service listening on port %s", *port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down refunds service...")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Refunds service exited")
}

// handleHealth handles health check requests
func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `{"status":"healthy","service":"refunds","timestamp":"%s"}`, time.Now().Format(time.RFC3339))
}

// handleRefunds handles refund creation and listing
func handleRefunds(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		// List refunds
		refunds, err := refundsService.ListRefunds(r.Context(), RefundFilters{})
		if err != nil {
			log.Printf("Failed to list refunds: %v", err)
			http.Error(w, "Failed to list refunds", http.StatusInternalServerError)
			return
		}
		
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(refunds)
		
	case "POST":
		// Create new refund
		var req CreateRefundRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}
		
		refund, err := refundsService.CreateRefund(r.Context(), &req)
		if err != nil {
			log.Printf("Failed to create refund: %v", err)
			http.Error(w, "Failed to create refund", http.StatusInternalServerError)
			return
		}
		
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(refund)
	}
}

// handleRefundByID handles individual refund operations
func handleRefundByID(w http.ResponseWriter, r *http.Request) {
	// Extract refund ID from URL path
	refundID := r.URL.Path[len("/v1/refunds/"):]
	
	switch r.Method {
	case "GET":
		refund, err := refundsService.GetRefund(r.Context(), refundID)
		if err != nil {
			log.Printf("Failed to get refund: %v", err)
			http.Error(w, "Refund not found", http.StatusNotFound)
			return
		}
		
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(refund)
		
	case "POST":
		// Handle refund processing
		if r.URL.Path[len("/v1/refunds/"):] == refundID+"/process" {
			handleProcessRefund(w, r, refundID)
		}
	}
}

// handleProcessRefund handles refund processing
func handleProcessRefund(w http.ResponseWriter, r *http.Request, refundID string) {
	var req ProcessRefundRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	
	req.RefundID = refundID
	
	if err := refundsService.ProcessRefund(r.Context(), &req); err != nil {
		log.Printf("Failed to process refund: %v", err)
		http.Error(w, "Failed to process refund", http.StatusInternalServerError)
		return
	}
	
	w.WriteHeader(http.StatusOK)
}

// handleReconcile handles refund reconciliation
func handleReconcile(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "POST":
		var req ReconcileRefundRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}
		
		result, err := refundsService.ReconcileRefund(r.Context(), &req)
		if err != nil {
			log.Printf("Failed to reconcile refund: %v", err)
			http.Error(w, "Failed to reconcile refund", http.StatusInternalServerError)
			return
		}
		
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(result)
	}
}
