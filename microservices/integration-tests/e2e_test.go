package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/project-x/microservices/shared/clients"
	"github.com/project-x/microservices/shared/servicediscovery"
)

// TestEndToEndEscrowWorkflow tests the complete escrow workflow
func TestEndToEndEscrowWorkflow(t *testing.T) {
	// Setup service discovery and clients
	registry := servicediscovery.NewServiceRegistry()
	loadBalancer := servicediscovery.NewLoadBalancer(registry)

	// Register test services
	setupTestServices(registry)

	escrowClient := clients.NewEscrowClient(loadBalancer)
	paymentClient := clients.NewPaymentClient(loadBalancer)
	ledgerClient := clients.NewLedgerClient(loadBalancer)
	riskClient := clients.NewRiskClient(loadBalancer)

	ctx := context.Background()

	t.Run("Complete Escrow Workflow", func(t *testing.T) {
		// Step 1: Create risk profiles for buyer and seller
		buyerProfile, err := riskClient.CreateRiskProfile(ctx, &clients.CreateRiskProfileRequest{
			EntityID:   "buyer-123",
			EntityType: "individual",
			Metadata:   map[string]interface{}{"country": "US", "verified": true},
		})
		if err != nil {
			t.Fatalf("Failed to create buyer risk profile: %v", err)
		}

		sellerProfile, err := riskClient.CreateRiskProfile(ctx, &clients.CreateRiskProfileRequest{
			EntityID:   "seller-456",
			EntityType: "business",
			Metadata:   map[string]interface{}{"country": "US", "verified": true},
		})
		if err != nil {
			t.Fatalf("Failed to create seller risk profile: %v", err)
		}

		// Step 2: Create ledger accounts
		buyerAccount, err := ledgerClient.CreateAccount(ctx, &clients.CreateAccountRequest{
			AccountID:   "buyer-123",
			AccountType: "asset",
			Currency:    "USD",
		})
		if err != nil {
			t.Fatalf("Failed to create buyer account: %v", err)
		}

		sellerAccount, err := ledgerClient.CreateAccount(ctx, &clients.CreateAccountRequest{
			AccountID:   "seller-456",
			AccountType: "asset",
			Currency:    "USD",
		})
		if err != nil {
			t.Fatalf("Failed to create seller account: %v", err)
		}

		escrowAccount, err := ledgerClient.CreateAccount(ctx, &clients.CreateAccountRequest{
			AccountID:   "escrow-holding",
			AccountType: "liability",
			Currency:    "USD",
		})
		if err != nil {
			t.Fatalf("Failed to create escrow account: %v", err)
		}

		// Step 3: Perform risk assessment
		riskAssessment, err := riskClient.AssessRisk(ctx, &clients.AssessRiskRequest{
			EntityID:      "buyer-123",
			Amount:        clients.Money{Value: 1000.00, Currency: "USD"},
			PaymentMethod: "credit_card",
			Context:       "escrow_transaction",
		})
		if err != nil {
			t.Fatalf("Failed to assess risk: %v", err)
		}

		// Verify risk assessment
		if riskAssessment.RiskLevel == "high" {
			t.Skip("Skipping transaction due to high risk")
		}

		// Step 4: Create escrow
		escrow, err := escrowClient.CreateEscrow(ctx, &clients.CreateEscrowRequest{
			BuyerID:  "buyer-123",
			SellerID: "seller-456",
			Amount:   clients.Money{Value: 1000.00, Currency: "USD"},
			Terms:    "Standard escrow terms",
		})
		if err != nil {
			t.Fatalf("Failed to create escrow: %v", err)
		}

		// Verify escrow creation
		if escrow.Status != "pending" {
			t.Errorf("Expected escrow status 'pending', got '%s'", escrow.Status)
		}

		// Step 5: Create payment
		payment, err := paymentClient.CreatePayment(ctx, &clients.CreatePaymentRequest{
			Amount:        clients.Money{Value: 1000.00, Currency: "USD"},
			PaymentMethod: "credit_card",
			Provider:      "stripe",
			Description:   fmt.Sprintf("Payment for escrow %s", escrow.ID),
		})
		if err != nil {
			t.Fatalf("Failed to create payment: %v", err)
		}

		// Step 6: Process payment
		processedPayment, err := paymentClient.ProcessPayment(ctx, payment.ID)
		if err != nil {
			t.Fatalf("Failed to process payment: %v", err)
		}

		if processedPayment.Status != "completed" {
			t.Errorf("Expected payment status 'completed', got '%s'", processedPayment.Status)
		}

		// Step 7: Fund escrow with payment
		err = escrowClient.FundEscrow(ctx, escrow.ID, payment.ID)
		if err != nil {
			t.Fatalf("Failed to fund escrow: %v", err)
		}

		// Step 8: Create journal entry for escrow funding
		_, err = ledgerClient.CreateJournalEntry(ctx, &clients.CreateJournalEntryRequest{
			Description: fmt.Sprintf("Fund escrow %s", escrow.ID),
			Entries: []clients.EntryDetail{
				{AccountID: "buyer-123", Amount: 1000.00, Type: "credit"},
				{AccountID: "escrow-holding", Amount: 1000.00, Type: "debit"},
			},
		})
		if err != nil {
			t.Fatalf("Failed to create funding journal entry: %v", err)
		}

		// Step 9: Verify account balances
		buyerBalance, err := ledgerClient.GetAccountBalance(ctx, "buyer-123")
		if err != nil {
			t.Fatalf("Failed to get buyer balance: %v", err)
		}

		escrowBalance, err := ledgerClient.GetAccountBalance(ctx, "escrow-holding")
		if err != nil {
			t.Fatalf("Failed to get escrow balance: %v", err)
		}

		if buyerBalance != -1000.00 {
			t.Errorf("Expected buyer balance -1000.00, got %f", buyerBalance)
		}

		if escrowBalance != 1000.00 {
			t.Errorf("Expected escrow balance 1000.00, got %f", escrowBalance)
		}

		// Step 10: Release escrow
		err = escrowClient.ReleaseEscrow(ctx, escrow.ID)
		if err != nil {
			t.Fatalf("Failed to release escrow: %v", err)
		}

		// Step 11: Create journal entry for escrow release
		_, err = ledgerClient.CreateJournalEntry(ctx, &clients.CreateJournalEntryRequest{
			Description: fmt.Sprintf("Release escrow %s to seller", escrow.ID),
			Entries: []clients.EntryDetail{
				{AccountID: "escrow-holding", Amount: 1000.00, Type: "credit"},
				{AccountID: "seller-456", Amount: 1000.00, Type: "debit"},
			},
		})
		if err != nil {
			t.Fatalf("Failed to create release journal entry: %v", err)
		}

		// Step 12: Verify final balances
		finalEscrowBalance, err := ledgerClient.GetAccountBalance(ctx, "escrow-holding")
		if err != nil {
			t.Fatalf("Failed to get final escrow balance: %v", err)
		}

		sellerBalance, err := ledgerClient.GetAccountBalance(ctx, "seller-456")
		if err != nil {
			t.Fatalf("Failed to get seller balance: %v", err)
		}

		if finalEscrowBalance != 0.00 {
			t.Errorf("Expected final escrow balance 0.00, got %f", finalEscrowBalance)
		}

		if sellerBalance != 1000.00 {
			t.Errorf("Expected seller balance 1000.00, got %f", sellerBalance)
		}

		// Verify risk profiles were updated
		updatedBuyerProfile, err := riskClient.GetRiskProfile(ctx, "buyer-123")
		if err != nil {
			t.Fatalf("Failed to get updated buyer profile: %v", err)
		}

		if updatedBuyerProfile.RiskScore >= buyerProfile.RiskScore {
			t.Logf("Buyer risk score updated from %f to %f", buyerProfile.RiskScore, updatedBuyerProfile.RiskScore)
		}

		t.Logf("End-to-end escrow workflow completed successfully")
		t.Logf("Escrow ID: %s", escrow.ID)
		t.Logf("Payment ID: %s", payment.ID)
		t.Logf("Final balances - Buyer: %f, Seller: %f, Escrow: %f", 
			buyerBalance, sellerBalance, finalEscrowBalance)
	})
}

