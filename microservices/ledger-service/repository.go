package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"math/big"
	"time"

	_ "github.com/lib/pq"
)

// PostgreSQLRepository implements Repository interface with PostgreSQL
type PostgreSQLRepository struct {
	db *sql.DB
}

// NewPostgreSQLRepository creates a new PostgreSQL repository
func NewPostgreSQLRepository(connectionString string) (*PostgreSQLRepository, error) {
	db, err := sql.Open("postgres", connectionString)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &PostgreSQLRepository{db: db}, nil
}

// CreateAccount creates a new ledger account
func (r *PostgreSQLRepository) CreateAccount(ctx context.Context, account *Account) error {
	query := `
		INSERT INTO ledger_accounts (
			id, name, account_type, parent_id, currency, 
			balance_value, balance_currency, status, metadata, 
			created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`

	metadataJSON, err := json.Marshal(account.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	balanceValue, _ := account.Balance.Value.Float64()

	_, err = r.db.ExecContext(ctx, query,
		account.ID,
		account.Name,
		account.Type,
		account.ParentID,
		account.Currency,
		balanceValue,
		account.Balance.Currency,
		account.Status,
		metadataJSON,
		account.CreatedAt,
		account.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create account: %w", err)
	}

	return nil
}

// GetAccount retrieves an account by ID
func (r *PostgreSQLRepository) GetAccount(ctx context.Context, id string) (*Account, error) {
	query := `
		SELECT id, name, account_type, parent_id, currency, 
			   balance_value, balance_currency, status, metadata, 
			   created_at, updated_at
		FROM ledger_accounts 
		WHERE id = $1`

	var account Account
	var balanceValue float64
	var metadataJSON []byte
	var parentID sql.NullString

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&account.ID,
		&account.Name,
		&account.Type,
		&parentID,
		&account.Currency,
		&balanceValue,
		&account.Balance.Currency,
		&account.Status,
		&metadataJSON,
		&account.CreatedAt,
		&account.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("account not found: %s", id)
		}
		return nil, fmt.Errorf("failed to get account: %w", err)
	}

	// Convert balance value to Money
	account.Balance = Money{
		Value:    big.NewFloat(balanceValue),
		Currency: account.Balance.Currency,
	}

	// Handle nullable parent ID
	if parentID.Valid {
		account.ParentID = parentID.String
	}

	// Parse metadata
	if len(metadataJSON) > 0 {
		var metadata map[string]interface{}
		if err := json.Unmarshal(metadataJSON, &metadata); err == nil {
			account.Metadata = metadata
		}
	}

	return &account, nil
}

// ListAccounts retrieves accounts with filters
func (r *PostgreSQLRepository) ListAccounts(ctx context.Context, filters AccountFilters) ([]*Account, error) {
	query := `
		SELECT id, name, account_type, parent_id, currency, 
			   balance_value, balance_currency, status, metadata, 
			   created_at, updated_at
		FROM ledger_accounts 
		WHERE 1=1`

	args := []interface{}{}
	argIndex := 1

	// Apply filters
	if filters.Type != "" {
		query += fmt.Sprintf(" AND account_type = $%d", argIndex)
		args = append(args, filters.Type)
		argIndex++
	}

	if filters.Currency != "" {
		query += fmt.Sprintf(" AND currency = $%d", argIndex)
		args = append(args, filters.Currency)
		argIndex++
	}

	if filters.Status != "" {
		query += fmt.Sprintf(" AND status = $%d", argIndex)
		args = append(args, filters.Status)
		argIndex++
	}

	if filters.ParentID != "" {
		query += fmt.Sprintf(" AND parent_id = $%d", argIndex)
		args = append(args, filters.ParentID)
		argIndex++
	}

	// Add ordering and pagination
	query += " ORDER BY created_at DESC"

	if filters.Limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", argIndex)
		args = append(args, filters.Limit)
		argIndex++
	}

	if filters.Offset > 0 {
		query += fmt.Sprintf(" OFFSET $%d", argIndex)
		args = append(args, filters.Offset)
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list accounts: %w", err)
	}
	defer rows.Close()

	var accounts []*Account
	for rows.Next() {
		var account Account
		var balanceValue float64
		var metadataJSON []byte
		var parentID sql.NullString

		err := rows.Scan(
			&account.ID,
			&account.Name,
			&account.Type,
			&parentID,
			&account.Currency,
			&balanceValue,
			&account.Balance.Currency,
			&account.Status,
			&metadataJSON,
			&account.CreatedAt,
			&account.UpdatedAt,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan account: %w", err)
		}

		// Convert balance value to Money
		account.Balance = Money{
			Value:    big.NewFloat(balanceValue),
			Currency: account.Balance.Currency,
		}

		// Handle nullable parent ID
		if parentID.Valid {
			account.ParentID = parentID.String
		}

		// Parse metadata
		if len(metadataJSON) > 0 {
			var metadata map[string]interface{}
			if err := json.Unmarshal(metadataJSON, &metadata); err == nil {
				account.Metadata = metadata
			}
		}

		accounts = append(accounts, &account)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating accounts: %w", err)
	}

	return accounts, nil
}

