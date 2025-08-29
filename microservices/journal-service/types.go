package main

import (
	"context"
	"errors"
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

// Journal represents an immutable journal entry
type Journal struct {
	ID            string    `json:"id"`
	TransactionID string    `json:"transaction_id"`
	Description   string    `json:"description"`
	Status        string    `json:"status"` // pending, posted, reversed
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// JournalEntry represents a single entry in a journal (debit or credit)
type JournalEntry struct {
	ID          string                 `json:"id"`
	JournalID   string                 `json:"journal_id"`
	AccountID   string                 `json:"account_id"`
	Type        string                 `json:"type"` // debit, credit
	Amount      Money                  `json:"amount"`
	Description string                 `json:"description"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt   time.Time              `json:"created_at"`
}

// BalanceProjection represents a balance projection for an account
type BalanceProjection struct {
	AccountID   string    `json:"account_id"`
	Currency    string    `json:"currency"`
	Balance     Money     `json:"balance"`
	AsOfDate    time.Time `json:"as_of_date"`
	LastEntryID string    `json:"last_entry_id"`
}

// CreateJournalRequest represents a request to create a journal
type CreateJournalRequest struct {
	TransactionID string                 `json:"transaction_id"`
	Description   string                 `json:"description"`
	Entries       []CreateEntryRequest   `json:"entries"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
}

// CreateEntryRequest represents a request to create a journal entry
type CreateEntryRequest struct {
	AccountID   string                 `json:"account_id"`
	Type        string                 `json:"type"` // debit, credit
	Amount      Money                  `json:"amount"`
	Description string                 `json:"description"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// PostEntryRequest represents a request to post a journal entry
type PostEntryRequest struct {
	JournalID   string                 `json:"journal_id"`
	AccountID   string                 `json:"account_id"`
	Type        string                 `json:"type"`
	Amount      Money                  `json:"amount"`
	Description string                 `json:"description"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// JournalFilters represents filters for listing journals
type JournalFilters struct {
	TransactionID string `json:"transaction_id"`
	Status        string `json:"status"`
	Limit         int    `json:"limit"`
	Offset        int    `json:"offset"`
}

// EntryFilters represents filters for listing entries
type EntryFilters struct {
	JournalID string `json:"journal_id"`
	AccountID string `json:"account_id"`
	Type      string `json:"type"`
	Limit     int    `json:"limit"`
	Offset    int    `json:"offset"`
}

// ProjectionFilters represents filters for balance projections
type ProjectionFilters struct {
	AccountID string    `json:"account_id"`
	Currency  string    `json:"currency"`
	AsOfDate  time.Time `json:"as_of_date"`
}

// Repository interface for journal data access
type Repository interface {
	CreateJournal(ctx context.Context, journal *Journal) error
	GetJournal(ctx context.Context, id string) (*Journal, error)
	ListJournals(ctx context.Context, filters JournalFilters) ([]*Journal, error)
	CreateEntry(ctx context.Context, entry *JournalEntry) error
	GetEntries(ctx context.Context, filters EntryFilters) ([]*JournalEntry, error)
	GetProjections(ctx context.Context, filters ProjectionFilters) ([]*BalanceProjection, error)
	ValidateDoubleEntry(ctx context.Context, entries []*JournalEntry) error
}

// Service represents the journal business logic
type Service struct {
	repo   Repository
	logger interface{}
}

// NewService creates a new journal service
func NewService(repo Repository, logger interface{}) *Service {
	return &Service{
		repo:   repo,
		logger: logger,
	}
}

// CreateJournal creates a new journal with double-entry validation
func (s *Service) CreateJournal(ctx context.Context, req *CreateJournalRequest) (*Journal, error) {
	// Create journal
	journal := &Journal{
		ID:            generateID(),
		TransactionID: req.TransactionID,
		Description:   req.Description,
		Status:        "pending",
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	// Create entries
	var entries []*JournalEntry
	for _, entryReq := range req.Entries {
		entry := &JournalEntry{
			ID:          generateID(),
			JournalID:   journal.ID,
			AccountID:   entryReq.AccountID,
			Type:        entryReq.Type,
			Amount:      entryReq.Amount,
			Description: entryReq.Description,
			Metadata:    entryReq.Metadata,
			CreatedAt:   time.Now(),
		}
		entries = append(entries, entry)
	}

	// Validate double-entry (debits = credits)
	if err := s.repo.ValidateDoubleEntry(ctx, entries); err != nil {
		return nil, err
	}

	// Save journal and entries
	if err := s.repo.CreateJournal(ctx, journal); err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if err := s.repo.CreateEntry(ctx, entry); err != nil {
			return nil, err
		}
	}

	journal.Status = "posted"
	journal.UpdatedAt = time.Now()

	return journal, nil
}

// GetJournal retrieves a journal by ID
func (s *Service) GetJournal(ctx context.Context, id string) (*Journal, error) {
	return s.repo.GetJournal(ctx, id)
}

// ListJournals lists journals with filters
func (s *Service) ListJournals(ctx context.Context, filters JournalFilters) ([]*Journal, error) {
	return s.repo.ListJournals(ctx, filters)
}

// PostEntry posts a journal entry
func (s *Service) PostEntry(ctx context.Context, req *PostEntryRequest) (*JournalEntry, error) {
	entry := &JournalEntry{
		ID:          generateID(),
		JournalID:   req.JournalID,
		AccountID:   req.AccountID,
		Type:        req.Type,
		Amount:      req.Amount,
		Description: req.Description,
		Metadata:    req.Metadata,
		CreatedAt:   time.Now(),
	}

	if err := s.repo.CreateEntry(ctx, entry); err != nil {
		return nil, err
	}

	return entry, nil
}

// GetEntries lists entries with filters
func (s *Service) GetEntries(ctx context.Context, filters EntryFilters) ([]*JournalEntry, error) {
	return s.repo.GetEntries(ctx, filters)
}

// GetProjections gets balance projections
func (s *Service) GetProjections(ctx context.Context, filters ProjectionFilters) ([]*BalanceProjection, error) {
	return s.repo.GetProjections(ctx, filters)
}

// MockRepository implements Repository for testing
type MockRepository struct{}

func (m *MockRepository) CreateJournal(ctx context.Context, journal *Journal) error {
	return nil
}

func (m *MockRepository) GetJournal(ctx context.Context, id string) (*Journal, error) {
	return &Journal{
		ID:            id,
		TransactionID: "txn_123",
		Description:   "Sample journal entry",
		Status:        "posted",
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}, nil
}

func (m *MockRepository) ListJournals(ctx context.Context, filters JournalFilters) ([]*Journal, error) {
	return []*Journal{}, nil
}

func (m *MockRepository) CreateEntry(ctx context.Context, entry *JournalEntry) error {
	return nil
}

func (m *MockRepository) GetEntries(ctx context.Context, filters EntryFilters) ([]*JournalEntry, error) {
	return []*JournalEntry{}, nil
}

func (m *MockRepository) GetProjections(ctx context.Context, filters ProjectionFilters) ([]*BalanceProjection, error) {
	return []*BalanceProjection{}, nil
}

func (m *MockRepository) ValidateDoubleEntry(ctx context.Context, entries []*JournalEntry) error {
	var totalDebits, totalCredits big.Float

	for _, entry := range entries {
		if entry.Type == "debit" {
			totalDebits.Add(&totalDebits, entry.Amount.Value)
		} else if entry.Type == "credit" {
			totalCredits.Add(&totalCredits, entry.Amount.Value)
		}
	}

	if totalDebits.Cmp(&totalCredits) != 0 {
		return errors.New("double-entry validation failed: debits must equal credits")
	}

	return nil
}

// Helper functions
func generateID() string {
	return fmt.Sprintf("journal_%d", time.Now().UnixNano())
}
