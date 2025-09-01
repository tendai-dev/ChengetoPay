package clients

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// EscrowClient handles communication with the escrow service
type EscrowClient struct {
	baseURL    string
	httpClient *http.Client
}

// NewEscrowClient creates a new escrow service client
func NewEscrowClient(baseURL string) *EscrowClient {
	return &EscrowClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
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
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", fmt.Sprintf("%s/api/v1/escrows", c.baseURL), bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var escrow Escrow
	if err := json.NewDecoder(resp.Body).Decode(&escrow); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &escrow, nil
}

// GetEscrow retrieves an escrow by ID
func (c *EscrowClient) GetEscrow(ctx context.Context, escrowID string) (*Escrow, error) {
	httpReq, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%s/api/v1/escrows/%s", c.baseURL, escrowID), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var escrow Escrow
	if err := json.NewDecoder(resp.Body).Decode(&escrow); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &escrow, nil
}

// FundEscrow funds an escrow
func (c *EscrowClient) FundEscrow(ctx context.Context, escrowID, paymentID string) error {
	body, _ := json.Marshal(map[string]string{"payment_id": paymentID})
	
	httpReq, err := http.NewRequestWithContext(ctx, "POST", fmt.Sprintf("%s/api/v1/escrows/%s/fund", c.baseURL, escrowID), bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("failed to fund escrow: status %d", resp.StatusCode)
	}

	return nil
}

// ReleaseEscrow releases funds from an escrow
func (c *EscrowClient) ReleaseEscrow(ctx context.Context, escrowID string) error {
	httpReq, err := http.NewRequestWithContext(ctx, "POST", fmt.Sprintf("%s/api/v1/escrows/%s/release", c.baseURL, escrowID), nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("failed to release escrow: status %d", resp.StatusCode)
	}

	return nil
}

// CancelEscrow cancels an escrow
func (c *EscrowClient) CancelEscrow(ctx context.Context, escrowID string) error {
	httpReq, err := http.NewRequestWithContext(ctx, "POST", fmt.Sprintf("%s/api/v1/escrows/%s/cancel", c.baseURL, escrowID), nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("failed to cancel escrow: status %d", resp.StatusCode)
	}

	return nil
}
