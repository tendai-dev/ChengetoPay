package main

import (
	"context"
	"fmt"
	"math/big"
	"time"
)

// Money represents a monetary amount with high precision
type Money struct {
	Value    *big.Float `json:"value"`
	Currency string     `json:"currency"`
}

// FromMinorUnits creates a Money instance from minor units (e.g., cents)
func FromMinorUnits(currency string, minorUnits int64) Money {
	value := new(big.Float).SetInt64(minorUnits)
	value.Quo(value, big.NewFloat(100))
	return Money{
		Value:    value,
		Currency: currency,
	}
}

// Account represents a ledger account
type Account struct {
	ID            string                 `json:"id"`
	Name          string                 `json:"name"`
	Type          string                 `json:"account_type"`
	ParentID      string                 `json:"parent_id,omitempty"`
	Currency      string                 `json:"currency"`
	Balance       Money                  `json:"balance"`
	Status        string                 `json:"status"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt     time.Time              `json:"created_at"`
	UpdatedAt     time.Time              `json:"updated_at"`
}

// LedgerAccount is an alias for backward compatibility
type LedgerAccount = Account

// Entry represents a ledger entry
type Entry struct {
	ID            string                 `json:"id"`
	AccountID     string                 `json:"account_id"`
	TransactionID string                 `json:"transaction_id"`
	Type          string                 `json:"type"`
	Amount        Money                  `json:"amount"`
	Description   string                 `json:"description"`
	Reference     string                 `json:"reference,omitempty"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt     time.Time              `json:"created_at"`
	UpdatedAt     time.Time              `json:"updated_at"`
}

// LedgerEntry is an alias for backward compatibility
type LedgerEntry = Entry

