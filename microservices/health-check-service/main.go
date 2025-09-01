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

type HealthService struct {
	postgresDB  *sql.DB
	mongoDB     *mongo.Database
	redisClient *redis.Client
	pgStatus    string
	mongoStatus string
	redisStatus string
}

func main() {
	service := &HealthService{
		pgStatus:    "initializing",
		mongoStatus: "initializing",
		redisStatus: "initializing",
	}

	// Try to connect to databases but don't fail if they're unavailable
	go service.initPostgreSQL()
	go service.initMongoDB()
	go service.initRedis()

	// Setup HTTP routes
	http.HandleFunc("/health", service.healthHandler)
	http.HandleFunc("/ready", service.readyHandler)
	http.HandleFunc("/api/v1/status", service.statusHandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("🚀 Health Check Service starting on port %s", port)
	log.Println("📊 Service will run in degraded mode if databases are unavailable")
	
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func (s *HealthService) initPostgreSQL() {
	postgresURL := os.Getenv("POSTGRES_URL")
	if postgresURL == "" {
		postgresURL = "postgresql://localhost/chengetopay?sslmode=disable"
	}

	var err error
	s.postgresDB, err = sql.Open("postgres", postgresURL)
	if err != nil {
		s.pgStatus = fmt.Sprintf("error: %v", err)
		log.Printf("PostgreSQL initialization failed: %v", err)
		return
	}

	s.postgresDB.SetMaxOpenConns(5)
	s.postgresDB.SetMaxIdleConns(2)
	s.postgresDB.SetConnMaxLifetime(5 * time.Minute)

	// Periodic health check
	ticker := time.NewTicker(30 * time.Second)
	go func() {
		for range ticker.C {
			if err := s.postgresDB.Ping(); err != nil {
				s.pgStatus = fmt.Sprintf("unhealthy: %v", err)
			} else {
				s.pgStatus = "healthy"
			}
		}
	}()

	// Initial check
	if err := s.postgresDB.Ping(); err != nil {
		s.pgStatus = fmt.Sprintf("unhealthy: %v", err)
		log.Printf("PostgreSQL initial ping failed: %v", err)
	} else {
		s.pgStatus = "healthy"
		log.Println("✅ PostgreSQL connected")
	}
}

func (s *HealthService) initMongoDB() {
	mongoURL := os.Getenv("MONGODB_URL")
	if mongoURL == "" {
		mongoURL = "mongodb://localhost:27017"
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	clientOptions := options.Client().ApplyURI(mongoURL).SetServerSelectionTimeout(10 * time.Second)
	mongoClient, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		s.mongoStatus = fmt.Sprintf("error: %v", err)
		log.Printf("MongoDB initialization failed: %v", err)
		return
	}

	s.mongoDB = mongoClient.Database("chengetopay")

	// Periodic health check
	ticker := time.NewTicker(30 * time.Second)
	go func() {
		for range ticker.C {
			ctx := context.Background()
			if err := mongoClient.Ping(ctx, nil); err != nil {
				s.mongoStatus = fmt.Sprintf("unhealthy: %v", err)
			} else {
				s.mongoStatus = "healthy"
			}
		}
	}()

	// Initial check
	if err := mongoClient.Ping(ctx, nil); err != nil {
		s.mongoStatus = fmt.Sprintf("unhealthy: %v", err)
		log.Printf("MongoDB initial ping failed: %v", err)
	} else {
		s.mongoStatus = "healthy"
		log.Println("✅ MongoDB connected")
	}
}

func (s *HealthService) initRedis() {
	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		redisURL = "redis://localhost:6379"
	}

	opt, err := redis.ParseURL(redisURL)
	if err != nil {
		s.redisStatus = fmt.Sprintf("error: %v", err)
		log.Printf("Redis URL parse failed: %v", err)
		return
	}

	s.redisClient = redis.NewClient(opt)

	// Periodic health check
	ticker := time.NewTicker(30 * time.Second)
	go func() {
		for range ticker.C {
			ctx := context.Background()
			if err := s.redisClient.Ping(ctx).Err(); err != nil {
				s.redisStatus = fmt.Sprintf("unhealthy: %v", err)
			} else {
				s.redisStatus = "healthy"
			}
		}
	}()

	// Initial check
	ctx := context.Background()
	if err := s.redisClient.Ping(ctx).Err(); err != nil {
		s.redisStatus = fmt.Sprintf("unhealthy: %v", err)
		log.Printf("Redis initial ping failed: %v", err)
	} else {
		s.redisStatus = "healthy"
		log.Println("✅ Redis connected")
	}
}

func (s *HealthService) healthHandler(w http.ResponseWriter, r *http.Request) {
	// Basic health check - service is running
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":    "ok",
		"timestamp": time.Now().Unix(),
		"service":   "health-check-service",
	})
}

func (s *HealthService) readyHandler(w http.ResponseWriter, r *http.Request) {
	// Readiness check - at least one database should be healthy
	isReady := s.pgStatus == "healthy" || s.mongoStatus == "healthy" || s.redisStatus == "healthy"
	
	response := map[string]interface{}{
		"ready":     isReady,
		"timestamp": time.Now().Unix(),
		"databases": map[string]string{
			"postgresql": s.pgStatus,
			"mongodb":    s.mongoStatus,
			"redis":      s.redisStatus,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	if !isReady {
		w.WriteHeader(http.StatusServiceUnavailable)
	}
	json.NewEncoder(w).Encode(response)
}

func (s *HealthService) statusHandler(w http.ResponseWriter, r *http.Request) {
	// Detailed status
	status := map[string]interface{}{
		"service":   "health-check-service",
		"version":   "1.0.0",
		"timestamp": time.Now().Unix(),
		"uptime":    time.Since(startTime).Seconds(),
		"databases": map[string]interface{}{
			"postgresql": map[string]interface{}{
				"status":      s.pgStatus,
				"configured":  os.Getenv("POSTGRES_URL") != "",
				"description": "Primary transactional database",
			},
			"mongodb": map[string]interface{}{
				"status":      s.mongoStatus,
				"configured":  os.Getenv("MONGODB_URL") != "",
				"description": "Document store for unstructured data",
			},
			"redis": map[string]interface{}{
				"status":      s.redisStatus,
				"configured":  os.Getenv("REDIS_URL") != "",
				"description": "Cache and session store",
			},
		},
		"environment": map[string]interface{}{
			"go_version": "1.21",
			"port":       os.Getenv("PORT"),
			"mode":       getMode(),
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}

var startTime = time.Now()

func getMode() string {
	if os.Getenv("POSTGRES_URL") != "" || os.Getenv("MONGODB_URL") != "" || os.Getenv("REDIS_URL") != "" {
		return "production"
	}
	return "development"
}
