package main

import (
	"context"
	"testing"
	"time"
)

func TestLedgerService_CreateAccount(t *testing.T) {
	repo := NewMockRepository()
	service := NewService(repo, nil)

	req := &CreateAccountRequest{
		AccountID: "acc_test_123",
		Type:      "asset",
		Currency:  "USD",
		Metadata:  map[string]interface{}{"purpose": "testing"},
	}

	account, err := service.CreateAccountWithValidation(context.Background(), req)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if account.ID != req.AccountID {
		t.Errorf("Expected account ID %s, got %s", req.AccountID, account.ID)
	}

	if account.Type != req.Type {
		t.Errorf("Expected account type %s, got %s", req.Type, account.Type)
	}

	if account.Currency != req.Currency {
		t.Errorf("Expected currency %s, got %s", req.Currency, account.Currency)
	}

	if account.Status != "active" {
		t.Errorf("Expected status 'active', got %s", account.Status)
	}
}

func TestLedgerService_CreateEntry(t *testing.T) {
	repo := NewMockRepository()
	service := NewService(repo, nil)

	// First create an account
	accountReq := &CreateAccountRequest{
		AccountID: "acc_test_456",
		Type:      "asset",
		Currency:  "USD",
	}

	_, err := service.CreateAccountWithValidation(context.Background(), accountReq)
	if err != nil {
		t.Fatalf("Failed to create account: %v", err)
	}

	// Create entry
	entryReq := &CreateEntryRequest{
		AccountID:   "acc_test_456",
		Type:        "debit",
		Amount:      FromMinorUnits("USD", 10000), // $100
		Description: "Test debit entry",
		Reference:   "ref_123",
	}

	entry, err := service.CreateEntryWithValidation(context.Background(), entryReq)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if entry.AccountID != entryReq.AccountID {
		t.Errorf("Expected account ID %s, got %s", entryReq.AccountID, entry.AccountID)
	}

	if entry.Type != entryReq.Type {
		t.Errorf("Expected type %s, got %s", entryReq.Type, entry.Type)
	}

	if entry.Description != entryReq.Description {
		t.Errorf("Expected description %s, got %s", entryReq.Description, entry.Description)
	}
}

func TestLedgerService_TransferFunds(t *testing.T) {
	repo := NewMockRepository()
	service := NewService(repo, nil)

	// Create source account
	fromReq := &CreateAccountRequest{
		AccountID: "acc_from_789",
		Type:      "asset",
		Currency:  "USD",
	}
	_, err := service.CreateAccountWithValidation(context.Background(), fromReq)
	if err != nil {
		t.Fatalf("Failed to create from account: %v", err)
	}

	// Create destination account
	toReq := &CreateAccountRequest{
		AccountID: "acc_to_101",
		Type:      "asset",
		Currency:  "USD",
	}
	_, err = service.CreateAccountWithValidation(context.Background(), toReq)
	if err != nil {
		t.Fatalf("Failed to create to account: %v", err)
	}

	// Add initial balance to from account
	initialEntry := &CreateEntryRequest{
		AccountID:   "acc_from_789",
		Type:        "credit",
		Amount:      FromMinorUnits("USD", 50000), // $500
		Description: "Initial balance",
	}
	_, err = service.CreateEntryWithValidation(context.Background(), initialEntry)
	if err != nil {
		t.Fatalf("Failed to create initial entry: %v", err)
	}

	// Transfer funds
	transferAmount := FromMinorUnits("USD", 20000) // $200
	err = service.TransferFunds(context.Background(), "acc_from_789", "acc_to_101", transferAmount, "Test transfer")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
}

func TestLedgerService_GetAccountBalance(t *testing.T) {
	repo := NewMockRepository()
	service := NewService(repo, nil)

	// Create account
	req := &CreateAccountRequest{
		AccountID: "acc_balance_test",
		Type:      "asset",
		Currency:  "USD",
	}
	_, err := service.CreateAccountWithValidation(context.Background(), req)
	if err != nil {
		t.Fatalf("Failed to create account: %v", err)
	}

	// Get balance
	balance, err := service.GetAccountBalance(context.Background(), "acc_balance_test")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if balance.Currency != "USD" {
		t.Errorf("Expected currency USD, got %s", balance.Currency)
	}

	// Balance should be zero for new account
	expectedZero := FromMinorUnits("USD", 0)
	if balance.Value.Cmp(expectedZero.Value) != 0 {
		t.Errorf("Expected zero balance, got %v", balance.Value)
	}
}

func TestLedgerService_CreateJournalEntry(t *testing.T) {
	repo := NewMockRepository()
	service := NewService(repo, nil)

	// Create asset account
	_, err := service.CreateAccount(context.Background(), &CreateAccountRequest{
		AccountID: "asset_account",
		Type:      "asset",
		Currency:  "USD",
	})
	if err != nil {
		t.Fatalf("Failed to create asset account: %v", err)
	}

	// Create revenue account
	_, err = service.CreateAccount(context.Background(), &CreateAccountRequest{
		AccountID: "revenue_account", 
		Type:      "revenue",
		Currency:  "USD",
	})
	if err != nil {
		t.Fatalf("Failed to create revenue account: %v", err)
	}

	// Get account IDs from the mock repository
	accounts, _ := repo.ListAccounts(context.Background(), AccountFilters{})
	assetAccountID := accounts[0].ID
	revenueAccountID := accounts[1].ID

	// Create journal entry (debit asset account, credit revenue account)
	entries := []*CreateEntryRequest{
		{
			AccountID:   assetAccountID,
			Type:        "debit",
			Amount:      FromMinorUnits("USD", 10000), // $100
			Description: "Debit entry",
		},
		{
			AccountID:   revenueAccountID,
			Type:        "credit",
			Amount:      FromMinorUnits("USD", -10000), // -$100 (credit should be negative to balance)
			Description: "Credit entry",
		},
	}

	entryIDs, err := service.CreateJournalEntry(context.Background(), entries, "Test journal entry")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(entryIDs) != 2 {
		t.Errorf("Expected 2 entry IDs, got %d", len(entryIDs))
	}
}

