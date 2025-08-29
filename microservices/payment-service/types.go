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

// Payment represents a payment transaction
type Payment struct {
	ID          string                 `json:"id"`
	AccountID   string                 `json:"account_id"`
	Provider    string                 `json:"provider"`
	Method      string                 `json:"method"`
	Amount      Money                  `json:"amount"`
	Currency    string                 `json:"currency"`
	Status      string                 `json:"status"`
	ExternalRef string                 `json:"external_ref,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

// CreatePaymentRequest represents a request to create a payment
type CreatePaymentRequest struct {
	AccountID     string                 `json:"account_id"`
	Provider      string                 `json:"provider"`
	PaymentMethod string                 `json:"payment_method"`
	Amount        Money                  `json:"amount"`
	Description   string                 `json:"description"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
}

// ProcessPaymentRequest represents a request to process a payment
type ProcessPaymentRequest struct {
	PaymentID string `json:"payment_id"`
	Provider  string `json:"provider"`
	Method    string `json:"method"`
}

// PaymentFilters represents filters for listing payments
type PaymentFilters struct {
	AccountID string `json:"account_id"`
	Provider  string `json:"provider"`
	Status    string `json:"status"`
	Limit     int    `json:"limit"`
	Offset    int    `json:"offset"`
}

// Provider represents a payment provider
type Provider struct {
	Name      string   `json:"name"`
	Methods   []string `json:"methods"`
	Currencies []string `json:"currencies"`
	Enabled   bool     `json:"enabled"`
}

// Repository interface for payment data access
type Repository interface {
	CreatePayment(ctx context.Context, payment *Payment) error
	GetPayment(ctx context.Context, id string) (*Payment, error)
	ListPayments(ctx context.Context, filters PaymentFilters) ([]*Payment, error)
	UpdatePayment(ctx context.Context, payment *Payment) error
	DeletePayment(ctx context.Context, id string) error
	GetProviders(ctx context.Context) ([]*Provider, error)
}

// Service represents the payment business logic
type Service struct {
	repo   Repository
	logger interface{}
}

// NewService creates a new payment service
func NewService(repo Repository, logger interface{}) *Service {
	return &Service{
		repo:   repo,
		logger: logger,
	}
}

// CreatePayment creates a new payment
func (s *Service) CreatePayment(ctx context.Context, req *CreatePaymentRequest) (*Payment, error) {
	// Validate request
	validator := NewPaymentValidator()
	if err := validator.ValidateCreatePaymentRequest(req); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Create payment with validation and business logic
	payment := &Payment{
		ID:          generateID(),
		AccountID:   req.AccountID,
		Amount:      req.Amount,
		Currency:    req.Amount.Currency,
		Method:      req.PaymentMethod,
		Provider:    req.Provider,
		Status:      "pending",
		Metadata:    req.Metadata,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// Add creation metadata
	if payment.Metadata == nil {
		payment.Metadata = make(map[string]interface{})
	}
	payment.Metadata["created_by"] = "payment-service"
	payment.Metadata["version"] = "1.0"

	// Calculate initial risk score
	businessValidator := NewBusinessRuleValidator()
	riskScore, err := businessValidator.CalculateRiskScore(payment)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate risk score: %w", err)
	}
	payment.Metadata["risk_score"] = riskScore

	// Validate business rules
	if err := businessValidator.ValidatePaymentLimits(ctx, payment); err != nil {
		return nil, fmt.Errorf("business rules validation failed: %w", err)
	}

	if err := businessValidator.ValidateFraudRules(ctx, payment); err != nil {
		return nil, fmt.Errorf("fraud validation failed: %w", err)
	}

	if err := businessValidator.ValidateComplianceRules(ctx, payment); err != nil {
		return nil, fmt.Errorf("compliance validation failed: %w", err)
	}

	// Create payment in repository
	if err := s.repo.CreatePayment(ctx, payment); err != nil {
		return nil, fmt.Errorf("failed to create payment: %w", err)
	}

	return payment, nil
}

// GetPayment retrieves a payment by ID
func (s *Service) GetPayment(ctx context.Context, id string) (*Payment, error) {
	return s.repo.GetPayment(ctx, id)
}

// ListPayments lists payments with filters
func (s *Service) ListPayments(ctx context.Context, filters PaymentFilters) ([]*Payment, error) {
	payments, err := s.repo.ListPayments(ctx, filters)
	if err != nil {
		return nil, err
	}

	return payments, nil
}


// GetProviders returns available payment providers
func (s *Service) GetProviders(ctx context.Context) ([]*Provider, error) {
	return s.repo.GetProviders(ctx)
}

// MockRepository implements Repository for testing
type MockRepository struct{}

func (m *MockRepository) CreatePayment(ctx context.Context, payment *Payment) error {
	return nil
}

func (m *MockRepository) GetPayment(ctx context.Context, id string) (*Payment, error) {
	return &Payment{
		ID:        id,
		AccountID: "acc_123",
		Provider:  "stripe",
		Method:    "card",
		Amount:    FromMinorUnits("USD", 5000), // $50
		Currency:  "USD",
		Status:    "completed",
		ExternalRef: "txn_" + id,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}, nil
}

func (m *MockRepository) ListPayments(ctx context.Context, filters PaymentFilters) ([]*Payment, error) {
	return []*Payment{}, nil
}

func (m *MockRepository) UpdatePayment(ctx context.Context, payment *Payment) error {
	return nil
}

func (m *MockRepository) DeletePayment(ctx context.Context, id string) error {
	return nil
}

func (m *MockRepository) GetProviders(ctx context.Context) ([]*Provider, error) {
	return []*Provider{
		{
			Name:       "stripe",
			Methods:    []string{"card", "bank_transfer"},
			Currencies: []string{"USD"},
			Enabled:    true,
		},
		{
			Name:       "paypal",
			Methods:    []string{"card", "paypal"},
			Currencies: []string{"USD"},
			Enabled:    true,
		},
		{
			Name:       "mpesa",
			Methods:    []string{"mobile_money"},
			Currencies: []string{"USD"},
			Enabled:    true,
		},
	}, nil
}

// Helper functions
func generateID() string {
	return fmt.Sprintf("payment_%d", time.Now().UnixNano())
}
