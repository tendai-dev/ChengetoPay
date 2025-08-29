package messagequeue

import (
	"context"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// EventStoreConfig holds EventStore configuration
type EventStoreConfig struct {
	DatabaseURL    string
	DatabaseName   string
	CollectionName string
	MaxBatchSize   int
}

// Event represents an event in the event store
type Event struct {
	ID          primitive.ObjectID     `bson:"_id,omitempty" json:"id"`
	AggregateID string                 `bson:"aggregate_id" json:"aggregate_id"`
	EventType   string                 `bson:"event_type" json:"event_type"`
	Version     int64                  `bson:"version" json:"version"`
	Data        map[string]interface{} `bson:"data" json:"data"`
	Metadata    map[string]interface{} `bson:"metadata,omitempty" json:"metadata,omitempty"`
	Timestamp   time.Time              `bson:"timestamp" json:"timestamp"`
	UserID      string                 `bson:"user_id,omitempty" json:"user_id,omitempty"`
	OrgID       string                 `bson:"org_id,omitempty" json:"org_id,omitempty"`
}

// EventStore represents the event store
type EventStore struct {
	client     *mongo.Client
	database   *mongo.Database
	collection *mongo.Collection
	Config     EventStoreConfig
}

// NewEventStore creates a new EventStore connection
func NewEventStore(config EventStoreConfig) (*EventStore, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(config.DatabaseURL))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MongoDB: %w", err)
	}

	// Test connection
	if err := client.Ping(ctx, nil); err != nil {
		return nil, fmt.Errorf("failed to ping MongoDB: %w", err)
	}

	database := client.Database(config.DatabaseName)
	collection := database.Collection(config.CollectionName)

	eventStore := &EventStore{
		client:     client,
		database:   database,
		collection: collection,
		Config:     config,
	}

	// Create indexes
	if err := eventStore.createIndexes(); err != nil {
		return nil, fmt.Errorf("failed to create indexes: %w", err)
	}

	log.Println("✅ EventStore connected successfully")
	return eventStore, nil
}

// createIndexes creates the necessary indexes
func (e *EventStore) createIndexes() error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	indexes := []mongo.IndexModel{
		{
			Keys: bson.D{
				{"aggregate_id", 1},
				{"version", 1},
			},
			Options: options.Index().SetUnique(true).SetName("idx_aggregate_version"),
		},
		{
			Keys: bson.D{
				{"event_type", 1},
			},
			Options: options.Index().SetName("idx_event_type"),
		},
		{
			Keys: bson.D{
				{"timestamp", -1},
			},
			Options: options.Index().SetName("idx_timestamp"),
		},
		{
			Keys: bson.D{
				{"org_id", 1},
			},
			Options: options.Index().SetName("idx_org_id"),
		},
		{
			Keys: bson.D{
				{"user_id", 1},
			},
			Options: options.Index().SetName("idx_user_id"),
		},
	}

	_, err := e.collection.Indexes().CreateMany(ctx, indexes)
	return err
}

// Close closes the EventStore connection
func (e *EventStore) Close(ctx context.Context) error {
	return e.client.Disconnect(ctx)
}

// AppendEvent appends an event to the event store
func (e *EventStore) AppendEvent(ctx context.Context, event Event) error {
	// Set timestamp if not provided
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}

	// Set version if not provided
	if event.Version == 0 {
		lastVersion, err := e.getLastVersion(ctx, event.AggregateID)
		if err != nil {
			return fmt.Errorf("failed to get last version: %w", err)
		}
		event.Version = lastVersion + 1
	}

	_, err := e.collection.InsertOne(ctx, event)
	if err != nil {
		return fmt.Errorf("failed to insert event: %w", err)
	}

	return nil
}

// AppendEvents appends multiple events to the event store
func (e *EventStore) AppendEvents(ctx context.Context, events []Event) error {
	if len(events) == 0 {
		return nil
	}

	// Prepare documents for bulk insert
	var documents []interface{}
	for i, event := range events {
		if event.Timestamp.IsZero() {
			event.Timestamp = time.Now()
		}
		if event.Version == 0 {
			lastVersion, err := e.getLastVersion(ctx, event.AggregateID)
			if err != nil {
				return fmt.Errorf("failed to get last version for event %d: %w", i, err)
			}
			event.Version = lastVersion + 1
		}
		documents = append(documents, event)
	}

	_, err := e.collection.InsertMany(ctx, documents)
	if err != nil {
		return fmt.Errorf("failed to insert events: %w", err)
	}

	return nil
}

// GetEvents retrieves events for an aggregate
func (e *EventStore) GetEvents(ctx context.Context, aggregateID string, fromVersion int64) ([]Event, error) {
	filter := bson.M{
		"aggregate_id": aggregateID,
		"version":      bson.M{"$gte": fromVersion},
	}

	opts := options.Find().SetSort(bson.D{{"version", 1}})

	cursor, err := e.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to find events: %w", err)
	}
	defer cursor.Close(ctx)

	var events []Event
	if err := cursor.All(ctx, &events); err != nil {
		return nil, fmt.Errorf("failed to decode events: %w", err)
	}

	return events, nil
}

