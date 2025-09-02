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

// LedgerClient handles communication with the ledger service
type LedgerClient struct {
	baseURL string
	client  *http.Client
}

// NewLedgerClient creates a new ledger service client
func NewLedgerClient(baseURL string) *LedgerClient {
	return &LedgerClient{
		baseURL: baseURL,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
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

// CreateAccount creates a new account in the ledger
func (c *LedgerClient) CreateAccount(ctx context.Context, req *CreateAccountRequest) (*Account, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	reqHTTP, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/v1/accounts", bytes.NewReader(body))
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
		return nil, fmt.Errorf("failed to create account: %d, body: %s", resp.StatusCode, string(body))
	}

	var account Account
	if err := json.NewDecoder(resp.Body).Decode(&account); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &account, nil
}

// GetAccount retrieves an account by ID
func (c *LedgerClient) GetAccount(ctx context.Context, accountID string) (*Account, error) {
	reqHTTP, err := http.NewRequestWithContext(ctx, "GET", c.baseURL+"/v1/accounts/"+accountID, nil)
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
		return nil, fmt.Errorf("failed to get account: %d, body: %s", resp.StatusCode, string(body))
	}

	var account Account
	if err := json.NewDecoder(resp.Body).Decode(&account); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &account, nil
}

// CreateJournalEntry creates a new journal entry
func (c *LedgerClient) CreateJournalEntry(ctx context.Context, req *CreateJournalEntryRequest) (*JournalEntry, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	reqHTTP, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/v1/journal-entries", bytes.NewReader(body))
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
		return nil, fmt.Errorf("failed to create journal entry: %d, body: %s", resp.StatusCode, string(body))
	}

	var entry JournalEntry
	if err := json.NewDecoder(resp.Body).Decode(&entry); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &entry, nil
}

// GetAccountBalance retrieves the current balance of an account
func (c *LedgerClient) GetAccountBalance(ctx context.Context, accountID string) (float64, error) {
	account, err := c.GetAccount(ctx, accountID)
	if err != nil {
		return 0, err
	}
	return account.Balance, nil
}
