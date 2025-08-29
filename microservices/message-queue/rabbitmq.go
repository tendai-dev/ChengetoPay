package messagequeue

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

// RabbitMQConfig holds RabbitMQ configuration
type RabbitMQConfig struct {
	URL           string
	Exchange      string
	QueueName     string
	RoutingKey    string
	MaxRetries    int
	RetryDelay    time.Duration
	PrefetchCount int
}

// RabbitMQ represents the RabbitMQ connection
type RabbitMQ struct {
	conn    *amqp.Connection
	channel *amqp.Channel
	Config  RabbitMQConfig
}

// Message represents a message in the queue
type Message struct {
	ID         string                 `json:"id"`
	Type       string                 `json:"type"`
	Data       map[string]interface{} `json:"data"`
	Timestamp  time.Time              `json:"timestamp"`
	Headers    map[string]interface{} `json:"headers,omitempty"`
	RetryCount int                    `json:"retry_count,omitempty"`
}

// DeadLetterMessage represents a failed message
type DeadLetterMessage struct {
	Message
	Error         string    `json:"error"`
	FailedAt      time.Time `json:"failed_at"`
	OriginalQueue string    `json:"original_queue"`
}

// NewRabbitMQ creates a new RabbitMQ connection
func NewRabbitMQ(config RabbitMQConfig) (*RabbitMQ, error) {
	conn, err := amqp.Dial(config.URL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("failed to open channel: %w", err)
	}

	// Set QoS
	err = ch.Qos(config.PrefetchCount, 0, false)
	if err != nil {
		return nil, fmt.Errorf("failed to set QoS: %w", err)
	}

	rabbitmq := &RabbitMQ{
		conn:    conn,
		channel: ch,
		Config:  config,
	}

	// Setup exchanges and queues
	if err := rabbitmq.setupExchangesAndQueues(); err != nil {
		return nil, fmt.Errorf("failed to setup exchanges and queues: %w", err)
	}

	log.Println("✅ RabbitMQ connected successfully")
	return rabbitmq, nil
}

// setupExchangesAndQueues sets up the required exchanges and queues
func (r *RabbitMQ) setupExchangesAndQueues() error {
	// Declare main exchange
	err := r.channel.ExchangeDeclare(
		r.Config.Exchange, // name
		"topic",           // type
		true,              // durable
		false,             // auto-deleted
		false,             // internal
		false,             // no-wait
		nil,               // arguments
	)
	if err != nil {
		return fmt.Errorf("failed to declare exchange: %w", err)
	}

	// Declare dead letter exchange
	err = r.channel.ExchangeDeclare(
		r.Config.Exchange+".dlx", // name
		"topic",                  // type
		true,                     // durable
		false,                    // auto-deleted
		false,                    // internal
		false,                    // no-wait
		nil,                      // arguments
	)
	if err != nil {
		return fmt.Errorf("failed to declare dead letter exchange: %w", err)
	}

	// Declare dead letter queue
	dlqName := r.Config.QueueName + ".dlq"
	_, err = r.channel.QueueDeclare(
		dlqName, // name
		true,    // durable
		false,   // delete when unused
		false,   // exclusive
		false,   // no-wait
		nil,     // arguments
	)
	if err != nil {
		return fmt.Errorf("failed to declare dead letter queue: %w", err)
	}

	// Bind dead letter queue to dead letter exchange
	err = r.channel.QueueBind(
		dlqName,                  // queue name
		"#",                      // routing key
		r.Config.Exchange+".dlx", // exchange
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to bind dead letter queue: %w", err)
	}

	// Declare main queue with dead letter configuration
	args := amqp.Table{
		"x-dead-letter-exchange":    r.Config.Exchange + ".dlx",
		"x-dead-letter-routing-key": "#",
		"x-message-ttl":             int32(24 * 60 * 60 * 1000), // 24 hours
	}

	_, err = r.channel.QueueDeclare(
		r.Config.QueueName, // name
		true,               // durable
		false,              // delete when unused
		false,              // exclusive
		false,              // no-wait
		args,               // arguments
	)
	if err != nil {
		return fmt.Errorf("failed to declare queue: %w", err)
	}

	// Bind main queue to exchange
	err = r.channel.QueueBind(
		r.Config.QueueName,  // queue name
		r.Config.RoutingKey, // routing key
		r.Config.Exchange,   // exchange
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to bind queue: %w", err)
	}

	return nil
}

// Close closes the RabbitMQ connection
func (r *RabbitMQ) Close() error {
	if r.channel != nil {
		r.channel.Close()
	}
	if r.conn != nil {
		return r.conn.Close()
	}
	return nil
}

// Publish publishes a message to the queue
func (r *RabbitMQ) Publish(ctx context.Context, message Message) error {
	body, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	headers := amqp.Table{}
	for k, v := range message.Headers {
		headers[k] = v
	}

	err = r.channel.PublishWithContext(ctx,
		r.Config.Exchange,   // exchange
		r.Config.RoutingKey, // routing key
		false,               // mandatory
		false,               // immediate
		amqp.Publishing{
			ContentType:  "application/json",
			Body:         body,
			Headers:      headers,
			DeliveryMode: amqp.Persistent,
			Timestamp:    time.Now(),
		},
	)
	if err != nil {
		return fmt.Errorf("failed to publish message: %w", err)
	}

	return nil
}