// GetEventsByType retrieves events by type
func (e *EventStore) GetEventsByType(ctx context.Context, eventType string, limit int64) ([]Event, error) {
	filter := bson.M{"event_type": eventType}

	opts := options.Find().
		SetSort(bson.D{{"timestamp", -1}}).
		SetLimit(limit)

	cursor, err := e.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to find events: %w", err)
	}
	defer cursor.Close(ctx)

	var events []Event
	if err := cursor.All(ctx, &events); err != nil {
		return nil, fmt.Errorf("failed to decode events: %w", err)
	}

	return events, nil
}

// GetEventsByTimeRange retrieves events within a time range
func (e *EventStore) GetEventsByTimeRange(ctx context.Context, start, end time.Time, limit int64) ([]Event, error) {
	filter := bson.M{
		"timestamp": bson.M{
			"$gte": start,
			"$lte": end,
		},
	}

	opts := options.Find().
		SetSort(bson.D{{"timestamp", -1}}).
		SetLimit(limit)

	cursor, err := e.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to find events: %w", err)
	}
	defer cursor.Close(ctx)

	var events []Event
	if err := cursor.All(ctx, &events); err != nil {
		return nil, fmt.Errorf("failed to decode events: %w", err)
	}

	return events, nil
}

// GetEventsByOrg retrieves events for an organization
func (e *EventStore) GetEventsByOrg(ctx context.Context, orgID string, limit int64) ([]Event, error) {
	filter := bson.M{"org_id": orgID}

	opts := options.Find().
		SetSort(bson.D{{"timestamp", -1}}).
		SetLimit(limit)

	cursor, err := e.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to find events: %w", err)
	}
	defer cursor.Close(ctx)

	var events []Event
	if err := cursor.All(ctx, &events); err != nil {
		return nil, fmt.Errorf("failed to decode events: %w", err)
	}

	return events, nil
}

// GetEventsByUser retrieves events for a user
func (e *EventStore) GetEventsByUser(ctx context.Context, userID string, limit int64) ([]Event, error) {
	filter := bson.M{"user_id": userID}

	opts := options.Find().
		SetSort(bson.D{{"timestamp", -1}}).
		SetLimit(limit)

	cursor, err := e.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to find events: %w", err)
	}
	defer cursor.Close(ctx)

	var events []Event
	if err := cursor.All(ctx, &events); err != nil {
		return nil, fmt.Errorf("failed to decode events: %w", err)
	}

	return events, nil
}

// getLastVersion gets the last version for an aggregate
func (e *EventStore) getLastVersion(ctx context.Context, aggregateID string) (int64, error) {
	filter := bson.M{"aggregate_id": aggregateID}

	opts := options.FindOne().SetSort(bson.D{{"version", -1}})

	var event Event
	err := e.collection.FindOne(ctx, filter, opts).Decode(&event)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return 0, nil
		}
		return 0, fmt.Errorf("failed to find last version: %w", err)
	}

	return event.Version, nil
}

// GetEventStats returns event store statistics
func (e *EventStore) GetEventStats(ctx context.Context) (map[string]interface{}, error) {
	// Total events
	totalEvents, err := e.collection.CountDocuments(ctx, bson.M{})
	if err != nil {
		return nil, fmt.Errorf("failed to count total events: %w", err)
	}

	// Events by type
	pipeline := []bson.M{
		{"$group": bson.M{
			"_id":   "$event_type",
			"count": bson.M{"$sum": 1},
		}},
		{"$sort": bson.M{"count": -1}},
	}

	cursor, err := e.collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, fmt.Errorf("failed to aggregate events by type: %w", err)
	}
	defer cursor.Close(ctx)

	var eventsByType []bson.M
	if err := cursor.All(ctx, &eventsByType); err != nil {
		return nil, fmt.Errorf("failed to decode aggregation: %w", err)
	}

	// Recent events (last 24 hours)
	yesterday := time.Now().Add(-24 * time.Hour)
	recentEvents, err := e.collection.CountDocuments(ctx, bson.M{
		"timestamp": bson.M{"$gte": yesterday},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to count recent events: %w", err)
	}

	return map[string]interface{}{
		"total_events":    totalEvents,
		"recent_events":   recentEvents,
		"events_by_type":  eventsByType,
		"collection_name": e.Config.CollectionName,
	}, nil
}

// CreateSnapshot creates a snapshot for an aggregate
func (e *EventStore) CreateSnapshot(ctx context.Context, aggregateID string, snapshot interface{}) error {
	snapshotDoc := bson.M{
		"aggregate_id": aggregateID,
		"snapshot":     snapshot,
		"created_at":   time.Now(),
	}

	// Use upsert to replace existing snapshot
	filter := bson.M{"aggregate_id": aggregateID}
	update := bson.M{"$set": snapshotDoc}
	opts := options.Update().SetUpsert(true)

	_, err := e.collection.UpdateOne(ctx, filter, update, opts)
	if err != nil {
		return fmt.Errorf("failed to create snapshot: %w", err)
	}

	return nil
}

// GetSnapshot retrieves a snapshot for an aggregate
func (e *EventStore) GetSnapshot(ctx context.Context, aggregateID string) (interface{}, error) {
	filter := bson.M{"aggregate_id": aggregateID}

	var result bson.M
	err := e.collection.FindOne(ctx, filter).Decode(&result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get snapshot: %w", err)
	}

	return result["snapshot"], nil
}
