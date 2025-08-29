package main

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"regexp"
	"strings"
	"time"
)

// PaymentValidator handles payment validation logic
type PaymentValidator struct{}

// NewPaymentValidator creates a new payment validator
func NewPaymentValidator() *PaymentValidator {
	return &PaymentValidator{}
}

// ValidateCreatePaymentRequest validates payment creation request
func (v *PaymentValidator) ValidateCreatePaymentRequest(req *CreatePaymentRequest) error {
	if req == nil {
		return errors.New("payment request cannot be nil")
	}

	// Validate amount
	if err := v.ValidateAmount(req.Amount); err != nil {
		return fmt.Errorf("invalid amount: %w", err)
	}

	// Validate currency
	if err := v.ValidateCurrency(req.Amount.Currency); err != nil {
		return fmt.Errorf("invalid currency: %w", err)
	}

	// Validate payment method
	if err := v.ValidatePaymentMethod(req.PaymentMethod); err != nil {
		return fmt.Errorf("invalid payment method: %w", err)
	}

	// Validate provider
	if err := v.ValidateProvider(req.Provider); err != nil {
		return fmt.Errorf("invalid provider: %w", err)
	}

	// Validate description
	if strings.TrimSpace(req.Description) == "" {
		return errors.New("description cannot be empty")
	}

	if len(req.Description) > 500 {
		return errors.New("description cannot exceed 500 characters")
	}

	// Validate metadata
	if err := v.ValidateMetadata(req.Metadata); err != nil {
		return fmt.Errorf("invalid metadata: %w", err)
	}

	return nil
}

// ValidateAmount validates payment amount
func (v *PaymentValidator) ValidateAmount(amount Money) error {
	if amount.Value == nil {
		return errors.New("amount value cannot be nil")
	}

	// Check if amount is positive
	zero := big.NewFloat(0)
	if amount.Value.Cmp(zero) <= 0 {
		return errors.New("amount must be positive")
	}

	// Check maximum amount (e.g., $1M per transaction)
	maxAmount := big.NewFloat(100000000) // $1M in cents
	if amount.Value.Cmp(maxAmount) > 0 {
		return errors.New("amount exceeds maximum limit")
	}

	// Check minimum amount (e.g., $0.50)
	minAmount := big.NewFloat(50) // $0.50 in cents
	if amount.Value.Cmp(minAmount) < 0 {
		return errors.New("amount below minimum limit")
	}

	return nil
}

// ValidateCurrency validates currency code
func (v *PaymentValidator) ValidateCurrency(currency string) error {
	if currency == "" {
		return errors.New("currency cannot be empty")
	}

	// List of supported currencies
	supportedCurrencies := map[string]bool{
		"USD": true,
		"EUR": true,
		"GBP": true,
		"CAD": true,
		"AUD": true,
		"JPY": true,
		"CHF": true,
		"SEK": true,
		"NOK": true,
		"DKK": true,
	}

	if !supportedCurrencies[currency] {
		return fmt.Errorf("unsupported currency: %s", currency)
	}

	return nil
}

// ValidatePaymentMethod validates payment method
func (v *PaymentValidator) ValidatePaymentMethod(method string) error {
	if method == "" {
		return errors.New("payment method cannot be empty")
	}

	validMethods := map[string]bool{
		"credit_card":    true,
		"debit_card":     true,
		"bank_transfer":  true,
		"ach":           true,
		"wire_transfer": true,
		"paypal":        true,
		"apple_pay":     true,
		"google_pay":    true,
		"sepa":          true,
		"ideal":         true,
	}

	if !validMethods[method] {
		return fmt.Errorf("unsupported payment method: %s", method)
	}

	return nil
}

// ValidateProvider validates payment provider
func (v *PaymentValidator) ValidateProvider(provider string) error {
	if provider == "" {
		return errors.New("provider cannot be empty")
	}

	validProviders := map[string]bool{
		"stripe":     true,
		"adyen":      true,
		"braintree":  true,
		"square":     true,
		"paypal":     true,
		"checkout":   true,
		"worldpay":   true,
		"authorize":  true,
	}

	if !validProviders[provider] {
		return fmt.Errorf("unsupported provider: %s", provider)
	}

	return nil
}

