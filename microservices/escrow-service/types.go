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

// Escrow represents an escrow transaction
type Escrow struct {
	ID        string                 `json:"id"`
	BuyerID   string                 `json:"buyer_id"`
	SellerID  string                 `json:"seller_id"`
	Amount    Money                  `json:"amount"`
	Currency  string                 `json:"currency"`
	Status    string                 `json:"status"`
	Terms     string                 `json:"terms"`
	HoldID    string                 `json:"hold_id,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt time.Time              `json:"created_at"`
	UpdatedAt time.Time              `json:"updated_at"`
}

// CreateEscrowRequest represents a request to create an escrow
type CreateEscrowRequest struct {
	BuyerID  string `json:"buyer_id"`
	SellerID string `json:"seller_id"`
	Amount   Money  `json:"amount"`
	Terms    string `json:"terms"`
}

// FundEscrowRequest represents a request to fund an escrow
type FundEscrowRequest struct {
	EscrowID      string `json:"escrow_id"`
	Amount        Money  `json:"amount"`
	PaymentMethod string `json:"payment_method,omitempty"`
}

// ConfirmDeliveryRequest represents a request to confirm delivery
type ConfirmDeliveryRequest struct {
	EscrowID string `json:"escrow_id"`
	Proof    string `json:"proof"`
}

// EscrowFilters represents filters for listing escrows
type EscrowFilters struct {
	Status   string `json:"status"`
	BuyerID  string `json:"buyer_id"`
	SellerID string `json:"seller_id"`
	Limit    int    `json:"limit"`
	Offset   int    `json:"offset"`
}

// Repository interface for escrow data access
type Repository interface {
	CreateEscrow(ctx context.Context, escrow *Escrow) error
	GetEscrow(ctx context.Context, id string) (*Escrow, error)
	ListEscrows(ctx context.Context, filters EscrowFilters) ([]*Escrow, error)
	UpdateEscrow(ctx context.Context, escrow *Escrow) error
	DeleteEscrow(ctx context.Context, id string) error
}

// Service represents the escrow business logic
type Service struct {
	repo   Repository
	logger interface{}
}

// NewService creates a new escrow service
func NewService(repo Repository, logger interface{}) *Service {
	return &Service{
		repo:   repo,
		logger: logger,
	}
}

// CreateEscrow creates a new escrow with comprehensive validation
func (s *Service) CreateEscrow(ctx context.Context, req *CreateEscrowRequest) (*Escrow, error) {
	// Validate request
	if err := ValidateCreateEscrowRequest(req); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Business rule validation
	validator := NewBusinessRuleValidator()
	
	// Check escrow limits
	if err := validator.ValidateEscrowLimits(req.Amount); err != nil {
		return nil, fmt.Errorf("business rule violation: %w", err)
	}

	// Check participant eligibility
	if err := validator.ValidateParticipantEligibility(req.BuyerID, req.SellerID); err != nil {
		return nil, fmt.Errorf("participant eligibility check failed: %w", err)
	}

	// Risk assessment
	if err := validator.ValidateEscrowRisk(req); err != nil {
		return nil, fmt.Errorf("risk assessment failed: %w", err)
	}

	// Create escrow with metadata
	escrow := &Escrow{
		ID:        generateID(),
		BuyerID:   req.BuyerID,
		SellerID:  req.SellerID,
		Amount:    req.Amount,
		Currency:  req.Amount.Currency,
		Status:    "pending",
		Terms:     req.Terms,
		Metadata: map[string]interface{}{
			"created_by": "system",
			"risk_score": calculateRiskScore(req),
			"validation_passed": true,
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := s.repo.CreateEscrow(ctx, escrow); err != nil {
		return nil, fmt.Errorf("failed to create escrow: %w", err)
	}

	return escrow, nil
}

// GetEscrow retrieves an escrow by ID
func (s *Service) GetEscrow(ctx context.Context, id string) (*Escrow, error) {
	return s.repo.GetEscrow(ctx, id)
}

// ListEscrows lists escrows with filters
func (s *Service) ListEscrows(ctx context.Context, filters EscrowFilters) ([]*Escrow, error) {
	return s.repo.ListEscrows(ctx, filters)
}

// FundEscrow funds an escrow with validation and state management
func (s *Service) FundEscrow(ctx context.Context, req *FundEscrowRequest) error {
	// Validate request
	if err := ValidateFundEscrowRequest(req); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	// Get current escrow
	escrow, err := s.repo.GetEscrow(ctx, req.EscrowID)
	if err != nil {
		return fmt.Errorf("failed to get escrow: %w", err)
	}

	// Validate state transition
	if err := ValidateEscrowAction(escrow, "fund"); err != nil {
		return fmt.Errorf("action not allowed: %w", err)
	}

	// Validate funding amount matches escrow amount
	if req.Amount.Currency != escrow.Currency {
		return fmt.Errorf("currency mismatch: expected %s, got %s", escrow.Currency, req.Amount.Currency)
	}

	if req.Amount.Value.Cmp(escrow.Amount.Value) != 0 {
		return fmt.Errorf("amount mismatch: expected %s, got %s", 
			escrow.Amount.Value.String(), req.Amount.Value.String())
	}

	// Update escrow status and metadata
	escrow.Status = "funded"
	escrow.UpdatedAt = time.Now()
	if escrow.Metadata == nil {
		escrow.Metadata = make(map[string]interface{})
	}
	escrow.Metadata["funded_at"] = time.Now()
	escrow.Metadata["funding_source"] = req.PaymentMethod

	// Update in repository
	if err := s.repo.UpdateEscrow(ctx, escrow); err != nil {
		return fmt.Errorf("failed to update escrow: %w", err)
	}

	return nil
}

// ConfirmDelivery confirms delivery of goods/services
func (s *Service) ConfirmDelivery(ctx context.Context, req *ConfirmDeliveryRequest) error {
	escrow, err := s.repo.GetEscrow(ctx, req.EscrowID)
	if err != nil {
		return err
	}

	escrow.Status = "delivered"
	escrow.UpdatedAt = time.Now()

	if err := s.repo.UpdateEscrow(ctx, escrow); err != nil {
		return err
	}

	return nil
}


// MockRepository implements Repository for testing
type MockRepository struct {
	escrows map[string]*Escrow
}

func NewMockRepository() *MockRepository {
	return &MockRepository{
		escrows: make(map[string]*Escrow),
	}
}

func (m *MockRepository) CreateEscrow(ctx context.Context, escrow *Escrow) error {
	if m.escrows == nil {
		m.escrows = make(map[string]*Escrow)
	}
	m.escrows[escrow.ID] = escrow
	return nil
}

func (m *MockRepository) GetEscrow(ctx context.Context, id string) (*Escrow, error) {
	if m.escrows == nil {
		m.escrows = make(map[string]*Escrow)
	}
	
	if escrow, exists := m.escrows[id]; exists {
		return escrow, nil
	}
	
	// Return default escrow for testing
	return &Escrow{
		ID:        id,
		BuyerID:   "buyer_123",
		SellerID:  "seller_456",
		Amount:    FromMinorUnits("USD", 10000),
		Currency:  "USD",
		Status:    "pending",
		Terms:     "Test terms",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}, nil
}

func (m *MockRepository) ListEscrows(ctx context.Context, filters EscrowFilters) ([]*Escrow, error) {
	return []*Escrow{}, nil
}

func (m *MockRepository) UpdateEscrow(ctx context.Context, escrow *Escrow) error {
	if m.escrows == nil {
		m.escrows = make(map[string]*Escrow)
	}
	m.escrows[escrow.ID] = escrow
	return nil
}

func (m *MockRepository) DeleteEscrow(ctx context.Context, id string) error {
	return nil
}

// Helper functions
func generateID() string {
	return fmt.Sprintf("escrow_%d", time.Now().UnixNano())
}