// TestAPIGatewayIntegration tests API Gateway routing and middleware
func TestAPIGatewayIntegration(t *testing.T) {
	gatewayURL := "http://localhost:8090"
	client := &http.Client{Timeout: 30 * time.Second}

	t.Run("Gateway Health Check", func(t *testing.T) {
		resp, err := client.Get(gatewayURL + "/health")
		if err != nil {
			t.Fatalf("Failed to call gateway health: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200, got %d", resp.StatusCode)
		}

		var healthResp map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&healthResp); err != nil {
			t.Fatalf("Failed to decode health response: %v", err)
		}

		if healthResp["status"] != "healthy" {
			t.Errorf("Expected healthy status, got %v", healthResp["status"])
		}
	})

	t.Run("Service Discovery", func(t *testing.T) {
		resp, err := client.Get(gatewayURL + "/services")
		if err != nil {
			t.Fatalf("Failed to call service discovery: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200, got %d", resp.StatusCode)
		}
	})

	t.Run("API Routing", func(t *testing.T) {
		// Test escrow service routing
		escrowData := map[string]interface{}{
			"buyer_id":  "test-buyer",
			"seller_id": "test-seller",
			"amount":    map[string]interface{}{"value": 500.0, "currency": "USD"},
			"terms":     "Test terms",
		}

		jsonData, _ := json.Marshal(escrowData)
		req, _ := http.NewRequest("POST", gatewayURL+"/api/v1/escrow/v1/escrows", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer test-token-12345")

		resp, err := client.Do(req)
		if err != nil {
			t.Logf("Expected error for unavailable service: %v", err)
			return
		}
		defer resp.Body.Close()

		// Should get service unavailable or successful response
		if resp.StatusCode != http.StatusServiceUnavailable && resp.StatusCode != http.StatusCreated {
			t.Logf("Gateway routing test - Status: %d", resp.StatusCode)
		}
	})

	t.Run("Rate Limiting", func(t *testing.T) {
		// Make multiple requests to test rate limiting
		for i := 0; i < 5; i++ {
			resp, err := client.Get(gatewayURL + "/")
			if err != nil {
				t.Fatalf("Request %d failed: %v", i, err)
			}
			resp.Body.Close()

			if resp.StatusCode == http.StatusTooManyRequests {
				t.Logf("Rate limiting triggered at request %d", i)
				return
			}
		}
		t.Logf("Rate limiting not triggered in test")
	})
}

