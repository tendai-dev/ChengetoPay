package main

import (
	"context"
	"testing"
	"time"
)

// TestEscrowWorkflow tests the complete escrow lifecycle
func TestEscrowWorkflow(t *testing.T) {
	repo := NewMockRepository()
	service := NewService(repo, nil)

	// Step 1: Create escrow
	createReq := &CreateEscrowRequest{
		BuyerID:  "buyer_alice",
		SellerID: "seller_bob",
		Amount:   FromMinorUnits("USD", 50000), // $500
		Terms:    "Delivery of premium headphones within 7 days",
	}

	escrow, err := service.CreateEscrow(context.Background(), createReq)
	if err != nil {
		t.Fatalf("Failed to create escrow: %v", err)
	}

	if escrow.Status != "pending" {
		t.Errorf("Expected status 'pending', got %s", escrow.Status)
	}

	// Step 2: Fund escrow
	fundReq := &FundEscrowRequest{
		EscrowID:      escrow.ID,
		Amount:        FromMinorUnits("USD", 50000),
		PaymentMethod: "credit_card",
	}

	err = service.FundEscrow(context.Background(), fundReq)
	if err != nil {
		t.Fatalf("Failed to fund escrow: %v", err)
	}

	// Verify escrow is funded
	fundedEscrow, err := service.GetEscrow(context.Background(), escrow.ID)
	if err != nil {
		t.Fatalf("Failed to get escrow: %v", err)
	}

	if fundedEscrow.Status != "funded" {
		t.Errorf("Expected status 'funded', got %s", fundedEscrow.Status)
	}

	// Step 3: Confirm delivery
	deliveryReq := &ConfirmDeliveryRequest{
		EscrowID: escrow.ID,
		Proof:    "Delivery confirmation #12345",
	}

	err = service.ConfirmDelivery(context.Background(), deliveryReq)
	if err != nil {
		t.Fatalf("Failed to confirm delivery: %v", err)
	}

	// Step 4: Release escrow
	err = service.ReleaseEscrow(context.Background(), escrow.ID)
	if err != nil {
		t.Fatalf("Failed to release escrow: %v", err)
	}

	// Verify final state
	finalEscrow, err := service.GetEscrow(context.Background(), escrow.ID)
	if err != nil {
		t.Fatalf("Failed to get final escrow: %v", err)
	}

	if finalEscrow.Status != "released" {
		t.Errorf("Expected status 'released', got %s", finalEscrow.Status)
	}
}

// TestEscrowDispute tests dispute workflow
func TestEscrowDispute(t *testing.T) {
	repo := NewMockRepository()
	service := NewService(repo, nil)

	// Create and fund escrow
	createReq := &CreateEscrowRequest{
		BuyerID:  "buyer_charlie",
		SellerID: "seller_dave",
		Amount:   FromMinorUnits("USD", 25000), // $250
		Terms:    "Custom software development",
	}

	escrow, err := service.CreateEscrow(context.Background(), createReq)
	if err != nil {
		t.Fatalf("Failed to create escrow: %v", err)
	}

	fundReq := &FundEscrowRequest{
		EscrowID: escrow.ID,
		Amount:   FromMinorUnits("USD", 25000),
	}

	err = service.FundEscrow(context.Background(), fundReq)
	if err != nil {
		t.Fatalf("Failed to fund escrow: %v", err)
	}

	// Initiate dispute
	err = service.DisputeEscrow(context.Background(), escrow.ID, "not_as_described")
	if err != nil {
		t.Fatalf("Failed to dispute escrow: %v", err)
	}

	// Verify dispute state
	disputedEscrow, err := service.GetEscrow(context.Background(), escrow.ID)
	if err != nil {
		t.Fatalf("Failed to get disputed escrow: %v", err)
	}

	if disputedEscrow.Status != "disputed" {
		t.Errorf("Expected status 'disputed', got %s", disputedEscrow.Status)
	}

	// Resolve dispute by releasing
	err = service.ReleaseEscrow(context.Background(), escrow.ID)
	if err != nil {
		t.Fatalf("Failed to resolve dispute: %v", err)
	}
}

// TestEscrowCancellation tests cancellation scenarios
func TestEscrowCancellation(t *testing.T) {
	repo := NewMockRepository()
	service := NewService(repo, nil)

	// Test 1: Cancel pending escrow
	createReq := &CreateEscrowRequest{
		BuyerID:  "buyer_eve",
		SellerID: "seller_frank",
		Amount:   FromMinorUnits("USD", 10000), // $100
		Terms:    "Test cancellation scenario",
	}

	escrow, err := service.CreateEscrow(context.Background(), createReq)
	if err != nil {
		t.Fatalf("Failed to create escrow: %v", err)
	}

	err = service.CancelEscrow(context.Background(), escrow.ID)
	if err != nil {
		t.Fatalf("Failed to cancel pending escrow: %v", err)
	}

	cancelledEscrow, err := service.GetEscrow(context.Background(), escrow.ID)
	if err != nil {
		t.Fatalf("Failed to get cancelled escrow: %v", err)
	}

	if cancelledEscrow.Status != "cancelled" {
		t.Errorf("Expected status 'cancelled', got %s", cancelledEscrow.Status)
	}
}

