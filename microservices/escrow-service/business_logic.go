package main

import (
	"context"
	"fmt"
	"time"
)

// EscrowStateMachine handles state transitions and business logic
type EscrowStateMachine struct {
	service *Service
}

// NewEscrowStateMachine creates a new state machine
func NewEscrowStateMachine(service *Service) *EscrowStateMachine {
	return &EscrowStateMachine{service: service}
}

// ReleaseEscrow releases funds to seller with comprehensive validation
func (s *Service) ReleaseEscrow(ctx context.Context, escrowID string) error {
	// Get current escrow
	escrow, err := s.repo.GetEscrow(ctx, escrowID)
	if err != nil {
		return fmt.Errorf("failed to get escrow: %w", err)
	}

	// Validate state transition
	if err := ValidateEscrowAction(escrow, "release"); err != nil {
		return fmt.Errorf("release not allowed: %w", err)
	}

	// Business logic checks
	if err := s.validateReleaseConditions(ctx, escrow); err != nil {
		return fmt.Errorf("release conditions not met: %w", err)
	}

	// Update escrow status
	escrow.Status = "released"
	escrow.UpdatedAt = time.Now()
	if escrow.Metadata == nil {
		escrow.Metadata = make(map[string]interface{})
	}
	escrow.Metadata["released_at"] = time.Now()
	escrow.Metadata["release_reason"] = "normal_completion"

	// Update in repository
	if err := s.repo.UpdateEscrow(ctx, escrow); err != nil {
		return fmt.Errorf("failed to update escrow: %w", err)
	}

	// TODO: Trigger payment to seller
	// TODO: Send notifications
	// TODO: Update ledger entries

	return nil
}

// CancelEscrow cancels an escrow with proper validation
func (s *Service) CancelEscrow(ctx context.Context, escrowID string) error {
	// Get current escrow
	escrow, err := s.repo.GetEscrow(ctx, escrowID)
	if err != nil {
		return fmt.Errorf("failed to get escrow: %w", err)
	}

	// Validate state transition
	if err := ValidateEscrowAction(escrow, "cancel"); err != nil {
		return fmt.Errorf("cancellation not allowed: %w", err)
	}

	// Business logic checks
	if err := s.validateCancellationConditions(ctx, escrow); err != nil {
		return fmt.Errorf("cancellation conditions not met: %w", err)
	}

	// Update escrow status
	escrow.Status = "cancelled"
	escrow.UpdatedAt = time.Now()
	if escrow.Metadata == nil {
		escrow.Metadata = make(map[string]interface{})
	}
	escrow.Metadata["cancelled_at"] = time.Now()
	escrow.Metadata["cancellation_reason"] = "user_requested"

	// Update in repository
	if err := s.repo.UpdateEscrow(ctx, escrow); err != nil {
		return fmt.Errorf("failed to update escrow: %w", err)
	}

	// TODO: Process refund if funded
	// TODO: Send notifications
	// TODO: Update ledger entries

	return nil
}

// DisputeEscrow initiates a dispute process
func (s *Service) DisputeEscrow(ctx context.Context, escrowID, reason string) error {
	// Get current escrow
	escrow, err := s.repo.GetEscrow(ctx, escrowID)
	if err != nil {
		return fmt.Errorf("failed to get escrow: %w", err)
	}

	// Validate state transition
	if err := ValidateEscrowAction(escrow, "dispute"); err != nil {
		return fmt.Errorf("dispute not allowed: %w", err)
	}

	// Validate dispute reason
	if err := s.validateDisputeReason(reason); err != nil {
		return fmt.Errorf("invalid dispute reason: %w", err)
	}

	// Update escrow status
	escrow.Status = "disputed"
	escrow.UpdatedAt = time.Now()
	if escrow.Metadata == nil {
		escrow.Metadata = make(map[string]interface{})
	}
	escrow.Metadata["disputed_at"] = time.Now()
	escrow.Metadata["dispute_reason"] = reason
	escrow.Metadata["dispute_status"] = "open"

	// Update in repository
	if err := s.repo.UpdateEscrow(ctx, escrow); err != nil {
		return fmt.Errorf("failed to update escrow: %w", err)
	}

	// TODO: Create dispute record
	// TODO: Notify dispute resolution team
	// TODO: Send notifications to parties

	return nil
}

// ProcessExpiredEscrows handles escrow expiration logic
func (s *Service) ProcessExpiredEscrows(ctx context.Context) error {
	// Get escrows that might be expired
	filters := EscrowFilters{
		Status: "funded",
		Limit:  100,
	}

	escrows, err := s.repo.ListEscrows(ctx, filters)
	if err != nil {
		return fmt.Errorf("failed to list escrows: %w", err)
	}

	for _, escrow := range escrows {
		if s.isEscrowExpired(escrow) {
			if err := s.expireEscrow(ctx, escrow); err != nil {
				// Log error but continue processing
				continue
			}
		}
	}

	return nil
}

