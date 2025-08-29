package main

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"time"
)

// ProcessPayment handles payment processing with business logic
func (s *Service) ProcessPayment(ctx context.Context, paymentID string) error {
	// Get payment
	payment, err := s.repo.GetPayment(ctx, paymentID)
	if err != nil {
		return fmt.Errorf("failed to get payment: %w", err)
	}

	// Validate payment can be processed
	if err := ValidatePaymentAction(payment, "process"); err != nil {
		return fmt.Errorf("process not allowed: %w", err)
	}

	// Validate business rules
	validator := NewBusinessRuleValidator()
	if err := validator.ValidatePaymentLimits(ctx, payment); err != nil {
		return fmt.Errorf("payment limits validation failed: %w", err)
	}

	if err := validator.ValidateFraudRules(ctx, payment); err != nil {
		return fmt.Errorf("fraud validation failed: %w", err)
	}

	if err := validator.ValidateComplianceRules(ctx, payment); err != nil {
		return fmt.Errorf("compliance validation failed: %w", err)
	}

	// Calculate and store risk score
	riskScore, err := validator.CalculateRiskScore(payment)
	if err != nil {
		return fmt.Errorf("failed to calculate risk score: %w", err)
	}

	if payment.Metadata == nil {
		payment.Metadata = make(map[string]interface{})
	}
	payment.Metadata["risk_score"] = riskScore
	payment.Metadata["processed_at"] = time.Now()

	// High risk payments require manual review
	if riskScore > 70 {
		payment.Status = "pending_review"
		payment.Metadata["review_required"] = true
		payment.Metadata["review_reason"] = "high_risk_score"
	} else {
		// Update status to processing
		payment.Status = "processing"
	}

	payment.UpdatedAt = time.Now()

	// Update payment in repository
	if err := s.repo.UpdatePayment(ctx, payment); err != nil {
		return fmt.Errorf("failed to update payment: %w", err)
	}

	// TODO: Integrate with actual payment processor
	// TODO: Handle webhook responses
	// TODO: Update ledger entries

	return nil
}

// CompletePayment marks payment as completed
func (s *Service) CompletePayment(ctx context.Context, paymentID string, providerTransactionID string) error {
	payment, err := s.repo.GetPayment(ctx, paymentID)
	if err != nil {
		return fmt.Errorf("failed to get payment: %w", err)
	}

	// Validate completion is allowed
	allowedStatuses := []string{"processing", "pending_review"}
	statusAllowed := false
	for _, status := range allowedStatuses {
		if payment.Status == status {
			statusAllowed = true
			break
		}
	}

	if !statusAllowed {
		return fmt.Errorf("payment completion not allowed for status: %s", payment.Status)
	}

	// Update payment status and metadata
	payment.Status = "completed"
	payment.UpdatedAt = time.Now()

	if payment.Metadata == nil {
		payment.Metadata = make(map[string]interface{})
	}
	payment.Metadata["completed_at"] = time.Now()
	payment.Metadata["provider_transaction_id"] = providerTransactionID

	// Calculate fees
	fees, err := s.CalculatePaymentFees(*payment)
	if err != nil {
		return fmt.Errorf("failed to calculate fees: %w", err)
	}
	payment.Metadata["fees"] = fees

	// Update in repository
	if err := s.repo.UpdatePayment(ctx, payment); err != nil {
		return fmt.Errorf("failed to update payment: %w", err)
	}

	// TODO: Create ledger entries
	// TODO: Send completion notifications
	// TODO: Update escrow if applicable

	return nil
}

// FailPayment marks payment as failed
func (s *Service) FailPayment(ctx context.Context, paymentID string, reason string) error {
	payment, err := s.repo.GetPayment(ctx, paymentID)
	if err != nil {
		return fmt.Errorf("failed to get payment: %w", err)
	}

	// Validate failure is allowed
	if err := ValidatePaymentAction(payment, "fail"); err != nil {
		return fmt.Errorf("payment failure not allowed: %w", err)
	}

	// Update payment status
	payment.Status = "failed"
	payment.UpdatedAt = time.Now()

	if payment.Metadata == nil {
		payment.Metadata = make(map[string]interface{})
	}
	payment.Metadata["failed_at"] = time.Now()
	payment.Metadata["failure_reason"] = reason

	// Increment retry count
	if retryCount, exists := payment.Metadata["retry_count"]; exists {
		if count, ok := retryCount.(float64); ok {
			payment.Metadata["retry_count"] = count + 1
		}
	} else {
		payment.Metadata["retry_count"] = 1
	}

	// Update in repository
	if err := s.repo.UpdatePayment(ctx, payment); err != nil {
		return fmt.Errorf("failed to update payment: %w", err)
	}

	// TODO: Send failure notifications
	// TODO: Handle automatic retry logic
	// TODO: Update related escrow status

	return nil
}