func TestLedgerService_ReconcileAccount(t *testing.T) {
	repo := NewMockRepository()
	service := NewService(repo, nil)

	// Create account
	req := &CreateAccountRequest{
		AccountID: "acc_reconcile_test",
		Type:      "asset",
		Currency:  "USD",
	}
	_, err := service.CreateAccountWithValidation(context.Background(), req)
	if err != nil {
		t.Fatalf("Failed to create account: %v", err)
	}

	// Reconcile with zero balance (should succeed for new account)
	expectedBalance := FromMinorUnits("USD", 0)
	err = service.ReconcileAccount(context.Background(), "acc_reconcile_test", expectedBalance)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
}

func TestLedgerService_CloseAccount(t *testing.T) {
	repo := NewMockRepository()
	service := NewService(repo, nil)

	// Create account with old timestamp to pass age validation
	req := &CreateAccountRequest{
		AccountID: "acc_close_test",
		Type:      "asset",
		Currency:  "USD",
	}
	account, err := service.CreateAccountWithValidation(context.Background(), req)
	if err != nil {
		t.Fatalf("Failed to create account: %v", err)
	}

	// Manually set created date to be old enough
	account.CreatedAt = time.Now().Add(-31 * 24 * time.Hour)
	repo.UpdateAccount(context.Background(), account)

	// Close account
	err = service.CloseAccount(context.Background(), "acc_close_test", "Test closure")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
}

func TestLedgerService_GetAccountStatement(t *testing.T) {
	repo := NewMockRepository()
	service := NewService(repo, nil)

	// Create account
	req := &CreateAccountRequest{
		AccountID: "acc_statement_test",
		Type:      "asset",
		Currency:  "USD",
	}
	_, err := service.CreateAccountWithValidation(context.Background(), req)
	if err != nil {
		t.Fatalf("Failed to create account: %v", err)
	}

	// Get statement
	fromDate := time.Now().Add(-30 * 24 * time.Hour)
	toDate := time.Now()
	statement, err := service.GetAccountStatement(context.Background(), "acc_statement_test", fromDate, toDate)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if statement.AccountID != "acc_statement_test" {
		t.Errorf("Expected account ID acc_statement_test, got %s", statement.AccountID)
	}

	if statement.Currency != "USD" {
		t.Errorf("Expected currency USD, got %s", statement.Currency)
	}
}

// Test validation errors
func TestLedgerService_ValidationErrors(t *testing.T) {
	repo := NewMockRepository()
	service := NewService(repo, nil)

	// Test invalid account type
	invalidReq := &CreateAccountRequest{
		AccountID: "acc_invalid",
		Type:      "invalid_type",
		Currency:  "USD",
	}

	_, err := service.CreateAccountWithValidation(context.Background(), invalidReq)
	if err == nil {
		t.Error("Expected validation error for invalid account type")
	}

	// Test invalid currency
	invalidCurrencyReq := &CreateAccountRequest{
		AccountID: "acc_invalid_currency",
		Type:      "asset",
		Currency:  "INVALID",
	}

	_, err = service.CreateAccountWithValidation(context.Background(), invalidCurrencyReq)
	if err == nil {
		t.Error("Expected validation error for invalid currency")
	}

	// Test empty account ID
	emptyIDReq := &CreateAccountRequest{
		AccountID: "",
		Type:      "asset",
		Currency:  "USD",
	}

	_, err = service.CreateAccountWithValidation(context.Background(), emptyIDReq)
	if err == nil {
		t.Error("Expected validation error for empty account ID")
	}
}

// Benchmark tests
func BenchmarkLedgerService_CreateAccount(b *testing.B) {
	repo := NewMockRepository()
	service := NewService(repo, nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := &CreateAccountRequest{
			AccountID: "acc_benchmark",
			Type:      "asset",
			Currency:  "USD",
		}

		_, err := service.CreateAccountWithValidation(context.Background(), req)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkLedgerService_CreateEntry(b *testing.B) {
	repo := NewMockRepository()
	service := NewService(repo, nil)

	// Create account first
	req := &CreateAccountRequest{
		AccountID: "acc_benchmark_entry",
		Type:      "asset",
		Currency:  "USD",
	}
	_, err := service.CreateAccountWithValidation(context.Background(), req)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		entryReq := &CreateEntryRequest{
			AccountID:   "acc_benchmark_entry",
			Type:        "debit",
			Amount:      FromMinorUnits("USD", 1000),
			Description: "Benchmark entry",
		}

		_, err := service.CreateEntryWithValidation(context.Background(), entryReq)
		if err != nil {
			b.Fatal(err)
		}
	}
}