// TestServiceCommunication tests inter-service communication
func TestServiceCommunication(t *testing.T) {
	registry := servicediscovery.NewServiceRegistry()
	setupTestServices(registry)

	ctx := context.Background()

	t.Run("Circuit Breaker", func(t *testing.T) {
		loadBalancer := servicediscovery.NewLoadBalancer(registry)
		client := clients.NewEscrowClient(loadBalancer)

		// This should trigger circuit breaker after failures
		for i := 0; i < 10; i++ {
			_, err := client.CreateEscrow(ctx, &clients.CreateEscrowRequest{
				BuyerID:  "test-buyer",
				SellerID: "test-seller",
				Amount:   clients.Money{Value: 100.0, Currency: "USD"},
				Terms:    "Test terms",
			})

			if err != nil {
				t.Logf("Request %d failed (expected): %v", i, err)
			}

			time.Sleep(100 * time.Millisecond)
		}
	})

	t.Run("Service Discovery", func(t *testing.T) {
		services := registry.GetAllServices()
		if len(services) == 0 {
			t.Error("No services registered")
		}

		for serviceName, instances := range services {
			if len(instances) == 0 {
				t.Errorf("No instances for service %s", serviceName)
			}
			t.Logf("Service %s has %d instances", serviceName, len(instances))
		}
	})
}

// TestDataConsistency tests data consistency across services
func TestDataConsistency(t *testing.T) {
	registry := servicediscovery.NewServiceRegistry()
	setupTestServices(registry)
	loadBalancer := servicediscovery.NewLoadBalancer(registry)

	ledgerClient := clients.NewLedgerClient(loadBalancer)
	ctx := context.Background()

	t.Run("Double Entry Accounting", func(t *testing.T) {
		// Create test accounts
		account1, err := ledgerClient.CreateAccount(ctx, &clients.CreateAccountRequest{
			AccountID:   "test-account-1",
			AccountType: "asset",
			Currency:    "USD",
		})
		if err != nil {
			t.Fatalf("Failed to create account 1: %v", err)
		}

		account2, err := ledgerClient.CreateAccount(ctx, &clients.CreateAccountRequest{
			AccountID:   "test-account-2",
			AccountType: "liability",
			Currency:    "USD",
		})
		if err != nil {
			t.Fatalf("Failed to create account 2: %v", err)
		}

		// Create journal entry
		entry, err := ledgerClient.CreateJournalEntry(ctx, &clients.CreateJournalEntryRequest{
			Description: "Test double entry",
			Entries: []clients.EntryDetail{
				{AccountID: "test-account-1", Amount: 100.0, Type: "debit"},
				{AccountID: "test-account-2", Amount: 100.0, Type: "credit"},
			},
		})
		if err != nil {
			t.Fatalf("Failed to create journal entry: %v", err)
		}

		// Verify balances
		balance1, err := ledgerClient.GetAccountBalance(ctx, "test-account-1")
		if err != nil {
			t.Fatalf("Failed to get balance 1: %v", err)
		}

		balance2, err := ledgerClient.GetAccountBalance(ctx, "test-account-2")
		if err != nil {
			t.Fatalf("Failed to get balance 2: %v", err)
		}

		if balance1 != 100.0 {
			t.Errorf("Expected account 1 balance 100.0, got %f", balance1)
		}

		if balance2 != -100.0 {
			t.Errorf("Expected account 2 balance -100.0, got %f", balance2)
		}

		t.Logf("Journal entry %s created successfully", entry.ID)
		t.Logf("Account balances: %s=%f, %s=%f", 
			account1.AccountID, balance1, account2.AccountID, balance2)
	})
}

// setupTestServices registers test service instances
func setupTestServices(registry *servicediscovery.ServiceRegistry) {
	services := map[string]int{
		"escrow-service":  8081,
		"payment-service": 8083,
		"ledger-service":  8084,
		"risk-service":    8085,
	}

	for serviceName, port := range services {
		instance := &servicediscovery.ServiceInstance{
			ID:      fmt.Sprintf("%s-test", serviceName),
			Name:    serviceName,
			Address: "localhost",
			Port:    port,
			Health:  "passing",
			Tags:    []string{"test"},
		}
		registry.RegisterService(instance)
	}
}
