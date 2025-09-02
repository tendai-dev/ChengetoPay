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
	authService *Service
)

func main() {
	// Parse command line flags
	port := flag.String("port", "8103", "Port to listen on")
	flag.Parse()

	log.Printf("Starting AuthN/AuthZ & Org/Tenant Microservice on port %s...", *port)

	// Initialize auth service with mock repository
	authService = NewService(&MockRepository{}, nil)

	// Create router
	mux := http.NewServeMux()

	// Health check endpoint
	mux.HandleFunc("/health", handleHealth)

	// API endpoints
	mux.HandleFunc("/v1/auth", handleAuth)
	mux.HandleFunc("/v1/token", handleToken)
	mux.HandleFunc("/v1/orgs", handleOrgs)
	mux.HandleFunc("/v1/users", handleUsers)
	mux.HandleFunc("/v1/roles", handleRoles)
	mux.HandleFunc("/v1/permissions", handlePermissions)

	// Create server with optimized settings for high-performance authentication
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
		log.Printf("AuthN/AuthZ & Org/Tenant service listening on port %s", *port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down auth service...")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("AuthN/AuthZ & Org/Tenant service exited")
}

// handleHealth handles health check requests
func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `{"status":"healthy","service":"auth","timestamp":"%s"}`, time.Now().Format(time.RFC3339))
}

// handleAuth handles authentication requests
func handleAuth(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "POST":
		var req AuthenticateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}
		
		result, err := authService.Authenticate(r.Context(), &req)
		if err != nil {
			log.Printf("Failed to authenticate: %v", err)
			http.Error(w, "Authentication failed", http.StatusUnauthorized)
			return
		}
		
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(result)
	}
}

// handleToken handles token operations
func handleToken(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "POST":
		var req TokenRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}
		
		token, err := authService.GenerateToken(r.Context(), &req)
		if err != nil {
			log.Printf("Failed to generate token: %v", err)
			http.Error(w, "Token generation failed", http.StatusInternalServerError)
			return
		}
		
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(token)
		
	case "GET":
		// Validate token
		token := r.Header.Get("Authorization")
		if token == "" {
			http.Error(w, "No token provided", http.StatusUnauthorized)
			return
		}
		
		claims, err := authService.ValidateToken(r.Context(), token)
		if err != nil {
			log.Printf("Failed to validate token: %v", err)
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}
		
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(claims)
	}
}

// handleOrgs handles organization management
func handleOrgs(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		orgs, err := authService.ListOrganizations(r.Context(), OrgFilters{})
		if err != nil {
			log.Printf("Failed to list organizations: %v", err)
			http.Error(w, "Failed to list organizations", http.StatusInternalServerError)
			return
		}
		
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(orgs)
		
	case "POST":
		var req CreateOrgRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}
		
		org, err := authService.CreateOrganization(r.Context(), &req)
		if err != nil {
			log.Printf("Failed to create organization: %v", err)
			http.Error(w, "Failed to create organization", http.StatusInternalServerError)
			return
		}
		
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(org)
	}
}

// handleUsers handles user management
func handleUsers(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		users, err := authService.ListUsers(r.Context(), UserFilters{})
		if err != nil {
			log.Printf("Failed to list users: %v", err)
			http.Error(w, "Failed to list users", http.StatusInternalServerError)
			return
		}
		
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(users)
		
	case "POST":
		var req CreateUserRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}
		
		user, err := authService.CreateUser(r.Context(), &req)
		if err != nil {
			log.Printf("Failed to create user: %v", err)
			http.Error(w, "Failed to create user", http.StatusInternalServerError)
			return
		}
		
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(user)
	}
}

// handleRoles handles role management
func handleRoles(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		roles, err := authService.ListRoles(r.Context(), RoleFilters{})
		if err != nil {
			log.Printf("Failed to list roles: %v", err)
			http.Error(w, "Failed to list roles", http.StatusInternalServerError)
			return
		}
		
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(roles)
	}
}

// handlePermissions handles permission checks
func handlePermissions(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "POST":
		var req CheckPermissionRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}
		
		result, err := authService.CheckPermission(r.Context(), &req)
		if err != nil {
			log.Printf("Failed to check permission: %v", err)
			http.Error(w, "Permission check failed", http.StatusInternalServerError)
			return
		}
		
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(result)
	}
}
