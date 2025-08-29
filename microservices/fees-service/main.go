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
	feesService *Service
)

func main() {
	// Parse command line flags
	port := flag.String("port", "8092", "Port to listen on")
	flag.Parse()

	log.Printf("Starting Fees & Pricing Microservice on port %s...", *port)

	// Initialize fees service with mock repository
	feesService = NewService(&MockRepository{}, nil)

	// Create router
	mux := http.NewServeMux()

	// Health check endpoint
	mux.HandleFunc("/health", handleHealth)

	// API endpoints
	mux.HandleFunc("/v1/fees", handleFees)
	mux.HandleFunc("/v1/fees/", handleFeeByID)
	mux.HandleFunc("/v1/calculate", handleCalculateFees)
	mux.HandleFunc("/v1/schedules", handleSchedules)
	mux.HandleFunc("/v1/taxes", handleTaxes)

	// Create server with optimized settings for high-performance fee calculations
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
		log.Printf("Fees & Pricing service listening on port %s", *port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down fees service...")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Fees & Pricing service exited")
}

// handleHealth handles health check requests
func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `{"status":"healthy","service":"fees","timestamp":"%s"}`, time.Now().Format(time.RFC3339))
}

// handleFees handles fee schedule management
func handleFees(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		// List fee schedules
		fees, err := feesService.ListFeeSchedules(r.Context(), FeeFilters{})
		if err != nil {
			log.Printf("Failed to list fees: %v", err)
			http.Error(w, "Failed to list fees", http.StatusInternalServerError)
			return
		}
		
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(fees)
		
	case "POST":
		// Create new fee schedule
		var req CreateFeeScheduleRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}
		
		fee, err := feesService.CreateFeeSchedule(r.Context(), &req)
		if err != nil {
			log.Printf("Failed to create fee schedule: %v", err)
			http.Error(w, "Failed to create fee schedule", http.StatusInternalServerError)
			return
		}
		
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(fee)
	}
}

// handleFeeByID handles individual fee schedule operations
func handleFeeByID(w http.ResponseWriter, r *http.Request) {
	// Extract fee ID from URL path
	feeID := r.URL.Path[len("/v1/fees/"):]
	
	switch r.Method {
	case "GET":
		fee, err := feesService.GetFeeSchedule(r.Context(), feeID)
		if err != nil {
			log.Printf("Failed to get fee schedule: %v", err)
			http.Error(w, "Fee schedule not found", http.StatusNotFound)
			return
		}
		
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(fee)
	}
}

// handleCalculateFees handles fee calculation
func handleCalculateFees(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "POST":
		var req CalculateFeesRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}
		
		result, err := feesService.CalculateFees(r.Context(), &req)
		if err != nil {
			log.Printf("Failed to calculate fees: %v", err)
			http.Error(w, "Failed to calculate fees", http.StatusInternalServerError)
			return
		}
		
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(result)
	}
}

// handleSchedules handles fee schedule management
func handleSchedules(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		schedules, err := feesService.GetSchedules(r.Context(), ScheduleFilters{})
		if err != nil {
			log.Printf("Failed to get schedules: %v", err)
			http.Error(w, "Failed to get schedules", http.StatusInternalServerError)
			return
		}
		
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(schedules)
	}
}

// handleTaxes handles tax calculations
func handleTaxes(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "POST":
		var req CalculateTaxRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}
		
		result, err := feesService.CalculateTax(r.Context(), &req)
		if err != nil {
			log.Printf("Failed to calculate tax: %v", err)
			http.Error(w, "Failed to calculate tax", http.StatusInternalServerError)
			return
		}
		
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(result)
	}
}
