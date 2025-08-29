package main

import (
	"context"
	"fmt"
	"math/big"
	"time"
)

// Money represents a monetary amount with currency
type Money struct {
	Value    *big.Float `json:"value"`
	Currency string     `json:"currency"`
}

// FromMinorUnits creates Money from minor units (e.g., cents)
func FromMinorUnits(currency string, minorUnits int64) *Money {
	value := big.NewFloat(float64(minorUnits) / 100.0)
	return &Money{
		Value:    value,
		Currency: currency,
	}
}

// RiskProfile represents a risk assessment profile for an entity
type RiskProfile struct {
	ID             string                 `json:"id"`
	EntityID       string                 `json:"entity_id"`
	EntityType     string                 `json:"entity_type"` // user, merchant, transaction, etc.
	RiskScore      float64                `json:"risk_score"`  // 0.0 to 1.0
	RiskLevel      string                 `json:"risk_level"`  // low, medium, high, critical
	Factors        map[string]interface{} `json:"factors"`
	RulesApplied   []string               `json:"rules_applied"`
	LastAssessment time.Time              `json:"last_assessment"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt      time.Time              `json:"created_at"`
	UpdatedAt      time.Time              `json:"updated_at"`
}

// RiskRule represents a risk assessment rule
type RiskRule struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	RuleType    string                 `json:"rule_type"` // threshold, pattern, ml_model, etc.
	Conditions  map[string]interface{} `json:"conditions"`
	Actions     []string               `json:"actions"`
	Priority    int                    `json:"priority"`
	IsActive    bool                   `json:"is_active"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

// RiskAssessment represents a risk assessment result
type RiskAssessment struct {
	ID           string                 `json:"id"`
	EntityID     string                 `json:"entity_id"`
	EntityType   string                 `json:"entity_type"`
	RiskScore    float64                `json:"risk_score"`
	RiskLevel    string                 `json:"risk_level"`
	Decision     string                 `json:"decision"` // allow, review, block
	Factors      map[string]interface{} `json:"factors"`
	RulesApplied []string               `json:"rules_applied"`
	Confidence   float64                `json:"confidence"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt    time.Time              `json:"created_at"`
}

// CreateRiskProfileRequest represents a request to create a risk profile
type CreateRiskProfileRequest struct {
	EntityID   string                 `json:"entity_id"`
	EntityType string                 `json:"entity_type"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}

// UpdateRiskProfileRequest represents a request to update a risk profile
type UpdateRiskProfileRequest struct {
	EntityID   string                 `json:"entity_id"`
	RiskScore  *float64               `json:"risk_score,omitempty"`
	RiskLevel  *string                `json:"risk_level,omitempty"`
	Factors    map[string]interface{} `json:"factors,omitempty"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}

// AssessRiskRequest represents a request for risk assessment
type AssessRiskRequest struct {
	EntityID     string                 `json:"entity_id"`
	EntityType   string                 `json:"entity_type"`
	Context      map[string]interface{} `json:"context"`
	Amount       *Money                 `json:"amount,omitempty"`
	PaymentMethod string                `json:"payment_method,omitempty"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

// CreateRiskRuleRequest represents a request to create a risk rule
type CreateRiskRuleRequest struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	RuleType    string                 `json:"rule_type"`
	Conditions  map[string]interface{} `json:"conditions"`
	Actions     []string               `json:"actions"`
	Priority    int                    `json:"priority"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// RiskFilters represents filters for querying risk data
type RiskFilters struct {
	EntityType string    `json:"entity_type,omitempty"`
	RiskLevel  string    `json:"risk_level,omitempty"`
	FromDate   time.Time `json:"from_date,omitempty"`
	ToDate     time.Time `json:"to_date,omitempty"`
	Limit      int       `json:"limit,omitempty"`
	Offset     int       `json:"offset,omitempty"`
}

// Repository interface for risk data operations
type Repository interface {
	// Risk Profile operations
	CreateRiskProfile(ctx context.Context, profile *RiskProfile) error
	GetRiskProfile(ctx context.Context, entityID string) (*RiskProfile, error)
	UpdateRiskProfile(ctx context.Context, profile *RiskProfile) error
	ListRiskProfiles(ctx context.Context, filters RiskFilters) ([]*RiskProfile, error)
	DeleteRiskProfile(ctx context.Context, entityID string) error

	// Risk Rule operations
	CreateRiskRule(ctx context.Context, rule *RiskRule) error
	GetRiskRule(ctx context.Context, ruleID string) (*RiskRule, error)
	UpdateRiskRule(ctx context.Context, rule *RiskRule) error
	ListRiskRules(ctx context.Context, filters RiskFilters) ([]*RiskRule, error)
	DeleteRiskRule(ctx context.Context, ruleID string) error

	// Risk Assessment operations
	CreateRiskAssessment(ctx context.Context, assessment *RiskAssessment) error
	GetRiskAssessment(ctx context.Context, assessmentID string) (*RiskAssessment, error)
	ListRiskAssessments(ctx context.Context, filters RiskFilters) ([]*RiskAssessment, error)
}

// Service represents the risk service
type Service struct {
	repo      Repository
	validator *RiskValidator
}

// NewService creates a new risk service
func NewService(repo Repository, validator *RiskValidator) *Service {
	if validator == nil {
		validator = NewRiskValidator()
	}
	return &Service{
		repo:      repo,
		validator: validator,
	}
}

// CreateRiskProfile creates a new risk profile
func (s *Service) CreateRiskProfile(ctx context.Context, req *CreateRiskProfileRequest) (*RiskProfile, error) {
	profile := &RiskProfile{
		ID:             generateID(),
		EntityID:       req.EntityID,
		EntityType:     req.EntityType,
		RiskScore:      0.0,
		RiskLevel:      "unknown",
		Factors:        make(map[string]interface{}),
		RulesApplied:   []string{},
		LastAssessment: time.Now(),
		Metadata:       req.Metadata,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	if err := s.repo.CreateRiskProfile(ctx, profile); err != nil {
		return nil, err
	}

	return profile, nil
}

// GetRiskProfile retrieves a risk profile
func (s *Service) GetRiskProfile(ctx context.Context, entityID string) (*RiskProfile, error) {
	return s.repo.GetRiskProfile(ctx, entityID)
}

// UpdateRiskProfile updates a risk profile
func (s *Service) UpdateRiskProfile(ctx context.Context, req *UpdateRiskProfileRequest) (*RiskProfile, error) {
	profile, err := s.repo.GetRiskProfile(ctx, req.EntityID)
	if err != nil {
		return nil, err
	}

	if req.RiskScore != nil {
		profile.RiskScore = *req.RiskScore
	}
	if req.RiskLevel != nil {
		profile.RiskLevel = *req.RiskLevel
	}
	if req.Factors != nil {
		profile.Factors = req.Factors
	}
	if req.Metadata != nil {
		profile.Metadata = req.Metadata
	}

	profile.UpdatedAt = time.Now()

	if err := s.repo.UpdateRiskProfile(ctx, profile); err != nil {
		return nil, err
	}

	return profile, nil
}

// AssessRisk performs risk assessment for an entity
func (s *Service) AssessRisk(ctx context.Context, req *AssessRiskRequest) (*RiskAssessment, error) {
	// This would be implemented with actual risk assessment logic
	// For now, return a basic assessment
	assessment := &RiskAssessment{
		ID:           generateID(),
		EntityID:     req.EntityID,
		EntityType:   req.EntityType,
		RiskScore:    0.2,
		RiskLevel:    "low",
		Decision:     "allow",
		Factors:      req.Context,
		RulesApplied: []string{},
		Confidence:   0.8,
		Metadata:     req.Metadata,
		CreatedAt:    time.Now(),
	}

	if err := s.repo.CreateRiskAssessment(ctx, assessment); err != nil {
		return nil, err
	}

	return assessment, nil
}

// MockRepository implements Repository for testing
type MockRepository struct {
	profiles    map[string]*RiskProfile
	rules       map[string]*RiskRule
	assessments map[string]*RiskAssessment
}

func NewMockRepository() *MockRepository {
	return &MockRepository{
		profiles:    make(map[string]*RiskProfile),
		rules:       make(map[string]*RiskRule),
		assessments: make(map[string]*RiskAssessment),
	}
}

func (m *MockRepository) CreateRiskProfile(ctx context.Context, profile *RiskProfile) error {
	m.profiles[profile.EntityID] = profile
	return nil
}

func (m *MockRepository) GetRiskProfile(ctx context.Context, entityID string) (*RiskProfile, error) {
	if profile, exists := m.profiles[entityID]; exists {
		return profile, nil
	}
	return nil, nil
}

func (m *MockRepository) UpdateRiskProfile(ctx context.Context, profile *RiskProfile) error {
	m.profiles[profile.EntityID] = profile
	return nil
}

func (m *MockRepository) ListRiskProfiles(ctx context.Context, filters RiskFilters) ([]*RiskProfile, error) {
	var profiles []*RiskProfile
	for _, profile := range m.profiles {
		profiles = append(profiles, profile)
	}
	return profiles, nil
}

func (m *MockRepository) DeleteRiskProfile(ctx context.Context, entityID string) error {
	delete(m.profiles, entityID)
	return nil
}

func (m *MockRepository) CreateRiskRule(ctx context.Context, rule *RiskRule) error {
	m.rules[rule.ID] = rule
	return nil
}

func (m *MockRepository) GetRiskRule(ctx context.Context, ruleID string) (*RiskRule, error) {
	if rule, exists := m.rules[ruleID]; exists {
		return rule, nil
	}
	return nil, nil
}

func (m *MockRepository) UpdateRiskRule(ctx context.Context, rule *RiskRule) error {
	m.rules[rule.ID] = rule
	return nil
}

func (m *MockRepository) ListRiskRules(ctx context.Context, filters RiskFilters) ([]*RiskRule, error) {
	var rules []*RiskRule
	for _, rule := range m.rules {
		rules = append(rules, rule)
	}
	return rules, nil
}

func (m *MockRepository) DeleteRiskRule(ctx context.Context, ruleID string) error {
	delete(m.rules, ruleID)
	return nil
}

func (m *MockRepository) CreateRiskAssessment(ctx context.Context, assessment *RiskAssessment) error {
	m.assessments[assessment.ID] = assessment
	return nil
}

func (m *MockRepository) GetRiskAssessment(ctx context.Context, assessmentID string) (*RiskAssessment, error) {
	if assessment, exists := m.assessments[assessmentID]; exists {
		return assessment, nil
	}
	return nil, nil
}

func (m *MockRepository) ListRiskAssessments(ctx context.Context, filters RiskFilters) ([]*RiskAssessment, error) {
	var assessments []*RiskAssessment
	for _, assessment := range m.assessments {
		assessments = append(assessments, assessment)
	}
	return assessments, nil
}

// Helper functions
func generateID() string {
	return fmt.Sprintf("risk_%d", time.Now().UnixNano())
}
