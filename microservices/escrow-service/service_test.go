package main

import (
	"context"
	"testing"
)

func TestEscrowService_CreateEscrow(t *testing.T) {
	repo := NewMockRepository()
	service := NewService(repo, nil)

	req := &CreateEscrowRequest{
		BuyerID:  "buyer_123",
		SellerID: "seller_456",
		Amount:   FromMinorUnits("USD", 10000), // $100
		Terms:    "Standard escrow terms",
	}

	escrow, err := service.CreateEscrow(context.Background(), req)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if escrow.BuyerID != req.BuyerID {
		t.Errorf("Expected buyer ID %s, got %s", req.BuyerID, escrow.BuyerID)
	}

	if escrow.SellerID != req.SellerID {
		t.Errorf("Expected seller ID %s, got %s", req.SellerID, escrow.SellerID)
	}

	if escrow.Status != "pending" {
		t.Errorf("Expected status 'pending', got %s", escrow.Status)
	}

	if escrow.Amount.Currency != "USD" {
		t.Errorf("Expected currency USD, got %s", escrow.Amount.Currency)
	}
}

func TestEscrowService_FundEscrow(t *testing.T) {
	repo := NewMockRepository()
	service := NewService(repo, nil)

	req := &FundEscrowRequest{
		EscrowID: "escrow_123",
		Amount:   FromMinorUnits("USD", 10000),
	}

	err := service.FundEscrow(context.Background(), req)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
}

func TestEscrowService_GetEscrow(t *testing.T) {
	repo := NewMockRepository()
	service := NewService(repo, nil)

	escrow, err := service.GetEscrow(context.Background(), "test_id")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if escrow.ID != "test_id" {
		t.Errorf("Expected ID 'test_id', got %s", escrow.ID)
	}

	if escrow.Status != "pending" {
		t.Errorf("Expected status 'pending', got %s", escrow.Status)
	}
}

func TestEscrowService_ListEscrows(t *testing.T) {
	repo := NewMockRepository()
	service := NewService(repo, nil)

	filters := EscrowFilters{
		Status: "pending",
		Limit:  10,
	}

	escrows, err := service.ListEscrows(context.Background(), filters)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if escrows == nil {
		t.Error("Expected escrows slice, got nil")
	}
}

func TestEscrowService_ReleaseEscrow(t *testing.T) {
	repo := NewMockRepository()
	service := NewService(repo, nil)

	// Create a funded escrow first
	req := &CreateEscrowRequest{
		BuyerID:  "buyer_123",
		SellerID: "seller_456",
		Amount:   FromMinorUnits("USD", 10000),
		Terms:    "Standard escrow terms for testing",
	}

	escrow, err := service.CreateEscrow(context.Background(), req)
	if err != nil {
		t.Fatalf("Failed to create escrow: %v", err)
	}

	// Fund the escrow
	fundReq := &FundEscrowRequest{
		EscrowID: escrow.ID,
		Amount:   FromMinorUnits("USD", 10000),
	}

	err = service.FundEscrow(context.Background(), fundReq)
	if err != nil {
		t.Fatalf("Failed to fund escrow: %v", err)
	}

	// Now test release
	err = service.ReleaseEscrow(context.Background(), escrow.ID)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
}

func TestEscrowService_CancelEscrow(t *testing.T) {
	repo := NewMockRepository()
	service := NewService(repo, nil)

	err := service.CancelEscrow(context.Background(), "escrow_123")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
}

func TestMoney_FromMinorUnits(t *testing.T) {
	tests := []struct {
		currency    string
		minorUnits  int64
		expectedVal float64
	}{
		{"USD", 10000, 100.0},
		{"EUR", 2550, 25.5},
		{"GBP", 199, 1.99},
	}

	for _, test := range tests {
		money := FromMinorUnits(test.currency, test.minorUnits)
		
		if money.Currency != test.currency {
			t.Errorf("Expected currency %s, got %s", test.currency, money.Currency)
		}

		val, _ := money.Value.Float64()
		if val != test.expectedVal {
			t.Errorf("Expected value %f, got %f", test.expectedVal, val)
		}
	}
}

// Benchmark tests
func BenchmarkEscrowService_CreateEscrow(b *testing.B) {
	repo := NewMockRepository()
	service := NewService(repo, nil)

	req := &CreateEscrowRequest{
		BuyerID:  "buyer_123",
		SellerID: "seller_456",
		Amount:   FromMinorUnits("USD", 10000),
		Terms:    "Standard escrow terms",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := service.CreateEscrow(context.Background(), req)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkEscrowService_GetEscrow(b *testing.B) {
	repo := NewMockRepository()
	service := NewService(repo, nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := service.GetEscrow(context.Background(), "test_id")
		if err != nil {
			b.Fatal(err)
		}
	}
}
