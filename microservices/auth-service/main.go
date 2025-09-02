package main

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"
)

type User struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	Username  string    `json:"username"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"created_at"`
}

type AuthResponse struct {
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
	User      User      `json:"user"`
}

func generateToken() string {
	b := make([]byte, 32)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)
}

func main() {
	log.Println("Starting Auth Service on port 8103...")

	// Health endpoint
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"status":"healthy","service":"auth","timestamp":"%s"}`, time.Now().Format(time.RFC3339))
	})

	// Status endpoint
	http.HandleFunc("/v1/status", func(w http.ResponseWriter, r *http.Request) {
		status := map[string]interface{}{
			"service": "auth",
			"status":  "active",
			"version": "1.0.0",
			"features": []string{
				"jwt_authentication",
				"oauth2",
				"api_keys",
				"role_based_access",
				"multi_factor_auth",
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(status)
	})

	// Login endpoint
	http.HandleFunc("/api/v1/auth/login", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var request struct {
			Email    string `json:"email"`
			Password string `json:"password"`
		}

		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			http.Error(w, "Invalid request", http.StatusBadRequest)
			return
		}

		// Mock authentication (in production, verify against database)
		if request.Email == "" || request.Password == "" {
			http.Error(w, "Invalid credentials", http.StatusUnauthorized)
			return
		}

		user := User{
			ID:        fmt.Sprintf("user_%d", time.Now().Unix()),
			Email:     request.Email,
			Username:  strings.Split(request.Email, "@")[0],
			Role:      "user",
			CreatedAt: time.Now(),
		}

		response := AuthResponse{
			Token:     generateToken(),
			ExpiresAt: time.Now().Add(24 * time.Hour),
			User:      user,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	})

	// Register endpoint
	http.HandleFunc("/api/v1/auth/register", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var request struct {
			Email    string `json:"email"`
			Password string `json:"password"`
			Username string `json:"username"`
		}

		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			http.Error(w, "Invalid request", http.StatusBadRequest)
			return
		}

		user := User{
			ID:        fmt.Sprintf("user_%d", time.Now().Unix()),
			Email:     request.Email,
			Username:  request.Username,
			Role:      "user",
			CreatedAt: time.Now(),
		}

		response := map[string]interface{}{
			"message": "User registered successfully",
			"user":    user,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(response)
	})

	// Verify token endpoint
	http.HandleFunc("/api/v1/auth/verify", func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Missing authorization header", http.StatusUnauthorized)
			return
		}

		token := strings.TrimPrefix(authHeader, "Bearer ")
		if token == authHeader {
			http.Error(w, "Invalid token format", http.StatusUnauthorized)
			return
		}

		// Mock token verification (in production, verify JWT signature)
		response := map[string]interface{}{
			"valid":     true,
			"user_id":   "user_123",
			"role":      "user",
			"expires_at": time.Now().Add(24 * time.Hour),
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	})

	// Refresh token endpoint
	http.HandleFunc("/api/v1/auth/refresh", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var request struct {
			RefreshToken string `json:"refresh_token"`
		}

		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			http.Error(w, "Invalid request", http.StatusBadRequest)
			return
		}

		response := map[string]interface{}{
			"token":      generateToken(),
			"expires_at": time.Now().Add(24 * time.Hour),
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	})

	// Logout endpoint
	http.HandleFunc("/api/v1/auth/logout", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		response := map[string]interface{}{
			"message": "Logged out successfully",
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	})

	log.Println("Auth service listening on port 8103")
	if err := http.ListenAndServe(":8103", nil); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
