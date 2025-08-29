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
	escrowService *Service
)

func main() {
	// Parse command line flags
	port := flag.String("port", "8081", "Port to listen on")
	flag.Parse()

	log.Printf("Starting Escrow Microservice on port %s...", *port)

	// Initialize escrow service with mock repository
	escrowService = NewService(&MockRepository{}, nil)

	// Create router
	mux := http.NewServeMux()

	// Health check endpoint
	mux.HandleFunc("/health", handleHealth)

	// API endpoints
	mux.HandleFunc("/v1/escrows", handleEscrows)
	mux.HandleFunc("/v1/escrows/", handleEscrowByID)

	// Create server with optimized settings
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
		log.Printf("Escrow service listening on port %s", *port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down escrow service...")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Escrow service exited")
}

// handleHealth handles health check requests
func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `{"status":"healthy","service":"escrow","timestamp":"%s"}`, time.Now().Format(time.RFC3339))
}

// handleEscrows handles escrow listing and creation
func handleEscrows(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		// List escrows
		escrows, err := escrowService.ListEscrows(r.Context(), EscrowFilters{})
		if err != nil {
			log.Printf("Failed to list escrows: %v", err)
			http.Error(w, "Failed to list escrows", http.StatusInternalServerError)
			return
		}
		
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(escrows)
		
	case "POST":
		// Create new escrow
		var req CreateEscrowRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}
		
		escrow, err := escrowService.CreateEscrow(r.Context(), &req)
		if err != nil {
			log.Printf("Failed to create escrow: %v", err)
			http.Error(w, "Failed to create escrow", http.StatusInternalServerError)
			return
		}
		
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(escrow)
	}
}

// handleEscrowByID handles individual escrow operations
func handleEscrowByID(w http.ResponseWriter, r *http.Request) {
	// Extract escrow ID from URL path
	escrowID := r.URL.Path[len("/v1/escrows/"):]
	
	switch r.Method {
	case "GET":
		escrow, err := escrowService.GetEscrow(r.Context(), escrowID)
		if err != nil {
			log.Printf("Failed to get escrow: %v", err)
			http.Error(w, "Escrow not found", http.StatusNotFound)
			return
		}
		
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(escrow)
		
	case "POST":
		// Handle escrow actions based on URL path
		if r.URL.Path[len("/v1/escrows/"):] == escrowID+"/fund" {
			handleFundEscrow(w, r, escrowID)
		} else if r.URL.Path[len("/v1/escrows/"):] == escrowID+"/confirm-delivery" {
			handleConfirmDelivery(w, r, escrowID)
		} else if r.URL.Path[len("/v1/escrows/"):] == escrowID+"/release" {
			handleReleaseEscrow(w, r, escrowID)
		} else if r.URL.Path[len("/v1/escrows/"):] == escrowID+"/cancel" {
			handleCancelEscrow(w, r, escrowID)
		}
	}
}

// handleFundEscrow handles escrow funding
func handleFundEscrow(w http.ResponseWriter, r *http.Request, escrowID string) {
	var req FundEscrowRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	
	req.EscrowID = escrowID
	
	if err := escrowService.FundEscrow(r.Context(), &req); err != nil {
		log.Printf("Failed to fund escrow: %v", err)
		http.Error(w, "Failed to fund escrow", http.StatusInternalServerError)
		return
	}
	
	w.WriteHeader(http.StatusOK)
}

// handleConfirmDelivery handles delivery confirmation
func handleConfirmDelivery(w http.ResponseWriter, r *http.Request, escrowID string) {
	var req ConfirmDeliveryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	
	req.EscrowID = escrowID
	
	if err := escrowService.ConfirmDelivery(r.Context(), &req); err != nil {
		log.Printf("Failed to confirm delivery: %v", err)
		http.Error(w, "Failed to confirm delivery", http.StatusInternalServerError)
		return
	}
	
	w.WriteHeader(http.StatusOK)
}

// handleReleaseEscrow handles escrow release
func handleReleaseEscrow(w http.ResponseWriter, r *http.Request, escrowID string) {
	if err := escrowService.ReleaseEscrow(r.Context(), escrowID); err != nil {
		log.Printf("Failed to release escrow: %v", err)
		http.Error(w, "Failed to release escrow", http.StatusInternalServerError)
		return
	}
	
	w.WriteHeader(http.StatusOK)
}

// handleCancelEscrow handles escrow cancellation
func handleCancelEscrow(w http.ResponseWriter, r *http.Request, escrowID string) {
	if err := escrowService.CancelEscrow(r.Context(), escrowID); err != nil {
		log.Printf("Failed to cancel escrow: %v", err)
		http.Error(w, "Failed to cancel escrow", http.StatusInternalServerError)
		return
	}
	
	w.WriteHeader(http.StatusOK)
}
