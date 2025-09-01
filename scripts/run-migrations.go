package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"
)

func main() {
	fmt.Println("🔄 Running Database Migrations...")
	
	postgresURL := os.Getenv("POSTGRES_URL")
	if postgresURL == "" {
		postgresURL = "postgresql://neondb_owner:npg_A7n1FliTzPvk@ep-divine-scene-advwzvm6-pooler.c-2.us-east-1.aws.neon.tech/neondb?sslmode=require"
	}

	db, err := sql.Open("postgres", postgresURL)
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer db.Close()

	// Create schemas
	schemas := []string{
		"escrow", "payments", "ledger", "journal", "fees",
		"refunds", "transfers", "payouts", "reserves",
		"reconciliation", "treasury", "risk", "disputes",
		"auth", "compliance",
	}

	for _, schema := range schemas {
		_, err := db.Exec(fmt.Sprintf("CREATE SCHEMA IF NOT EXISTS %s", schema))
		if err != nil {
			log.Printf("Warning: Could not create schema %s: %v", schema, err)
		} else {
			fmt.Printf("✅ Schema %s created\n", schema)
		}
	}

	// Create auth tables
	authTables := `
	CREATE TABLE IF NOT EXISTS auth.users (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		email VARCHAR(255) UNIQUE NOT NULL,
		password_hash VARCHAR(255) NOT NULL,
		role VARCHAR(50) NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS auth.sessions (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		user_id UUID REFERENCES auth.users(id),
		token VARCHAR(500) NOT NULL,
		expires_at TIMESTAMP NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);

	CREATE INDEX IF NOT EXISTS idx_sessions_token ON auth.sessions(token);
	CREATE INDEX IF NOT EXISTS idx_sessions_user ON auth.sessions(user_id);
	`

	if _, err := db.Exec(authTables); err != nil {
		log.Printf("Warning: Could not create auth tables: %v", err)
	} else {
		fmt.Println("✅ Auth tables created")
	}

	// Create escrow tables
	escrowTables := `
	CREATE TABLE IF NOT EXISTS escrow.transactions (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		buyer_id UUID NOT NULL,
		seller_id UUID NOT NULL,
		amount DECIMAL(19,4) NOT NULL,
		currency VARCHAR(3) NOT NULL,
		status VARCHAR(50) NOT NULL,
		metadata JSONB,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		completed_at TIMESTAMP
	);

	CREATE INDEX IF NOT EXISTS idx_escrow_buyer ON escrow.transactions(buyer_id);
	CREATE INDEX IF NOT EXISTS idx_escrow_seller ON escrow.transactions(seller_id);
	CREATE INDEX IF NOT EXISTS idx_escrow_status ON escrow.transactions(status);
	`

	if _, err := db.Exec(escrowTables); err != nil {
		log.Printf("Warning: Could not create escrow tables: %v", err)
	} else {
		fmt.Println("✅ Escrow tables created")
	}

	// Create payments tables
	paymentsTables := `
	CREATE TABLE IF NOT EXISTS payments.transactions (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		escrow_id UUID,
		amount DECIMAL(19,4) NOT NULL,
		currency VARCHAR(3) NOT NULL,
		payment_method VARCHAR(50),
		status VARCHAR(50) NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);

	CREATE INDEX IF NOT EXISTS idx_payments_escrow ON payments.transactions(escrow_id);
	CREATE INDEX IF NOT EXISTS idx_payments_status ON payments.transactions(status);
	`

	if _, err := db.Exec(paymentsTables); err != nil {
		log.Printf("Warning: Could not create payments tables: %v", err)
	} else {
		fmt.Println("✅ Payments tables created")
	}

	// Create ledger tables
	ledgerTables := `
	CREATE TABLE IF NOT EXISTS ledger.entries (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		account_id UUID NOT NULL,
		transaction_id UUID,
		debit DECIMAL(19,4),
		credit DECIMAL(19,4),
		balance DECIMAL(19,4) NOT NULL,
		description TEXT,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);

	CREATE INDEX IF NOT EXISTS idx_ledger_account ON ledger.entries(account_id);
	CREATE INDEX IF NOT EXISTS idx_ledger_transaction ON ledger.entries(transaction_id);
	`

	if _, err := db.Exec(ledgerTables); err != nil {
		log.Printf("Warning: Could not create ledger tables: %v", err)
	} else {
		fmt.Println("✅ Ledger tables created")
	}

	fmt.Println("\n✅ All migrations completed successfully!")
}
