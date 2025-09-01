package database

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/lib/pq"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"github.com/redis/go-redis/v9"
)

// Config holds all database configurations
type Config struct {
	PostgresURL string
	MongoDBURL  string
	RedisURL    string
	ServiceName string
	Schema      string
}

// NewConfig creates a new database configuration from environment variables
func NewConfig(serviceName string, schema string) *Config {
	return &Config{
		PostgresURL: getEnv("POSTGRES_URL", "postgres://localhost:5432/chengetopay"),
		MongoDBURL:  getEnv("MONGODB_URL", "mongodb://localhost:27017/chengetopay"),
		RedisURL:    getEnv("REDIS_URL", "redis://localhost:6379/0"),
		ServiceName: serviceName,
		Schema:      schema,
	}
}

// PostgresConnection establishes a PostgreSQL connection with schema
func (c *Config) PostgresConnection() (*sql.DB, error) {
	db, err := sql.Open("postgres", c.PostgresURL)
	if err != nil {
		return nil, fmt.Errorf("failed to open postgres: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(20)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)
	db.SetConnMaxIdleTime(5 * time.Minute)

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping postgres: %w", err)
	}

	// Set schema if provided
	if c.Schema != "" {
		_, err = db.Exec(fmt.Sprintf("CREATE SCHEMA IF NOT EXISTS %s", c.Schema))
		if err != nil {
			log.Printf("Warning: Could not create schema %s: %v", c.Schema, err)
		}
		
		_, err = db.Exec(fmt.Sprintf("SET search_path TO %s", c.Schema))
		if err != nil {
			return nil, fmt.Errorf("failed to set schema: %w", err)
		}
	}

	log.Printf("✅ PostgreSQL connected for service: %s (schema: %s)", c.ServiceName, c.Schema)
	return db, nil
}

// MongoConnection establishes a MongoDB connection
func (c *Config) MongoConnection() (*mongo.Client, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	clientOptions := options.Client().
		ApplyURI(c.MongoDBURL).
		SetMaxPoolSize(10).
		SetMinPoolSize(2).
		SetRetryWrites(true).
		SetRetryReads(true)

	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to mongodb: %w", err)
	}

	// Ping the database
	if err := client.Ping(ctx, nil); err != nil {
		return nil, fmt.Errorf("failed to ping mongodb: %w", err)
	}

	log.Printf("✅ MongoDB connected for service: %s", c.ServiceName)
	return client, nil
}

// RedisConnection establishes a Redis connection
func (c *Config) RedisConnection() (*redis.Client, error) {
	opt, err := redis.ParseURL(c.RedisURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse redis URL: %w", err)
	}

	// Configure connection pool
	opt.MaxRetries = 3
	opt.PoolSize = 10
	opt.MinIdleConns = 2
	opt.MaxIdleConns = 5
	opt.ConnMaxIdleTime = 5 * time.Minute
	opt.ConnMaxLifetime = 30 * time.Minute

	client := redis.NewClient(opt)

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to ping redis: %w", err)
	}

	log.Printf("✅ Redis connected for service: %s", c.ServiceName)
	return client, nil
}

// getEnv gets an environment variable with a fallback
func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

// ServiceSchemaMap defines which schema each service should use
var ServiceSchemaMap = map[string]string{
	"escrow-service":         "escrow",
	"payment-service":        "payments",
	"ledger-service":         "ledger",
	"journal-service":        "journal",
	"fees-service":           "fees",
	"refunds-service":        "refunds",
	"transfers-service":      "transfers",
	"payouts-service":        "payouts",
	"reserves-service":       "reserves",
	"reconciliation-service": "reconciliation",
	"treasury-service":       "treasury",
	"risk-service":           "risk",
	"disputes-service":       "disputes",
	"auth-service":           "auth",
	"compliance-service":     "compliance",
}

// MongoCollectionMap defines which collection each service should use
var MongoCollectionMap = map[string]string{
	"kyb-service":            "kyb_documents",
	"evidence-service":       "evidence",
	"workflow-service":       "workflows",
	"config-service":         "configurations",
	"portal-service":         "portal_data",
	"webhooks-service":       "webhooks",
	"data-platform-service":  "analytics",
	"compliance-ops-service": "compliance_ops",
	"dx-service":             "developer_portal",
	"workers-service":        "jobs",
	"observability-service":  "metrics",
	"eventbus-service":       "events",
	"saga-service":           "sagas",
	"sca-service":            "sca_data",
}
