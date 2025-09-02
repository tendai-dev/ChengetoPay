package clients

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// EscrowClient handles communication with the escrow service
type EscrowClient struct {
	baseURL string
	client  *http.Client
}

// NewEscrowClient creates a new escrow service client
func NewEscrowClient(baseURL string) *EscrowClient {
	return &EscrowClient{
		baseURL: baseURL,
		client: &http.Client{
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
	data, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	reqHTTP, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/v1/escrows", bytes.NewBuffer(data))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	reqHTTP.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(reqHTTP)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(body))
	}

	var escrow Escrow
	if err := json.NewDecoder(resp.Body).Decode(&escrow); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &escrow, nil
}

// GetEscrow retrieves an escrow by ID
func (c *EscrowClient) GetEscrow(ctx context.Context, escrowID string) (*Escrow, error) {
	reqHTTP, err := http.NewRequestWithContext(ctx, "GET", c.baseURL+"/v1/escrows/"+escrowID, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.client.Do(reqHTTP)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(body))
	}

	var escrow Escrow
	if err := json.NewDecoder(resp.Body).Decode(&escrow); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &escrow, nil
}

// FundEscrow funds an escrow
func (c *EscrowClient) FundEscrow(ctx context.Context, escrowID string) error {
	req := map[string]string{"escrow_id": escrowID}
	data, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	reqHTTP, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/v1/escrows/"+escrowID+"/fund", bytes.NewBuffer(data))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	reqHTTP.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(reqHTTP)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to fund escrow: %d, body: %s", resp.StatusCode, string(body))
	}

	return nil
}

// ReleaseEscrow releases funds from an escrow
func (c *EscrowClient) ReleaseEscrow(ctx context.Context, escrowID string) error {
	reqHTTP, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/v1/escrows/"+escrowID+"/release", nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.client.Do(reqHTTP)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to release escrow: %d, body: %s", resp.StatusCode, string(body))
	}

	return nil
}

// CancelEscrow cancels an escrow
func (c *EscrowClient) CancelEscrow(ctx context.Context, escrowID string) error {
	reqHTTP, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/v1/escrows/"+escrowID+"/cancel", nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.client.Do(reqHTTP)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to cancel escrow: %d, body: %s", resp.StatusCode, string(body))
	}

	return nil
}
