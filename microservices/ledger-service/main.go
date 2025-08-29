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
	ledgerService *Service
)

func main() {
	// Parse command line flags
	port := flag.String("port", "8084", "Port to listen on")
	flag.Parse()

	log.Printf("Starting Ledger Microservice on port %s...", *port)

	// Initialize ledger service with mock repository
	ledgerService = NewService(&MockRepository{}, nil)

	// Create router
	mux := http.NewServeMux()

	// Health check endpoint
	mux.HandleFunc("/health", handleHealth)

	// API endpoints
	mux.HandleFunc("/v1/accounts", handleAccounts)
	mux.HandleFunc("/v1/accounts/", handleAccountByID)
	mux.HandleFunc("/v1/entries", handleEntries)
	mux.HandleFunc("/v1/balance/", handleBalance)

	// Create server with optimized settings for high-performance calculations
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
		log.Printf("Ledger service listening on port %s", *port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down ledger service...")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Ledger service exited")
}

// handleHealth handles health check requests
func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `{"status":"healthy","service":"ledger","timestamp":"%s"}`, time.Now().Format(time.RFC3339))
}

// handleAccounts handles account listing and creation
func handleAccounts(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		// List accounts
		accounts, err := ledgerService.ListAccounts(r.Context(), AccountFilters{})
		if err != nil {
			log.Printf("Failed to list accounts: %v", err)
			http.Error(w, "Failed to list accounts", http.StatusInternalServerError)
			return
		}
		
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(accounts)
		
	case "POST":
		// Create new account
		var req CreateAccountRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}
		
		account, err := ledgerService.CreateAccount(r.Context(), &req)
		if err != nil {
			log.Printf("Failed to create account: %v", err)
			http.Error(w, "Failed to create account", http.StatusInternalServerError)
			return
		}
		
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(account)
	}
}

// handleAccountByID handles individual account operations
func handleAccountByID(w http.ResponseWriter, r *http.Request) {
	// Extract account ID from URL path
	accountID := r.URL.Path[len("/v1/accounts/"):]
	
	switch r.Method {
	case "GET":
		account, err := ledgerService.GetAccount(r.Context(), accountID)
		if err != nil {
			log.Printf("Failed to get account: %v", err)
			http.Error(w, "Account not found", http.StatusNotFound)
			return
		}
		
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(account)
	}
}

// handleEntries handles ledger entry operations
func handleEntries(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		// List entries
		entries, err := ledgerService.GetEntries(r.Context(), EntryFilters{})
		if err != nil {
			log.Printf("Failed to list entries: %v", err)
			http.Error(w, "Failed to list entries", http.StatusInternalServerError)
			return
		}
		
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(entries)
		
	case "POST":
		// Post new entry
		var req PostEntryRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}
		
		entry, err := ledgerService.PostEntry(r.Context(), &req)
		if err != nil {
			log.Printf("Failed to post entry: %v", err)
			http.Error(w, "Failed to post entry", http.StatusInternalServerError)
			return
		}
		
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(entry)
	}
}

// handleBalance handles balance retrieval
func handleBalance(w http.ResponseWriter, r *http.Request) {
	// Extract account ID from URL path
	accountID := r.URL.Path[len("/v1/balance/"):]
	
	balance, err := ledgerService.GetBalance(r.Context(), accountID)
	if err != nil {
		log.Printf("Failed to get balance: %v", err)
		http.Error(w, "Failed to get balance", http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(balance)
}