// RefundPayment processes payment refund
func (s *Service) RefundPayment(ctx context.Context, paymentID string, refundAmount Money, reason string) error {
	payment, err := s.repo.GetPayment(ctx, paymentID)
	if err != nil {
		return fmt.Errorf("failed to get payment: %w", err)
	}

	// Validate refund is allowed
	if err := ValidatePaymentAction(payment, "refund"); err != nil {
		return fmt.Errorf("refund not allowed: %w", err)
	}

	// Validate refund amount
	if err := s.validateRefundAmount(payment, refundAmount); err != nil {
		return fmt.Errorf("invalid refund amount: %w", err)
	}

	// Check if partial or full refund
	isFullRefund := refundAmount.Value.Cmp(payment.Amount.Value) == 0

	// Update payment status
	if isFullRefund {
		payment.Status = "refunded"
	} else {
		payment.Status = "partially_refunded"
	}

	payment.UpdatedAt = time.Now()

	if payment.Metadata == nil {
		payment.Metadata = make(map[string]interface{})
	}

	// Track refund details
	refundData := map[string]interface{}{
		"amount":     refundAmount.Value.String(),
		"currency":   refundAmount.Currency,
		"reason":     reason,
		"refunded_at": time.Now(),
		"is_full_refund": isFullRefund,
	}

	// Handle multiple refunds
	if existingRefunds, exists := payment.Metadata["refunds"]; exists {
		if refunds, ok := existingRefunds.([]interface{}); ok {
			payment.Metadata["refunds"] = append(refunds, refundData)
		}
	} else {
		payment.Metadata["refunds"] = []interface{}{refundData}
	}

	// Calculate total refunded amount
	totalRefunded := big.NewFloat(0)
	if refunds, exists := payment.Metadata["refunds"]; exists {
		if refundList, ok := refunds.([]interface{}); ok {
			for _, refund := range refundList {
				if refundMap, ok := refund.(map[string]interface{}); ok {
					if amountStr, ok := refundMap["amount"].(string); ok {
						amount := big.NewFloat(0)
						amount.SetString(amountStr)
						totalRefunded.Add(totalRefunded, amount)
					}
				}
			}
		}
	}

	payment.Metadata["total_refunded"] = totalRefunded.String()

	// Update in repository
	if err := s.repo.UpdatePayment(ctx, payment); err != nil {
		return fmt.Errorf("failed to update payment: %w", err)
	}

	// TODO: Process actual refund with payment provider
	// TODO: Create refund ledger entries
	// TODO: Send refund notifications

	return nil
}

// CancelPayment cancels a payment
func (s *Service) CancelPayment(ctx context.Context, paymentID string, reason string) error {
	payment, err := s.repo.GetPayment(ctx, paymentID)
	if err != nil {
		return fmt.Errorf("failed to get payment: %w", err)
	}

	// Validate cancellation is allowed
	if err := ValidatePaymentAction(payment, "cancel"); err != nil {
		return fmt.Errorf("cancellation not allowed: %w", err)
	}

	// Update payment status
	payment.Status = "cancelled"
	payment.UpdatedAt = time.Now()

	if payment.Metadata == nil {
		payment.Metadata = make(map[string]interface{})
	}
	payment.Metadata["cancelled_at"] = time.Now()
	payment.Metadata["cancellation_reason"] = reason

	// Update in repository
	if err := s.repo.UpdatePayment(ctx, payment); err != nil {
		return fmt.Errorf("failed to update payment: %w", err)
	}

	// TODO: Cancel with payment provider if already processing
	// TODO: Send cancellation notifications
	// TODO: Update related escrow status

	return nil
}

// RetryPayment retries a failed payment
func (s *Service) RetryPayment(ctx context.Context, paymentID string) error {
	payment, err := s.repo.GetPayment(ctx, paymentID)
	if err != nil {
		return fmt.Errorf("failed to get payment: %w", err)
	}

	// Validate retry is allowed
	if err := ValidatePaymentAction(payment, "retry"); err != nil {
		return fmt.Errorf("retry not allowed: %w", err)
	}

	// Check retry limits
	maxRetries := 3
	retryCount := 0
	if count, exists := payment.Metadata["retry_count"]; exists {
		if c, ok := count.(float64); ok {
			retryCount = int(c)
		}
	}

	if retryCount >= maxRetries {
		return fmt.Errorf("maximum retry attempts (%d) exceeded", maxRetries)
	}

	// Reset payment to pending for retry
	payment.Status = "pending"
	payment.UpdatedAt = time.Now()

	if payment.Metadata == nil {
		payment.Metadata = make(map[string]interface{})
	}
	payment.Metadata["retry_attempted_at"] = time.Now()

	// Update in repository
	if err := s.repo.UpdatePayment(ctx, payment); err != nil {
		return fmt.Errorf("failed to update payment: %w", err)
	}

	// Process the payment again
	return s.ProcessPayment(ctx, paymentID)
}

