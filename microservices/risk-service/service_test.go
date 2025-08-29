package main

import (
	"context"
	"fmt"
	"testing"
)

func TestRiskService_CreateRiskProfile(t *testing.T) {
	repo := NewMockRepository()
	service := NewService(repo, nil)

	req := &CreateRiskProfileRequest{
		EntityID:   "user_123",
		EntityType: "user",
		Metadata: map[string]interface{}{
			"source": "test",
		},
	}

	profile, err := service.CreateRiskProfileWithValidation(context.Background(), req)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if profile.EntityID != req.EntityID {
		t.Errorf("Expected entity ID %s, got %s", req.EntityID, profile.EntityID)
	}

	if profile.EntityType != req.EntityType {
		t.Errorf("Expected entity type %s, got %s", req.EntityType, profile.EntityType)
	}

	if profile.RiskLevel != "low" {
		t.Errorf("Expected initial risk level 'low', got %s", profile.RiskLevel)
	}
}

func TestRiskService_UpdateRiskProfile(t *testing.T) {
	repo := NewMockRepository()
	service := NewService(repo, nil)

	// Create initial profile
	createReq := &CreateRiskProfileRequest{
		EntityID:   "user_456",
		EntityType: "user",
	}

	_, err := service.CreateRiskProfileWithValidation(context.Background(), createReq)
	if err != nil {
		t.Fatalf("Failed to create initial profile: %v", err)
	}

	// Update profile
	newScore := 0.5
	updateReq := &UpdateRiskProfileRequest{
		EntityID:  "user_456",
		RiskScore: &newScore,
		Factors: map[string]interface{}{
			"updated": true,
		},
	}

	updatedProfile, err := service.UpdateRiskProfileWithValidation(context.Background(), updateReq)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if updatedProfile.RiskScore != newScore {
		t.Errorf("Expected risk score %f, got %f", newScore, updatedProfile.RiskScore)
	}

	if updatedProfile.RiskLevel != "medium" {
		t.Errorf("Expected risk level 'medium', got %s", updatedProfile.RiskLevel)
	}
}

func TestRiskService_AssessRisk(t *testing.T) {
	repo := NewMockRepository()
	service := NewService(repo, nil)

	req := &AssessRiskRequest{
		EntityID:   "transaction_789",
		EntityType: "transaction",
		Context: map[string]interface{}{
			"amount":               1000.0,
			"country":             "US",
			"transaction_count_24h": 3.0,
		},
		Amount: FromMinorUnits("USD", 100000), // $1000
		PaymentMethod: "credit_card",
	}

	assessment, err := service.AssessRiskWithValidation(context.Background(), req)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if assessment.EntityID != req.EntityID {
		t.Errorf("Expected entity ID %s, got %s", req.EntityID, assessment.EntityID)
	}

	if assessment.RiskScore < 0 || assessment.RiskScore > 1 {
		t.Errorf("Risk score should be between 0 and 1, got %f", assessment.RiskScore)
	}

	if assessment.Confidence < 0.5 {
		t.Errorf("Confidence should be at least 0.5, got %f", assessment.Confidence)
	}

	validDecisions := []string{"allow", "review", "block"}
	validDecision := false
	for _, valid := range validDecisions {
		if assessment.Decision == valid {
			validDecision = true
			break
		}
	}
	if !validDecision {
		t.Errorf("Invalid decision: %s", assessment.Decision)
	}
}

func TestRiskService_CreateRiskRule(t *testing.T) {
	repo := NewMockRepository()
	service := NewService(repo, nil)

	req := &CreateRiskRuleRequest{
		Name:        "High Amount Rule",
		Description: "Flag transactions over $5000",
		RuleType:    "threshold",
		Conditions: map[string]interface{}{
			"amount_threshold": 5000.0,
			"currency":        "USD",
		},
		Actions:  []string{"review", "flag"},
		Priority: 10,
	}

	rule, err := service.CreateRiskRuleWithValidation(context.Background(), req)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if rule.Name != req.Name {
		t.Errorf("Expected name %s, got %s", req.Name, rule.Name)
	}

	if rule.RuleType != req.RuleType {
		t.Errorf("Expected rule type %s, got %s", req.RuleType, rule.RuleType)
	}

	if !rule.IsActive {
		t.Error("Expected rule to be active")
	}
}