// UpdateAccount updates an existing account
func (r *PostgreSQLRepository) UpdateAccount(ctx context.Context, account *Account) error {
	query := `
		UPDATE ledger_accounts 
		SET name = $2, account_type = $3, parent_id = $4, currency = $5,
			balance_value = $6, balance_currency = $7, status = $8, 
			metadata = $9, updated_at = $10
		WHERE id = $1`

	metadataJSON, err := json.Marshal(account.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	balanceValue, _ := account.Balance.Value.Float64()
	account.UpdatedAt = time.Now()

	result, err := r.db.ExecContext(ctx, query,
		account.ID,
		account.Name,
		account.Type,
		account.ParentID,
		account.Currency,
		balanceValue,
		account.Balance.Currency,
		account.Status,
		metadataJSON,
		account.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to update account: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("account not found: %s", account.ID)
	}

	return nil
}

// DeleteAccount deletes an account by ID
func (r *PostgreSQLRepository) DeleteAccount(ctx context.Context, id string) error {
	query := `DELETE FROM ledger_accounts WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete account: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("account not found: %s", id)
	}

	return nil
}

// CreateEntry creates a new ledger entry
func (r *PostgreSQLRepository) CreateEntry(ctx context.Context, entry *Entry) error {
	query := `
		INSERT INTO ledger_entries (
			id, account_id, transaction_id, entry_type, amount_value, 
			amount_currency, description, reference, metadata, 
			created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`

	metadataJSON, err := json.Marshal(entry.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	amountValue, _ := entry.Amount.Value.Float64()

	_, err = r.db.ExecContext(ctx, query,
		entry.ID,
		entry.AccountID,
		entry.TransactionID,
		entry.Type,
		amountValue,
		entry.Amount.Currency,
		entry.Description,
		entry.Reference,
		metadataJSON,
		entry.CreatedAt,
		entry.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create entry: %w", err)
	}

	return nil
}

// GetEntry retrieves an entry by ID
func (r *PostgreSQLRepository) GetEntry(ctx context.Context, id string) (*Entry, error) {
	query := `
		SELECT id, account_id, transaction_id, entry_type, amount_value, 
			   amount_currency, description, reference, metadata, 
			   created_at, updated_at
		FROM ledger_entries 
		WHERE id = $1`

	var entry Entry
	var amountValue float64
	var metadataJSON []byte

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&entry.ID,
		&entry.AccountID,
		&entry.TransactionID,
		&entry.Type,
		&amountValue,
		&entry.Amount.Currency,
		&entry.Description,
		&entry.Reference,
		&metadataJSON,
		&entry.CreatedAt,
		&entry.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("entry not found: %s", id)
		}
		return nil, fmt.Errorf("failed to get entry: %w", err)
	}

	// Convert amount value to Money
	entry.Amount = Money{
		Value:    big.NewFloat(amountValue),
		Currency: entry.Amount.Currency,
	}

	// Parse metadata
	if len(metadataJSON) > 0 {
		var metadata map[string]interface{}
		if err := json.Unmarshal(metadataJSON, &metadata); err == nil {
			entry.Metadata = metadata
		}
	}

	return &entry, nil
}

// ListEntries retrieves entries with filters
func (r *PostgreSQLRepository) ListEntries(ctx context.Context, filters EntryFilters) ([]*Entry, error) {
	query := `
		SELECT id, account_id, transaction_id, entry_type, amount_value, 
			   amount_currency, description, reference, metadata, 
			   created_at, updated_at
		FROM ledger_entries 
		WHERE 1=1`

	args := []interface{}{}
	argIndex := 1

	// Apply filters
	if filters.AccountID != "" {
		query += fmt.Sprintf(" AND account_id = $%d", argIndex)
		args = append(args, filters.AccountID)
		argIndex++
	}

	if filters.TransactionID != "" {
		query += fmt.Sprintf(" AND transaction_id = $%d", argIndex)
		args = append(args, filters.TransactionID)
		argIndex++
	}

	if filters.Type != "" {
		query += fmt.Sprintf(" AND entry_type = $%d", argIndex)
		args = append(args, filters.Type)
		argIndex++
	}

	// Add ordering and pagination
	query += " ORDER BY created_at DESC"

	if filters.Limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", argIndex)
		args = append(args, filters.Limit)
		argIndex++
	}

	if filters.Offset > 0 {
		query += fmt.Sprintf(" OFFSET $%d", argIndex)
		args = append(args, filters.Offset)
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list entries: %w", err)
	}
	defer rows.Close()

	var entries []*Entry
	for rows.Next() {
		var entry Entry
		var amountValue float64
		var metadataJSON []byte

		err := rows.Scan(
			&entry.ID,
			&entry.AccountID,
			&entry.TransactionID,
			&entry.Type,
			&amountValue,
			&entry.Amount.Currency,
			&entry.Description,
			&entry.Reference,
			&metadataJSON,
			&entry.CreatedAt,
			&entry.UpdatedAt,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan entry: %w", err)
		}

		// Convert amount value to Money
		entry.Amount = Money{
			Value:    big.NewFloat(amountValue),
			Currency: entry.Amount.Currency,
		}

		// Parse metadata
		if len(metadataJSON) > 0 {
			var metadata map[string]interface{}
			if err := json.Unmarshal(metadataJSON, &metadata); err == nil {
				entry.Metadata = metadata
			}
		}

		entries = append(entries, &entry)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating entries: %w", err)
	}

	return entries, nil
}

// Close closes the database connection
func (r *PostgreSQLRepository) Close() error {
	return r.db.Close()
}

// Health checks database connectivity
func (r *PostgreSQLRepository) Health(ctx context.Context) error {
	return r.db.PingContext(ctx)
}
