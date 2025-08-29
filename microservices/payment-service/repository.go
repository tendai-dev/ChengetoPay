package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"math/big"
	"time"

	"github.com/lib/pq"
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

// CreatePayment creates a new payment in the database
func (r *PostgreSQLRepository) CreatePayment(ctx context.Context, payment *Payment) error {
	query := `
		INSERT INTO payments (
			id, account_id, provider, method, amount_value, amount_currency,
			status, external_ref, provider_ref, metadata, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`

	metadataJSON, err := json.Marshal(payment.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	amountValue, _ := payment.Amount.Value.Float64()

	_, err = r.db.ExecContext(ctx, query,
		payment.ID,
		payment.AccountID,
		payment.Provider,
		payment.Method,
		amountValue,
		payment.Amount.Currency,
		payment.Status,
		payment.ExternalRef,
		nil, // provider_ref
		metadataJSON,
		payment.CreatedAt,
		payment.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create payment: %w", err)
	}

	return nil
}

// GetPayment retrieves a payment by ID
func (r *PostgreSQLRepository) GetPayment(ctx context.Context, id string) (*Payment, error) {
	query := `
		SELECT id, account_id, provider, method, amount_value, amount_currency,
			   status, external_ref, provider_ref, metadata, created_at, updated_at
		FROM payments 
		WHERE id = $1`

	var payment Payment
	var amountValue float64
	var metadataJSON []byte
	var externalRef, providerRef sql.NullString

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&payment.ID,
		&payment.AccountID,
		&payment.Provider,
		&payment.Method,
		&amountValue,
		&payment.Currency,
		&payment.Status,
		&externalRef,
		&providerRef,
		&metadataJSON,
		&payment.CreatedAt,
		&payment.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("payment not found: %s", id)
		}
		return nil, fmt.Errorf("failed to get payment: %w", err)
	}

	// Convert amount value to Money
	payment.Amount = Money{
		Value:    big.NewFloat(amountValue),
		Currency: payment.Currency,
	}

	// Handle nullable fields
	if externalRef.Valid {
		payment.ExternalRef = externalRef.String
	}

	// Parse metadata
	if len(metadataJSON) > 0 {
		var metadata map[string]interface{}
		if err := json.Unmarshal(metadataJSON, &metadata); err == nil {
			payment.Metadata = metadata
		}
	}

	return &payment, nil
}

// ListPayments retrieves payments with filters
func (r *PostgreSQLRepository) ListPayments(ctx context.Context, filters PaymentFilters) ([]*Payment, error) {
	query := `
		SELECT id, account_id, provider, method, amount_value, amount_currency,
			   status, external_ref, provider_ref, metadata, created_at, updated_at
		FROM payments 
		WHERE 1=1`

	args := []interface{}{}
	argIndex := 1

	// Apply filters
	if filters.AccountID != "" {
		query += fmt.Sprintf(" AND account_id = $%d", argIndex)
		args = append(args, filters.AccountID)
		argIndex++
	}

	if filters.Provider != "" {
		query += fmt.Sprintf(" AND provider = $%d", argIndex)
		args = append(args, filters.Provider)
		argIndex++
	}

	if filters.Status != "" {
		query += fmt.Sprintf(" AND status = $%d", argIndex)
		args = append(args, filters.Status)
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
		return nil, fmt.Errorf("failed to list payments: %w", err)
	}
	defer rows.Close()

	var payments []*Payment
	for rows.Next() {
		var payment Payment
		var amountValue float64
		var metadataJSON []byte
		var externalRef, providerRef sql.NullString

		err := rows.Scan(
			&payment.ID,
			&payment.AccountID,
			&payment.Provider,
			&payment.Method,
			&amountValue,
			&payment.Currency,
			&payment.Status,
			&externalRef,
			&providerRef,
			&metadataJSON,
			&payment.CreatedAt,
			&payment.UpdatedAt,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan payment: %w", err)
		}

		// Convert amount value to Money
		payment.Amount = Money{
			Value:    big.NewFloat(amountValue),
			Currency: payment.Currency,
		}

		// Handle nullable fields
		if externalRef.Valid {
			payment.ExternalRef = externalRef.String
		}

		// Parse metadata
		if len(metadataJSON) > 0 {
			var metadata map[string]interface{}
			if err := json.Unmarshal(metadataJSON, &metadata); err == nil {
				payment.Metadata = metadata
			}
		}

		payments = append(payments, &payment)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating payments: %w", err)
	}

	return payments, nil
}

// UpdatePayment updates an existing payment
func (r *PostgreSQLRepository) UpdatePayment(ctx context.Context, payment *Payment) error {
	query := `
		UPDATE payments 
		SET account_id = $2, provider = $3, method = $4, amount_value = $5, 
			amount_currency = $6, status = $7, external_ref = $8, provider_ref = $9,
			metadata = $10, updated_at = $11
		WHERE id = $1`

	metadataJSON, err := json.Marshal(payment.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	amountValue, _ := payment.Amount.Value.Float64()
	payment.UpdatedAt = time.Now()

	result, err := r.db.ExecContext(ctx, query,
		payment.ID,
		payment.AccountID,
		payment.Provider,
		payment.Method,
		amountValue,
		payment.Amount.Currency,
		payment.Status,
		payment.ExternalRef,
		nil, // provider_ref
		metadataJSON,
		payment.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to update payment: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("payment not found: %s", payment.ID)
	}

	return nil
}

// DeletePayment deletes a payment by ID
func (r *PostgreSQLRepository) DeletePayment(ctx context.Context, id string) error {
	query := `DELETE FROM payments WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete payment: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("payment not found: %s", id)
	}

	return nil
}

// GetProviders retrieves all payment providers
func (r *PostgreSQLRepository) GetProviders(ctx context.Context) ([]*Provider, error) {
	query := `
		SELECT id, name, methods, currencies, enabled, config, created_at, updated_at
		FROM payment_providers 
		WHERE enabled = true
		ORDER BY name`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get providers: %w", err)
	}
	defer rows.Close()

	var providers []*Provider
	for rows.Next() {
		var provider Provider
		var id string
		var configJSON []byte
		var createdAt, updatedAt time.Time

		err := rows.Scan(
			&id,
			&provider.Name,
			(*pq.StringArray)(&provider.Methods),
			(*pq.StringArray)(&provider.Currencies),
			&provider.Enabled,
			&configJSON,
			&createdAt,
			&updatedAt,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan provider: %w", err)
		}

		providers = append(providers, &provider)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating providers: %w", err)
	}

	return providers, nil
}

// Close closes the database connection
func (r *PostgreSQLRepository) Close() error {
	return r.db.Close()
}

// Health checks database connectivity
func (r *PostgreSQLRepository) Health(ctx context.Context) error {
	return r.db.PingContext(ctx)
}