// CalculatePaymentFees calculates fees for a payment
func (s *Service) CalculatePaymentFees(payment Payment) (Money, error) {
	if payment.Amount.Value == nil {
		return Money{}, errors.New("payment amount cannot be nil")
	}

	// Base fee structure
	baseFeePercent := 0.029 // 2.9%
	fixedFee := big.NewFloat(30) // $0.30

	// Payment method specific fees
	methodFees := map[string]float64{
		"credit_card":    0.029,
		"debit_card":     0.024,
		"bank_transfer":  0.008,
		"ach":           0.008,
		"wire_transfer": 0.015,
		"paypal":        0.034,
		"apple_pay":     0.029,
		"google_pay":    0.029,
	}

	if methodFee, exists := methodFees[payment.Method]; exists {
		baseFeePercent = methodFee
	}

	// Calculate percentage fee
	percentageFee := big.NewFloat(0)
	percentageFee.Mul(payment.Amount.Value, big.NewFloat(baseFeePercent))

	// Total fee = percentage fee + fixed fee
	totalFee := big.NewFloat(0)
	totalFee.Add(percentageFee, fixedFee)

	// Minimum fee
	minFee := big.NewFloat(50) // $0.50
	if totalFee.Cmp(minFee) < 0 {
		totalFee = minFee
	}

	// Maximum fee cap (e.g., $500)
	maxFee := big.NewFloat(50000) // $500
	if totalFee.Cmp(maxFee) > 0 {
		totalFee = maxFee
	}

	return Money{
		Value:    totalFee,
		Currency: payment.Amount.Currency,
	}, nil
}

// GetPaymentMetrics returns payment metrics
func (s *Service) GetPaymentMetrics(ctx context.Context) (*PaymentMetrics, error) {
	// TODO: Implement actual metrics calculation from database
	// This is a placeholder implementation

	return &PaymentMetrics{
		TotalPayments:     1000,
		TotalVolume:       FromMinorUnits("USD", 50000000), // $500k
		SuccessRate:       0.95,
		AverageAmount:     FromMinorUnits("USD", 50000),    // $500
		TopPaymentMethod:  "credit_card",
		TopProvider:       "stripe",
		ProcessingTime:    2.5, // seconds
		RefundRate:        0.03,
		ChargebackRate:    0.01,
		FraudRate:         0.005,
	}, nil
}

// validateRefundAmount validates refund amount against payment
func (s *Service) validateRefundAmount(payment *Payment, refundAmount Money) error {
	if refundAmount.Currency != payment.Amount.Currency {
		return errors.New("refund currency must match payment currency")
	}

	if refundAmount.Value.Cmp(big.NewFloat(0)) <= 0 {
		return errors.New("refund amount must be positive")
	}

	// Calculate total already refunded
	totalRefunded := big.NewFloat(0)
	if refunds, exists := payment.Metadata["refunds"]; exists {
		if refundList, ok := refunds.([]interface{}); ok {
			for _, refund := range refundList {
				if refundMap, ok := refund.(map[string]interface{}); ok {
					if amountStr, ok := refundMap["amount"].(string); ok {
						amount := big.NewFloat(0)
						amount.SetString(amountStr)
						totalRefunded.Add(totalRefunded, amount)
					}
				}
			}
		}
	}

	// Check if refund would exceed payment amount
	totalAfterRefund := big.NewFloat(0)
	totalAfterRefund.Add(totalRefunded, refundAmount.Value)

	if totalAfterRefund.Cmp(payment.Amount.Value) > 0 {
		return errors.New("total refund amount cannot exceed payment amount")
	}

	return nil
}

// PaymentMetrics represents payment analytics
type PaymentMetrics struct {
	TotalPayments     int     `json:"total_payments"`
	TotalVolume       Money   `json:"total_volume"`
	SuccessRate       float64 `json:"success_rate"`
	AverageAmount     Money   `json:"average_amount"`
	TopPaymentMethod  string  `json:"top_payment_method"`
	TopProvider       string  `json:"top_provider"`
	ProcessingTime    float64 `json:"processing_time_seconds"`
	RefundRate        float64 `json:"refund_rate"`
	ChargebackRate    float64 `json:"chargeback_rate"`
	FraudRate         float64 `json:"fraud_rate"`
}
