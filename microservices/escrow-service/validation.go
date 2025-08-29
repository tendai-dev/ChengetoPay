package main

import (
	"errors"
	"fmt"
	"math/big"
	"regexp"
	"strings"
)

var (
	// Business rule errors
	ErrInvalidAmount        = errors.New("amount must be greater than zero")
	ErrInvalidCurrency      = errors.New("invalid currency code")
	ErrInvalidParticipant   = errors.New("invalid participant ID")
	ErrInvalidTerms         = errors.New("terms cannot be empty")
	ErrInsufficientFunds    = errors.New("insufficient funds for escrow")
	ErrInvalidStateTransition = errors.New("invalid state transition")
	ErrEscrowExpired        = errors.New("escrow has expired")
	ErrUnauthorizedAction   = errors.New("unauthorized action")
	ErrDuplicateEscrow      = errors.New("duplicate escrow detected")
	
	// Validation patterns
	currencyPattern = regexp.MustCompile(`^[A-Z]{3}$`)
	participantPattern = regexp.MustCompile(`^[a-zA-Z0-9_-]{3,50}$`)
)

// ValidateCreateEscrowRequest validates escrow creation request
func ValidateCreateEscrowRequest(req *CreateEscrowRequest) error {
	if req == nil {
		return errors.New("request cannot be nil")
	}

	// Validate buyer ID
	if err := validateParticipantID(req.BuyerID, "buyer"); err != nil {
		return err
	}

	// Validate seller ID
	if err := validateParticipantID(req.SellerID, "seller"); err != nil {
		return err
	}

	// Buyer and seller cannot be the same
	if req.BuyerID == req.SellerID {
		return errors.New("buyer and seller cannot be the same")
	}

	// Validate amount
	if err := validateAmount(req.Amount); err != nil {
		return err
	}

	// Validate terms
	if err := validateTerms(req.Terms); err != nil {
		return err
	}

	return nil
}

// ValidateFundEscrowRequest validates escrow funding request
func ValidateFundEscrowRequest(req *FundEscrowRequest) error {
	if req == nil {
		return errors.New("request cannot be nil")
	}

	if strings.TrimSpace(req.EscrowID) == "" {
		return errors.New("escrow ID cannot be empty")
	}

	if err := validateAmount(req.Amount); err != nil {
		return err
	}

	return nil
}

// ValidateEscrowStateTransition validates if state transition is allowed
func ValidateEscrowStateTransition(currentStatus, newStatus string) error {
	validTransitions := map[string][]string{
		"pending":   {"funded", "cancelled"},
		"funded":    {"released", "disputed", "cancelled"},
		"disputed":  {"released", "refunded", "cancelled"},
		"released":  {}, // Terminal state
		"refunded":  {}, // Terminal state
		"cancelled": {}, // Terminal state
		"expired":   {}, // Terminal state
	}

	allowedStates, exists := validTransitions[currentStatus]
	if !exists {
		return fmt.Errorf("unknown current status: %s", currentStatus)
	}

	for _, allowed := range allowedStates {
		if allowed == newStatus {
			return nil
		}
	}

	return fmt.Errorf("cannot transition from %s to %s", currentStatus, newStatus)
}

// ValidateEscrowAction validates if action is allowed for current state
func ValidateEscrowAction(escrow *Escrow, action string) error {
	if escrow == nil {
		return errors.New("escrow cannot be nil")
	}

	allowedActions := map[string][]string{
		"pending":   {"fund", "cancel"},
		"funded":    {"release", "dispute", "cancel"},
		"disputed":  {"release", "refund", "cancel"},
		"released":  {},
		"refunded":  {},
		"cancelled": {},
		"expired":   {},
	}

	actions, exists := allowedActions[escrow.Status]
	if !exists {
		return fmt.Errorf("unknown escrow status: %s", escrow.Status)
	}

	for _, allowed := range actions {
		if allowed == action {
			return nil
		}
	}

	return fmt.Errorf("action '%s' not allowed for escrow in status '%s'", action, escrow.Status)
}

// validateParticipantID validates participant ID format
func validateParticipantID(id, role string) error {
	if strings.TrimSpace(id) == "" {
		return fmt.Errorf("%s ID cannot be empty", role)
	}

	if !participantPattern.MatchString(id) {
		return fmt.Errorf("%s ID must be 3-50 characters, alphanumeric with hyphens/underscores", role)
	}

	return nil
}

// validateAmount validates monetary amount
func validateAmount(amount Money) error {
	if amount.Value == nil {
		return ErrInvalidAmount
	}

	// Check if amount is positive
	zero := big.NewFloat(0)
	if amount.Value.Cmp(zero) <= 0 {
		return ErrInvalidAmount
	}

	// Validate currency code
	if !currencyPattern.MatchString(amount.Currency) {
		return ErrInvalidCurrency
	}

	// Check for reasonable maximum (e.g., $1 billion)
	maxAmount := big.NewFloat(1000000000)
	if amount.Value.Cmp(maxAmount) > 0 {
		return errors.New("amount exceeds maximum allowed limit")
	}

	return nil
}

