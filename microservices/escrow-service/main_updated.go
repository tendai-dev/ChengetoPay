package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/chengetopay/shared/database"
	"github.com/chengetopay/shared/health"
	"github.com/chengetopay/shared/tracing"
	"github.com/chengetopay/shared/errors"
	"github.com/gorilla/mux"
	"github.com/google/uuid"
)

type EscrowService struct {
	db     *database.Config
	health *health.Checker
}

type Escrow struct {
	ID        string    `json:"id"`
	BuyerID   string    `json:"buyer_id"`
	SellerID  string    `json:"seller_id"`
	Amount    float64   `json:"amount"`
	Currency  string    `json:"currency"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func NewEscrowService() *EscrowService {
	// Initialize database configuration
	dbConfig := database.NewConfig("escrow-service", "escrow")
	
	// Initialize health checker
	healthChecker := health.NewChecker("escrow-service", "1.0.0")
	
	// Connect to PostgreSQL
	postgresDB, err := dbConfig.PostgresConnection()
	if err != nil {
		log.Printf("Warning: PostgreSQL connection failed: %v", err)
	} else {
		healthChecker.PostgresDB = postgresDB
	}
	
	// Connect to Redis
	redisClient, err := dbConfig.RedisConnection()
	if err != nil {
		log.Printf("Warning: Redis connection failed: %v", err)
	} else {
		healthChecker.RedisClient = redisClient
	}
	
	return &EscrowService{
		db:     dbConfig,
		health: healthChecker,
	}
}

func (s *EscrowService) CreateEscrow(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracing.StartSpan(r.Context(), "CreateEscrow")
	defer span.End()
	
	var escrow Escrow
	if err := json.NewDecoder(r.Body).Decode(&escrow); err != nil {
		errors.NewValidationError("Invalid request body").Send(w)
		return
	}
	
	// Validate required fields
	if escrow.BuyerID == "" || escrow.SellerID == "" || escrow.Amount <= 0 {
		errors.NewValidationError("Missing required fields").Send(w)
		return
	}
	
	// Generate ID and timestamps
	escrow.ID = uuid.New().String()
	escrow.Status = "pending"
	escrow.CreatedAt = time.Now()
	escrow.UpdatedAt = time.Now()
	
	// TODO: Save to database when connection string is provided
	
	// Add event to trace
	tracing.AddEvent(ctx, "Escrow created", 
		"escrow.id", escrow.ID,
		"escrow.amount", escrow.Amount,
	)
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(escrow)
}

func (s *EscrowService) GetEscrow(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracing.StartSpan(r.Context(), "GetEscrow")
	defer span.End()
	
	vars := mux.Vars(r)
	escrowID := vars["id"]
	
	if escrowID == "" {
		errors.NewValidationError("Escrow ID required").Send(w)
		return
	}
	
	// TODO: Fetch from database when connection string is provided
	
	// Mock response for now
	escrow := Escrow{
		ID:        escrowID,
		BuyerID:   "buyer-123",
		SellerID:  "seller-456",
		Amount:    1000.00,
		Currency:  "USD",
		Status:    "pending",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(escrow)
}

func (s *EscrowService) ListEscrows(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracing.StartSpan(r.Context(), "ListEscrows")
	defer span.End()
	
	// TODO: Implement pagination and filtering
	
	escrows := []Escrow{
		{
			ID:        uuid.New().String(),
			BuyerID:   "buyer-123",
			SellerID:  "seller-456",
			Amount:    1000.00,
			Currency:  "USD",
			Status:    "pending",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"escrows": escrows,
		"total":   len(escrows),
	})
}

func main() {
	// Initialize tracing
	cleanup, err := tracing.InitTracer("escrow-service")
	if err != nil {
		log.Printf("Failed to initialize tracer: %v", err)
	} else {
		defer cleanup()
	}
	
	// Create service
	service := NewEscrowService()
	
	// Setup routes
	router := mux.NewRouter()
	
	// Health check
	router.HandleFunc("/health", service.health.Handler()).Methods("GET")
	
	// Escrow endpoints
	router.HandleFunc("/api/v1/escrows", service.CreateEscrow).Methods("POST")
	router.HandleFunc("/api/v1/escrows", service.ListEscrows).Methods("GET")
	router.HandleFunc("/api/v1/escrows/{id}", service.GetEscrow).Methods("GET")
	
	// Metrics endpoint
	router.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("# Metrics endpoint ready\n"))
	}).Methods("GET")
	
	// Server configuration
	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}
	
	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}
	
	// Start server in goroutine
	go func() {
		log.Printf("🚀 Escrow Service starting on port %s", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()
	
	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	
	log.Println("Shutting down server...")
	
	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}
	
	log.Println("Server shutdown complete")
}