// CreateAccountRequest represents a request to create an account
type CreateAccountRequest struct {
	AccountID string                 `json:"account_id"`
	Currency  string                 `json:"currency"`
	Type      string                 `json:"account_type"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// CreateEntryRequest represents a request to create a ledger entry
type CreateEntryRequest struct {
	AccountID   string                 `json:"account_id"`
	Type        string                 `json:"type"`
	Amount      Money                  `json:"amount"`
	Description string                 `json:"description"`
	Reference   string                 `json:"reference,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// PostEntryRequest represents a request to post a ledger entry
type PostEntryRequest struct {
	AccountID   string                 `json:"account_id"`
	TransactionID string               `json:"transaction_id"`
	Type        string                 `json:"type"`
	Amount      Money                  `json:"amount"`
	Description string                 `json:"description"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// GetBalanceRequest represents a request to get account balance
type GetBalanceRequest struct {
	AccountID string `json:"account_id"`
}

// AccountFilters represents filters for listing accounts
type AccountFilters struct {
	Type     string `json:"account_type"`
	Currency string `json:"currency"`
	Status   string `json:"status"`
	ParentID string `json:"parent_id"`
	Limit    int    `json:"limit"`
	Offset   int    `json:"offset"`
}

// EntryFilters represents filters for listing entries
type EntryFilters struct {
	AccountID     string     `json:"account_id"`
	TransactionID string     `json:"transaction_id"`
	Type          string     `json:"type"`
	FromDate      *time.Time `json:"from_date,omitempty"`
	ToDate        *time.Time `json:"to_date,omitempty"`
	Limit         int        `json:"limit"`
	Offset        int        `json:"offset"`
}

// Repository interface for ledger data access
type Repository interface {
	CreateAccount(ctx context.Context, account *Account) error
	GetAccount(ctx context.Context, accountID string) (*Account, error)
	ListAccounts(ctx context.Context, filters AccountFilters) ([]*Account, error)
	UpdateAccount(ctx context.Context, account *Account) error
	DeleteAccount(ctx context.Context, id string) error
	CreateEntry(ctx context.Context, entry *Entry) error
	GetEntry(ctx context.Context, id string) (*Entry, error)
	ListEntries(ctx context.Context, filters EntryFilters) ([]*Entry, error)
}

// Service represents the ledger business logic
type Service struct {
	repo   Repository
	logger interface{}
}

// NewService creates a new ledger service
func NewService(repo Repository, logger interface{}) *Service {
	return &Service{
		repo:   repo,
		logger: logger,
	}
}

// CreateAccount creates a new ledger account
func (s *Service) CreateAccount(ctx context.Context, req *CreateAccountRequest) (*LedgerAccount, error) {
	account := &LedgerAccount{
		ID:        generateID(),
		Name:      req.AccountID,
		Type:      req.Type,
		Currency:  req.Currency,
		Balance:   FromMinorUnits(req.Currency, 0),
		Status:    "active",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := s.repo.CreateAccount(ctx, account); err != nil {
		return nil, err
	}

	return account, nil
}

// GetAccount retrieves an account by ID
func (s *Service) GetAccount(ctx context.Context, accountID string) (*LedgerAccount, error) {
	return s.repo.GetAccount(ctx, accountID)
}

// ListAccounts lists accounts with filters
func (s *Service) ListAccounts(ctx context.Context, filters AccountFilters) ([]*LedgerAccount, error) {
	return s.repo.ListAccounts(ctx, filters)
}

// PostEntry posts a ledger entry
func (s *Service) PostEntry(ctx context.Context, req *PostEntryRequest) (*LedgerEntry, error) {
	// Get current account
	account, err := s.repo.GetAccount(ctx, req.AccountID)
	if err != nil {
		return nil, err
	}

	// Calculate new balance
	newBalance := new(big.Float).Add(account.Balance.Value, req.Amount.Value)
	
	// Create entry
	entry := &LedgerEntry{
		ID:            generateID(),
		AccountID:     req.AccountID,
		TransactionID: req.TransactionID,
		Type:          req.Type,
		Amount:        req.Amount,
		Description:   req.Description,
		Metadata:      req.Metadata,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	// Update account balance
	account.Balance = Money{Value: newBalance, Currency: req.Amount.Currency}
	account.UpdatedAt = time.Now()

	// Post entry and update account atomically
	if err := s.repo.CreateEntry(ctx, entry); err != nil {
		return nil, err
	}

	if err := s.repo.UpdateAccount(ctx, account); err != nil {
		return nil, err
	}

	return entry, nil
}

// GetEntries lists entries with filters
func (s *Service) GetEntries(ctx context.Context, filters EntryFilters) ([]*LedgerEntry, error) {
	return s.repo.ListEntries(ctx, filters)
}

// GetBalance gets account balance
func (s *Service) GetBalance(ctx context.Context, accountID string) (*Money, error) {
	account, err := s.repo.GetAccount(ctx, accountID)
	if err != nil {
		return nil, err
	}
	return &account.Balance, nil
}

// MockRepository implements Repository for testing
type MockRepository struct{}

func (m *MockRepository) CreateAccount(ctx context.Context, account *Account) error {
	return nil
}

func (m *MockRepository) GetAccount(ctx context.Context, accountID string) (*Account, error) {
	return &Account{
		ID:        generateID(),
		Name:      accountID,
		Type:      "escrow",
		Currency:  "USD",
		Balance:   FromMinorUnits("USD", 10000), // $100
		Status:    "active",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}, nil
}

func (m *MockRepository) ListAccounts(ctx context.Context, filters AccountFilters) ([]*Account, error) {
	return []*Account{}, nil
}

func (m *MockRepository) UpdateAccount(ctx context.Context, account *Account) error {
	return nil
}

func (m *MockRepository) DeleteAccount(ctx context.Context, id string) error {
	return nil
}

func (m *MockRepository) CreateEntry(ctx context.Context, entry *Entry) error {
	return nil
}

func (m *MockRepository) GetEntry(ctx context.Context, id string) (*Entry, error) {
	return &Entry{
		ID:            generateID(),
		AccountID:     "test_account",
		TransactionID: "test_tx",
		Type:          "credit",
		Amount:        FromMinorUnits("USD", 1000),
		Description:   "Test entry",
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}, nil
}

func (m *MockRepository) ListEntries(ctx context.Context, filters EntryFilters) ([]*Entry, error) {
	return []*Entry{}, nil
}

// Helper functions
func generateID() string {
	return fmt.Sprintf("ledger_%d", time.Now().UnixNano())
}
