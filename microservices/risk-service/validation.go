package main

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"
)

// RiskValidator handles validation for risk service operations
type RiskValidator struct {
	businessRuleValidator *BusinessRuleValidator
}

// NewRiskValidator creates a new risk validator
func NewRiskValidator() *RiskValidator {
	return &RiskValidator{
		businessRuleValidator: NewBusinessRuleValidator(),
	}
}

// ValidateCreateRiskProfileRequest validates risk profile creation request
func (v *RiskValidator) ValidateCreateRiskProfileRequest(req *CreateRiskProfileRequest) error {
	if req == nil {
		return errors.New("request cannot be nil")
	}

	if err := v.validateEntityID(req.EntityID); err != nil {
		return fmt.Errorf("entity_id validation failed: %w", err)
	}

	if err := v.validateEntityType(req.EntityType); err != nil {
		return fmt.Errorf("entity_type validation failed: %w", err)
	}

	if err := v.validateMetadata(req.Metadata); err != nil {
		return fmt.Errorf("metadata validation failed: %w", err)
	}

	return nil
}

// ValidateUpdateRiskProfileRequest validates risk profile update request
func (v *RiskValidator) ValidateUpdateRiskProfileRequest(req *UpdateRiskProfileRequest) error {
	if req == nil {
		return errors.New("request cannot be nil")
	}

	if err := v.validateEntityID(req.EntityID); err != nil {
		return fmt.Errorf("entity_id validation failed: %w", err)
	}

	if req.RiskScore != nil {
		if err := v.validateRiskScore(*req.RiskScore); err != nil {
			return fmt.Errorf("risk_score validation failed: %w", err)
		}
	}

	if req.RiskLevel != nil {
		if err := v.validateRiskLevel(*req.RiskLevel); err != nil {
			return fmt.Errorf("risk_level validation failed: %w", err)
		}
	}

	if err := v.validateMetadata(req.Metadata); err != nil {
		return fmt.Errorf("metadata validation failed: %w", err)
	}

	return nil
}

// ValidateAssessRiskRequest validates risk assessment request
func (v *RiskValidator) ValidateAssessRiskRequest(req *AssessRiskRequest) error {
	if req == nil {
		return errors.New("request cannot be nil")
	}

	if err := v.validateEntityID(req.EntityID); err != nil {
		return fmt.Errorf("entity_id validation failed: %w", err)
	}

	if err := v.validateEntityType(req.EntityType); err != nil {
		return fmt.Errorf("entity_type validation failed: %w", err)
	}

	if req.Amount != nil {
		if err := v.validateAmount(req.Amount); err != nil {
			return fmt.Errorf("amount validation failed: %w", err)
		}
	}

	if req.PaymentMethod != "" {
		if err := v.validatePaymentMethod(req.PaymentMethod); err != nil {
			return fmt.Errorf("payment_method validation failed: %w", err)
		}
	}

	if err := v.validateContext(req.Context); err != nil {
		return fmt.Errorf("context validation failed: %w", err)
	}

	if err := v.validateMetadata(req.Metadata); err != nil {
		return fmt.Errorf("metadata validation failed: %w", err)
	}

	return nil
}

// ValidateCreateRiskRuleRequest validates risk rule creation request
func (v *RiskValidator) ValidateCreateRiskRuleRequest(req *CreateRiskRuleRequest) error {
	if req == nil {
		return errors.New("request cannot be nil")
	}

	if err := v.validateRuleName(req.Name); err != nil {
		return fmt.Errorf("name validation failed: %w", err)
	}

	if err := v.validateRuleDescription(req.Description); err != nil {
		return fmt.Errorf("description validation failed: %w", err)
	}

	if err := v.validateRuleType(req.RuleType); err != nil {
		return fmt.Errorf("rule_type validation failed: %w", err)
	}

	if err := v.validateRuleConditions(req.Conditions); err != nil {
		return fmt.Errorf("conditions validation failed: %w", err)
	}

	if err := v.validateRuleActions(req.Actions); err != nil {
		return fmt.Errorf("actions validation failed: %w", err)
	}

	if err := v.validateRulePriority(req.Priority); err != nil {
		return fmt.Errorf("priority validation failed: %w", err)
	}

	if err := v.validateMetadata(req.Metadata); err != nil {
		return fmt.Errorf("metadata validation failed: %w", err)
	}

	return nil
}

// validateEntityID validates entity ID format
func (v *RiskValidator) validateEntityID(entityID string) error {
	if entityID == "" {
		return errors.New("entity_id cannot be empty")
	}

	if len(entityID) < 3 || len(entityID) > 100 {
		return errors.New("entity_id must be between 3 and 100 characters")
	}

	// Check for valid characters (alphanumeric, hyphens, underscores)
	validPattern := regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
	if !validPattern.MatchString(entityID) {
		return errors.New("entity_id contains invalid characters")
	}

	return nil
}

