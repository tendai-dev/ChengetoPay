package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	_ "github.com/lib/pq"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Service struct {
	postgresDB *sql.DB
	mongoDB    *mongo.Database
	redisClient *redis.Client
}

func main() {
	// Load environment variables
	postgresURL := os.Getenv("POSTGRES_URL")
	if postgresURL == "" {
		postgresURL = "postgresql://localhost/chengetopay?sslmode=disable"
	}
	
	mongoURL := os.Getenv("MONGODB_URL")
	if mongoURL == "" {
		mongoURL = "mongodb://localhost:27017"
	}
	
	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		redisURL = "redis://localhost:6379"
	}

	// Initialize service
	service := &Service{}
	
	// Connect to PostgreSQL
	var err error
	service.postgresDB, err = sql.Open("postgres", postgresURL)
	if err != nil {
		log.Fatalf("Failed to connect to PostgreSQL: %v", err)
	}
	defer service.postgresDB.Close()
	
	// Configure connection pool
	service.postgresDB.SetMaxOpenConns(25)
	service.postgresDB.SetMaxIdleConns(5)
	service.postgresDB.SetConnMaxLifetime(5 * time.Minute)
	
	// Test PostgreSQL connection with retry
	for i := 0; i < 3; i++ {
		if err := service.postgresDB.Ping(); err != nil {
			if i == 2 {
				log.Printf("Warning: PostgreSQL connection failed after 3 attempts: %v", err)
				log.Println("⚠️  Continuing without PostgreSQL (degraded mode)")
				break
			}
			log.Printf("PostgreSQL ping attempt %d failed: %v, retrying...", i+1, err)
			time.Sleep(2 * time.Second)
		} else {
			log.Println("✅ Connected to PostgreSQL")
			break
		}
	}
	
	// Connect to MongoDB
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	
	mongoClient, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoURL))
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	defer mongoClient.Disconnect(context.Background())
	
	// Test MongoDB connection
	if err := mongoClient.Ping(ctx, nil); err != nil {
		log.Fatalf("Failed to ping MongoDB: %v", err)
	}
	service.mongoDB = mongoClient.Database("chengetopay")
	log.Println("✅ Connected to MongoDB")
	
	// Connect to Redis
	opt, err := redis.ParseURL(redisURL)
	if err != nil {
		log.Fatalf("Failed to parse Redis URL: %v", err)
	}
	
	service.redisClient = redis.NewClient(opt)
	
	// Test Redis connection
	ctx = context.Background()
	if err := service.redisClient.Ping(ctx).Err(); err != nil {
		log.Fatalf("Failed to ping Redis: %v", err)
	}
	log.Println("✅ Connected to Redis")
	
	// Setup HTTP routes
	http.HandleFunc("/health", service.healthHandler)
	http.HandleFunc("/api/v1/test", service.testHandler)
	
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	
	log.Printf("🚀 Example service starting on port %s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func (s *Service) healthHandler(w http.ResponseWriter, r *http.Request) {
	health := map[string]interface{}{
		"status": "healthy",
		"timestamp": time.Now().Unix(),
		"databases": map[string]bool{
			"postgres": s.postgresDB.Ping() == nil,
			"mongodb": s.mongoDB.Client().Ping(context.Background(), nil) == nil,
			"redis": s.redisClient.Ping(context.Background()).Err() == nil,
		},
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(health)
}

func (s *Service) testHandler(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	
	// Test PostgreSQL
	var pgVersion string
	err := s.postgresDB.QueryRow("SELECT version()").Scan(&pgVersion)
	if err != nil {
		http.Error(w, fmt.Sprintf("PostgreSQL error: %v", err), http.StatusInternalServerError)
		return
	}
	
	// Test MongoDB
	mongoResult := s.mongoDB.Client().Ping(ctx, nil)
	
	// Test Redis
	redisResult := s.redisClient.Set(ctx, "test:key", "test-value", 10*time.Second).Err()
	redisValue, _ := s.redisClient.Get(ctx, "test:key").Result()
	
	response := map[string]interface{}{
		"postgres": map[string]interface{}{
			"connected": true,
			"version": pgVersion,
		},
		"mongodb": map[string]interface{}{
			"connected": mongoResult == nil,
		},
		"redis": map[string]interface{}{
			"connected": redisResult == nil,
			"test_value": redisValue,
		},
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