func TestRiskService_GetRiskProfileWithHistory(t *testing.T) {
	repo := NewMockRepository()
	service := NewService(repo, nil)

	// Create profile
	createReq := &CreateRiskProfileRequest{
		EntityID:   "merchant_123",
		EntityType: "merchant",
	}

	_, err := service.CreateRiskProfileWithValidation(context.Background(), createReq)
	if err != nil {
		t.Fatalf("Failed to create profile: %v", err)
	}

	// Perform assessment to create history
	assessReq := &AssessRiskRequest{
		EntityID:   "merchant_123",
		EntityType: "merchant",
		Context: map[string]interface{}{
			"test": "data",
		},
	}

	_, err = service.AssessRiskWithValidation(context.Background(), assessReq)
	if err != nil {
		t.Fatalf("Failed to assess risk: %v", err)
	}

	// Get profile with history
	profile, assessments, err := service.GetRiskProfileWithHistory(context.Background(), "merchant_123")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if profile == nil {
		t.Fatal("Expected profile, got nil")
	}

	if len(assessments) == 0 {
		t.Error("Expected at least one assessment in history")
	}
}

func TestRiskService_BulkAssessRisk(t *testing.T) {
	repo := NewMockRepository()
	service := NewService(repo, nil)

	requests := []*AssessRiskRequest{
		{
			EntityID:   "bulk_1",
			EntityType: "user",
			Context: map[string]interface{}{
				"test": "bulk1",
			},
		},
		{
			EntityID:   "bulk_2",
			EntityType: "user",
			Context: map[string]interface{}{
				"test": "bulk2",
			},
		},
	}

	assessments, err := service.BulkAssessRisk(context.Background(), requests)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(assessments) != len(requests) {
		t.Errorf("Expected %d assessments, got %d", len(requests), len(assessments))
	}

	for i, assessment := range assessments {
		if assessment.EntityID != requests[i].EntityID {
			t.Errorf("Assessment %d: expected entity ID %s, got %s", 
				i, requests[i].EntityID, assessment.EntityID)
		}
	}
}

func TestRiskService_ValidationErrors(t *testing.T) {
	repo := NewMockRepository()
	service := NewService(repo, nil)

	// Test invalid entity ID
	req := &CreateRiskProfileRequest{
		EntityID:   "", // Invalid empty ID
		EntityType: "user",
	}

	_, err := service.CreateRiskProfileWithValidation(context.Background(), req)
	if err == nil {
		t.Error("Expected validation error for empty entity ID")
	}

	// Test invalid entity type
	req2 := &CreateRiskProfileRequest{
		EntityID:   "valid_id",
		EntityType: "invalid_type",
	}

	_, err = service.CreateRiskProfileWithValidation(context.Background(), req2)
	if err == nil {
		t.Error("Expected validation error for invalid entity type")
	}

	// Test invalid risk score
	updateReq := &UpdateRiskProfileRequest{
		EntityID:  "valid_id",
		RiskScore: func() *float64 { s := 1.5; return &s }(), // Invalid score > 1
	}

	_, err = service.UpdateRiskProfileWithValidation(context.Background(), updateReq)
	if err == nil {
		t.Error("Expected validation error for invalid risk score")
	}
}

func TestRiskService_BusinessRuleValidation(t *testing.T) {
	repo := NewMockRepository()
	service := NewService(repo, nil)

	// Create profile
	createReq := &CreateRiskProfileRequest{
		EntityID:   "business_rule_test",
		EntityType: "user",
	}

	_, err := service.CreateRiskProfileWithValidation(context.Background(), createReq)
	if err != nil {
		t.Fatalf("Failed to create profile: %v", err)
	}

	// Test inconsistent risk score and level
	highScore := 0.9
	updateReq := &UpdateRiskProfileRequest{
		EntityID:  "business_rule_test",
		RiskScore: &highScore,
		RiskLevel: func() *string { s := "low"; return &s }(), // Inconsistent with high score
	}

	_, err = service.UpdateRiskProfileWithValidation(context.Background(), updateReq)
	if err == nil {
		t.Error("Expected business rule validation error for inconsistent score/level")
	}
}