// TestEscrowValidation tests validation scenarios
func TestEscrowValidation(t *testing.T) {
	repo := NewMockRepository()
	service := NewService(repo, nil)

	// Test invalid amount
	invalidReq := &CreateEscrowRequest{
		BuyerID:  "buyer_test",
		SellerID: "seller_test",
		Amount:   FromMinorUnits("USD", 0), // Invalid: zero amount
		Terms:    "Test validation",
	}

	_, err := service.CreateEscrow(context.Background(), invalidReq)
	if err == nil {
		t.Error("Expected validation error for zero amount")
	}

	// Test same buyer and seller
	samePartyReq := &CreateEscrowRequest{
		BuyerID:  "same_user",
		SellerID: "same_user", // Invalid: same as buyer
		Amount:   FromMinorUnits("USD", 10000),
		Terms:    "Test validation",
	}

	_, err = service.CreateEscrow(context.Background(), samePartyReq)
	if err == nil {
		t.Error("Expected validation error for same buyer and seller")
	}

	// Test invalid currency
	invalidCurrencyReq := &CreateEscrowRequest{
		BuyerID:  "buyer_test",
		SellerID: "seller_test",
		Amount:   Money{Value: FromMinorUnits("USD", 10000).Value, Currency: "INVALID"},
		Terms:    "Test validation",
	}

	_, err = service.CreateEscrow(context.Background(), invalidCurrencyReq)
	if err == nil {
		t.Error("Expected validation error for invalid currency")
	}
}

// TestEscrowMetrics tests business metrics calculation
func TestEscrowMetrics(t *testing.T) {
	repo := NewMockRepository()
	service := NewService(repo, nil)

	metrics, err := service.GetEscrowMetrics(context.Background())
	if err != nil {
		t.Fatalf("Failed to get metrics: %v", err)
	}

	if metrics == nil {
		t.Error("Expected metrics, got nil")
	}
}

// TestEscrowFeeCalculation tests fee calculation
func TestEscrowFeeCalculation(t *testing.T) {
	repo := NewMockRepository()
	service := NewService(repo, nil)

	testCases := []struct {
		amount      Money
		expectedMin float64
		expectedMax float64
	}{
		{FromMinorUnits("USD", 10000), 1.0, 2.5},   // $100 -> $1-2.50 fee
		{FromMinorUnits("USD", 100000), 25.0, 25.0}, // $1000 -> $25 fee
		{FromMinorUnits("USD", 50), 1.0, 1.0},       // $0.50 -> $1 min fee
	}

	for _, tc := range testCases {
		fee, err := service.CalculateEscrowFees(tc.amount)
		if err != nil {
			t.Errorf("Failed to calculate fee for %v: %v", tc.amount, err)
			continue
		}

		feeFloat, _ := fee.Value.Float64()
		if feeFloat < tc.expectedMin || feeFloat > tc.expectedMax {
			t.Errorf("Fee %f not in expected range [%f, %f] for amount %v",
				feeFloat, tc.expectedMin, tc.expectedMax, tc.amount)
		}
	}
}

// TestConcurrentEscrowOperations tests thread safety
func TestConcurrentEscrowOperations(t *testing.T) {
	repo := NewMockRepository()
	service := NewService(repo, nil)

	// Create multiple escrows concurrently
	numGoroutines := 10
	results := make(chan error, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			req := &CreateEscrowRequest{
				BuyerID:  "buyer_concurrent",
				SellerID: "seller_concurrent",
				Amount:   FromMinorUnits("USD", int64(1000*(id+1))),
				Terms:    "Concurrent test escrow",
			}

			_, err := service.CreateEscrow(context.Background(), req)
			results <- err
		}(i)
	}

	// Check all results
	for i := 0; i < numGoroutines; i++ {
		err := <-results
		if err != nil {
			t.Errorf("Concurrent escrow creation failed: %v", err)
		}
	}
}

// TestEscrowExpiration tests expiration logic
func TestEscrowExpiration(t *testing.T) {
	repo := NewMockRepository()
	service := NewService(repo, nil)

	// Create escrow with custom expiration
	createReq := &CreateEscrowRequest{
		BuyerID:  "buyer_expire",
		SellerID: "seller_expire",
		Amount:   FromMinorUnits("USD", 10000),
		Terms:    "Test expiration",
	}

	escrow, err := service.CreateEscrow(context.Background(), createReq)
	if err != nil {
		t.Fatalf("Failed to create escrow: %v", err)
	}

	// Set custom expiration in metadata
	escrow.Metadata["expires_at"] = time.Now().Add(-1 * time.Hour) // Expired 1 hour ago
	repo.UpdateEscrow(context.Background(), escrow)

	// Test expiration check
	if !service.isEscrowExpired(escrow) {
		t.Error("Expected escrow to be expired")
	}

	// Process expired escrows
	err = service.ProcessExpiredEscrows(context.Background())
	if err != nil {
		t.Fatalf("Failed to process expired escrows: %v", err)
	}
}
