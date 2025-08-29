package main

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"time"
)

// CreateAccountWithValidation creates an account with comprehensive validation
func (s *Service) CreateAccountWithValidation(ctx context.Context, req *CreateAccountRequest) (*Account, error) {
	// Validate request
	validator := NewLedgerValidator()
	if err := validator.ValidateCreateAccountRequest(req); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Check if account already exists
	existingAccount, err := s.repo.GetAccount(ctx, req.AccountID)
	if err == nil && existingAccount != nil {
		return nil, fmt.Errorf("account %s already exists", req.AccountID)
	}

	// Create account with business logic
	account := &Account{
		ID:       req.AccountID,
		Type:     req.Type,
		Currency: req.Currency,
		Balance:  Money{Value: big.NewFloat(0), Currency: req.Currency},
		Status:   "active",
		Metadata: req.Metadata,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Add creation metadata
	if account.Metadata == nil {
		account.Metadata = make(map[string]interface{})
	}
	account.Metadata["created_by"] = "ledger-service"
	account.Metadata["version"] = "1.0"
	account.Metadata["account_class"] = s.getAccountClass(req.Type)

	// Create account in repository
	if err := s.repo.CreateAccount(ctx, account); err != nil {
		return nil, fmt.Errorf("failed to create account: %w", err)
	}

	return account, nil
}

// CreateEntryWithValidation creates a ledger entry with comprehensive validation
func (s *Service) CreateEntryWithValidation(ctx context.Context, req *CreateEntryRequest) (*Entry, error) {
	// Validate request
	validator := NewLedgerValidator()
	if err := validator.ValidateCreateEntryRequest(req); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Get account and validate it exists
	account, err := s.repo.GetAccount(ctx, req.AccountID)
	if err != nil {
		return nil, fmt.Errorf("failed to get account: %w", err)
	}
	if account == nil {
		return nil, fmt.Errorf("account %s not found", req.AccountID)
	}

	// Validate account is active
	if account.Status != "active" {
		return nil, fmt.Errorf("account %s is not active", req.AccountID)
	}

	// Create entry
	entry := &Entry{
		ID:          generateID(),
		AccountID:   req.AccountID,
		Amount:      req.Amount,
		Type:        req.Type,
		Description: req.Description,
		Reference:   req.Reference,
		Metadata:    req.Metadata,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// Add entry metadata
	if entry.Metadata == nil {
		entry.Metadata = make(map[string]interface{})
	}
	entry.Metadata["created_by"] = "ledger-service"
	entry.Metadata["account_type"] = account.Type
	entry.Metadata["currency"] = req.Amount.Currency

	// Validate business rules
	businessValidator := NewBusinessRuleValidator()
	if err := businessValidator.ValidateAccountBalance(ctx, account, entry); err != nil {
		return nil, fmt.Errorf("balance validation failed: %w", err)
	}

	// Create entry in repository
	if err := s.repo.CreateEntry(ctx, entry); err != nil {
		return nil, fmt.Errorf("failed to create entry: %w", err)
	}

	// Update account balance
	newBalance := big.NewFloat(0)
	newBalance.Add(account.Balance.Value, entry.Amount.Value)
	account.Balance.Value = newBalance
	account.UpdatedAt = time.Now()

	if err := s.repo.UpdateAccount(ctx, account); err != nil {
		return nil, fmt.Errorf("failed to update account balance: %w", err)
	}

	return entry, nil
}

// TransferFunds transfers funds between accounts with validation
func (s *Service) TransferFunds(ctx context.Context, fromAccountID, toAccountID string, amount Money, description string) error {
	// Get both accounts
	fromAccount, err := s.repo.GetAccount(ctx, fromAccountID)
	if err != nil {
		return fmt.Errorf("failed to get from account: %w", err)
	}
	if fromAccount == nil {
		return fmt.Errorf("from account %s not found", fromAccountID)
	}

	toAccount, err := s.repo.GetAccount(ctx, toAccountID)
	if err != nil {
		return fmt.Errorf("failed to get to account: %w", err)
	}
	if toAccount == nil {
		return fmt.Errorf("to account %s not found", toAccountID)
	}

	// Validate both accounts are active
	if fromAccount.Status != "active" {
		return fmt.Errorf("from account %s is not active", fromAccountID)
	}
	if toAccount.Status != "active" {
		return fmt.Errorf("to account %s is not active", toAccountID)
	}

	// Validate transfer limits
	businessValidator := NewBusinessRuleValidator()
	if err := businessValidator.ValidateTransferLimits(ctx, fromAccount, toAccount, amount); err != nil {
		return fmt.Errorf("transfer limits validation failed: %w", err)
	}

	// Create transfer reference
	transferRef := generateID()

	// Create debit entry for from account
	debitEntry := &Entry{
		ID:          generateID(),
		AccountID:   fromAccountID,
		Amount:      Money{Value: big.NewFloat(0).Neg(amount.Value), Currency: amount.Currency},
		Type:        "transfer",
		Description: fmt.Sprintf("Transfer to %s: %s", toAccountID, description),
		Reference:   transferRef,
		Metadata: map[string]interface{}{
			"transfer_type": "debit",
			"counterparty":  toAccountID,
			"created_by":    "ledger-service",
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Create credit entry for to account
	creditEntry := &Entry{
		ID:          generateID(),
		AccountID:   toAccountID,
		Amount:      amount,
		Type:        "transfer",
		Description: fmt.Sprintf("Transfer from %s: %s", fromAccountID, description),
		Reference:   transferRef,
		Metadata: map[string]interface{}{
			"transfer_type": "credit",
			"counterparty":  fromAccountID,
			"created_by":    "ledger-service",
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Validate double-entry
	entries := []*Entry{debitEntry, creditEntry}
	if err := businessValidator.ValidateDoubleEntry(entries); err != nil {
		return fmt.Errorf("double-entry validation failed: %w", err)
	}

	// Validate account balances after transfer
	if err := businessValidator.ValidateAccountBalance(ctx, fromAccount, debitEntry); err != nil {
		return fmt.Errorf("from account balance validation failed: %w", err)
	}

	if err := businessValidator.ValidateAccountBalance(ctx, toAccount, creditEntry); err != nil {
		return fmt.Errorf("to account balance validation failed: %w", err)
	}

	// Create both entries
	if err := s.repo.CreateEntry(ctx, debitEntry); err != nil {
		return fmt.Errorf("failed to create debit entry: %w", err)
	}

	if err := s.repo.CreateEntry(ctx, creditEntry); err != nil {
		return fmt.Errorf("failed to create credit entry: %w", err)
	}

	// Update account balances
	fromAccount.Balance.Value.Add(fromAccount.Balance.Value, debitEntry.Amount.Value)
	fromAccount.UpdatedAt = time.Now()

	toAccount.Balance.Value.Add(toAccount.Balance.Value, creditEntry.Amount.Value)
	toAccount.UpdatedAt = time.Now()

	if err := s.repo.UpdateAccount(ctx, fromAccount); err != nil {
		return fmt.Errorf("failed to update from account: %w", err)
	}

	if err := s.repo.UpdateAccount(ctx, toAccount); err != nil {
		return fmt.Errorf("failed to update to account: %w", err)
	}

	return nil
}

// CreateJournalEntry creates a complete journal entry with multiple legs
func (s *Service) CreateJournalEntry(ctx context.Context, entries []*CreateEntryRequest, description string) ([]string, error) {
	if len(entries) == 0 {
		return nil, errors.New("journal entry must have at least one entry")
	}

	// Validate all entries first
	validator := NewLedgerValidator()
	for i, req := range entries {
		if err := validator.ValidateCreateEntryRequest(req); err != nil {
			return nil, fmt.Errorf("entry %d validation failed: %w", i, err)
		}
	}

	// Create entry objects
	var ledgerEntries []*Entry
	var entryIDs []string
	journalRef := generateID()

	for _, req := range entries {
		entry := &Entry{
			ID:          generateID(),
			AccountID:   req.AccountID,
			Amount:      req.Amount,
			Type:        req.Type,
			Description: fmt.Sprintf("%s: %s", description, req.Description),
			Reference:   journalRef,
			Metadata:    req.Metadata,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		if entry.Metadata == nil {
			entry.Metadata = make(map[string]interface{})
		}
		entry.Metadata["journal_entry"] = true
		entry.Metadata["created_by"] = "ledger-service"

		ledgerEntries = append(ledgerEntries, entry)
		entryIDs = append(entryIDs, entry.ID)
	}

	// Validate double-entry bookkeeping
	businessValidator := NewBusinessRuleValidator()
	if err := businessValidator.ValidateJournalEntry(ctx, ledgerEntries); err != nil {
		return nil, fmt.Errorf("journal entry validation failed: %w", err)
	}

	// Validate account balances for each entry
	for i, entry := range ledgerEntries {
		account, err := s.repo.GetAccount(ctx, entry.AccountID)
		if err != nil {
			return nil, fmt.Errorf("failed to get account %s: %w", entry.AccountID, err)
		}
		if account == nil {
			return nil, fmt.Errorf("account %s not found", entry.AccountID)
		}

		if err := businessValidator.ValidateAccountBalance(ctx, account, entry); err != nil {
			return nil, fmt.Errorf("entry %d balance validation failed: %w", i, err)
		}
	}

	// Create all entries
	for _, entry := range ledgerEntries {
		if err := s.repo.CreateEntry(ctx, entry); err != nil {
			return nil, fmt.Errorf("failed to create entry %s: %w", entry.ID, err)
		}
	}

	// Update all account balances
	accountUpdates := make(map[string]*big.Float)
	for _, entry := range ledgerEntries {
		if accountUpdates[entry.AccountID] == nil {
			accountUpdates[entry.AccountID] = big.NewFloat(0)
		}
		accountUpdates[entry.AccountID].Add(accountUpdates[entry.AccountID], entry.Amount.Value)
	}

	for accountID, balanceChange := range accountUpdates {
		account, err := s.repo.GetAccount(ctx, accountID)
		if err != nil {
			return nil, fmt.Errorf("failed to get account %s for balance update: %w", accountID, err)
		}

		account.Balance.Value.Add(account.Balance.Value, balanceChange)
		account.UpdatedAt = time.Now()

		if err := s.repo.UpdateAccount(ctx, account); err != nil {
			return nil, fmt.Errorf("failed to update account %s balance: %w", accountID, err)
		}
	}

	return entryIDs, nil
}

// ReconcileAccount reconciles an account balance
func (s *Service) ReconcileAccount(ctx context.Context, accountID string, expectedBalance Money) error {
	account, err := s.repo.GetAccount(ctx, accountID)
	if err != nil {
		return fmt.Errorf("failed to get account: %w", err)
	}
	if account == nil {
		return fmt.Errorf("account %s not found", accountID)
	}

	// Validate reconciliation
	businessValidator := NewBusinessRuleValidator()
	if err := businessValidator.ValidateReconciliation(ctx, account, expectedBalance); err != nil {
		return fmt.Errorf("reconciliation validation failed: %w", err)
	}

	// Update account metadata with reconciliation info
	if account.Metadata == nil {
		account.Metadata = make(map[string]interface{})
	}
	account.Metadata["last_reconciled"] = time.Now()
	account.Metadata["reconciliation_status"] = "success"
	account.UpdatedAt = time.Now()

	if err := s.repo.UpdateAccount(ctx, account); err != nil {
		return fmt.Errorf("failed to update account: %w", err)
	}

	return nil
}

// CloseAccount closes an account with validation
func (s *Service) CloseAccount(ctx context.Context, accountID string, reason string) error {
	account, err := s.repo.GetAccount(ctx, accountID)
	if err != nil {
		return fmt.Errorf("failed to get account: %w", err)
	}
	if account == nil {
		return fmt.Errorf("account %s not found", accountID)
	}

	// Validate account closure
	businessValidator := NewBusinessRuleValidator()
	if err := businessValidator.ValidateAccountClosure(ctx, account); err != nil {
		return fmt.Errorf("account closure validation failed: %w", err)
	}

	// Update account status
	account.Status = "closed"
	account.UpdatedAt = time.Now()

	if account.Metadata == nil {
		account.Metadata = make(map[string]interface{})
	}
	account.Metadata["closed_at"] = time.Now()
	account.Metadata["closure_reason"] = reason
	account.Metadata["closed_by"] = "ledger-service"

	if err := s.repo.UpdateAccount(ctx, account); err != nil {
		return fmt.Errorf("failed to close account: %w", err)
	}

	return nil
}

// GetAccountBalance gets current account balance with validation
func (s *Service) GetAccountBalance(ctx context.Context, accountID string) (*Money, error) {
	account, err := s.repo.GetAccount(ctx, accountID)
	if err != nil {
		return nil, fmt.Errorf("failed to get account: %w", err)
	}
	if account == nil {
		return nil, fmt.Errorf("account %s not found", accountID)
	}

	return &account.Balance, nil
}

// GetAccountStatement generates account statement
func (s *Service) GetAccountStatement(ctx context.Context, accountID string, fromDate, toDate time.Time) (*AccountStatement, error) {
	account, err := s.repo.GetAccount(ctx, accountID)
	if err != nil {
		return nil, fmt.Errorf("failed to get account: %w", err)
	}
	if account == nil {
		return nil, fmt.Errorf("account %s not found", accountID)
	}

	// Get entries for the period
	filters := EntryFilters{
		AccountID: accountID,
		FromDate:  &fromDate,
		ToDate:    &toDate,
		Limit:     1000, // Reasonable limit
	}

	entries, err := s.repo.ListEntries(ctx, filters)
	if err != nil {
		return nil, fmt.Errorf("failed to get entries: %w", err)
	}

	// Calculate running balance
	runningBalance := big.NewFloat(0)
	var statementEntries []StatementEntry

	for _, entry := range entries {
		runningBalance.Add(runningBalance, entry.Amount.Value)
		
		statementEntries = append(statementEntries, StatementEntry{
			Date:        entry.CreatedAt,
			Description: entry.Description,
			Reference:   entry.Reference,
			Amount:      entry.Amount,
			Balance:     Money{Value: big.NewFloat(0).Set(runningBalance), Currency: entry.Amount.Currency},
		})
	}

	statement := &AccountStatement{
		AccountID:    accountID,
		AccountType:  account.Type,
		Currency:     account.Currency,
		FromDate:     fromDate,
		ToDate:       toDate,
		OpenBalance:  account.Balance, // This should be calculated from entries before fromDate
		CloseBalance: account.Balance,
		Entries:      statementEntries,
		GeneratedAt:  time.Now(),
	}

	return statement, nil
}

// getAccountClass returns the accounting class for an account type
func (s *Service) getAccountClass(accountType string) string {
	classes := map[string]string{
		"asset":     "asset",
		"liability": "liability",
		"equity":    "equity",
		"revenue":   "revenue",
		"expense":   "expense",
		"escrow":    "asset",
		"reserve":   "asset",
		"fee":       "revenue",
		"suspense":  "asset",
	}

	if class, exists := classes[accountType]; exists {
		return class
	}
	return "unknown"
}

// AccountStatement represents an account statement
type AccountStatement struct {
	AccountID    string           `json:"account_id"`
	AccountType  string           `json:"account_type"`
	Currency     string           `json:"currency"`
	FromDate     time.Time        `json:"from_date"`
	ToDate       time.Time        `json:"to_date"`
	OpenBalance  Money            `json:"open_balance"`
	CloseBalance Money            `json:"close_balance"`
	Entries      []StatementEntry `json:"entries"`
	GeneratedAt  time.Time        `json:"generated_at"`
}

// StatementEntry represents an entry in an account statement
type StatementEntry struct {
	Date        time.Time `json:"date"`
	Description string    `json:"description"`
	Reference   string    `json:"reference"`
	Amount      Money     `json:"amount"`
	Balance     Money     `json:"balance"`
}
