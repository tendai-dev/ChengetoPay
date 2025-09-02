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

	"messagequeue"
)

// MessageQueueService represents the message queue service
type MessageQueueService struct {
	rabbitmq   *messagequeue.RabbitMQ
	eventStore *messagequeue.EventStore
}

// Global service instance
var mqService *MessageQueueService

func main() {
	port := flag.String("port", "8117", "Port to listen on")
	flag.Parse()

	log.Printf("Starting Message Queue & Event Streaming Service on port %s...", *port)

	// Initialize message queue service
	if err := initializeMessageQueue(); err != nil {
		log.Fatalf("Failed to initialize message queue: %v", err)
	}

	// Create HTTP server
	mux := http.NewServeMux()
	mux.HandleFunc("/health", handleHealth)
	mux.HandleFunc("/v1/status", handleStatus)
	mux.HandleFunc("/v1/publish", handlePublish)
	mux.HandleFunc("/v1/events", handleEvents)
	mux.HandleFunc("/v1/dlq", handleDeadLetterQueue)
	mux.HandleFunc("/v1/stats", handleStats)

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
		log.Printf("Message Queue service listening on port %s", *port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down message queue service...")

	// Close connections
	if mqService != nil {
		if mqService.rabbitmq != nil {
			mqService.rabbitmq.Close()
		}
		if mqService.eventStore != nil {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			mqService.eventStore.Close(ctx)
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Message Queue service exited")
}

// initializeMessageQueue initializes RabbitMQ and EventStore
func initializeMessageQueue() error {
	// RabbitMQ configuration
	rabbitmqConfig := messagequeue.RabbitMQConfig{
		URL:           "amqp://guest:guest@localhost:5672/",
		Exchange:      "financial_events",
		QueueName:     "payment_events",
		RoutingKey:    "payment.*",
		MaxRetries:    3,
		RetryDelay:    5 * time.Second,
		PrefetchCount: 10,
	}

	// EventStore configuration
	eventStoreConfig := messagequeue.EventStoreConfig{
		DatabaseURL:    "mongodb+srv://tendai_db_user:aEmut0m48FtaES1E@cluster0.csdtbuo.mongodb.net/?retryWrites=true&w=majority&appName=Cluster0",
		DatabaseName:   "financial_platform",
		CollectionName: "events",
		MaxBatchSize:   100,
	}

	// Initialize RabbitMQ
	rabbitmq, err := messagequeue.NewRabbitMQ(rabbitmqConfig)
	if err != nil {
		return fmt.Errorf("failed to initialize RabbitMQ: %w", err)
	}

	// Initialize EventStore
	eventStore, err := messagequeue.NewEventStore(eventStoreConfig)
	if err != nil {
		return fmt.Errorf("failed to initialize EventStore: %w", err)
	}

	mqService = &MessageQueueService{
		rabbitmq:   rabbitmq,
		eventStore: eventStore,
	}

	// Start consuming messages
	go startMessageConsumer()

	log.Println("✅ Message Queue initialized successfully")
	return nil
}

// startMessageConsumer starts consuming messages from RabbitMQ
func startMessageConsumer() {
	ctx := context.Background()

	err := mqService.rabbitmq.Consume(ctx, func(msg messagequeue.Message) error {
		// Store event in EventStore
		event := messagequeue.Event{
			AggregateID: msg.ID,
			EventType:   msg.Type,
			Data:        msg.Data,
			Metadata:    msg.Headers,
			Timestamp:   msg.Timestamp,
		}

		if err := mqService.eventStore.AppendEvent(ctx, event); err != nil {
			return fmt.Errorf("failed to store event: %w", err)
		}

		log.Printf("✅ Processed message: %s (%s)", msg.ID, msg.Type)
		return nil
	})

	if err != nil {
		log.Printf("❌ Failed to start message consumer: %v", err)
	}
}

// handleHealth handles health check requests
func handleHealth(w http.ResponseWriter, r *http.Request) {
	health := map[string]interface{}{
		"status":    "healthy",
		"service":   "message-queue",
		"timestamp": time.Now().Format(time.RFC3339),
		"components": map[string]string{
			"rabbitmq":   "connected",
			"eventstore": "connected",
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(health)
}

// handleStatus handles status requests
func handleStatus(w http.ResponseWriter, r *http.Request) {
	status := map[string]interface{}{
		"message_queue": map[string]interface{}{
			"status": "active",
			"rabbitmq": map[string]interface{}{
				"exchange":    mqService.rabbitmq.Config.Exchange,
				"queue":       mqService.rabbitmq.Config.QueueName,
				"routing_key": mqService.rabbitmq.Config.RoutingKey,
			},
			"event_store": map[string]interface{}{
				"database":   mqService.eventStore.Config.DatabaseName,
				"collection": mqService.eventStore.Config.CollectionName,
			},
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}

// handlePublish handles message publishing requests
func handlePublish(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var request struct {
		Type    string                 `json:"type"`
		Data    map[string]interface{} `json:"data"`
		Headers map[string]interface{} `json:"headers,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	message := messagequeue.Message{
		ID:        generateMessageID(),
		Type:      request.Type,
		Data:      request.Data,
		Headers:   request.Headers,
		Timestamp: time.Now(),
	}

	if err := mqService.rabbitmq.Publish(r.Context(), message); err != nil {
		http.Error(w, fmt.Sprintf("Failed to publish message: %v", err), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"status":  "success",
		"message": "Message published successfully",
		"data": map[string]interface{}{
			"message_id": message.ID,
			"type":       message.Type,
			"timestamp":  message.Timestamp,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleEvents handles event store requests
func handleEvents(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		handleGetEvents(w, r)
	case "POST":
		handleStoreEvent(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleGetEvents handles event retrieval requests
func handleGetEvents(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	eventType := query.Get("type")
	orgID := query.Get("org_id")
	userID := query.Get("user_id")
	limit := int64(100) // Default limit

	if limitStr := query.Get("limit"); limitStr != "" {
		if l, err := fmt.Sscanf(limitStr, "%d", &limit); err != nil || l != 1 {
			limit = 100
		}
	}

	var events []messagequeue.Event
	var err error

	switch {
	case eventType != "":
		events, err = mqService.eventStore.GetEventsByType(r.Context(), eventType, limit)
	case orgID != "":
		events, err = mqService.eventStore.GetEventsByOrg(r.Context(), orgID, limit)
	case userID != "":
		events, err = mqService.eventStore.GetEventsByUser(r.Context(), userID, limit)
	default:
		// Get recent events
		end := time.Now()
		start := end.Add(-24 * time.Hour)
		events, err = mqService.eventStore.GetEventsByTimeRange(r.Context(), start, end, limit)
	}

	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get events: %v", err), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"status": "success",
		"data": map[string]interface{}{
			"events": events,
			"count":  len(events),
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleStoreEvent handles direct event storage requests
func handleStoreEvent(w http.ResponseWriter, r *http.Request) {
	var event messagequeue.Event
	if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := mqService.eventStore.AppendEvent(r.Context(), event); err != nil {
		http.Error(w, fmt.Sprintf("Failed to store event: %v", err), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"status":  "success",
		"message": "Event stored successfully",
		"data": map[string]interface{}{
			"event_id": event.ID,
			"type":     event.EventType,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleDeadLetterQueue handles dead letter queue operations
func handleDeadLetterQueue(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		handleGetDLQMessages(w, r)
	case "POST":
		handleRepublishDLQMessage(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleGetDLQMessages handles dead letter queue message retrieval
func handleGetDLQMessages(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	limit := 10 // Default limit

	if limitStr := query.Get("limit"); limitStr != "" {
		if l, err := fmt.Sscanf(limitStr, "%d", &limit); err != nil || l != 1 {
			limit = 10
		}
	}

	messages, err := mqService.rabbitmq.GetDeadLetterMessages(limit)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get DLQ messages: %v", err), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"status": "success",
		"data": map[string]interface{}{
			"messages": messages,
			"count":    len(messages),
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleRepublishDLQMessage handles dead letter queue message republishing
func handleRepublishDLQMessage(w http.ResponseWriter, r *http.Request) {
	var request struct {
		MessageID string `json:"message_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := mqService.rabbitmq.RepublishDeadLetterMessage(request.MessageID); err != nil {
		http.Error(w, fmt.Sprintf("Failed to republish message: %v", err), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"status":  "success",
		"message": "Message republished successfully",
		"data": map[string]interface{}{
			"message_id": request.MessageID,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleStats handles statistics requests
func handleStats(w http.ResponseWriter, r *http.Request) {
	// Get RabbitMQ stats
	rabbitmqStats, err := mqService.rabbitmq.GetQueueStats()
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get RabbitMQ stats: %v", err), http.StatusInternalServerError)
		return
	}

	// Get EventStore stats
	eventStoreStats, err := mqService.eventStore.GetEventStats(r.Context())
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get EventStore stats: %v", err), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"status": "success",
		"data": map[string]interface{}{
			"rabbitmq":   rabbitmqStats,
			"eventstore": eventStoreStats,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// generateMessageID generates a unique message ID
func generateMessageID() string {
	return fmt.Sprintf("msg_%d", time.Now().UnixNano())
}
