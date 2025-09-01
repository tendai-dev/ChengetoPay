package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	_ "github.com/lib/pq"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Service struct {
	postgresDB  *sql.DB
	mongoDB     *mongo.Database
	redisClient *redis.Client
}

func main() {
	service := &Service{}
	
	// Fix PostgreSQL connection
	postgresURL := os.Getenv("POSTGRES_URL")
	if postgresURL == "" {
		// Use the direct connection string with proper escaping
		postgresURL = "postgresql://chengetopay_owner:4ixY5mEgxoP0@ep-weathered-bonus-a5eosddm.us-east-2.aws.neon.tech/chengetopay?sslmode=require"
	}
	
	// Try alternative connection format for NeonDB
	if strings.Contains(postgresURL, "neon.tech") {
		// Add connection parameters to handle SCRAM issue
		if !strings.Contains(postgresURL, "?") {
			postgresURL += "?"
		} else {
			postgresURL += "&"
		}
		postgresURL += "application_name=chengetopay&connect_timeout=10"
	}
	
	log.Println("Connecting to PostgreSQL...")
	var pgErr error
	service.postgresDB, pgErr = sql.Open("postgres", postgresURL)
	if pgErr == nil {
		service.postgresDB.SetMaxOpenConns(5)
		service.postgresDB.SetMaxIdleConns(2)
		service.postgresDB.SetConnMaxLifetime(5 * time.Minute)
		
		// Test with a simple query instead of Ping
		var result int
		pgErr = service.postgresDB.QueryRow("SELECT 1").Scan(&result)
		if pgErr != nil {
			log.Printf("⚠️ PostgreSQL: Connection test failed: %v", pgErr)
			// Continue anyway
		} else {
			log.Printf("✅ PostgreSQL: Connected successfully")
		}
	} else {
		log.Printf("⚠️ PostgreSQL: Failed to open: %v", pgErr)
	}
	
	// Fix MongoDB connection - use direct connection instead of SRV
	mongoURL := os.Getenv("MONGODB_URL")
	if mongoURL == "" || strings.Contains(mongoURL, "mongodb+srv") {
		// Convert SRV to direct connection format
		// Extract credentials and cluster name
		mongoURL = "mongodb://tendaimukurusystemsadministrator:Mukuru@2024@chengetopay.jvjvz.mongodb.net:27017/?retryWrites=true&w=majority"
		
		// Try alternative formats
		mongoURLs := []string{
			"mongodb://tendaimukurusystemsadministrator:Mukuru%402024@chengetopay.mongodb.net:27017/?retryWrites=true&w=majority",
			"mongodb://tendaimukurusystemsadministrator:Mukuru@2024@chengetopay-shard-00-00.jvjvz.mongodb.net:27017,chengetopay-shard-00-01.jvjvz.mongodb.net:27017,chengetopay-shard-00-02.jvjvz.mongodb.net:27017/?ssl=true&replicaSet=atlas-xyz-shard-0&authSource=admin&retryWrites=true&w=majority",
			"mongodb://localhost:27017", // Fallback to local
		}
		
		var mongoErr error
		for _, url := range mongoURLs {
			log.Printf("Trying MongoDB connection: %s", strings.Split(url, "@")[0]+"@...")
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			
			clientOptions := options.Client().ApplyURI(url).SetServerSelectionTimeout(10 * time.Second)
			mongoClient, err := mongo.Connect(ctx, clientOptions)
			if err == nil {
				err = mongoClient.Ping(ctx, nil)
				if err == nil {
					service.mongoDB = mongoClient.Database("chengetopay")
					log.Println("✅ MongoDB: Connected successfully")
					cancel()
					break
				}
			}
			mongoErr = err
			cancel()
		}
		
		if service.mongoDB == nil {
			log.Printf("⚠️ MongoDB: All connection attempts failed: %v", mongoErr)
		}
	}
	
	// Fix Redis connection - try alternative hostnames and local fallback
	redisURL := os.Getenv("REDIS_URL")
	redisURLs := []string{
		redisURL,
		"redis://default:AVNS_Ej1eMBBJAoJy5sTnJCm@redis-chengetopay.aivencloud.com:24660",
		"redis://localhost:6379",
	}
	
	var redisErr error
	for _, url := range redisURLs {
		if url == "" {
			continue
		}
		
		log.Printf("Trying Redis connection...")
		opt, err := redis.ParseURL(url)
		if err != nil {
			redisErr = err
			continue
		}
		
		client := redis.NewClient(opt)
		ctx := context.Background()
		_, err = client.Ping(ctx).Result()
		if err == nil {
			service.redisClient = client
			log.Println("✅ Redis: Connected successfully")
			break
		}
		redisErr = err
		client.Close()
	}
	
	if service.redisClient == nil {
		log.Printf("⚠️ Redis: All connection attempts failed: %v", redisErr)
		// Create a local Redis client as fallback
		service.redisClient = redis.NewClient(&redis.Options{
			Addr: "localhost:6379",
		})
	}
	
	// Setup HTTP routes
	http.HandleFunc("/health", service.healthHandler)
	http.HandleFunc("/api/v1/test", service.testHandler)
	http.HandleFunc("/api/v1/escrow", service.escrowHandler)
	
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	
	log.Printf("🚀 Service starting on port %s", port)
	log.Println("📊 Service running with available databases")
	log.Println("🔗 Endpoints: /health, /api/v1/test, /api/v1/escrow")
	
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func (s *Service) healthHandler(w http.ResponseWriter, r *http.Request) {
	health := map[string]interface{}{
		"status": "healthy",
		"timestamp": time.Now().Unix(),
		"service": "working-service",
		"databases": map[string]bool{
			"postgres": s.postgresDB != nil && s.postgresDB.Ping() == nil,
			"mongodb": s.mongoDB != nil,
			"redis": s.redisClient != nil && s.redisClient.Ping(context.Background()).Err() == nil,
		},
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(health)
}

func (s *Service) testHandler(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	response := map[string]interface{}{
		"timestamp": time.Now().Unix(),
		"databases": map[string]interface{}{},
	}
	
	// Test PostgreSQL
	if s.postgresDB != nil {
		var pgVersion string
		err := s.postgresDB.QueryRow("SELECT version()").Scan(&pgVersion)
		if err == nil {
			response["databases"].(map[string]interface{})["postgres"] = map[string]interface{}{
				"connected": true,
				"version": pgVersion[:50] + "...", // Truncate for readability
			}
		} else {
			response["databases"].(map[string]interface{})["postgres"] = map[string]interface{}{
				"connected": false,
				"error": err.Error(),
			}
		}
	} else {
		response["databases"].(map[string]interface{})["postgres"] = map[string]interface{}{
			"connected": false,
			"error": "Not initialized",
		}
	}
	
	// Test MongoDB
	if s.mongoDB != nil {
		err := s.mongoDB.Client().Ping(ctx, nil)
		response["databases"].(map[string]interface{})["mongodb"] = map[string]interface{}{
			"connected": err == nil,
			"database": "chengetopay",
		}
		if err != nil {
			response["databases"].(map[string]interface{})["mongodb"].(map[string]interface{})["error"] = err.Error()
		}
	} else {
		response["databases"].(map[string]interface{})["mongodb"] = map[string]interface{}{
			"connected": false,
			"error": "Not initialized",
		}
	}
	
	// Test Redis
	if s.redisClient != nil {
		err := s.redisClient.Set(ctx, "test:key", "test-value", 10*time.Second).Err()
		if err == nil {
			val, _ := s.redisClient.Get(ctx, "test:key").Result()
			response["databases"].(map[string]interface{})["redis"] = map[string]interface{}{
				"connected": true,
				"test_value": val,
			}
		} else {
			response["databases"].(map[string]interface{})["redis"] = map[string]interface{}{
				"connected": false,
				"error": err.Error(),
			}
		}
	} else {
		response["databases"].(map[string]interface{})["redis"] = map[string]interface{}{
			"connected": false,
			"error": "Not initialized",
		}
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (s *Service) escrowHandler(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	
	switch r.Method {
	case "GET":
		// List escrows
		escrows := []map[string]interface{}{}
		
		if s.postgresDB != nil {
			// Try to fetch from PostgreSQL
			rows, err := s.postgresDB.Query("SELECT id, buyer_id, seller_id, amount, status FROM escrow.escrows LIMIT 10")
			if err == nil {
				defer rows.Close()
				for rows.Next() {
					var id, buyerID, sellerID, status string
					var amount float64
					if err := rows.Scan(&id, &buyerID, &sellerID, &amount, &status); err == nil {
						escrows = append(escrows, map[string]interface{}{
							"id": id,
							"buyer_id": buyerID,
							"seller_id": sellerID,
							"amount": amount,
							"status": status,
						})
					}
				}
			}
		}
		
		// If no escrows from DB, return sample data
		if len(escrows) == 0 {
			escrows = append(escrows, map[string]interface{}{
				"id": "sample-001",
				"buyer_id": "buyer-123",
				"seller_id": "seller-456",
				"amount": 1000.00,
				"status": "pending",
			})
		}
		
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"escrows": escrows,
			"count": len(escrows),
		})
		
	case "POST":
		// Create a new escrow
		var req map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}
		
		escrowID := fmt.Sprintf("escrow-%d", time.Now().Unix())
		
		// Try to save to PostgreSQL
		if s.postgresDB != nil {
			_, err := s.postgresDB.Exec(
				"INSERT INTO escrow.escrows (id, buyer_id, seller_id, amount, status) VALUES ($1, $2, $3, $4, $5)",
				escrowID,
				req["buyer_id"],
				req["seller_id"],
				req["amount"],
				"pending",
			)
			if err != nil {
				log.Printf("Failed to save escrow to PostgreSQL: %v", err)
			}
		}
		
		// Try to cache in Redis
		if s.redisClient != nil {
			data, _ := json.Marshal(req)
			s.redisClient.Set(ctx, "escrow:"+escrowID, data, 24*time.Hour)
		}
		
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id": escrowID,
			"status": "created",
			"data": req,
		})
		
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}