func TestRiskService_RiskCalculations(t *testing.T) {
	repo := NewMockRepository()
	service := NewService(repo, nil)

	// Test risk level calculation
	testCases := []struct {
		score    float64
		expected string
	}{
		{0.1, "low"},
		{0.5, "medium"},
		{0.8, "high"},
		{0.95, "critical"},
	}

	for _, tc := range testCases {
		level := service.calculateRiskLevel(tc.score)
		if level != tc.expected {
			t.Errorf("Score %f: expected level %s, got %s", tc.score, tc.expected, level)
		}
	}

	// Test decision calculation
	decisionCases := []struct {
		score    float64
		expected string
	}{
		{0.2, "allow"},
		{0.5, "review"},
		{0.8, "block"},
	}

	for _, tc := range decisionCases {
		decision := service.calculateDecision(tc.score)
		if decision != tc.expected {
			t.Errorf("Score %f: expected decision %s, got %s", tc.score, tc.expected, decision)
		}
	}
}

func TestRiskService_AmountRiskCalculation(t *testing.T) {
	repo := NewMockRepository()
	service := NewService(repo, nil)

	testCases := []struct {
		amount   *Money
		expected float64
	}{
		{FromMinorUnits("USD", 50000), 0.3},   // $500 - medium
		{FromMinorUnits("USD", 600000), 0.6},  // $6000 - high
		{FromMinorUnits("USD", 1500000), 0.9}, // $15000 - very high
	}

	for _, tc := range testCases {
		risk := service.calculateAmountRisk(tc.amount)
		if risk != tc.expected {
			amountFloat, _ := tc.amount.Value.Float64()
			t.Errorf("Amount $%.2f: expected risk %f, got %f", amountFloat, tc.expected, risk)
		}
	}
}

func TestRiskService_PaymentMethodRisk(t *testing.T) {
	repo := NewMockRepository()
	service := NewService(repo, nil)

	testCases := []struct {
		method   string
		expected float64
	}{
		{"credit_card", 0.2},
		{"bank_transfer", 0.05},
		{"crypto", 0.8},
		{"unknown_method", 0.5},
	}

	for _, tc := range testCases {
		risk := service.calculatePaymentMethodRisk(tc.method)
		if risk != tc.expected {
			t.Errorf("Payment method %s: expected risk %f, got %f", tc.method, tc.expected, risk)
		}
	}
}

// Benchmark tests
func BenchmarkRiskService_AssessRisk(b *testing.B) {
	repo := NewMockRepository()
	service := NewService(repo, nil)

	req := &AssessRiskRequest{
		EntityID:   "benchmark_test",
		EntityType: "transaction",
		Context: map[string]interface{}{
			"amount": 1000.0,
			"country": "US",
		},
		Amount: FromMinorUnits("USD", 100000),
		PaymentMethod: "credit_card",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := service.AssessRiskWithValidation(context.Background(), req)
		if err != nil {
			b.Fatalf("Benchmark failed: %v", err)
		}
	}
}

func BenchmarkRiskService_BulkAssessRisk(b *testing.B) {
	repo := NewMockRepository()
	service := NewService(repo, nil)

	// Create 10 requests for bulk assessment
	requests := make([]*AssessRiskRequest, 10)
	for i := 0; i < 10; i++ {
		requests[i] = &AssessRiskRequest{
			EntityID:   fmt.Sprintf("bulk_bench_%d", i),
			EntityType: "user",
			Context: map[string]interface{}{
				"test": i,
			},
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := service.BulkAssessRisk(context.Background(), requests)
		if err != nil {
			b.Fatalf("Bulk benchmark failed: %v", err)
		}
	}
}
