package main

import (
	"context"
	"testing"
)

func TestPaymentService_CreatePayment(t *testing.T) {
	repo := &MockRepository{}
	service := NewService(repo, nil)

	req := &CreatePaymentRequest{
		AccountID: "acc_123",
		Provider:  "stripe",
		Method:    "card",
		Amount:    FromMinorUnits("USD", 5000), // $50
		Metadata:  map[string]interface{}{"order_id": "order_123"},
	}

	payment, err := service.CreatePayment(context.Background(), req)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if payment.AccountID != req.AccountID {
		t.Errorf("Expected account ID %s, got %s", req.AccountID, payment.AccountID)
	}

	if payment.Provider != req.Provider {
		t.Errorf("Expected provider %s, got %s", req.Provider, payment.Provider)
	}

	if payment.Status != "pending" {
		t.Errorf("Expected status 'pending', got %s", payment.Status)
	}

	if payment.Amount.Currency != "USD" {
		t.Errorf("Expected currency USD, got %s", payment.Amount.Currency)
	}
}

func TestPaymentService_ProcessPayment(t *testing.T) {
	repo := &MockRepository{}
	service := NewService(repo, nil)

	req := &ProcessPaymentRequest{
		PaymentID: "payment_123",
		Provider:  "stripe",
		Method:    "card",
	}

	err := service.ProcessPayment(context.Background(), req)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
}

func TestPaymentService_GetPayment(t *testing.T) {
	repo := &MockRepository{}
	service := NewService(repo, nil)

	payment, err := service.GetPayment(context.Background(), "test_id")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if payment.ID != "test_id" {
		t.Errorf("Expected ID 'test_id', got %s", payment.ID)
	}

	if payment.Status != "completed" {
		t.Errorf("Expected status 'completed', got %s", payment.Status)
	}
}

func TestPaymentService_ListPayments(t *testing.T) {
	repo := &MockRepository{}
	service := NewService(repo, nil)

	filters := PaymentFilters{
		AccountID: "acc_123",
		Status:    "completed",
		Limit:     10,
	}

	payments, err := service.ListPayments(context.Background(), filters)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if payments == nil {
		t.Error("Expected payments slice, got nil")
	}
}

func TestPaymentService_GetProviders(t *testing.T) {
	repo := &MockRepository{}
	service := NewService(repo, nil)

	providers, err := service.GetProviders(context.Background())
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(providers) == 0 {
		t.Error("Expected providers, got empty slice")
	}

	// Check if default providers are present
	providerNames := make(map[string]bool)
	for _, provider := range providers {
		providerNames[provider.Name] = true
	}

	expectedProviders := []string{"stripe", "paypal", "mpesa"}
	for _, expected := range expectedProviders {
		if !providerNames[expected] {
			t.Errorf("Expected provider %s not found", expected)
		}
	}
}

// Test provider validation
func TestPaymentService_ValidateProvider(t *testing.T) {
	repo := &MockRepository{}
	service := NewService(repo, nil)

	providers, err := service.GetProviders(context.Background())
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	for _, provider := range providers {
		if !provider.Enabled {
			t.Errorf("Provider %s should be enabled", provider.Name)
		}

		if len(provider.Methods) == 0 {
			t.Errorf("Provider %s should have payment methods", provider.Name)
		}

		if len(provider.Currencies) == 0 {
			t.Errorf("Provider %s should support currencies", provider.Name)
		}
	}
}

// Benchmark tests
func BenchmarkPaymentService_CreatePayment(b *testing.B) {
	repo := &MockRepository{}
	service := NewService(repo, nil)

	req := &CreatePaymentRequest{
		AccountID: "acc_123",
		Provider:  "stripe",
		Method:    "card",
		Amount:    FromMinorUnits("USD", 5000),
		Metadata:  map[string]interface{}{"order_id": "order_123"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := service.CreatePayment(context.Background(), req)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkPaymentService_GetPayment(b *testing.B) {
	repo := &MockRepository{}
	service := NewService(repo, nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := service.GetPayment(context.Background(), "test_id")
		if err != nil {
			b.Fatal(err)
		}
	}
}
