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

// Refund represents a refund transaction
type Refund struct {
	ID            string    `json:"id"`
	PaymentID     string    `json:"payment_id"`
	TransactionID string    `json:"transaction_id"`
	Amount        Money     `json:"amount"`
	Currency      string    `json:"currency"`
	Type          string    `json:"type"` // full, partial, multiple
	Reason        string    `json:"reason"`
	Status        string    `json:"status"` // pending, processing, completed, failed
	ProviderRef   string    `json:"provider_ref,omitempty"`
	IdempotencyKey string   `json:"idempotency_key"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// RefundItem represents an individual refund item in a multiple refund
type RefundItem struct {
	ID            string    `json:"id"`
	RefundID      string    `json:"refund_id"`
	PaymentID     string    `json:"payment_id"`
	Amount        Money     `json:"amount"`
	Reason        string    `json:"reason"`
	Status        string    `json:"status"`
	CreatedAt     time.Time `json:"created_at"`
}

// Reconciliation represents a refund reconciliation record
type Reconciliation struct {
	ID            string    `json:"id"`
	RefundID      string    `json:"refund_id"`
	ProviderRef   string    `json:"provider_ref"`
	Status        string    `json:"status"` // pending, matched, unmatched
	Amount        Money     `json:"amount"`
	ReconciledAt  *time.Time `json:"reconciled_at,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
}