// validateEntityType validates entity type
func (v *RiskValidator) validateEntityType(entityType string) error {
	if entityType == "" {
		return errors.New("entity_type cannot be empty")
	}

	validTypes := []string{
		"user", "merchant", "transaction", "payment", "account", 
		"device", "session", "ip_address", "card", "bank_account",
	}

	for _, validType := range validTypes {
		if entityType == validType {
			return nil
		}
	}

	return fmt.Errorf("invalid entity_type: %s", entityType)
}

// validateRiskScore validates risk score range
func (v *RiskValidator) validateRiskScore(score float64) error {
	if score < 0.0 || score > 1.0 {
		return errors.New("risk_score must be between 0.0 and 1.0")
	}
	return nil
}

// validateRiskLevel validates risk level
func (v *RiskValidator) validateRiskLevel(level string) error {
	validLevels := []string{"unknown", "low", "medium", "high", "critical"}
	
	for _, validLevel := range validLevels {
		if level == validLevel {
			return nil
		}
	}

	return fmt.Errorf("invalid risk_level: %s", level)
}

// validateAmount validates monetary amount
func (v *RiskValidator) validateAmount(amount *Money) error {
	if amount == nil {
		return errors.New("amount cannot be nil")
	}

	if amount.Value == nil {
		return errors.New("amount value cannot be nil")
	}

	if amount.Value.Sign() < 0 {
		return errors.New("amount cannot be negative")
	}

	if amount.Currency == "" {
		return errors.New("currency cannot be empty")
	}

	validCurrencies := []string{"USD", "EUR", "GBP", "CAD", "AUD", "JPY"}
	for _, validCurrency := range validCurrencies {
		if amount.Currency == validCurrency {
			return nil
		}
	}

	return fmt.Errorf("unsupported currency: %s", amount.Currency)
}

// validatePaymentMethod validates payment method
func (v *RiskValidator) validatePaymentMethod(method string) error {
	validMethods := []string{
		"credit_card", "debit_card", "bank_transfer", "wire_transfer",
		"ach", "paypal", "apple_pay", "google_pay", "crypto",
	}

	for _, validMethod := range validMethods {
		if method == validMethod {
			return nil
		}
	}

	return fmt.Errorf("invalid payment_method: %s", method)
}

// validateContext validates assessment context
func (v *RiskValidator) validateContext(context map[string]interface{}) error {
	if context == nil {
		return errors.New("context cannot be nil")
	}

	if len(context) == 0 {
		return errors.New("context cannot be empty")
	}

	// Check context size limit
	if len(context) > 100 {
		return errors.New("context cannot have more than 100 keys")
	}

	// Validate each key-value pair
	for key, value := range context {
		if len(key) > 100 {
			return fmt.Errorf("context key '%s' exceeds 100 characters", key)
		}

		if value == nil {
			continue
		}

		// Convert value to string for length check
		valueStr := fmt.Sprintf("%v", value)
		if len(valueStr) > 1000 {
			return fmt.Errorf("context value for key '%s' exceeds 1000 characters", key)
		}
	}

	return nil
}

// validateRuleName validates rule name
func (v *RiskValidator) validateRuleName(name string) error {
	if name == "" {
		return errors.New("name cannot be empty")
	}

	if len(name) < 3 || len(name) > 100 {
		return errors.New("name must be between 3 and 100 characters")
	}

	// Check for prohibited terms
	prohibitedTerms := []string{"test", "debug", "temp", "delete"}
	nameLower := strings.ToLower(name)
	for _, term := range prohibitedTerms {
		if strings.Contains(nameLower, term) {
			return fmt.Errorf("name contains prohibited term: %s", term)
		}
	}

	return nil
}

// validateRuleDescription validates rule description
func (v *RiskValidator) validateRuleDescription(description string) error {
	if description == "" {
		return errors.New("description cannot be empty")
	}

	if len(description) > 500 {
		return errors.New("description cannot exceed 500 characters")
	}

	return nil
}

// validateRuleType validates rule type
func (v *RiskValidator) validateRuleType(ruleType string) error {
	validTypes := []string{
		"threshold", "pattern", "ml_model", "blacklist", "whitelist",
		"velocity", "geolocation", "device_fingerprint", "behavioral",
	}

	for _, validType := range validTypes {
		if ruleType == validType {
			return nil
		}
	}

	return fmt.Errorf("invalid rule_type: %s", ruleType)
}

// validateRuleConditions validates rule conditions
func (v *RiskValidator) validateRuleConditions(conditions map[string]interface{}) error {
	if conditions == nil {
		return errors.New("conditions cannot be nil")
	}

	if len(conditions) == 0 {
		return errors.New("conditions cannot be empty")
	}

	// Check conditions size limit
	if len(conditions) > 50 {
		return errors.New("conditions cannot have more than 50 keys")
	}

	return nil
}

