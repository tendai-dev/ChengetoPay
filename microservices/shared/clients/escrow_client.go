package clients

import (
	"context"
	"fmt"
	"time"

	"github.com/project-x/microservices/shared/circuitbreaker"
	"github.com/project-x/microservices/shared/httpclient"
	"github.com/project-x/microservices/shared/servicediscovery"
)

// EscrowClient handles communication with the escrow service
type EscrowClient struct {
	client         *httpclient.ServiceClient
	circuitBreaker *circuitbreaker.CircuitBreaker
	loadBalancer   *servicediscovery.LoadBalancer
}

// NewEscrowClient creates a new escrow service client
func NewEscrowClient(loadBalancer *servicediscovery.LoadBalancer) *EscrowClient {
	cb := circuitbreaker.NewCircuitBreaker(circuitbreaker.Config{
		FailureThreshold: 5,
		SuccessThreshold: 3,
		Timeout:          60 * time.Second,
	})

	return &EscrowClient{
		circuitBreaker: cb,
		loadBalancer:   loadBalancer,
	}
}

// CreateEscrowRequest represents an escrow creation request
type CreateEscrowRequest struct {
	BuyerID  string      `json:"buyer_id"`
	SellerID string      `json:"seller_id"`
	Amount   Money       `json:"amount"`
	Terms    string      `json:"terms"`
	Metadata interface{} `json:"metadata,omitempty"`
}

// Escrow represents an escrow
type Escrow struct {
	ID        string      `json:"id"`
	BuyerID   string      `json:"buyer_id"`
	SellerID  string      `json:"seller_id"`
	Amount    Money       `json:"amount"`
	Currency  string      `json:"currency"`
	Status    string      `json:"status"`
	Terms     string      `json:"terms"`
	Metadata  interface{} `json:"metadata,omitempty"`
	CreatedAt time.Time   `json:"created_at"`
	UpdatedAt time.Time   `json:"updated_at"`
}

// Money represents a monetary amount
type Money struct {
	Value    float64 `json:"value"`
	Currency string  `json:"currency"`
}

// CreateEscrow creates a new escrow
func (c *EscrowClient) CreateEscrow(ctx context.Context, req *CreateEscrowRequest) (*Escrow, error) {
	var escrow *Escrow
	
	err := c.circuitBreaker.Execute(ctx, func() error {
		endpoint, err := c.loadBalancer.GetEndpoint("escrow-service")
		if err != nil {
			return fmt.Errorf("failed to get escrow service endpoint: %w", err)
		}

		if c.client == nil {
			c.client = httpclient.NewServiceClient(endpoint, 30*time.Second)
		}

		resp, err := c.client.Post(ctx, "/v1/escrows", req, nil)
		if err != nil {
			return err
		}

		escrow = &Escrow{}
		return resp.UnmarshalResponse(escrow)
	})

	return escrow, err
}

// GetEscrow retrieves an escrow by ID
func (c *EscrowClient) GetEscrow(ctx context.Context, escrowID string) (*Escrow, error) {
	var escrow *Escrow
	
	err := c.circuitBreaker.Execute(ctx, func() error {
		endpoint, err := c.loadBalancer.GetEndpoint("escrow-service")
		if err != nil {
			return fmt.Errorf("failed to get escrow service endpoint: %w", err)
		}

		if c.client == nil || c.client.BaseURL() != endpoint {
			c.client = httpclient.NewServiceClient(endpoint, 30*time.Second)
		}

		resp, err := c.client.Get(ctx, fmt.Sprintf("/v1/escrows/%s", escrowID), nil)
		if err != nil {
			return err
		}

		escrow = &Escrow{}
		return resp.UnmarshalResponse(escrow)
	})

	return escrow, err
}

// FundEscrow funds an escrow
func (c *EscrowClient) FundEscrow(ctx context.Context, escrowID, paymentID string) error {
	return c.circuitBreaker.Execute(ctx, func() error {
		endpoint, err := c.loadBalancer.GetEndpoint("escrow-service")
		if err != nil {
			return fmt.Errorf("failed to get escrow service endpoint: %w", err)
		}

		if c.client == nil || c.client.BaseURL() != endpoint {
			c.client = httpclient.NewServiceClient(endpoint, 30*time.Second)
		}

		body := map[string]string{"payment_id": paymentID}
		resp, err := c.client.Post(ctx, fmt.Sprintf("/v1/escrows/%s/fund", escrowID), body, nil)
		if err != nil {
			return err
		}

		if !resp.IsSuccess() {
			return fmt.Errorf("failed to fund escrow: %s", string(resp.Body))
		}

		return nil
	})
}

// ReleaseEscrow releases funds from an escrow
func (c *EscrowClient) ReleaseEscrow(ctx context.Context, escrowID string) error {
	return c.circuitBreaker.Execute(ctx, func() error {
		endpoint, err := c.loadBalancer.GetEndpoint("escrow-service")
		if err != nil {
			return fmt.Errorf("failed to get escrow service endpoint: %w", err)
		}

		if c.client == nil || c.client.BaseURL() != endpoint {
			c.client = httpclient.NewServiceClient(endpoint, 30*time.Second)
		}

		resp, err := c.client.Post(ctx, fmt.Sprintf("/v1/escrows/%s/release", escrowID), nil, nil)
		if err != nil {
			return err
		}

		if !resp.IsSuccess() {
			return fmt.Errorf("failed to release escrow: %s", string(resp.Body))
		}

		return nil
	})
}

// CancelEscrow cancels an escrow
func (c *EscrowClient) CancelEscrow(ctx context.Context, escrowID string) error {
	return c.circuitBreaker.Execute(ctx, func() error {
		endpoint, err := c.loadBalancer.GetEndpoint("escrow-service")
		if err != nil {
			return fmt.Errorf("failed to get escrow service endpoint: %w", err)
		}

		if c.client == nil || c.client.BaseURL() != endpoint {
			c.client = httpclient.NewServiceClient(endpoint, 30*time.Second)
		}

		resp, err := c.client.Post(ctx, fmt.Sprintf("/v1/escrows/%s/cancel", escrowID), nil, nil)
		if err != nil {
			return err
		}

		if !resp.IsSuccess() {
			return fmt.Errorf("failed to cancel escrow: %s", string(resp.Body))
		}

		return nil
	})
}
