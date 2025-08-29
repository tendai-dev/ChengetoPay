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

// CreateEscrow creates a new escrow in the database
func (r *PostgreSQLRepository) CreateEscrow(ctx context.Context, escrow *Escrow) error {
	query := `
		INSERT INTO escrows (
			id, buyer_id, seller_id, amount_value, amount_currency, 
			status, terms, hold_id, external_ref, metadata, 
			created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`

	metadataJSON, err := json.Marshal(escrow.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	amountValue, _ := escrow.Amount.Value.Float64()

	_, err = r.db.ExecContext(ctx, query,
		escrow.ID,
		escrow.BuyerID,
		escrow.SellerID,
		amountValue,
		escrow.Amount.Currency,
		escrow.Status,
		escrow.Terms,
		escrow.HoldID,
		nil, // external_ref
		metadataJSON,
		escrow.CreatedAt,
		escrow.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create escrow: %w", err)
	}

	return nil
}

// GetEscrow retrieves an escrow by ID
func (r *PostgreSQLRepository) GetEscrow(ctx context.Context, id string) (*Escrow, error) {
	query := `
		SELECT id, buyer_id, seller_id, amount_value, amount_currency,
			   status, terms, hold_id, external_ref, metadata,
			   created_at, updated_at
		FROM escrows 
		WHERE id = $1`

	var escrow Escrow
	var amountValue float64
	var metadataJSON []byte
	var holdID, externalRef sql.NullString

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&escrow.ID,
		&escrow.BuyerID,
		&escrow.SellerID,
		&amountValue,
		&escrow.Currency,
		&escrow.Status,
		&escrow.Terms,
		&holdID,
		&externalRef,
		&metadataJSON,
		&escrow.CreatedAt,
		&escrow.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("escrow not found: %s", id)
		}
		return nil, fmt.Errorf("failed to get escrow: %w", err)
	}

	// Convert amount value to Money
	escrow.Amount = Money{
		Value:    big.NewFloat(amountValue),
		Currency: escrow.Currency,
	}

	// Handle nullable fields
	if holdID.Valid {
		escrow.HoldID = holdID.String
	}

	// Parse metadata
	if len(metadataJSON) > 0 {
		var metadata map[string]interface{}
		if err := json.Unmarshal(metadataJSON, &metadata); err == nil {
			escrow.Metadata = metadata
		}
	}

	return &escrow, nil
}

// ListEscrows retrieves escrows with filters
func (r *PostgreSQLRepository) ListEscrows(ctx context.Context, filters EscrowFilters) ([]*Escrow, error) {
	query := `
		SELECT id, buyer_id, seller_id, amount_value, amount_currency,
			   status, terms, hold_id, external_ref, metadata,
			   created_at, updated_at
		FROM escrows 
		WHERE 1=1`

	args := []interface{}{}
	argIndex := 1

	// Apply filters
	if filters.Status != "" {
		query += fmt.Sprintf(" AND status = $%d", argIndex)
		args = append(args, filters.Status)
		argIndex++
	}

	if filters.BuyerID != "" {
		query += fmt.Sprintf(" AND buyer_id = $%d", argIndex)
		args = append(args, filters.BuyerID)
		argIndex++
	}

	if filters.SellerID != "" {
		query += fmt.Sprintf(" AND seller_id = $%d", argIndex)
		args = append(args, filters.SellerID)
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
		return nil, fmt.Errorf("failed to list escrows: %w", err)
	}
	defer rows.Close()

	var escrows []*Escrow
	for rows.Next() {
		var escrow Escrow
		var amountValue float64
		var metadataJSON []byte
		var holdID, externalRef sql.NullString

		err := rows.Scan(
			&escrow.ID,
			&escrow.BuyerID,
			&escrow.SellerID,
			&amountValue,
			&escrow.Currency,
			&escrow.Status,
			&escrow.Terms,
			&holdID,
			&externalRef,
			&metadataJSON,
			&escrow.CreatedAt,
			&escrow.UpdatedAt,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan escrow: %w", err)
		}

		// Convert amount value to Money
		escrow.Amount = Money{
			Value:    big.NewFloat(amountValue),
			Currency: escrow.Currency,
		}

		// Handle nullable fields
		if holdID.Valid {
			escrow.HoldID = holdID.String
		}

		// Parse metadata
		if len(metadataJSON) > 0 {
			var metadata map[string]interface{}
			if err := json.Unmarshal(metadataJSON, &metadata); err == nil {
				escrow.Metadata = metadata
			}
		}

		escrows = append(escrows, &escrow)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating escrows: %w", err)
	}

	return escrows, nil
}

// UpdateEscrow updates an existing escrow
func (r *PostgreSQLRepository) UpdateEscrow(ctx context.Context, escrow *Escrow) error {
	query := `
		UPDATE escrows 
		SET buyer_id = $2, seller_id = $3, amount_value = $4, amount_currency = $5,
			status = $6, terms = $7, hold_id = $8, external_ref = $9, 
			metadata = $10, updated_at = $11
		WHERE id = $1`

	metadataJSON, err := json.Marshal(escrow.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	amountValue, _ := escrow.Amount.Value.Float64()
	escrow.UpdatedAt = time.Now()

	result, err := r.db.ExecContext(ctx, query,
		escrow.ID,
		escrow.BuyerID,
		escrow.SellerID,
		amountValue,
		escrow.Amount.Currency,
		escrow.Status,
		escrow.Terms,
		escrow.HoldID,
		nil, // external_ref
		metadataJSON,
		escrow.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to update escrow: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("escrow not found: %s", escrow.ID)
	}

	return nil
}

// DeleteEscrow deletes an escrow by ID
func (r *PostgreSQLRepository) DeleteEscrow(ctx context.Context, id string) error {
	query := `DELETE FROM escrows WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete escrow: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("escrow not found: %s", id)
	}

	return nil
}

// Close closes the database connection
func (r *PostgreSQLRepository) Close() error {
	return r.db.Close()
}

// Health checks database connectivity
func (r *PostgreSQLRepository) Health(ctx context.Context) error {
	return r.db.PingContext(ctx)
}