// validateRuleActions validates rule actions
func (v *RiskValidator) validateRuleActions(actions []string) error {
	if len(actions) == 0 {
		return errors.New("actions cannot be empty")
	}

	validActions := []string{
		"allow", "review", "block", "flag", "score_increase", "score_decrease",
		"require_verification", "limit_amount", "notify_admin",
	}

	for _, action := range actions {
		found := false
		for _, validAction := range validActions {
			if action == validAction {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("invalid action: %s", action)
		}
	}

	return nil
}

// validateRulePriority validates rule priority
func (v *RiskValidator) validateRulePriority(priority int) error {
	if priority < 1 || priority > 100 {
		return errors.New("priority must be between 1 and 100")
	}
	return nil
}

// validateMetadata validates metadata map
func (v *RiskValidator) validateMetadata(metadata map[string]interface{}) error {
	if metadata == nil {
		return nil // Metadata is optional
	}

	// Check metadata size limit
	if len(metadata) > 50 {
		return errors.New("metadata cannot have more than 50 keys")
	}

	// Validate each key-value pair
	for key, value := range metadata {
		if len(key) > 100 {
			return fmt.Errorf("metadata key '%s' exceeds 100 characters", key)
		}

		if value == nil {
			continue
		}

		// Convert value to string for length check
		valueStr := fmt.Sprintf("%v", value)
		if len(valueStr) > 1000 {
			return fmt.Errorf("metadata value for key '%s' exceeds 1000 characters", key)
		}
	}

	return nil
}

// BusinessRuleValidator handles business rule validation for risk service
type BusinessRuleValidator struct {
	validator *RiskValidator
}

// NewBusinessRuleValidator creates a new business rule validator
func NewBusinessRuleValidator() *BusinessRuleValidator {
	return &BusinessRuleValidator{
		validator: NewRiskValidator(),
	}
}

// ValidateRiskThresholds validates risk score thresholds
func (v *BusinessRuleValidator) ValidateRiskThresholds(profile *RiskProfile) error {
	if profile == nil {
		return errors.New("profile cannot be nil")
	}

	// Validate risk score consistency with risk level
	switch profile.RiskLevel {
	case "low":
		if profile.RiskScore > 0.3 {
			return errors.New("low risk level should have score <= 0.3")
		}
	case "medium":
		if profile.RiskScore <= 0.3 || profile.RiskScore > 0.7 {
			return errors.New("medium risk level should have score between 0.3 and 0.7")
		}
	case "high":
		if profile.RiskScore <= 0.7 || profile.RiskScore > 0.9 {
			return errors.New("high risk level should have score between 0.7 and 0.9")
		}
	case "critical":
		if profile.RiskScore <= 0.9 {
			return errors.New("critical risk level should have score > 0.9")
		}
	}

	return nil
}

// ValidateAssessmentFrequency validates assessment frequency rules
func (v *BusinessRuleValidator) ValidateAssessmentFrequency(profile *RiskProfile) error {
	if profile == nil {
		return errors.New("profile cannot be nil")
	}

	// Check if assessment is too frequent
	timeSinceLastAssessment := time.Since(profile.LastAssessment)
	
	switch profile.RiskLevel {
	case "low":
		if timeSinceLastAssessment < 24*time.Hour {
			return errors.New("low risk profiles should not be assessed more than once per day")
		}
	case "medium":
		if timeSinceLastAssessment < 12*time.Hour {
			return errors.New("medium risk profiles should not be assessed more than twice per day")
		}
	case "high", "critical":
		if timeSinceLastAssessment < 1*time.Hour {
			return errors.New("high/critical risk profiles should not be assessed more than once per hour")
		}
	}

	return nil
}

// ValidateRiskFactors validates risk factors for consistency
func (v *BusinessRuleValidator) ValidateRiskFactors(factors map[string]interface{}) error {
	if factors == nil {
		return nil
	}

	// Validate specific risk factors
	if velocity, exists := factors["velocity_score"]; exists {
		if velocityFloat, ok := velocity.(float64); ok {
			if velocityFloat < 0 || velocityFloat > 1 {
				return errors.New("velocity_score must be between 0 and 1")
			}
		}
	}

	if geoScore, exists := factors["geo_risk_score"]; exists {
		if geoFloat, ok := geoScore.(float64); ok {
			if geoFloat < 0 || geoFloat > 1 {
				return errors.New("geo_risk_score must be between 0 and 1")
			}
		}
	}

	if deviceScore, exists := factors["device_risk_score"]; exists {
		if deviceFloat, ok := deviceScore.(float64); ok {
			if deviceFloat < 0 || deviceFloat > 1 {
				return errors.New("device_risk_score must be between 0 and 1")
			}
		}
	}

	return nil
}

// ValidateBusinessRules validates overall business rules for risk assessment
func (v *BusinessRuleValidator) ValidateBusinessRules(assessment *RiskAssessment) error {
	if assessment == nil {
		return errors.New("assessment cannot be nil")
	}

	// Validate decision consistency with risk score
	switch assessment.Decision {
	case "allow":
		if assessment.RiskScore > 0.7 {
			return errors.New("allow decision should not have risk score > 0.7")
		}
	case "review":
		if assessment.RiskScore <= 0.3 || assessment.RiskScore > 0.9 {
			return errors.New("review decision should have risk score between 0.3 and 0.9")
		}
	case "block":
		if assessment.RiskScore <= 0.7 {
			return errors.New("block decision should have risk score > 0.7")
		}
	}

	// Validate confidence level
	if assessment.Confidence < 0.5 {
		return errors.New("assessment confidence should be at least 0.5")
	}

	return nil
}