// CreateRefundRequest represents a request to create a refund
type CreateRefundRequest struct {
	PaymentID      string                 `json:"payment_id"`
	TransactionID  string                 `json:"transaction_id"`
	Amount         Money                  `json:"amount"`
	Type           string                 `json:"type"`
	Reason         string                 `json:"reason"`
	IdempotencyKey string                 `json:"idempotency_key"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
}

// ProcessRefundRequest represents a request to process a refund
type ProcessRefundRequest struct {
	RefundID       string `json:"refund_id"`
	ProviderRef    string `json:"provider_ref,omitempty"`
	IdempotencyKey string `json:"idempotency_key"`
}

// ReconcileRefundRequest represents a request to reconcile a refund
type ReconcileRefundRequest struct {
	RefundID     string `json:"refund_id"`
	ProviderRef  string `json:"provider_ref"`
	Amount       Money  `json:"amount"`
	Status       string `json:"status"`
}

// RefundFilters represents filters for listing refunds
type RefundFilters struct {
	PaymentID     string `json:"payment_id"`
	TransactionID string `json:"transaction_id"`
	Type          string `json:"type"`
	Status        string `json:"status"`
	Limit         int    `json:"limit"`
	Offset        int    `json:"offset"`
}

// Repository interface for refunds data access
type Repository interface {
	CreateRefund(ctx context.Context, refund *Refund) error
	GetRefund(ctx context.Context, id string) (*Refund, error)
	ListRefunds(ctx context.Context, filters RefundFilters) ([]*Refund, error)
	UpdateRefund(ctx context.Context, refund *Refund) error
	CreateRefundItem(ctx context.Context, item *RefundItem) error
	GetRefundByKey(ctx context.Context, idempotencyKey string) (*Refund, error)
	CreateReconciliation(ctx context.Context, reconciliation *Reconciliation) error
	UpdateReconciliation(ctx context.Context, reconciliation *Reconciliation) error
}

// Service represents the refunds business logic
type Service struct {
	repo   Repository
	logger interface{}
}

// NewService creates a new refunds service
func NewService(repo Repository, logger interface{}) *Service {
	return &Service{
		repo:   repo,
		logger: logger,
	}
}

// CreateRefund creates a new refund with idempotency check
func (s *Service) CreateRefund(ctx context.Context, req *CreateRefundRequest) (*Refund, error) {
	// Check idempotency key
	if existingRefund, err := s.repo.GetRefundByKey(ctx, req.IdempotencyKey); err == nil && existingRefund != nil {
		return existingRefund, nil
	}

	refund := &Refund{
		ID:            generateID(),
		PaymentID:     req.PaymentID,
		TransactionID: req.TransactionID,
		Amount:        req.Amount,
		Currency:      req.Amount.Currency,
		Type:          req.Type,
		Reason:        req.Reason,
		Status:        "pending",
		IdempotencyKey: req.IdempotencyKey,
		Metadata:      req.Metadata,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	if err := s.repo.CreateRefund(ctx, refund); err != nil {
		return nil, err
	}

	return refund, nil
}

// GetRefund retrieves a refund by ID
func (s *Service) GetRefund(ctx context.Context, id string) (*Refund, error) {
	return s.repo.GetRefund(ctx, id)
}

// ListRefunds lists refunds with filters
func (s *Service) ListRefunds(ctx context.Context, filters RefundFilters) ([]*Refund, error) {
	return s.repo.ListRefunds(ctx, filters)
}

// ProcessRefund processes a refund
func (s *Service) ProcessRefund(ctx context.Context, req *ProcessRefundRequest) error {
	refund, err := s.repo.GetRefund(ctx, req.RefundID)
	if err != nil {
		return err
	}

	// Check idempotency key
	if existingRefund, err := s.repo.GetRefundByKey(ctx, req.IdempotencyKey); err == nil && existingRefund != nil {
		return nil // Already processed
	}

	refund.Status = "processing"
	refund.ProviderRef = req.ProviderRef
	refund.UpdatedAt = time.Now()

	if err := s.repo.UpdateRefund(ctx, refund); err != nil {
		return err
	}

	// Simulate async processing
	go s.processRefundAsync(refund)

	return nil
}

// ReconcileRefund reconciles a refund with provider data
func (s *Service) ReconcileRefund(ctx context.Context, req *ReconcileRefundRequest) (*Reconciliation, error) {
	reconciliation := &Reconciliation{
		ID:           generateID(),
		RefundID:     req.RefundID,
		ProviderRef:  req.ProviderRef,
		Status:       req.Status,
		Amount:       req.Amount,
		CreatedAt:    time.Now(),
	}

	if req.Status == "matched" {
		now := time.Now()
		reconciliation.ReconciledAt = &now
	}

	if err := s.repo.CreateReconciliation(ctx, reconciliation); err != nil {
		return nil, err
	}

	// Update refund status if reconciled
	if req.Status == "matched" {
		refund, err := s.repo.GetRefund(ctx, req.RefundID)
		if err == nil {
			refund.Status = "completed"
			refund.UpdatedAt = time.Now()
			s.repo.UpdateRefund(ctx, refund)
		}
	}

	return reconciliation, nil
}

// processRefundAsync simulates async refund processing
func (s *Service) processRefundAsync(refund *Refund) {
	// Simulate processing delay
	time.Sleep(2 * time.Second)

	// Update refund status
	refund.Status = "completed"
	refund.UpdatedAt = time.Now()
	s.repo.UpdateRefund(context.Background(), refund)
}

// MockRepository implements Repository for testing
type MockRepository struct{}

func (m *MockRepository) CreateRefund(ctx context.Context, refund *Refund) error {
	return nil
}

func (m *MockRepository) GetRefund(ctx context.Context, id string) (*Refund, error) {
	return &Refund{
		ID:            id,
		PaymentID:     "payment_123",
		TransactionID: "txn_123",
		Amount:        FromMinorUnits("USD", 5000), // $50
		Currency:      "USD",
		Type:          "full",
		Reason:        "Customer request",
		Status:        "pending",
		IdempotencyKey: "key_123",
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}, nil
}

func (m *MockRepository) ListRefunds(ctx context.Context, filters RefundFilters) ([]*Refund, error) {
	return []*Refund{}, nil
}

func (m *MockRepository) UpdateRefund(ctx context.Context, refund *Refund) error {
	return nil
}

func (m *MockRepository) CreateRefundItem(ctx context.Context, item *RefundItem) error {
	return nil
}

func (m *MockRepository) GetRefundByKey(ctx context.Context, idempotencyKey string) (*Refund, error) {
	return nil, nil // No existing refund found
}

func (m *MockRepository) CreateReconciliation(ctx context.Context, reconciliation *Reconciliation) error {
	return nil
}

func (m *MockRepository) UpdateReconciliation(ctx context.Context, reconciliation *Reconciliation) error {
	return nil
}

// Helper functions
func generateID() string {
	return fmt.Sprintf("refund_%d", time.Now().UnixNano())
}
