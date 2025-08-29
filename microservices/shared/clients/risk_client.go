package clients

import (
	"context"
	"fmt"
	"time"

	"github.com/project-x/microservices/shared/circuitbreaker"
	"github.com/project-x/microservices/shared/httpclient"
	"github.com/project-x/microservices/shared/servicediscovery"
)

// RiskClient handles communication with the risk service
type RiskClient struct {
	client         *httpclient.ServiceClient
	circuitBreaker *circuitbreaker.CircuitBreaker
	loadBalancer   *servicediscovery.LoadBalancer
}

// NewRiskClient creates a new risk service client
func NewRiskClient(loadBalancer *servicediscovery.LoadBalancer) *RiskClient {
	cb := circuitbreaker.NewCircuitBreaker(circuitbreaker.Config{
		FailureThreshold: 5,
		SuccessThreshold: 3,
		Timeout:          60 * time.Second,
	})

	return &RiskClient{
		circuitBreaker: cb,
		loadBalancer:   loadBalancer,
	}
}

// CreateRiskProfileRequest represents a risk profile creation request
type CreateRiskProfileRequest struct {
	EntityID   string      `json:"entity_id"`
	EntityType string      `json:"entity_type"`
	Metadata   interface{} `json:"metadata,omitempty"`
}

// AssessRiskRequest represents a risk assessment request
type AssessRiskRequest struct {
	EntityID      string      `json:"entity_id"`
	Amount        Money       `json:"amount"`
	PaymentMethod string      `json:"payment_method"`
	Context       string      `json:"context"`
	Metadata      interface{} `json:"metadata,omitempty"`
}

// RiskProfile represents a risk profile
type RiskProfile struct {
	ID         string      `json:"id"`
	EntityID   string      `json:"entity_id"`
	EntityType string      `json:"entity_type"`
	RiskLevel  string      `json:"risk_level"`
	RiskScore  float64     `json:"risk_score"`
	Metadata   interface{} `json:"metadata,omitempty"`
	CreatedAt  time.Time   `json:"created_at"`
	UpdatedAt  time.Time   `json:"updated_at"`
}

// RiskAssessment represents a risk assessment
type RiskAssessment struct {
	ID           string      `json:"id"`
	EntityID     string      `json:"entity_id"`
	RiskScore    float64     `json:"risk_score"`
	RiskLevel    string      `json:"risk_level"`
	Confidence   float64     `json:"confidence"`
	RulesApplied []string    `json:"rules_applied"`
	Metadata     interface{} `json:"metadata,omitempty"`
	CreatedAt    time.Time   `json:"created_at"`
}

// CreateRiskProfile creates a new risk profile
func (c *RiskClient) CreateRiskProfile(ctx context.Context, req *CreateRiskProfileRequest) (*RiskProfile, error) {
	var profile *RiskProfile
	
	err := c.circuitBreaker.Execute(ctx, func() error {
		endpoint, err := c.loadBalancer.GetEndpoint("risk-service")
		if err != nil {
			return fmt.Errorf("failed to get risk service endpoint: %w", err)
		}

		if c.client == nil || c.client.BaseURL() != endpoint {
			c.client = httpclient.NewServiceClient(endpoint, 30*time.Second)
		}

		resp, err := c.client.Post(ctx, "/v1/risk-profiles", req, nil)
		if err != nil {
			return err
		}

		profile = &RiskProfile{}
		return resp.UnmarshalResponse(profile)
	})

	return profile, err
}

// AssessRisk performs a risk assessment
func (c *RiskClient) AssessRisk(ctx context.Context, req *AssessRiskRequest) (*RiskAssessment, error) {
	var assessment *RiskAssessment
	
	err := c.circuitBreaker.Execute(ctx, func() error {
		endpoint, err := c.loadBalancer.GetEndpoint("risk-service")
		if err != nil {
			return fmt.Errorf("failed to get risk service endpoint: %w", err)
		}

		if c.client == nil || c.client.BaseURL() != endpoint {
			c.client = httpclient.NewServiceClient(endpoint, 30*time.Second)
		}

		resp, err := c.client.Post(ctx, "/v1/risk-assessments", req, nil)
		if err != nil {
			return err
		}

		assessment = &RiskAssessment{}
		return resp.UnmarshalResponse(assessment)
	})

	return assessment, err
}

// GetRiskProfile retrieves a risk profile by entity ID
func (c *RiskClient) GetRiskProfile(ctx context.Context, entityID string) (*RiskProfile, error) {
	var profile *RiskProfile
	
	err := c.circuitBreaker.Execute(ctx, func() error {
		endpoint, err := c.loadBalancer.GetEndpoint("risk-service")
		if err != nil {
			return fmt.Errorf("failed to get risk service endpoint: %w", err)
		}

		if c.client == nil || c.client.BaseURL() != endpoint {
			c.client = httpclient.NewServiceClient(endpoint, 30*time.Second)
		}

		resp, err := c.client.Get(ctx, fmt.Sprintf("/v1/risk-profiles/%s", entityID), nil)
		if err != nil {
			return err
		}

		profile = &RiskProfile{}
		return resp.UnmarshalResponse(profile)
	})

	return profile, err
}
