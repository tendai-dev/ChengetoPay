package database

import (
	"context"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// MongoConfig holds MongoDB configuration
type MongoConfig struct {
	URL                string
	Database           string
	MaxPoolSize        uint64
	MinPoolSize        uint64
	MaxConnIdleTime    time.Duration
	ServerSelectionTimeout time.Duration
}

// MongoDB represents the MongoDB connection
type MongoDB struct {
	client   *mongo.Client
	database *mongo.Database
}

// NewMongoDB creates a new MongoDB connection
func NewMongoDB(config MongoConfig) (*MongoDB, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	clientOptions := options.Client().ApplyURI(config.URL).
		SetMaxPoolSize(config.MaxPoolSize).
		SetMinPoolSize(config.MinPoolSize).
		SetMaxConnIdleTime(config.MaxConnIdleTime).
		SetServerSelectionTimeout(config.ServerSelectionTimeout)

	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MongoDB: %w", err)
	}

	// Test connection
	if err := client.Ping(ctx, nil); err != nil {
		return nil, fmt.Errorf("failed to ping MongoDB: %w", err)
	}

	database := client.Database(config.Database)

	log.Println("✅ MongoDB connected successfully")

	return &MongoDB{
		client:   client,
		database: database,
	}, nil
}

// Close closes the MongoDB connection
func (m *MongoDB) Close(ctx context.Context) error {
	return m.client.Disconnect(ctx)
}

// GetDatabase returns the database instance
func (m *MongoDB) GetDatabase() *mongo.Database {
	return m.database
}

// GetClient returns the MongoDB client
func (m *MongoDB) GetClient() *mongo.Client {
	return m.client
}

// HealthCheck performs a health check on the database
func (m *MongoDB) HealthCheck(ctx context.Context) error {
	return m.client.Ping(ctx, nil)
}

// CreateCollections creates the necessary collections with indexes
func (m *MongoDB) CreateCollections() error {
	collections := map[string][]mongo.IndexModel{
		"audit_logs": {
			{
				Keys: map[string]interface{}{
					"timestamp": -1,
				},
				Options: options.Index().SetName("idx_audit_logs_timestamp"),
			},
			{
				Keys: map[string]interface{}{
					"org_id": 1,
				},
				Options: options.Index().SetName("idx_audit_logs_org_id"),
			},
			{
				Keys: map[string]interface{}{
					"user_id": 1,
				},
				Options: options.Index().SetName("idx_audit_logs_user_id"),
			},
			{
				Keys: map[string]interface{}{
					"action": 1,
				},
				Options: options.Index().SetName("idx_audit_logs_action"),
			},
		},
		"evidence": {
			{
				Keys: map[string]interface{}{
					"org_id": 1,
				},
				Options: options.Index().SetName("idx_evidence_org_id"),
			},
			{
				Keys: map[string]interface{}{
					"created_at": -1,
				},
				Options: options.Index().SetName("idx_evidence_created_at"),
			},
			{
				Keys: map[string]interface{}{
					"type": 1,
				},
				Options: options.Index().SetName("idx_evidence_type"),
			},
		},
		"disputes": {
			{
				Keys: map[string]interface{}{
					"org_id": 1,
				},
				Options: options.Index().SetName("idx_disputes_org_id"),
			},
			{
				Keys: map[string]interface{}{
					"status": 1,
				},
				Options: options.Index().SetName("idx_disputes_status"),
			},
			{
				Keys: map[string]interface{}{
					"created_at": -1,
				},
				Options: options.Index().SetName("idx_disputes_created_at"),
			},
		},
		"compliance_cases": {
			{
				Keys: map[string]interface{}{
					"org_id": 1,
				},
				Options: options.Index().SetName("idx_compliance_cases_org_id"),
			},
			{
				Keys: map[string]interface{}{
					"case_type": 1,
				},
				Options: options.Index().SetName("idx_compliance_cases_type"),
			},
			{
				Keys: map[string]interface{}{
					"status": 1,
				},
				Options: options.Index().SetName("idx_compliance_cases_status"),
			},
		},
		"kyc_documents": {
			{
				Keys: map[string]interface{}{
					"user_id": 1,
				},
				Options: options.Index().SetName("idx_kyc_documents_user_id"),
			},
			{
				Keys: map[string]interface{}{
					"org_id": 1,
				},
				Options: options.Index().SetName("idx_kyc_documents_org_id"),
			},
			{
				Keys: map[string]interface{}{
					"document_type": 1,
				},
				Options: options.Index().SetName("idx_kyc_documents_type"),
			},
		},
		"api_logs": {
			{
				Keys: map[string]interface{}{
					"timestamp": -1,
				},
				Options: options.Index().SetName("idx_api_logs_timestamp"),
			},
			{
				Keys: map[string]interface{}{
					"org_id": 1,
				},
				Options: options.Index().SetName("idx_api_logs_org_id"),
			},
			{
				Keys: map[string]interface{}{
					"endpoint": 1,
				},
				Options: options.Index().SetName("idx_api_logs_endpoint"),
			},
		},
		"performance_metrics": {
			{
				Keys: map[string]interface{}{
					"timestamp": -1,
				},
				Options: options.Index().SetName("idx_performance_metrics_timestamp"),
			},
			{
				Keys: map[string]interface{}{
					"service": 1,
				},
				Options: options.Index().SetName("idx_performance_metrics_service"),
			},
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	for collectionName, indexes := range collections {
		log.Printf("Creating collection: %s", collectionName)
		
		// Create collection
		err := m.database.CreateCollection(ctx, collectionName)
		if err != nil {
			// Collection might already exist, continue
			log.Printf("Collection %s might already exist: %v", collectionName, err)
		}

		// Create indexes
		collection := m.database.Collection(collectionName)
		if len(indexes) > 0 {
			_, err := collection.Indexes().CreateMany(ctx, indexes)
			if err != nil {
				return fmt.Errorf("failed to create indexes for %s: %w", collectionName, err)
			}
		}
	}

	log.Println("✅ MongoDB collections and indexes created successfully")
	return nil
}

// CreateTTLIndexes creates TTL indexes for data retention
func (m *MongoDB) CreateTTLIndexes() error {
	ttlIndexes := map[string]time.Duration{
		"audit_logs":     90 * 24 * time.Hour,  // 90 days
		"api_logs":       30 * 24 * time.Hour,  // 30 days
		"performance_metrics": 7 * 24 * time.Hour, // 7 days
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	for collectionName, ttl := range ttlIndexes {
		collection := m.database.Collection(collectionName)
		
		indexModel := mongo.IndexModel{
			Keys: map[string]interface{}{
				"created_at": 1,
			},
			Options: options.Index().
				SetName(fmt.Sprintf("idx_%s_ttl", collectionName)).
				SetExpireAfterSeconds(int32(ttl.Seconds())),
		}

		_, err := collection.Indexes().CreateOne(ctx, indexModel)
		if err != nil {
			return fmt.Errorf("failed to create TTL index for %s: %w", collectionName, err)
		}
	}

	log.Println("✅ MongoDB TTL indexes created successfully")
	return nil
}