// Consume consumes messages from the queue
func (r *RabbitMQ) Consume(ctx context.Context, handler func(Message) error) error {
	msgs, err := r.channel.Consume(
		r.Config.QueueName, // queue
		"",                 // consumer
		false,              // auto-ack
		false,              // exclusive
		false,              // no-local
		false,              // no-wait
		nil,                // args
	)
	if err != nil {
		return fmt.Errorf("failed to start consuming: %w", err)
	}

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case msg := <-msgs:
				var message Message
				if err := json.Unmarshal(msg.Body, &message); err != nil {
					log.Printf("Failed to unmarshal message: %v", err)
					msg.Nack(false, false)
					continue
				}

				// Process message with retry logic
				if err := r.processMessageWithRetry(message, handler); err != nil {
					log.Printf("Failed to process message after retries: %v", err)
					msg.Nack(false, false)
				} else {
					msg.Ack(false)
				}
			}
		}
	}()

	return nil
}

// processMessageWithRetry processes a message with retry logic
func (r *RabbitMQ) processMessageWithRetry(message Message, handler func(Message) error) error {
	var lastErr error
	for attempt := 0; attempt <= r.Config.MaxRetries; attempt++ {
		if attempt > 0 {
			time.Sleep(r.Config.RetryDelay * time.Duration(attempt))
		}

		if err := handler(message); err != nil {
			lastErr = err
			log.Printf("Message processing attempt %d failed: %v", attempt+1, err)
			continue
		}

		return nil
	}

	return fmt.Errorf("message processing failed after %d attempts: %w", r.Config.MaxRetries+1, lastErr)
}

// GetDeadLetterMessages retrieves messages from the dead letter queue
func (r *RabbitMQ) GetDeadLetterMessages(limit int) ([]DeadLetterMessage, error) {
	var messages []DeadLetterMessage
	dlqName := r.Config.QueueName + ".dlq"

	for i := 0; i < limit; i++ {
		msg, ok, err := r.channel.Get(dlqName, false)
		if err != nil {
			return messages, fmt.Errorf("failed to get message from DLQ: %w", err)
		}
		if !ok {
			break
		}

		var dlqMessage DeadLetterMessage
		if err := json.Unmarshal(msg.Body, &dlqMessage); err != nil {
			log.Printf("Failed to unmarshal DLQ message: %v", err)
			msg.Nack(false, false)
			continue
		}

		messages = append(messages, dlqMessage)
		msg.Ack(false)
	}

	return messages, nil
}

// RepublishDeadLetterMessage republishes a message from the dead letter queue
func (r *RabbitMQ) RepublishDeadLetterMessage(messageID string) error {
	dlqName := r.Config.QueueName + ".dlq"

	// Get message from DLQ
	msg, ok, err := r.channel.Get(dlqName, false)
	if err != nil {
		return fmt.Errorf("failed to get message from DLQ: %w", err)
	}
	if !ok {
		return fmt.Errorf("message not found in DLQ")
	}

	var dlqMessage DeadLetterMessage
	if err := json.Unmarshal(msg.Body, &dlqMessage); err != nil {
		msg.Nack(false, false)
		return fmt.Errorf("failed to unmarshal DLQ message: %w", err)
	}

	// Check if this is the message we want to republish
	if dlqMessage.ID != messageID {
		msg.Nack(false, true) // Requeue the message
		return fmt.Errorf("message ID mismatch")
	}

	// Reset retry count and republish
	dlqMessage.RetryCount = 0
	dlqMessage.Headers["republished"] = true

	// Publish back to main queue
	if err := r.Publish(context.Background(), dlqMessage.Message); err != nil {
		msg.Nack(false, true)
		return fmt.Errorf("failed to republish message: %w", err)
	}

	// Acknowledge the DLQ message
	msg.Ack(false)
	return nil
}

// GetQueueStats returns queue statistics
func (r *RabbitMQ) GetQueueStats() (map[string]interface{}, error) {
	queue, err := r.channel.QueueInspect(r.Config.QueueName)
	if err != nil {
		return nil, fmt.Errorf("failed to inspect queue: %w", err)
	}

	dlqName := r.Config.QueueName + ".dlq"
	dlq, err := r.channel.QueueInspect(dlqName)
	if err != nil {
		return nil, fmt.Errorf("failed to inspect DLQ: %w", err)
	}

	return map[string]interface{}{
		"main_queue": map[string]interface{}{
			"name":      queue.Name,
			"messages":  queue.Messages,
			"consumers": queue.Consumers,
		},
		"dead_letter_queue": map[string]interface{}{
			"name":      dlq.Name,
			"messages":  dlq.Messages,
			"consumers": dlq.Consumers,
		},
	}, nil
}
