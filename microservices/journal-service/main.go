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
	journalService *Service
)

func main() {
	// Parse command line flags
	port := flag.String("port", "8091", "Port to listen on")
	flag.Parse()

	log.Printf("Starting Journal Microservice on port %s...", *port)

	// Initialize journal service with mock repository
	journalService = NewService(&MockRepository{}, nil)

	// Create router
	mux := http.NewServeMux()

	// Health check endpoint
	mux.HandleFunc("/health", handleHealth)

	// API endpoints
	mux.HandleFunc("/v1/journals", handleJournals)
	mux.HandleFunc("/v1/journals/", handleJournalByID)
	mux.HandleFunc("/v1/entries", handleEntries)
	mux.HandleFunc("/v1/projections", handleProjections)

	// Create server with optimized settings for high-performance journaling
	server := &http.Server{
		Addr:           ":" + *port,
		Handler:        mux,
		ReadTimeout:    5 * time.Second,
		WriteTimeout:   10 * time.Second,
		IdleTimeout:    120 * time.Second,
		MaxHeaderBytes: 1 << 20, // 1MB
	}

	// Start server in goroutine
	go func() {
		log.Printf("Journal service listening on port %s", *port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down journal service...")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Journal service exited")
}

// handleHealth handles health check requests
func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `{"status":"healthy","service":"journal","timestamp":"%s"}`, time.Now().Format(time.RFC3339))
}

// handleJournals handles journal creation and listing
func handleJournals(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		// List journals
		journals, err := journalService.ListJournals(r.Context(), JournalFilters{})
		if err != nil {
			log.Printf("Failed to list journals: %v", err)
			http.Error(w, "Failed to list journals", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(journals)

	case "POST":
		// Create new journal
		var req CreateJournalRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		journal, err := journalService.CreateJournal(r.Context(), &req)
		if err != nil {
			log.Printf("Failed to create journal: %v", err)
			http.Error(w, "Failed to create journal", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(journal)
	}
}

// handleJournalByID handles individual journal operations
func handleJournalByID(w http.ResponseWriter, r *http.Request) {
	// Extract journal ID from URL path
	journalID := r.URL.Path[len("/v1/journals/"):]

	switch r.Method {
	case "GET":
		journal, err := journalService.GetJournal(r.Context(), journalID)
		if err != nil {
			log.Printf("Failed to get journal: %v", err)
			http.Error(w, "Journal not found", http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(journal)
	}
}

// handleEntries handles journal entry operations
func handleEntries(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		// List entries
		entries, err := journalService.GetEntries(r.Context(), EntryFilters{})
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

		entry, err := journalService.PostEntry(r.Context(), &req)
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

// handleProjections handles balance projections
func handleProjections(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		// Get balance projections
		projections, err := journalService.GetProjections(r.Context(), ProjectionFilters{})
		if err != nil {
			log.Printf("Failed to get projections: %v", err)
			http.Error(w, "Failed to get projections", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(projections)
	}
}