// validateReleaseConditions checks if escrow can be released
func (s *Service) validateReleaseConditions(ctx context.Context, escrow *Escrow) error {
	// Check if escrow is funded
	if escrow.Status != "funded" && escrow.Status != "disputed" {
		return fmt.Errorf("escrow must be funded or disputed to release")
	}

	// Check if not expired
	if s.isEscrowExpired(escrow) {
		return fmt.Errorf("cannot release expired escrow")
	}

	// TODO: Check delivery confirmation
	// TODO: Check dispute resolution if disputed
	// TODO: Validate release authorization

	return nil
}

// validateCancellationConditions checks if escrow can be cancelled
func (s *Service) validateCancellationConditions(ctx context.Context, escrow *Escrow) error {
	// Allow cancellation in early stages
	allowedStatuses := []string{"pending", "funded"}
	statusAllowed := false
	for _, status := range allowedStatuses {
		if escrow.Status == status {
			statusAllowed = true
			break
		}
	}

	if !statusAllowed {
		return fmt.Errorf("escrow in status '%s' cannot be cancelled", escrow.Status)
	}

	// Check cancellation window (e.g., within 24 hours of funding)
	if escrow.Status == "funded" {
		fundedAt, exists := escrow.Metadata["funded_at"]
		if exists {
			if fundedTime, ok := fundedAt.(time.Time); ok {
				if time.Since(fundedTime) > 24*time.Hour {
					return fmt.Errorf("cancellation window expired")
				}
			}
		}
	}

	return nil
}

// validateDisputeReason validates dispute reason
func (s *Service) validateDisputeReason(reason string) error {
	validReasons := []string{
		"non_delivery",
		"defective_goods",
		"not_as_described",
		"unauthorized_transaction",
		"fraud_suspected",
		"other",
	}

	for _, valid := range validReasons {
		if reason == valid {
			return nil
		}
	}

	return fmt.Errorf("invalid dispute reason: %s", reason)
}

// isEscrowExpired checks if escrow has expired
func (s *Service) isEscrowExpired(escrow *Escrow) bool {
	// Default expiration: 30 days from creation
	expirationPeriod := 30 * 24 * time.Hour
	
	// Check for custom expiration in metadata
	if escrow.Metadata != nil {
		if customExpiry, exists := escrow.Metadata["expires_at"]; exists {
			if expiryTime, ok := customExpiry.(time.Time); ok {
				return time.Now().After(expiryTime)
			}
		}
	}

	// Default expiration check
	return time.Since(escrow.CreatedAt) > expirationPeriod
}

// expireEscrow handles escrow expiration
func (s *Service) expireEscrow(ctx context.Context, escrow *Escrow) error {
	escrow.Status = "expired"
	escrow.UpdatedAt = time.Now()
	if escrow.Metadata == nil {
		escrow.Metadata = make(map[string]interface{})
	}
	escrow.Metadata["expired_at"] = time.Now()
	escrow.Metadata["expiration_reason"] = "time_limit_exceeded"

	return s.repo.UpdateEscrow(ctx, escrow)
}

// CalculateEscrowFees calculates fees for escrow transaction
func (s *Service) CalculateEscrowFees(amount Money) (Money, error) {
	// Base fee: 2.5% of transaction amount
	baseFeeRate := 0.025
	
	// Minimum fee: $1.00
	minFee := FromMinorUnits(amount.Currency, 100)
	
	// Maximum fee: $100.00
	maxFee := FromMinorUnits(amount.Currency, 10000)
	
	// Calculate fee
	amountFloat, _ := amount.Value.Float64()
	feeAmount := amountFloat * baseFeeRate
	
	fee := FromMinorUnits(amount.Currency, int64(feeAmount*100))
	
	// Apply min/max limits
	if fee.Value.Cmp(minFee.Value) < 0 {
		fee = minFee
	}
	if fee.Value.Cmp(maxFee.Value) > 0 {
		fee = maxFee
	}
	
	return fee, nil
}

// GetEscrowMetrics returns business metrics for escrows
func (s *Service) GetEscrowMetrics(ctx context.Context) (*EscrowMetrics, error) {
	// This would typically query aggregated data
	// For now, return basic metrics structure
	return &EscrowMetrics{
		TotalEscrows:     0,
		ActiveEscrows:    0,
		CompletedEscrows: 0,
		DisputedEscrows:  0,
		TotalVolume:      FromMinorUnits("USD", 0),
	}, nil
}

// EscrowMetrics represents business metrics
type EscrowMetrics struct {
	TotalEscrows     int64 `json:"total_escrows"`
	ActiveEscrows    int64 `json:"active_escrows"`
	CompletedEscrows int64 `json:"completed_escrows"`
	DisputedEscrows  int64 `json:"disputed_escrows"`
	TotalVolume      Money `json:"total_volume"`
}
