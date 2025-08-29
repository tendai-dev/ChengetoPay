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
	paymentService *Service
)

func main() {
	// Parse command line flags
	port := flag.String("port", "8083", "Port to listen on")
	flag.Parse()

	log.Printf("Starting Payment Microservice on port %s...", *port)

	// Initialize payment service with mock repository
	paymentService = NewService(&MockRepository{}, nil)

	// Create router
	mux := http.NewServeMux()

	// Health check endpoint
	mux.HandleFunc("/health", handleHealth)

	// API endpoints
	mux.HandleFunc("/v1/payments", handlePayments)
	mux.HandleFunc("/v1/payments/", handlePaymentByID)
	mux.HandleFunc("/v1/providers", handleProviders)

	// Create server with optimized settings for high throughput
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
		log.Printf("Payment service listening on port %s", *port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down payment service...")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Payment service exited")
}

// handleHealth handles health check requests
func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `{"status":"healthy","service":"payment","timestamp":"%s"}`, time.Now().Format(time.RFC3339))
}

// handlePayments handles payment listing and creation
func handlePayments(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		// List payments
		payments, err := paymentService.ListPayments(r.Context(), PaymentFilters{})
		if err != nil {
			log.Printf("Failed to list payments: %v", err)
			http.Error(w, "Failed to list payments", http.StatusInternalServerError)
			return
		}
		
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(payments)
		
	case "POST":
		// Create new payment
		var req CreatePaymentRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}
		
		payment, err := paymentService.CreatePayment(r.Context(), &req)
		if err != nil {
			log.Printf("Failed to create payment: %v", err)
			http.Error(w, "Failed to create payment", http.StatusInternalServerError)
			return
		}
		
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(payment)
	}
}

// handlePaymentByID handles individual payment operations
func handlePaymentByID(w http.ResponseWriter, r *http.Request) {
	// Extract payment ID from URL path
	paymentID := r.URL.Path[len("/v1/payments/"):]
	
	switch r.Method {
	case "GET":
		payment, err := paymentService.GetPayment(r.Context(), paymentID)
		if err != nil {
			log.Printf("Failed to get payment: %v", err)
			http.Error(w, "Payment not found", http.StatusNotFound)
			return
		}
		
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(payment)
		
	case "POST":
		// Handle payment processing
		if r.URL.Path[len("/v1/payments/"):] == paymentID+"/process" {
			handleProcessPayment(w, r, paymentID)
		}
	}
}

// handleProcessPayment handles payment processing
func handleProcessPayment(w http.ResponseWriter, r *http.Request, paymentID string) {
	var req ProcessPaymentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	
	req.PaymentID = paymentID
	
	if err := paymentService.ProcessPayment(r.Context(), req.PaymentID); err != nil {
		log.Printf("Failed to process payment: %v", err)
		http.Error(w, "Failed to process payment", http.StatusInternalServerError)
		return
	}
	
	w.WriteHeader(http.StatusOK)
}

// handleProviders handles payment provider listing
func handleProviders(w http.ResponseWriter, r *http.Request) {
	providers, err := paymentService.GetProviders(r.Context())
	if err != nil {
		log.Printf("Failed to get providers: %v", err)
		http.Error(w, "Failed to get providers", http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(providers)
}
