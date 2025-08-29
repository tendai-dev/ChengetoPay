package clients

import (
	"context"
	"fmt"
	"time"

	"github.com/project-x/microservices/shared/circuitbreaker"
	"github.com/project-x/microservices/shared/httpclient"
	"github.com/project-x/microservices/shared/servicediscovery"
)

// LedgerClient handles communication with the ledger service
type LedgerClient struct {
	client         *httpclient.ServiceClient
	circuitBreaker *circuitbreaker.CircuitBreaker
	loadBalancer   *servicediscovery.LoadBalancer
}

// NewLedgerClient creates a new ledger service client
func NewLedgerClient(loadBalancer *servicediscovery.LoadBalancer) *LedgerClient {
	cb := circuitbreaker.NewCircuitBreaker(circuitbreaker.Config{
		FailureThreshold: 5,
		SuccessThreshold: 3,
		Timeout:          60 * time.Second,
	})

	return &LedgerClient{
		circuitBreaker: cb,
		loadBalancer:   loadBalancer,
	}
}

// CreateAccountRequest represents an account creation request
type CreateAccountRequest struct {
	AccountID   string      `json:"account_id"`
	AccountType string      `json:"account_type"`
	Currency    string      `json:"currency"`
	Metadata    interface{} `json:"metadata,omitempty"`
}

// CreateJournalEntryRequest represents a journal entry creation request
type CreateJournalEntryRequest struct {
	Description string        `json:"description"`
	Entries     []EntryDetail `json:"entries"`
	Metadata    interface{}   `json:"metadata,omitempty"`
}

// EntryDetail represents a journal entry detail
type EntryDetail struct {
	AccountID string  `json:"account_id"`
	Amount    float64 `json:"amount"`
	Type      string  `json:"type"` // "debit" or "credit"
}

// Account represents a ledger account
type Account struct {
	ID          string      `json:"id"`
	AccountID   string      `json:"account_id"`
	AccountType string      `json:"account_type"`
	Balance     float64     `json:"balance"`
	Currency    string      `json:"currency"`
	Metadata    interface{} `json:"metadata,omitempty"`
	CreatedAt   time.Time   `json:"created_at"`
	UpdatedAt   time.Time   `json:"updated_at"`
}

// JournalEntry represents a journal entry
type JournalEntry struct {
	ID          string        `json:"id"`
	Description string        `json:"description"`
	Entries     []EntryDetail `json:"entries"`
	TotalAmount float64       `json:"total_amount"`
	Metadata    interface{}   `json:"metadata,omitempty"`
	CreatedAt   time.Time     `json:"created_at"`
}

// CreateAccount creates a new ledger account
func (c *LedgerClient) CreateAccount(ctx context.Context, req *CreateAccountRequest) (*Account, error) {
	var account *Account
	
	err := c.circuitBreaker.Execute(ctx, func() error {
		endpoint, err := c.loadBalancer.GetEndpoint("ledger-service")
		if err != nil {
			return fmt.Errorf("failed to get ledger service endpoint: %w", err)
		}

		if c.client == nil || c.client.BaseURL() != endpoint {
			c.client = httpclient.NewServiceClient(endpoint, 30*time.Second)
		}

		resp, err := c.client.Post(ctx, "/v1/accounts", req, nil)
		if err != nil {
			return err
		}

		account = &Account{}
		return resp.UnmarshalResponse(account)
	})

	return account, err
}

// GetAccount retrieves an account by ID
func (c *LedgerClient) GetAccount(ctx context.Context, accountID string) (*Account, error) {
	var account *Account
	
	err := c.circuitBreaker.Execute(ctx, func() error {
		endpoint, err := c.loadBalancer.GetEndpoint("ledger-service")
		if err != nil {
			return fmt.Errorf("failed to get ledger service endpoint: %w", err)
		}

		if c.client == nil || c.client.BaseURL() != endpoint {
			c.client = httpclient.NewServiceClient(endpoint, 30*time.Second)
		}

		resp, err := c.client.Get(ctx, fmt.Sprintf("/v1/accounts/%s", accountID), nil)
		if err != nil {
			return err
		}

		account = &Account{}
		return resp.UnmarshalResponse(account)
	})

	return account, err
}

// CreateJournalEntry creates a new journal entry
func (c *LedgerClient) CreateJournalEntry(ctx context.Context, req *CreateJournalEntryRequest) (*JournalEntry, error) {
	var entry *JournalEntry
	
	err := c.circuitBreaker.Execute(ctx, func() error {
		endpoint, err := c.loadBalancer.GetEndpoint("ledger-service")
		if err != nil {
			return fmt.Errorf("failed to get ledger service endpoint: %w", err)
		}

		if c.client == nil || c.client.BaseURL() != endpoint {
			c.client = httpclient.NewServiceClient(endpoint, 30*time.Second)
		}

		resp, err := c.client.Post(ctx, "/v1/journal-entries", req, nil)
		if err != nil {
			return err
		}

		entry = &JournalEntry{}
		return resp.UnmarshalResponse(entry)
	})

	return entry, err
}

// GetAccountBalance retrieves the current balance of an account
func (c *LedgerClient) GetAccountBalance(ctx context.Context, accountID string) (float64, error) {
	account, err := c.GetAccount(ctx, accountID)
	if err != nil {
		return 0, err
	}
	return account.Balance, nil
}
