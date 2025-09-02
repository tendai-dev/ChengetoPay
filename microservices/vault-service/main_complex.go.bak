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

	"vault"
)

// VaultService represents the Vault service
type VaultService struct {
	vaultClient *vault.VaultClient
}

// Global service instance
var vaultService *VaultService

func main() {
	port := flag.String("port", "8119", "Port to listen on")
	vaultAddr := flag.String("vault-addr", "http://localhost:8200", "Vault address")
	vaultToken := flag.String("vault-token", "dev-token", "Vault token")
	mountPath := flag.String("mount-path", "secret", "Vault mount path")
	flag.Parse()

	log.Printf("Starting Vault Service on port %s...", *port)

	// Initialize Vault service
	if err := initializeVaultService(*vaultAddr, *vaultToken, *mountPath); err != nil {
		log.Fatalf("Failed to initialize Vault service: %v", err)
	}

	// Create HTTP server
	mux := http.NewServeMux()
	mux.HandleFunc("/health", handleHealth)
	mux.HandleFunc("/v1/status", handleStatus)
	mux.HandleFunc("/v1/secrets", handleSecrets)
	mux.HandleFunc("/v1/database", handleDatabase)
	mux.HandleFunc("/v1/api", handleAPI)
	mux.HandleFunc("/v1/certificates", handleCertificates)
	mux.HandleFunc("/v1/generate", handleGenerate)

	server := &http.Server{
		Addr:           ":" + *port,
		Handler:        mux,
		ReadTimeout:    5 * time.Second,
		WriteTimeout:   10 * time.Second,
		IdleTimeout:    120 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	// Start server
	go func() {
		log.Printf("Vault service listening on port %s", *port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down Vault service...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Vault service exited")
}

// initializeVaultService initializes the Vault client
func initializeVaultService(vaultAddr, vaultToken, mountPath string) error {
	config := vault.VaultConfig{
		Address:    vaultAddr,
		Token:      vaultToken,
		MountPath:  mountPath,
		Timeout:    10 * time.Second,
		MaxRetries: 3,
	}

	vaultClient, err := vault.NewVaultClient(config)
	if err != nil {
		return fmt.Errorf("failed to initialize Vault client: %w", err)
	}

	vaultService = &VaultService{
		vaultClient: vaultClient,
	}

	// Initialize default secrets
	if err := initializeDefaultSecrets(); err != nil {
		log.Printf("Warning: Failed to initialize default secrets: %v", err)
	}

	log.Println("✅ Vault service initialized successfully")
	return nil
}

// initializeDefaultSecrets initializes default secrets for the platform
func initializeDefaultSecrets() error {
	ctx := context.Background()

	// Store database credentials
	dbCred := vault.DatabaseCredential{
		Username: "neondb_owner",
		Password: "npg_6oAPnbj5zIKN",
		Host:     "ep-wispy-union-adi5og8a-pooler.c-2.us-east-1.aws.neon.tech",
		Port:     5432,
		Database: "neondb",
		SSLMode:  "require",
	}

	if err := vaultService.vaultClient.StoreDatabaseCredential(ctx, "postgresql", dbCred); err != nil {
		return fmt.Errorf("failed to store PostgreSQL credentials: %w", err)
	}

	// Store MongoDB credentials
	mongoCred := vault.DatabaseCredential{
		Username: "tendai_db_user",
		Password: "aEmut0m48FtaES1E",
		Host:     "cluster0.csdtbuo.mongodb.net",
		Port:     27017,
		Database: "financial_platform",
		SSLMode:  "require",
	}

	if err := vaultService.vaultClient.StoreDatabaseCredential(ctx, "mongodb", mongoCred); err != nil {
		return fmt.Errorf("failed to store MongoDB credentials: %w", err)
	}

	// Store API credentials
	apiCred := vault.APICredential{
		APIKey:    "sk_test_1234567890abcdef",
		APISecret: "sk_test_secret_abcdef1234567890",
		Endpoint:  "https://api.stripe.com",
		Version:   "2023-10-16",
	}

	if err := vaultService.vaultClient.StoreAPICredential(ctx, "stripe", apiCred); err != nil {
		return fmt.Errorf("failed to store Stripe API credentials: %w", err)
	}

	// Store RabbitMQ credentials
	rabbitmqCred := vault.APICredential{
		APIKey:    "guest",
		APISecret: "guest",
		Endpoint:  "amqp://localhost:5672",
		Version:   "3.9",
	}

	if err := vaultService.vaultClient.StoreAPICredential(ctx, "rabbitmq", rabbitmqCred); err != nil {
		return fmt.Errorf("failed to store RabbitMQ credentials: %w", err)
	}

	log.Println("✅ Default secrets initialized")
	return nil
}

// handleHealth handles health check requests
func handleHealth(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	if err := vaultService.vaultClient.HealthCheck(ctx); err != nil {
		http.Error(w, fmt.Sprintf("Vault health check failed: %v", err), http.StatusServiceUnavailable)
		return
	}

	health := map[string]interface{}{
		"status":    "healthy",
		"service":   "vault",
		"timestamp": time.Now().Format(time.RFC3339),
		"vault":     "connected",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(health)
}

// handleStatus handles Vault status requests
func handleStatus(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	status, err := vaultService.vaultClient.GetVaultStatus(ctx)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get Vault status: %v", err), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"status": "success",
		"data":   status,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleSecrets handles secret operations
func handleSecrets(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		handleGetSecrets(w, r)
	case "POST":
		handleStoreSecret(w, r)
	case "DELETE":
		handleDeleteSecret(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleGetSecrets handles secret retrieval requests
func handleGetSecrets(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	path := query.Get("path")

	if path == "" {
		http.Error(w, "Path parameter is required", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	secret, err := vaultService.vaultClient.GetSecret(ctx, path)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get secret: %v", err), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"status": "success",
		"data":   secret,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleStoreSecret handles secret storage requests
func handleStoreSecret(w http.ResponseWriter, r *http.Request) {
	var request struct {
		Path string                 `json:"path"`
		Data map[string]interface{} `json:"data"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if request.Path == "" {
		http.Error(w, "Path is required", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	if err := vaultService.vaultClient.StoreSecret(ctx, request.Path, request.Data); err != nil {
		http.Error(w, fmt.Sprintf("Failed to store secret: %v", err), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"status":  "success",
		"message": "Secret stored successfully",
		"data": map[string]interface{}{
			"path": request.Path,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleDeleteSecret handles secret deletion requests
func handleDeleteSecret(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	path := query.Get("path")

	if path == "" {
		http.Error(w, "Path parameter is required", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	if err := vaultService.vaultClient.DeleteSecret(ctx, path); err != nil {
		http.Error(w, fmt.Sprintf("Failed to delete secret: %v", err), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"status":  "success",
		"message": "Secret deleted successfully",
		"data": map[string]interface{}{
			"path": path,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleDatabase handles database credential operations
func handleDatabase(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		handleGetDatabaseCredential(w, r)
	case "POST":
		handleStoreDatabaseCredential(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleGetDatabaseCredential handles database credential retrieval
func handleGetDatabaseCredential(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	name := query.Get("name")

	if name == "" {
		http.Error(w, "Name parameter is required", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	cred, err := vaultService.vaultClient.GetDatabaseCredential(ctx, name)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get database credential: %v", err), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"status": "success",
		"data":   cred,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleStoreDatabaseCredential handles database credential storage
func handleStoreDatabaseCredential(w http.ResponseWriter, r *http.Request) {
	var request struct {
		Name string                    `json:"name"`
		Cred vault.DatabaseCredential `json:"credential"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if request.Name == "" {
		http.Error(w, "Name is required", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	if err := vaultService.vaultClient.StoreDatabaseCredential(ctx, request.Name, request.Cred); err != nil {
		http.Error(w, fmt.Sprintf("Failed to store database credential: %v", err), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"status":  "success",
		"message": "Database credential stored successfully",
		"data": map[string]interface{}{
			"name": request.Name,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleAPI handles API credential operations
func handleAPI(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		handleGetAPICredential(w, r)
	case "POST":
		handleStoreAPICredential(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleGetAPICredential handles API credential retrieval
func handleGetAPICredential(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	name := query.Get("name")

	if name == "" {
		http.Error(w, "Name parameter is required", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	cred, err := vaultService.vaultClient.GetAPICredential(ctx, name)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get API credential: %v", err), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"status": "success",
		"data":   cred,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleStoreAPICredential handles API credential storage
func handleStoreAPICredential(w http.ResponseWriter, r *http.Request) {
	var request struct {
		Name string               `json:"name"`
		Cred vault.APICredential `json:"credential"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if request.Name == "" {
		http.Error(w, "Name is required", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	if err := vaultService.vaultClient.StoreAPICredential(ctx, request.Name, request.Cred); err != nil {
		http.Error(w, fmt.Sprintf("Failed to store API credential: %v", err), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"status":  "success",
		"message": "API credential stored successfully",
		"data": map[string]interface{}{
			"name": request.Name,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleCertificates handles certificate operations
func handleCertificates(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		handleGetCertificate(w, r)
	case "POST":
		handleStoreCertificate(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleGetCertificate handles certificate retrieval
func handleGetCertificate(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	name := query.Get("name")

	if name == "" {
		http.Error(w, "Name parameter is required", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	cert, err := vaultService.vaultClient.GetCertificate(ctx, name)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get certificate: %v", err), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"status": "success",
		"data":   cert,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleStoreCertificate handles certificate storage
func handleStoreCertificate(w http.ResponseWriter, r *http.Request) {
	var request struct {
		Name string              `json:"name"`
		Cert vault.Certificate `json:"certificate"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if request.Name == "" {
		http.Error(w, "Name is required", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	if err := vaultService.vaultClient.StoreCertificate(ctx, request.Name, request.Cert); err != nil {
		http.Error(w, fmt.Sprintf("Failed to store certificate: %v", err), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"status":  "success",
		"message": "Certificate stored successfully",
		"data": map[string]interface{}{
			"name": request.Name,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleGenerate handles password generation requests
func handleGenerate(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var request struct {
		Length int `json:"length"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if request.Length <= 0 {
		request.Length = 16 // Default length
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	password, err := vaultService.vaultClient.GeneratePassword(ctx, request.Length)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to generate password: %v", err), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"status": "success",
		"data": map[string]interface{}{
			"password": password,
			"length":   request.Length,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
