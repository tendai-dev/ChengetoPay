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

// FeeSchedule represents a configurable fee schedule
type FeeSchedule struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Type        string    `json:"type"` // flat, tiered, volume, fx_margin
	Rate        *big.Float `json:"rate"`
	FixedAmount Money     `json:"fixed_amount,omitempty"`
	Tiers       []FeeTier `json:"tiers,omitempty"`
	Currency    string    `json:"currency"`
	EffectiveFrom time.Time `json:"effective_from"`
	EffectiveTo   *time.Time `json:"effective_to,omitempty"`
	Status      string    `json:"status"` // active, inactive
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// FeeTier represents a tier in a tiered fee structure
type FeeTier struct {
	MinAmount   Money     `json:"min_amount"`
	MaxAmount   *Money    `json:"max_amount,omitempty"`
	Rate        *big.Float `json:"rate"`
	FixedAmount Money     `json:"fixed_amount,omitempty"`
}

// TaxRule represents a tax rule for calculations
type TaxRule struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Rate        *big.Float `json:"rate"`
	Type        string    `json:"type"` // inclusive, exclusive
	Country     string    `json:"country"`
	Region      string    `json:"region,omitempty"`
	EffectiveFrom time.Time `json:"effective_from"`
	EffectiveTo   *time.Time `json:"effective_to,omitempty"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
}

// FeeCalculation represents a fee calculation result
type FeeCalculation struct {
	TransactionID string    `json:"transaction_id"`
	Amount        Money     `json:"amount"`
	Fees          []FeeItem `json:"fees"`
	Taxes         []TaxItem `json:"taxes"`
	TotalFees     Money     `json:"total_fees"`
	TotalTaxes    Money     `json:"total_taxes"`
	NetAmount     Money     `json:"net_amount"`
	CreatedAt     time.Time `json:"created_at"`
}

// FeeItem represents an individual fee item
type FeeItem struct {
	ScheduleID string    `json:"schedule_id"`
	Name       string    `json:"name"`
	Type       string    `json:"type"`
	Amount     Money     `json:"amount"`
	Rate       *big.Float `json:"rate,omitempty"`
}

// TaxItem represents an individual tax item
type TaxItem struct {
	RuleID   string    `json:"rule_id"`
	Name     string    `json:"name"`
	Rate     *big.Float `json:"rate"`
	Amount   Money     `json:"amount"`
	Type     string    `json:"type"`
}

// CreateFeeScheduleRequest represents a request to create a fee schedule
type CreateFeeScheduleRequest struct {
	Name         string    `json:"name"`
	Type         string    `json:"type"`
	Rate         *big.Float `json:"rate,omitempty"`
	FixedAmount  Money     `json:"fixed_amount,omitempty"`
	Tiers        []FeeTier `json:"tiers,omitempty"`
	Currency     string    `json:"currency"`
	EffectiveFrom time.Time `json:"effective_from"`
	EffectiveTo   *time.Time `json:"effective_to,omitempty"`
}

// CalculateFeesRequest represents a request to calculate fees
type CalculateFeesRequest struct {
	TransactionID string    `json:"transaction_id"`
	Amount        Money     `json:"amount"`
	Currency      string    `json:"currency"`
	Country       string    `json:"country"`
	Region        string    `json:"region,omitempty"`
	ScheduleIDs   []string  `json:"schedule_ids,omitempty"`
	Date          time.Time `json:"date"`
}

// CalculateTaxRequest represents a request to calculate taxes
type CalculateTaxRequest struct {
	Amount   Money     `json:"amount"`
	Country  string    `json:"country"`
	Region   string    `json:"region,omitempty"`
	Date     time.Time `json:"date"`
	Type     string    `json:"type"` // inclusive, exclusive
}

// FeeFilters represents filters for listing fee schedules
type FeeFilters struct {
	Type         string `json:"type"`
	Currency     string `json:"currency"`
	Status       string `json:"status"`
	EffectiveDate *time.Time `json:"effective_date"`
	Limit        int    `json:"limit"`
	Offset       int    `json:"offset"`
}

// ScheduleFilters represents filters for listing schedules
type ScheduleFilters struct {
	Type         string `json:"type"`
	Status       string `json:"status"`
	Limit        int    `json:"limit"`
	Offset       int    `json:"offset"`
}

// Repository interface for fees data access
type Repository interface {
	CreateFeeSchedule(ctx context.Context, schedule *FeeSchedule) error
	GetFeeSchedule(ctx context.Context, id string) (*FeeSchedule, error)
	ListFeeSchedules(ctx context.Context, filters FeeFilters) ([]*FeeSchedule, error)
	GetSchedules(ctx context.Context, filters ScheduleFilters) ([]*FeeSchedule, error)
	GetTaxRules(ctx context.Context, country, region string, date time.Time) ([]*TaxRule, error)
	SaveFeeCalculation(ctx context.Context, calculation *FeeCalculation) error
}

// Service represents the fees business logic
type Service struct {
	repo   Repository
	logger interface{}
}

// NewService creates a new fees service
func NewService(repo Repository, logger interface{}) *Service {
	return &Service{
		repo:   repo,
		logger: logger,
	}
}

// CreateFeeSchedule creates a new fee schedule
func (s *Service) CreateFeeSchedule(ctx context.Context, req *CreateFeeScheduleRequest) (*FeeSchedule, error) {
	schedule := &FeeSchedule{
		ID:           generateID(),
		Name:         req.Name,
		Type:         req.Type,
		Rate:         req.Rate,
		FixedAmount:  req.FixedAmount,
		Tiers:        req.Tiers,
		Currency:     req.Currency,
		EffectiveFrom: req.EffectiveFrom,
		EffectiveTo:   req.EffectiveTo,
		Status:       "active",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if err := s.repo.CreateFeeSchedule(ctx, schedule); err != nil {
		return nil, err
	}

	return schedule, nil
}

// GetFeeSchedule retrieves a fee schedule by ID
func (s *Service) GetFeeSchedule(ctx context.Context, id string) (*FeeSchedule, error) {
	return s.repo.GetFeeSchedule(ctx, id)
}

// ListFeeSchedules lists fee schedules with filters
func (s *Service) ListFeeSchedules(ctx context.Context, filters FeeFilters) ([]*FeeSchedule, error) {
	return s.repo.ListFeeSchedules(ctx, filters)
}

// GetSchedules gets fee schedules
func (s *Service) GetSchedules(ctx context.Context, filters ScheduleFilters) ([]*FeeSchedule, error) {
	return s.repo.GetSchedules(ctx, filters)
}

// CalculateFees calculates fees for a transaction
func (s *Service) CalculateFees(ctx context.Context, req *CalculateFeesRequest) (*FeeCalculation, error) {
	// Get applicable fee schedules
	schedules, err := s.repo.ListFeeSchedules(ctx, FeeFilters{
		Currency: req.Currency,
		Status:   "active",
		EffectiveDate: &req.Date,
	})
	if err != nil {
		return nil, err
	}

	// Calculate fees
	var feeItems []FeeItem
	var totalFees Money = FromMinorUnits(req.Currency, 0)

	for _, schedule := range schedules {
		feeAmount := s.calculateFeeAmount(schedule, req.Amount)
		feeItems = append(feeItems, FeeItem{
			ScheduleID: schedule.ID,
			Name:       schedule.Name,
			Type:       schedule.Type,
			Amount:     feeAmount,
			Rate:       schedule.Rate,
		})
		totalFees.Value.Add(totalFees.Value, feeAmount.Value)
	}

	// Calculate taxes
	taxRules, err := s.repo.GetTaxRules(ctx, req.Country, req.Region, req.Date)
	if err != nil {
		return nil, err
	}

	var taxItems []TaxItem
	var totalTaxes Money = FromMinorUnits(req.Currency, 0)

	for _, rule := range taxRules {
		taxAmount := s.calculateTaxAmount(rule, req.Amount)
		taxItems = append(taxItems, TaxItem{
			RuleID: rule.ID,
			Name:   rule.Name,
			Rate:   rule.Rate,
			Amount: taxAmount,
			Type:   rule.Type,
		})
		totalTaxes.Value.Add(totalTaxes.Value, taxAmount.Value)
	}

	// Calculate net amount
	netAmount := Money{
		Value:    new(big.Float).Sub(req.Amount.Value, totalFees.Value),
		Currency: req.Currency,
	}
	netAmount.Value.Sub(netAmount.Value, totalTaxes.Value)

	calculation := &FeeCalculation{
		TransactionID: req.TransactionID,
		Amount:        req.Amount,
		Fees:          feeItems,
		Taxes:         taxItems,
		TotalFees:     totalFees,
		TotalTaxes:    totalTaxes,
		NetAmount:     netAmount,
		CreatedAt:     time.Now(),
	}

	// Save calculation
	if err := s.repo.SaveFeeCalculation(ctx, calculation); err != nil {
		return nil, err
	}

	return calculation, nil
}

// CalculateTax calculates taxes for an amount
func (s *Service) CalculateTax(ctx context.Context, req *CalculateTaxRequest) (*TaxItem, error) {
	taxRules, err := s.repo.GetTaxRules(ctx, req.Country, req.Region, req.Date)
	if err != nil {
		return nil, err
	}

	if len(taxRules) == 0 {
		return &TaxItem{
			RuleID: "no_tax",
			Name:   "No Tax",
			Rate:   big.NewFloat(0),
			Amount: FromMinorUnits(req.Amount.Currency, 0),
			Type:   req.Type,
		}, nil
	}

	// Use first applicable tax rule
	rule := taxRules[0]
	taxAmount := s.calculateTaxAmount(rule, req.Amount)

	return &TaxItem{
		RuleID: rule.ID,
		Name:   rule.Name,
		Rate:   rule.Rate,
		Amount: taxAmount,
		Type:   rule.Type,
	}, nil
}

// calculateFeeAmount calculates the fee amount for a given schedule and amount
func (s *Service) calculateFeeAmount(schedule *FeeSchedule, amount Money) Money {
	switch schedule.Type {
	case "flat":
		return schedule.FixedAmount
	case "percentage":
		if schedule.Rate != nil {
			feeValue := new(big.Float).Mul(amount.Value, schedule.Rate)
			return Money{
				Value:    feeValue,
				Currency: amount.Currency,
			}
		}
	case "tiered":
		return s.calculateTieredFee(schedule, amount)
	}
	return FromMinorUnits(amount.Currency, 0)
}

// calculateTieredFee calculates tiered fee
func (s *Service) calculateTieredFee(schedule *FeeSchedule, amount Money) Money {
	for _, tier := range schedule.Tiers {
		if amount.Value.Cmp(tier.MinAmount.Value) >= 0 {
			if tier.MaxAmount == nil || amount.Value.Cmp(tier.MaxAmount.Value) <= 0 {
				if tier.Rate != nil {
					feeValue := new(big.Float).Mul(amount.Value, tier.Rate)
					return Money{
						Value:    feeValue,
						Currency: amount.Currency,
					}
				}
				return tier.FixedAmount
			}
		}
	}
	return FromMinorUnits(amount.Currency, 0)
}

// calculateTaxAmount calculates tax amount
func (s *Service) calculateTaxAmount(rule *TaxRule, amount Money) Money {
	if rule.Rate != nil {
		taxValue := new(big.Float).Mul(amount.Value, rule.Rate)
		return Money{
			Value:    taxValue,
			Currency: amount.Currency,
		}
	}
	return FromMinorUnits(amount.Currency, 0)
}

// MockRepository implements Repository for testing
type MockRepository struct{}

func (m *MockRepository) CreateFeeSchedule(ctx context.Context, schedule *FeeSchedule) error {
	return nil
}

func (m *MockRepository) GetFeeSchedule(ctx context.Context, id string) (*FeeSchedule, error) {
	return &FeeSchedule{
		ID:          id,
		Name:        "Standard Fee",
		Type:        "percentage",
		Rate:        big.NewFloat(0.029), // 2.9%
		Currency:    "USD",
		EffectiveFrom: time.Now(),
		Status:      "active",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}, nil
}

func (m *MockRepository) ListFeeSchedules(ctx context.Context, filters FeeFilters) ([]*FeeSchedule, error) {
	return []*FeeSchedule{}, nil
}

func (m *MockRepository) GetSchedules(ctx context.Context, filters ScheduleFilters) ([]*FeeSchedule, error) {
	return []*FeeSchedule{}, nil
}

func (m *MockRepository) GetTaxRules(ctx context.Context, country, region string, date time.Time) ([]*TaxRule, error) {
	return []*TaxRule{}, nil
}

func (m *MockRepository) SaveFeeCalculation(ctx context.Context, calculation *FeeCalculation) error {
	return nil
}

// Helper functions
func generateID() string {
	return fmt.Sprintf("fee_%d", time.Now().UnixNano())
}
