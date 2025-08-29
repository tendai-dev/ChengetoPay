package database

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/lib/pq"
)

// PostgresConfig holds PostgreSQL configuration
type PostgresConfig struct {
	URL                string
	MaxOpenConnections int
	MaxIdleConnections int
	ConnMaxLifetime    time.Duration
}

// PostgresDB represents the PostgreSQL database connection
type PostgresDB struct {
	db *sql.DB
}

// NewPostgresDB creates a new PostgreSQL connection
func NewPostgresDB(config PostgresConfig) (*PostgresDB, error) {
	db, err := sql.Open("postgres", config.URL)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(config.MaxOpenConnections)
	db.SetMaxIdleConns(config.MaxIdleConnections)
	db.SetConnMaxLifetime(config.ConnMaxLifetime)

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	log.Println("✅ PostgreSQL connected successfully")

	return &PostgresDB{db: db}, nil
}

// Close closes the database connection
func (p *PostgresDB) Close() error {
	return p.db.Close()
}

// GetDB returns the underlying sql.DB
func (p *PostgresDB) GetDB() *sql.DB {
	return p.db
}

// HealthCheck performs a health check on the database
func (p *PostgresDB) HealthCheck(ctx context.Context) error {
	return p.db.PingContext(ctx)
}

// RunMigrations runs database migrations
func (p *PostgresDB) RunMigrations() error {
	migrations := []string{
		// Users and Organizations
		`CREATE TABLE IF NOT EXISTS organizations (
			id VARCHAR(255) PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			domain VARCHAR(255) UNIQUE,
			status VARCHAR(50) DEFAULT 'active',
			plan VARCHAR(50) DEFAULT 'basic',
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		
		`CREATE TABLE IF NOT EXISTS users (
			id VARCHAR(255) PRIMARY KEY,
			org_id VARCHAR(255) REFERENCES organizations(id),
			email VARCHAR(255) UNIQUE NOT NULL,
			first_name VARCHAR(255),
			last_name VARCHAR(255),
			password_hash VARCHAR(255),
			status VARCHAR(50) DEFAULT 'active',
			last_login_at TIMESTAMP,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		
		`CREATE TABLE IF NOT EXISTS roles (
			id VARCHAR(255) PRIMARY KEY,
			org_id VARCHAR(255) REFERENCES organizations(id),
			name VARCHAR(255) NOT NULL,
			description TEXT,
			permissions JSONB,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		
		`CREATE TABLE IF NOT EXISTS user_roles (
			id VARCHAR(255) PRIMARY KEY,
			user_id VARCHAR(255) REFERENCES users(id),
			role_id VARCHAR(255) REFERENCES roles(id),
			org_id VARCHAR(255) REFERENCES organizations(id),
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		
		// API Keys and Tokens
		`CREATE TABLE IF NOT EXISTS api_keys (
			id VARCHAR(255) PRIMARY KEY,
			org_id VARCHAR(255) REFERENCES organizations(id),
			user_id VARCHAR(255) REFERENCES users(id),
			name VARCHAR(255) NOT NULL,
			key_hash VARCHAR(255) UNIQUE NOT NULL,
			scopes JSONB,
			status VARCHAR(50) DEFAULT 'active',
			expires_at TIMESTAMP,
			last_used TIMESTAMP,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		
		`CREATE TABLE IF NOT EXISTS tokens (
			id VARCHAR(255) PRIMARY KEY,
			org_id VARCHAR(255) REFERENCES organizations(id),
			user_id VARCHAR(255) REFERENCES users(id),
			token_type VARCHAR(50) NOT NULL,
			token_hash VARCHAR(255) UNIQUE NOT NULL,
			scopes JSONB,
			expires_at TIMESTAMP NOT NULL,
			refresh_token VARCHAR(255),
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		
		// Financial Data
		`CREATE TABLE IF NOT EXISTS accounts (
			id VARCHAR(255) PRIMARY KEY,
			org_id VARCHAR(255) REFERENCES organizations(id),
			user_id VARCHAR(255) REFERENCES users(id),
			account_type VARCHAR(50) NOT NULL,
			currency VARCHAR(3) DEFAULT 'USD',
			available_balance DECIMAL(20,8) DEFAULT 0,
			reserved_balance DECIMAL(20,8) DEFAULT 0,
			total_balance DECIMAL(20,8) DEFAULT 0,
			status VARCHAR(50) DEFAULT 'active',
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		
		`CREATE TABLE IF NOT EXISTS payments (
			id VARCHAR(255) PRIMARY KEY,
			org_id VARCHAR(255) REFERENCES organizations(id),
			account_id VARCHAR(255) REFERENCES accounts(id),
			amount DECIMAL(20,8) NOT NULL,
			currency VARCHAR(3) DEFAULT 'USD',
			status VARCHAR(50) DEFAULT 'pending',
			provider VARCHAR(100),
			provider_payment_id VARCHAR(255),
			metadata JSONB,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		
		`CREATE TABLE IF NOT EXISTS escrows (
			id VARCHAR(255) PRIMARY KEY,
			org_id VARCHAR(255) REFERENCES organizations(id),
			buyer_id VARCHAR(255) REFERENCES users(id),
			seller_id VARCHAR(255) REFERENCES users(id),
			amount DECIMAL(20,8) NOT NULL,
			currency VARCHAR(3) DEFAULT 'USD',
			status VARCHAR(50) DEFAULT 'pending',
			description TEXT,
			metadata JSONB,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		
		`CREATE TABLE IF NOT EXISTS journal_entries (
			id VARCHAR(255) PRIMARY KEY,
			journal_id VARCHAR(255) NOT NULL,
			account_id VARCHAR(255) REFERENCES accounts(id),
			entry_type VARCHAR(50) NOT NULL,
			amount DECIMAL(20,8) NOT NULL,
			currency VARCHAR(3) DEFAULT 'USD',
			description TEXT,
			metadata JSONB,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		
		// Idempotency
		`CREATE TABLE IF NOT EXISTS idempotency_keys (
			id VARCHAR(255) PRIMARY KEY,
			key_hash VARCHAR(255) UNIQUE NOT NULL,
			org_id VARCHAR(255) REFERENCES organizations(id),
			user_id VARCHAR(255) REFERENCES users(id),
			request_hash VARCHAR(255),
			response_data JSONB,
			status VARCHAR(50) DEFAULT 'pending',
			expires_at TIMESTAMP NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		
		// Events and Webhooks
		`CREATE TABLE IF NOT EXISTS events (
			id VARCHAR(255) PRIMARY KEY,
			event_type VARCHAR(255) NOT NULL,
			org_id VARCHAR(255) REFERENCES organizations(id),
			user_id VARCHAR(255) REFERENCES users(id),
			event_data JSONB NOT NULL,
			status VARCHAR(50) DEFAULT 'pending',
			published_at TIMESTAMP,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		
		`CREATE TABLE IF NOT EXISTS webhooks (
			id VARCHAR(255) PRIMARY KEY,
			org_id VARCHAR(255) REFERENCES organizations(id),
			url VARCHAR(500) NOT NULL,
			events JSONB,
			secret VARCHAR(255),
			status VARCHAR(50) DEFAULT 'active',
			last_delivery_at TIMESTAMP,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		
		`CREATE TABLE IF NOT EXISTS webhook_deliveries (
			id VARCHAR(255) PRIMARY KEY,
			webhook_id VARCHAR(255) REFERENCES webhooks(id),
			event_id VARCHAR(255) REFERENCES events(id),
			status VARCHAR(50) DEFAULT 'pending',
			attempt_count INTEGER DEFAULT 0,
			last_attempt_at TIMESTAMP,
			next_attempt_at TIMESTAMP,
			response_code INTEGER,
			response_body TEXT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		
		// Configuration and Feature Flags
		`CREATE TABLE IF NOT EXISTS feature_flags (
			id VARCHAR(255) PRIMARY KEY,
			name VARCHAR(255) UNIQUE NOT NULL,
			description TEXT,
			enabled BOOLEAN DEFAULT false,
			rollout_percentage INTEGER DEFAULT 0,
			org_ids JSONB,
			user_ids JSONB,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		
		`CREATE TABLE IF NOT EXISTS config_values (
			id VARCHAR(255) PRIMARY KEY,
			key VARCHAR(255) UNIQUE NOT NULL,
			value JSONB NOT NULL,
			org_id VARCHAR(255) REFERENCES organizations(id),
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
	}

	for i, migration := range migrations {
		log.Printf("Running migration %d/%d", i+1, len(migrations))
		if _, err := p.db.Exec(migration); err != nil {
			return fmt.Errorf("migration %d failed: %w", i+1, err)
		}
	}

	log.Println("✅ Database migrations completed successfully")
	return nil
}

// CreateIndexes creates database indexes for performance
func (p *PostgresDB) CreateIndexes() error {
	indexes := []string{
		"CREATE INDEX IF NOT EXISTS idx_users_email ON users(email)",
		"CREATE INDEX IF NOT EXISTS idx_users_org_id ON users(org_id)",
		"CREATE INDEX IF NOT EXISTS idx_payments_org_id ON payments(org_id)",
		"CREATE INDEX IF NOT EXISTS idx_payments_status ON payments(status)",
		"CREATE INDEX IF NOT EXISTS idx_escrows_org_id ON escrows(org_id)",
		"CREATE INDEX IF NOT EXISTS idx_escrows_status ON escrows(status)",
		"CREATE INDEX IF NOT EXISTS idx_journal_entries_account_id ON journal_entries(account_id)",
		"CREATE INDEX IF NOT EXISTS idx_journal_entries_created_at ON journal_entries(created_at)",
		"CREATE INDEX IF NOT EXISTS idx_idempotency_keys_hash ON idempotency_keys(key_hash)",
		"CREATE INDEX IF NOT EXISTS idx_idempotency_keys_expires ON idempotency_keys(expires_at)",
		"CREATE INDEX IF NOT EXISTS idx_events_org_id ON events(org_id)",
		"CREATE INDEX IF NOT EXISTS idx_events_type ON events(event_type)",
		"CREATE INDEX IF NOT EXISTS idx_events_created_at ON events(created_at)",
		"CREATE INDEX IF NOT EXISTS idx_webhook_deliveries_status ON webhook_deliveries(status)",
		"CREATE INDEX IF NOT EXISTS idx_webhook_deliveries_next_attempt ON webhook_deliveries(next_attempt_at)",
	}

	for i, index := range indexes {
		log.Printf("Creating index %d/%d", i+1, len(indexes))
		if _, err := p.db.Exec(index); err != nil {
			return fmt.Errorf("index %d failed: %w", i+1, err)
		}
	}

	log.Println("✅ Database indexes created successfully")
	return nil
}
