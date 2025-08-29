package main

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"time"
)

// LedgerValidator handles ledger validation logic
type LedgerValidator struct{}

// NewLedgerValidator creates a new ledger validator
func NewLedgerValidator() *LedgerValidator {
	return &LedgerValidator{}
}

// ValidateCreateAccountRequest validates account creation request
func (v *LedgerValidator) ValidateCreateAccountRequest(req *CreateAccountRequest) error {
	if req == nil {
		return errors.New("account request cannot be nil")
	}

	// Validate account ID
	if strings.TrimSpace(req.AccountID) == "" {
		return errors.New("account ID cannot be empty")
	}

	if len(req.AccountID) > 100 {
		return errors.New("account ID cannot exceed 100 characters")
	}

	// Validate account type
	if err := v.ValidateAccountType(req.Type); err != nil {
		return fmt.Errorf("invalid account type: %w", err)
	}

	// Validate currency
	if err := v.ValidateCurrency(req.Currency); err != nil {
		return fmt.Errorf("invalid currency: %w", err)
	}

	// Validate metadata
	if err := v.ValidateMetadata(req.Metadata); err != nil {
		return fmt.Errorf("invalid metadata: %w", err)
	}

	return nil
}

// ValidateCreateEntryRequest validates entry creation request
func (v *LedgerValidator) ValidateCreateEntryRequest(req *CreateEntryRequest) error {
	if req == nil {
		return errors.New("entry request cannot be nil")
	}

	// Validate account ID
	if strings.TrimSpace(req.AccountID) == "" {
		return errors.New("account ID cannot be empty")
	}

	// Validate amount
	if err := v.ValidateAmount(req.Amount); err != nil {
		return fmt.Errorf("invalid amount: %w", err)
	}

	// Validate entry type
	if err := v.ValidateEntryType(req.Type); err != nil {
		return fmt.Errorf("invalid entry type: %w", err)
	}

	// Validate description
	if strings.TrimSpace(req.Description) == "" {
		return errors.New("description cannot be empty")
	}

	if len(req.Description) > 500 {
		return errors.New("description cannot exceed 500 characters")
	}

	// Validate reference
	if req.Reference != "" && len(req.Reference) > 100 {
		return errors.New("reference cannot exceed 100 characters")
	}

	// Validate metadata
	if err := v.ValidateMetadata(req.Metadata); err != nil {
		return fmt.Errorf("invalid metadata: %w", err)
	}

	return nil
}

// ValidateAccountType validates account type
func (v *LedgerValidator) ValidateAccountType(accountType string) error {
	if accountType == "" {
		return errors.New("account type cannot be empty")
	}

	validTypes := map[string]bool{
		"asset":     true,
		"liability": true,
		"equity":    true,
		"revenue":   true,
		"expense":   true,
		"escrow":    true,
		"reserve":   true,
		"fee":       true,
		"suspense":  true,
	}

	if !validTypes[accountType] {
		return fmt.Errorf("unsupported account type: %s", accountType)
	}

	return nil
}

// ValidateEntryType validates entry type
func (v *LedgerValidator) ValidateEntryType(entryType string) error {
	if entryType == "" {
		return errors.New("entry type cannot be empty")
	}

	validTypes := map[string]bool{
		"debit":      true,
		"credit":     true,
		"transfer":   true,
		"adjustment": true,
		"fee":        true,
		"refund":     true,
		"chargeback": true,
		"settlement": true,
	}

	if !validTypes[entryType] {
		return fmt.Errorf("unsupported entry type: %s", entryType)
	}

	return nil
}

// ValidateAmount validates monetary amount
func (v *LedgerValidator) ValidateAmount(amount Money) error {
	if amount.Value == nil {
		return errors.New("amount value cannot be nil")
	}

	// Check if amount is zero (allowed for some operations)
	zero := big.NewFloat(0)
	if amount.Value.Cmp(zero) == 0 {
		return nil // Zero amounts are allowed
	}

	// Check maximum amount (e.g., $10M per entry)
	maxAmount := big.NewFloat(1000000000) // $10M in cents
	absAmount := big.NewFloat(0).Abs(amount.Value)
	if absAmount.Cmp(maxAmount) > 0 {
		return errors.New("amount exceeds maximum limit")
	}

	return nil
}