// validateTerms validates escrow terms
func validateTerms(terms string) error {
	terms = strings.TrimSpace(terms)
	if terms == "" {
		return ErrInvalidTerms
	}

	if len(terms) < 10 {
		return errors.New("terms must be at least 10 characters")
	}

	if len(terms) > 5000 {
		return errors.New("terms cannot exceed 5000 characters")
	}

	// Check for prohibited content
	prohibitedWords := []string{"illegal", "fraud", "scam", "money laundering"}
	lowerTerms := strings.ToLower(terms)
	for _, word := range prohibitedWords {
		if strings.Contains(lowerTerms, word) {
			return fmt.Errorf("terms contain prohibited content: %s", word)
		}
	}

	return nil
}

// BusinessRuleValidator provides business rule validation
type BusinessRuleValidator struct {
	minEscrowAmount map[string]*big.Float // Currency -> minimum amount
	maxEscrowAmount map[string]*big.Float // Currency -> maximum amount
}

// NewBusinessRuleValidator creates a new validator with default rules
func NewBusinessRuleValidator() *BusinessRuleValidator {
	return &BusinessRuleValidator{
		minEscrowAmount: map[string]*big.Float{
			"USD": big.NewFloat(1.00),    // $1 minimum
			"EUR": big.NewFloat(1.00),    // €1 minimum
			"GBP": big.NewFloat(1.00),    // £1 minimum
		},
		maxEscrowAmount: map[string]*big.Float{
			"USD": big.NewFloat(1000000), // $1M maximum
			"EUR": big.NewFloat(1000000), // €1M maximum
			"GBP": big.NewFloat(1000000), // £1M maximum
		},
	}
}

// ValidateEscrowLimits validates escrow amount against business limits
func (v *BusinessRuleValidator) ValidateEscrowLimits(amount Money) error {
	currency := amount.Currency
	
	// Check minimum amount
	if minAmount, exists := v.minEscrowAmount[currency]; exists {
		if amount.Value.Cmp(minAmount) < 0 {
			return fmt.Errorf("amount below minimum limit for %s: %s", currency, minAmount.String())
		}
	}

	// Check maximum amount
	if maxAmount, exists := v.maxEscrowAmount[currency]; exists {
		if amount.Value.Cmp(maxAmount) > 0 {
			return fmt.Errorf("amount exceeds maximum limit for %s: %s", currency, maxAmount.String())
		}
	}

	return nil
}

// ValidateParticipantEligibility validates if participants can create escrow
func (v *BusinessRuleValidator) ValidateParticipantEligibility(buyerID, sellerID string) error {
	// In a real system, this would check:
	// - KYC/KYB status
	// - Account standing
	// - Regulatory compliance
	// - Risk assessment scores
	
	// Placeholder validation
	blockedParticipants := []string{"blocked_user", "suspended_account"}
	
	for _, blocked := range blockedParticipants {
		if buyerID == blocked || sellerID == blocked {
			return fmt.Errorf("participant %s is not eligible for escrow transactions", blocked)
		}
	}

	return nil
}

// ValidateEscrowRisk performs risk assessment validation
func (v *BusinessRuleValidator) ValidateEscrowRisk(req *CreateEscrowRequest) error {
	// High-risk indicators
	riskFactors := 0

	// Large amount risk
	largeAmountThreshold := big.NewFloat(100000) // $100k
	if req.Amount.Value.Cmp(largeAmountThreshold) > 0 {
		riskFactors++
	}

	// Suspicious terms patterns
	suspiciousPatterns := []string{"urgent", "confidential", "no questions", "cash only"}
	lowerTerms := strings.ToLower(req.Terms)
	for _, pattern := range suspiciousPatterns {
		if strings.Contains(lowerTerms, pattern) {
			riskFactors++
		}
	}

	// High risk threshold
	if riskFactors >= 2 {
		return errors.New("escrow flagged for manual review due to risk factors")
	}

	return nil
}

// calculateRiskScore calculates a risk score for escrow creation
func calculateRiskScore(req *CreateEscrowRequest) float64 {
	score := 0.0
	
	// Amount-based risk (higher amounts = higher risk)
	amountFloat, _ := req.Amount.Value.Float64()
	if amountFloat > 100000 {
		score += 0.3
	} else if amountFloat > 10000 {
		score += 0.2
	} else if amountFloat > 1000 {
		score += 0.1
	}
	
	// Terms-based risk
	lowerTerms := strings.ToLower(req.Terms)
	riskKeywords := []string{"urgent", "confidential", "no questions", "cash only", "bitcoin", "crypto"}
	for _, keyword := range riskKeywords {
		if strings.Contains(lowerTerms, keyword) {
			score += 0.15
		}
	}
	
	// Participant pattern risk (simplified)
	if strings.Contains(req.BuyerID, "temp") || strings.Contains(req.SellerID, "temp") {
		score += 0.2
	}
	
	// Cap at 1.0
	if score > 1.0 {
		score = 1.0
	}
	
	return score
}
