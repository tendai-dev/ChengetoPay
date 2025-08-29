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

	"database"
)

// DatabaseService represents the database service
type DatabaseService struct {
	postgres *database.PostgresDB
	mongodb  *database.MongoDB
	redis    *database.RedisDB
}

// Global service instance
var dbService *DatabaseService

func main() {
	port := flag.String("port", "8115", "Port to listen on")
	flag.Parse()

	log.Printf("Starting Database Initialization Service on port %s...", *port)

	// Initialize databases
	if err := initializeDatabases(); err != nil {
		log.Fatalf("Failed to initialize databases: %v", err)
	}

	// Create HTTP server
	mux := http.NewServeMux()
	mux.HandleFunc("/health", handleHealth)
	mux.HandleFunc("/v1/status", handleStatus)
	mux.HandleFunc("/v1/migrate", handleMigrate)
	mux.HandleFunc("/v1/seed", handleSeed)

	server := &http.Server{
		Addr:         ":" + *port,
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	// Start server
	go func() {
		log.Printf("Database service listening on port %s", *port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down database service...")
	
	// Close database connections
	if dbService != nil {
		if dbService.postgres != nil {
			dbService.postgres.Close()
		}
		if dbService.mongodb != nil {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			dbService.mongodb.Close(ctx)
		}
		if dbService.redis != nil {
			dbService.redis.Close()
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Database service exited")
}

// initializeDatabases initializes all database connections
func initializeDatabases() error {
	// PostgreSQL configuration
	postgresConfig := database.PostgresConfig{
		URL:                "postgresql://neondb_owner:npg_6oAPnbj5zIKN@ep-wispy-union-adi5og8a-pooler.c-2.us-east-1.aws.neon.tech/neondb?sslmode=require&channel_binding=require",
		MaxOpenConnections: 25,
		MaxIdleConnections: 5,
		ConnMaxLifetime:    5 * time.Minute,
	}

	// MongoDB configuration
	mongoConfig := database.MongoConfig{
		URL:                   "mongodb+srv://tendai_db_user:aEmut0m48FtaES1E@cluster0.csdtbuo.mongodb.net/?retryWrites=true&w=majority&appName=Cluster0",
		Database:              "financial_platform",
		MaxPoolSize:           100,
		MinPoolSize:           5,
		MaxConnIdleTime:       30 * time.Second,
		ServerSelectionTimeout: 5 * time.Second,
	}

	// Redis configuration (using Redis Cloud or local Redis)
	redisConfig := database.RedisConfig{
		URL:          "redis://localhost:6379", // You'll need to set up Redis
		PoolSize:     10,
		MinIdleConns: 2,
		MaxRetries:   3,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
	}

	// Initialize PostgreSQL
	postgres, err := database.NewPostgresDB(postgresConfig)
	if err != nil {
		return fmt.Errorf("failed to initialize PostgreSQL: %w", err)
	}

	// Run PostgreSQL migrations
	if err := postgres.RunMigrations(); err != nil {
		return fmt.Errorf("failed to run PostgreSQL migrations: %w", err)
	}

	// Create PostgreSQL indexes
	if err := postgres.CreateIndexes(); err != nil {
		return fmt.Errorf("failed to create PostgreSQL indexes: %w", err)
	}

	// Initialize MongoDB
	mongodb, err := database.NewMongoDB(mongoConfig)
	if err != nil {
		return fmt.Errorf("failed to initialize MongoDB: %w", err)
	}

	// Create MongoDB collections and indexes
	if err := mongodb.CreateCollections(); err != nil {
		return fmt.Errorf("failed to create MongoDB collections: %w", err)
	}

	// Create MongoDB TTL indexes
	if err := mongodb.CreateTTLIndexes(); err != nil {
		return fmt.Errorf("failed to create MongoDB TTL indexes: %w", err)
	}

	// Initialize Redis (skip if not available)
	redis, err := database.NewRedisDB(redisConfig)
	if err != nil {
		log.Printf("Warning: Redis not available: %v", err)
		redis = nil
	}

	dbService = &DatabaseService{
		postgres: postgres,
		mongodb:  mongodb,
		redis:    redis,
	}

	log.Println("✅ All databases initialized successfully")
	return nil
}

// handleHealth handles health check requests
func handleHealth(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	health := map[string]interface{}{
		"status":    "healthy",
		"service":   "database",
		"timestamp": time.Now().Format(time.RFC3339),
		"databases": map[string]string{},
	}

	// Check PostgreSQL
	if dbService.postgres != nil {
		if err := dbService.postgres.HealthCheck(ctx); err != nil {
			health["databases"].(map[string]string)["postgresql"] = "unhealthy"
			health["status"] = "unhealthy"
		} else {
			health["databases"].(map[string]string)["postgresql"] = "healthy"
		}
	}

	// Check MongoDB
	if dbService.mongodb != nil {
		if err := dbService.mongodb.HealthCheck(ctx); err != nil {
			health["databases"].(map[string]string)["mongodb"] = "unhealthy"
			health["status"] = "unhealthy"
		} else {
			health["databases"].(map[string]string)["mongodb"] = "healthy"
		}
	}

	// Check Redis
	if dbService.redis != nil {
		if err := dbService.redis.HealthCheck(ctx); err != nil {
			health["databases"].(map[string]string)["redis"] = "unhealthy"
			health["status"] = "unhealthy"
		} else {
			health["databases"].(map[string]string)["redis"] = "healthy"
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(health)
}

// handleStatus handles database status requests
func handleStatus(w http.ResponseWriter, r *http.Request) {
	status := map[string]interface{}{
		"postgresql": map[string]interface{}{
			"connected": dbService.postgres != nil,
			"type":      "PostgreSQL",
			"purpose":   "Transactional data, financial records",
		},
		"mongodb": map[string]interface{}{
			"connected": dbService.mongodb != nil,
			"type":      "MongoDB",
			"purpose":   "Document storage, audit logs, evidence",
		},
		"redis": map[string]interface{}{
			"connected": dbService.redis != nil,
			"type":      "Redis",
			"purpose":   "Caching, sessions, distributed locking",
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}

// handleMigrate handles migration requests
func handleMigrate(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	response := map[string]interface{}{
		"status":  "success",
		"message": "Migrations completed",
		"details": map[string]string{},
	}

	// Run PostgreSQL migrations
	if dbService.postgres != nil {
		if err := dbService.postgres.RunMigrations(); err != nil {
			response["details"].(map[string]string)["postgresql"] = "failed: " + err.Error()
			response["status"] = "error"
		} else {
			response["details"].(map[string]string)["postgresql"] = "completed"
		}
	}

	// Create MongoDB collections
	if dbService.mongodb != nil {
		if err := dbService.mongodb.CreateCollections(); err != nil {
			response["details"].(map[string]string)["mongodb"] = "failed: " + err.Error()
			response["status"] = "error"
		} else {
			response["details"].(map[string]string)["mongodb"] = "completed"
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleSeed handles database seeding requests
func handleSeed(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	response := map[string]interface{}{
		"status":  "success",
		"message": "Database seeded",
		"details": map[string]string{},
	}

	// Seed PostgreSQL with initial data
	if dbService.postgres != nil {
		if err := seedPostgreSQL(); err != nil {
			response["details"].(map[string]string)["postgresql"] = "failed: " + err.Error()
			response["status"] = "error"
		} else {
			response["details"].(map[string]string)["postgresql"] = "completed"
		}
	}

	// Seed MongoDB with initial data
	if dbService.mongodb != nil {
		if err := seedMongoDB(); err != nil {
			response["details"].(map[string]string)["mongodb"] = "failed: " + err.Error()
			response["status"] = "error"
		} else {
			response["details"].(map[string]string)["mongodb"] = "completed"
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// seedPostgreSQL seeds PostgreSQL with initial data
func seedPostgreSQL() error {
	// Add sample organizations
	organizations := []string{
		`INSERT INTO organizations (id, name, domain, status, plan) VALUES 
		('org_demo_1', 'Demo Organization 1', 'demo1.com', 'active', 'pro')
		ON CONFLICT (id) DO NOTHING`,
		
		`INSERT INTO organizations (id, name, domain, status, plan) VALUES 
		('org_demo_2', 'Demo Organization 2', 'demo2.com', 'active', 'basic')
		ON CONFLICT (id) DO NOTHING`,
	}

	// Add sample users
	users := []string{
		`INSERT INTO users (id, org_id, email, first_name, last_name, status) VALUES 
		('user_demo_1', 'org_demo_1', 'admin@demo1.com', 'John', 'Doe', 'active')
		ON CONFLICT (id) DO NOTHING`,
		
		`INSERT INTO users (id, org_id, email, first_name, last_name, status) VALUES 
		('user_demo_2', 'org_demo_2', 'admin@demo2.com', 'Jane', 'Smith', 'active')
		ON CONFLICT (id) DO NOTHING`,
	}

	// Add sample roles
	roles := []string{
		`INSERT INTO roles (id, org_id, name, description, permissions) VALUES 
		('role_admin', 'org_demo_1', 'Administrator', 'Full system access', '["read", "write", "admin"]')
		ON CONFLICT (id) DO NOTHING`,
		
		`INSERT INTO roles (id, org_id, name, description, permissions) VALUES 
		('role_user', 'org_demo_1', 'User', 'Basic user access', '["read", "write"]')
		ON CONFLICT (id) DO NOTHING`,
	}

	// Add sample feature flags
	featureFlags := []string{
		`INSERT INTO feature_flags (id, name, description, enabled, rollout_percentage) VALUES 
		('new_payment_flow', 'New Payment Flow', 'Enable new payment processing flow', true, 50)
		ON CONFLICT (id) DO NOTHING`,
		
		`INSERT INTO feature_flags (id, name, description, enabled, rollout_percentage) VALUES 
		('advanced_analytics', 'Advanced Analytics', 'Enable advanced analytics dashboard', false, 0)
		ON CONFLICT (id) DO NOTHING`,
	}

	// Execute all seed queries
	allQueries := append(organizations, append(users, append(roles, featureFlags...)...)...)

	for i, query := range allQueries {
		if _, err := dbService.postgres.GetDB().Exec(query); err != nil {
			return fmt.Errorf("seed query %d failed: %w", i+1, err)
		}
	}

	log.Println("✅ PostgreSQL seeded successfully")
	return nil
}

// seedMongoDB seeds MongoDB with initial data
func seedMongoDB() error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Sample audit logs
	auditLogs := []interface{}{
		map[string]interface{}{
			"org_id":    "org_demo_1",
			"user_id":   "user_demo_1",
			"action":    "user.login",
			"timestamp": time.Now(),
			"details":   map[string]interface{}{"ip": "192.168.1.1"},
		},
		map[string]interface{}{
			"org_id":    "org_demo_1",
			"user_id":   "user_demo_1",
			"action":    "payment.create",
			"timestamp": time.Now(),
			"details":   map[string]interface{}{"amount": 100.00, "currency": "USD"},
		},
	}

	// Insert audit logs
	auditCollection := dbService.mongodb.GetDatabase().Collection("audit_logs")
	for _, log := range auditLogs {
		if _, err := auditCollection.InsertOne(ctx, log); err != nil {
			return fmt.Errorf("failed to insert audit log: %w", err)
		}
	}

	// Sample performance metrics
	metrics := []interface{}{
		map[string]interface{}{
			"service":   "payment",
			"timestamp": time.Now(),
			"response_time": 15.5,
			"requests_per_second": 1000,
		},
		map[string]interface{}{
			"service":   "escrow",
			"timestamp": time.Now(),
			"response_time": 12.3,
			"requests_per_second": 500,
		},
	}

	// Insert performance metrics
	metricsCollection := dbService.mongodb.GetDatabase().Collection("performance_metrics")
	for _, metric := range metrics {
		if _, err := metricsCollection.InsertOne(ctx, metric); err != nil {
			return fmt.Errorf("failed to insert performance metric: %w", err)
		}
	}

	log.Println("✅ MongoDB seeded successfully")
	return nil
}