// ValidateCurrency validates currency code
func (v *LedgerValidator) ValidateCurrency(currency string) error {
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

// ValidateMetadata validates metadata
func (v *LedgerValidator) ValidateMetadata(metadata map[string]interface{}) error {
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

// BusinessRuleValidator handles business rule validation
type BusinessRuleValidator struct {
	validator *LedgerValidator
}

// NewBusinessRuleValidator creates a new business rule validator
func NewBusinessRuleValidator() *BusinessRuleValidator {
	return &BusinessRuleValidator{
		validator: NewLedgerValidator(),
	}
}

// ValidateAccountBalance validates account balance constraints
func (v *BusinessRuleValidator) ValidateAccountBalance(ctx context.Context, account *Account, newEntry *Entry) error {
	if account == nil {
		return errors.New("account cannot be nil")
	}

	if newEntry == nil {
		return errors.New("entry cannot be nil")
	}

	// Calculate new balance
	newBalance := big.NewFloat(0)
	newBalance.Add(account.Balance.Value, newEntry.Amount.Value)

	// Check account type specific rules
	switch account.Type {
	case "asset", "expense":
		// Asset and expense accounts should have positive balances
		if newBalance.Cmp(big.NewFloat(0)) < 0 {
			return fmt.Errorf("asset/expense account cannot have negative balance")
		}

	case "liability", "equity", "revenue":
		// Liability, equity, and revenue accounts typically have negative balances
		// But we allow positive balances for flexibility

	case "escrow":
		// Escrow accounts should never go negative
		if newBalance.Cmp(big.NewFloat(0)) < 0 {
			return fmt.Errorf("escrow account cannot have negative balance")
		}

	case "reserve":
		// Reserve accounts should maintain minimum balance
		minReserve := big.NewFloat(100000) // $1000 minimum
		if newBalance.Cmp(minReserve) < 0 {
			return fmt.Errorf("reserve account cannot go below minimum balance")
		}
	}

	return nil
}

// ValidateDoubleEntry validates double-entry bookkeeping rules
func (v *BusinessRuleValidator) ValidateDoubleEntry(entries []*Entry) error {
	if len(entries) < 2 {
		return errors.New("double-entry requires at least 2 entries")
	}

	// Group entries by currency
	currencyTotals := make(map[string]*big.Float)

	for _, entry := range entries {
		if entry == nil {
			return errors.New("entry cannot be nil")
		}

		currency := entry.Amount.Currency
		if currencyTotals[currency] == nil {
			currencyTotals[currency] = big.NewFloat(0)
		}

		currencyTotals[currency].Add(currencyTotals[currency], entry.Amount.Value)
	}

	// Check that debits equal credits for each currency
	zero := big.NewFloat(0)
	for currency, total := range currencyTotals {
		if total.Cmp(zero) != 0 {
			return fmt.Errorf("double-entry not balanced for currency %s: total = %s", 
				currency, total.String())
		}
	}

	return nil
}

// ValidateTransferLimits validates transfer amount limits
func (v *BusinessRuleValidator) ValidateTransferLimits(ctx context.Context, fromAccount, toAccount *Account, amount Money) error {
	if fromAccount == nil || toAccount == nil {
		return errors.New("both accounts must be provided")
	}

	// Daily transfer limits
	dailyLimit := big.NewFloat(10000000) // $100k daily limit
	if amount.Value.Cmp(dailyLimit) > 0 {
		return errors.New("transfer exceeds daily limit")
	}

	// Cross-currency transfer validation
	if fromAccount.Currency != toAccount.Currency {
		// TODO: Implement FX rate validation
		// For now, just ensure both currencies are supported
		if err := v.validator.ValidateCurrency(fromAccount.Currency); err != nil {
			return fmt.Errorf("invalid from currency: %w", err)
		}
		if err := v.validator.ValidateCurrency(toAccount.Currency); err != nil {
			return fmt.Errorf("invalid to currency: %w", err)
		}
	}

	// Account type transfer restrictions
	restrictedTransfers := map[string][]string{
		"suspense": {"revenue", "expense"}, // Suspense can't transfer to revenue/expense
		"reserve":  {"expense"},            // Reserve can't transfer to expense
	}

	if restricted, exists := restrictedTransfers[fromAccount.Type]; exists {
		for _, restrictedType := range restricted {
			if toAccount.Type == restrictedType {
				return fmt.Errorf("transfer from %s to %s account not allowed", 
					fromAccount.Type, toAccount.Type)
			}
		}
	}

	return nil
}

// ValidateReconciliation validates account reconciliation
func (v *BusinessRuleValidator) ValidateReconciliation(ctx context.Context, account *Account, expectedBalance Money) error {
	if account == nil {
		return errors.New("account cannot be nil")
	}

	// Check currency match
	if account.Currency != expectedBalance.Currency {
		return fmt.Errorf("currency mismatch: account has %s, expected %s", 
			account.Currency, expectedBalance.Currency)
	}

	// Calculate variance
	variance := big.NewFloat(0)
	variance.Sub(account.Balance.Value, expectedBalance.Value)
	variance.Abs(variance)

	// Allow small variance due to rounding (e.g., 1 cent)
	tolerance := big.NewFloat(1)
	if variance.Cmp(tolerance) > 0 {
		return fmt.Errorf("reconciliation failed: variance of %s exceeds tolerance", 
			variance.String())
	}

	return nil
}

// ValidateJournalEntry validates complete journal entry
func (v *BusinessRuleValidator) ValidateJournalEntry(ctx context.Context, entries []*Entry) error {
	if len(entries) == 0 {
		return errors.New("journal entry must have at least one entry")
	}

	// Validate double-entry rules
	if err := v.ValidateDoubleEntry(entries); err != nil {
		return fmt.Errorf("double-entry validation failed: %w", err)
	}

	// Validate each individual entry
	for i, entry := range entries {
		if entry == nil {
			return fmt.Errorf("entry at index %d cannot be nil", i)
		}

		// Validate entry fields
		if err := v.validator.ValidateAmount(entry.Amount); err != nil {
			return fmt.Errorf("entry %d amount validation failed: %w", i, err)
		}

		if err := v.validator.ValidateEntryType(entry.Type); err != nil {
			return fmt.Errorf("entry %d type validation failed: %w", i, err)
		}

		// Validate entry has required reference for certain types
		requiresReference := map[string]bool{
			"transfer":   true,
			"adjustment": true,
			"chargeback": true,
		}

		if requiresReference[entry.Type] && entry.Reference == "" {
			return fmt.Errorf("entry %d of type '%s' requires a reference", i, entry.Type)
		}
	}

	// Validate entry timestamps are reasonable
	now := time.Now()
	for i, entry := range entries {
		// Entries shouldn't be too far in the future
		if entry.CreatedAt.After(now.Add(1 * time.Hour)) {
			return fmt.Errorf("entry %d timestamp too far in future", i)
		}

		// Entries shouldn't be too old (e.g., more than 1 year)
		if entry.CreatedAt.Before(now.Add(-365 * 24 * time.Hour)) {
			return fmt.Errorf("entry %d timestamp too old", i)
		}
	}

	return nil
}

// ValidateAccountClosure validates account closure
func (v *BusinessRuleValidator) ValidateAccountClosure(ctx context.Context, account *Account) error {
	if account == nil {
		return errors.New("account cannot be nil")
	}

	// Check if account has zero balance
	zero := big.NewFloat(0)
	if account.Balance.Value.Cmp(zero) != 0 {
		return fmt.Errorf("account must have zero balance to close, current balance: %s", 
			account.Balance.Value.String())
	}

	// Check account type restrictions
	restrictedTypes := map[string]bool{
		"reserve": true, // Reserve accounts cannot be closed
		"escrow":  true, // Escrow accounts need special handling
	}

	if restrictedTypes[account.Type] {
		return fmt.Errorf("account type '%s' cannot be closed", account.Type)
	}

	// Check account age (must be at least 30 days old)
	minAge := time.Now().Add(-30 * 24 * time.Hour)
	if account.CreatedAt.After(minAge) {
		return errors.New("account must be at least 30 days old to close")
	}

	return nil
}