// ValidateMetadata validates payment metadata
func (v *PaymentValidator) ValidateMetadata(metadata map[string]interface{}) error {
	if metadata == nil {
		return nil
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

// ValidatePaymentAction validates if action is allowed for current state
func ValidatePaymentAction(payment *Payment, action string) error {
	if payment == nil {
		return errors.New("payment cannot be nil")
	}

	allowedActions := map[string][]string{
		"pending":    {"process", "cancel"},
		"processing": {"complete", "fail", "cancel"},
		"completed":  {"refund"},
		"failed":     {"retry"},
		"cancelled":  {},
		"refunded":   {},
	}

	actions, exists := allowedActions[payment.Status]
	if !exists {
		return fmt.Errorf("unknown payment status: %s", payment.Status)
	}

	for _, allowedAction := range actions {
		if allowedAction == action {
			return nil
		}
	}

	return fmt.Errorf("action '%s' not allowed for payment in status '%s'", action, payment.Status)
}

// ValidatePaymentStateTransition validates if state transition is allowed
func ValidatePaymentStateTransition(currentStatus, newStatus string) error {
	validTransitions := map[string][]string{
		"pending":    {"processing", "cancelled"},
		"processing": {"completed", "failed", "cancelled"},
		"completed":  {"refunded"},
		"failed":     {"pending", "cancelled"}, // Allow retry
		"cancelled":  {},                       // Terminal state
		"refunded":   {},                       // Terminal state
	}

	allowedStates, exists := validTransitions[currentStatus]
	if !exists {
		return fmt.Errorf("unknown current status: %s", currentStatus)
	}

	for _, allowedState := range allowedStates {
		if allowedState == newStatus {
			return nil
		}
	}

	return fmt.Errorf("transition from '%s' to '%s' not allowed", currentStatus, newStatus)
}

// BusinessRuleValidator handles business rule validation
type BusinessRuleValidator struct {
	validator *PaymentValidator
}

// NewBusinessRuleValidator creates a new business rule validator
func NewBusinessRuleValidator() *BusinessRuleValidator {
	return &BusinessRuleValidator{
		validator: NewPaymentValidator(),
	}
}

// ValidatePaymentLimits validates payment against business limits
func (v *BusinessRuleValidator) ValidatePaymentLimits(ctx context.Context, payment *Payment) error {
	// Daily limit check (placeholder - would query actual transactions)
	dailyLimit := big.NewFloat(1000000) // $10,000 daily limit
	if payment.Amount.Value.Cmp(dailyLimit) > 0 {
		return errors.New("payment exceeds daily limit")
	}

	// Velocity check - too many payments in short time
	// TODO: Implement velocity checking with actual transaction history

	return nil
}

// ValidateFraudRules validates payment against fraud detection rules
func (v *BusinessRuleValidator) ValidateFraudRules(ctx context.Context, payment *Payment) error {
	// Check for suspicious patterns
	if payment.Metadata != nil {
		// Check for blocked countries
		if country, exists := payment.Metadata["country"]; exists {
			blockedCountries := []string{"XX", "YY"} // Placeholder
			countryStr := fmt.Sprintf("%v", country)
			for _, blocked := range blockedCountries {
				if countryStr == blocked {
					return fmt.Errorf("payments from country %s are blocked", countryStr)
				}
			}
		}

		// Check for suspicious email patterns
		if email, exists := payment.Metadata["email"]; exists {
			emailStr := fmt.Sprintf("%v", email)
			if strings.Contains(emailStr, "suspicious") {
				return errors.New("suspicious email pattern detected")
			}
		}
	}

	// Check amount patterns
	suspiciousAmounts := []*big.Float{
		big.NewFloat(999999), // Just under $10k
		big.NewFloat(499999), // Just under $5k
	}

	for _, suspiciousAmount := range suspiciousAmounts {
		if payment.Amount.Value.Cmp(suspiciousAmount) == 0 {
			return errors.New("suspicious amount pattern detected")
		}
	}

	return nil
}

// ValidateComplianceRules validates payment against compliance requirements
func (v *BusinessRuleValidator) ValidateComplianceRules(ctx context.Context, payment *Payment) error {
	// AML (Anti-Money Laundering) checks
	amlThreshold := big.NewFloat(1000000) // $10,000 threshold
	if payment.Amount.Value.Cmp(amlThreshold) >= 0 {
		// TODO: Trigger AML reporting
		if payment.Metadata == nil {
			payment.Metadata = make(map[string]interface{})
		}
		payment.Metadata["aml_required"] = true
	}

	// KYC (Know Your Customer) requirements
	kycThreshold := big.NewFloat(500000) // $5,000 threshold
	if payment.Amount.Value.Cmp(kycThreshold) >= 0 {
		// TODO: Verify KYC status
		if payment.Metadata == nil {
			payment.Metadata = make(map[string]interface{})
		}
		payment.Metadata["kyc_required"] = true
	}

	return nil
}

// CalculateRiskScore calculates risk score for payment
func (v *BusinessRuleValidator) CalculateRiskScore(payment *Payment) (float64, error) {
	if payment == nil {
		return 0, errors.New("payment cannot be nil")
	}

	score := 0.0

	// Amount-based risk
	amountFloat, _ := payment.Amount.Value.Float64()
	if amountFloat > 100000 { // > $1000
		score += 20
	} else if amountFloat > 50000 { // > $500
		score += 10
	}

	// Payment method risk
	methodRisk := map[string]float64{
		"credit_card":    5,
		"debit_card":     3,
		"bank_transfer":  2,
		"wire_transfer":  8,
		"paypal":         4,
		"apple_pay":      2,
		"google_pay":     2,
	}

	if risk, exists := methodRisk[payment.Method]; exists {
		score += risk
	}

	// Provider risk
	providerRisk := map[string]float64{
		"stripe":    2,
		"adyen":     3,
		"braintree": 4,
		"square":    3,
		"paypal":    5,
	}

	if risk, exists := providerRisk[payment.Provider]; exists {
		score += risk
	}

	// Time-based risk (late night transactions)
	hour := time.Now().Hour()
	if hour < 6 || hour > 22 {
		score += 10
	}

	// Metadata-based risk
	if payment.Metadata != nil {
		if _, exists := payment.Metadata["vpn_detected"]; exists {
			score += 15
		}
		if _, exists := payment.Metadata["new_device"]; exists {
			score += 10
		}
	}

	// Normalize score to 0-100 range
	if score > 100 {
		score = 100
	}

	return score, nil
}

// ValidateCardNumber validates credit card number format
func ValidateCardNumber(cardNumber string) error {
	if cardNumber == "" {
		return errors.New("card number cannot be empty")
	}

	// Remove spaces and dashes
	cardNumber = regexp.MustCompile(`[\s-]`).ReplaceAllString(cardNumber, "")

	// Check if all digits
	if !regexp.MustCompile(`^\d+$`).MatchString(cardNumber) {
		return errors.New("card number must contain only digits")
	}

	// Check length
	if len(cardNumber) < 13 || len(cardNumber) > 19 {
		return errors.New("card number must be between 13 and 19 digits")
	}

	// Luhn algorithm check
	if !isValidLuhn(cardNumber) {
		return errors.New("invalid card number (failed Luhn check)")
	}

	return nil
}

// isValidLuhn implements the Luhn algorithm for card validation
func isValidLuhn(cardNumber string) bool {
	sum := 0
	alternate := false

	for i := len(cardNumber) - 1; i >= 0; i-- {
		digit := int(cardNumber[i] - '0')

		if alternate {
			digit *= 2
			if digit > 9 {
				digit = (digit % 10) + 1
			}
		}

		sum += digit
		alternate = !alternate
	}

	return sum%10 == 0
}
