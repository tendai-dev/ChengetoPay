package clients

import (
	"context"
	"fmt"
	"time"

	"github.com/project-x/microservices/shared/circuitbreaker"
	"github.com/project-x/microservices/shared/httpclient"
	"github.com/project-x/microservices/shared/servicediscovery"
)

// PaymentClient handles communication with the payment service
type PaymentClient struct {
	client         *httpclient.ServiceClient
	circuitBreaker *circuitbreaker.CircuitBreaker
	loadBalancer   *servicediscovery.LoadBalancer
}

// NewPaymentClient creates a new payment service client
func NewPaymentClient(loadBalancer *servicediscovery.LoadBalancer) *PaymentClient {
	cb := circuitbreaker.NewCircuitBreaker(circuitbreaker.Config{
		FailureThreshold: 5,
		SuccessThreshold: 3,
		Timeout:          60 * time.Second,
	})

	return &PaymentClient{
		circuitBreaker: cb,
		loadBalancer:   loadBalancer,
	}
}

// CreatePaymentRequest represents a payment creation request
type CreatePaymentRequest struct {
	Amount        Money       `json:"amount"`
	PaymentMethod string      `json:"payment_method"`
	Provider      string      `json:"provider"`
	Description   string      `json:"description"`
	Metadata      interface{} `json:"metadata,omitempty"`
}

// ProcessPaymentRequest represents a payment processing request
type ProcessPaymentRequest struct {
	PaymentID string      `json:"payment_id"`
	Metadata  interface{} `json:"metadata,omitempty"`
}

// Payment represents a payment
type Payment struct {
	ID            string      `json:"id"`
	Amount        Money       `json:"amount"`
	PaymentMethod string      `json:"payment_method"`
	Provider      string      `json:"provider"`
	Status        string      `json:"status"`
	Description   string      `json:"description"`
	Metadata      interface{} `json:"metadata,omitempty"`
	CreatedAt     time.Time   `json:"created_at"`
	UpdatedAt     time.Time   `json:"updated_at"`
}

// CreatePayment creates a new payment
func (c *PaymentClient) CreatePayment(ctx context.Context, req *CreatePaymentRequest) (*Payment, error) {
	var payment *Payment
	
	err := c.circuitBreaker.Execute(ctx, func() error {
		endpoint, err := c.loadBalancer.GetEndpoint("payment-service")
		if err != nil {
			return fmt.Errorf("failed to get payment service endpoint: %w", err)
		}

		if c.client == nil || c.client.BaseURL() != endpoint {
			c.client = httpclient.NewServiceClient(endpoint, 30*time.Second)
		}

		resp, err := c.client.Post(ctx, "/v1/payments", req, nil)
		if err != nil {
			return err
		}

		payment = &Payment{}
		return resp.UnmarshalResponse(payment)
	})

	return payment, err
}

// ProcessPayment processes a payment
func (c *PaymentClient) ProcessPayment(ctx context.Context, paymentID string) (*Payment, error) {
	var payment *Payment
	
	err := c.circuitBreaker.Execute(ctx, func() error {
		endpoint, err := c.loadBalancer.GetEndpoint("payment-service")
		if err != nil {
			return fmt.Errorf("failed to get payment service endpoint: %w", err)
		}

		if c.client == nil || c.client.BaseURL() != endpoint {
			c.client = httpclient.NewServiceClient(endpoint, 30*time.Second)
		}

		body := map[string]string{"payment_id": paymentID}
		resp, err := c.client.Post(ctx, "/v1/payments/process", body, nil)
		if err != nil {
			return err
		}

		payment = &Payment{}
		return resp.UnmarshalResponse(payment)
	})

	return payment, err
}

// GetPayment retrieves a payment by ID
func (c *PaymentClient) GetPayment(ctx context.Context, paymentID string) (*Payment, error) {
	var payment *Payment
	
	err := c.circuitBreaker.Execute(ctx, func() error {
		endpoint, err := c.loadBalancer.GetEndpoint("payment-service")
		if err != nil {
			return fmt.Errorf("failed to get payment service endpoint: %w", err)
		}

		if c.client == nil || c.client.BaseURL() != endpoint {
			c.client = httpclient.NewServiceClient(endpoint, 30*time.Second)
		}

		resp, err := c.client.Get(ctx, fmt.Sprintf("/v1/payments/%s", paymentID), nil)
		if err != nil {
			return err
		}

		payment = &Payment{}
		return resp.UnmarshalResponse(payment)
	})

	return payment, err
}

// RefundPayment refunds a payment
func (c *PaymentClient) RefundPayment(ctx context.Context, paymentID string, amount *Money) (*Payment, error) {
	var payment *Payment
	
	err := c.circuitBreaker.Execute(ctx, func() error {
		endpoint, err := c.loadBalancer.GetEndpoint("payment-service")
		if err != nil {
			return fmt.Errorf("failed to get payment service endpoint: %w", err)
		}

		if c.client == nil || c.client.BaseURL() != endpoint {
			c.client = httpclient.NewServiceClient(endpoint, 30*time.Second)
		}

		body := map[string]interface{}{
			"payment_id": paymentID,
			"amount":     amount,
		}
		resp, err := c.client.Post(ctx, "/v1/payments/refund", body, nil)
		if err != nil {
			return err
		}

		payment = &Payment{}
		return resp.UnmarshalResponse(payment)
	})

	return payment, err
}
