-- =====================================================
-- CORE FINANCIAL PLATFORM DATABASE SCHEMAS
-- =====================================================
-- PostgreSQL schemas for core financial services

-- Enable required extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- =====================================================
-- ESCROW SERVICE SCHEMA
-- =====================================================

CREATE TABLE escrows (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    buyer_id VARCHAR(255) NOT NULL,
    seller_id VARCHAR(255) NOT NULL,
    amount_value DECIMAL(20,8) NOT NULL,
    amount_currency VARCHAR(3) NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    terms TEXT,
    hold_id VARCHAR(255),
    external_ref VARCHAR(255),
    metadata JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    CONSTRAINT escrows_status_check CHECK (status IN ('pending', 'funded', 'delivered', 'released', 'cancelled', 'disputed')),
    CONSTRAINT escrows_amount_positive CHECK (amount_value > 0)
);

CREATE INDEX idx_escrows_buyer_id ON escrows(buyer_id);
CREATE INDEX idx_escrows_seller_id ON escrows(seller_id);
CREATE INDEX idx_escrows_status ON escrows(status);
CREATE INDEX idx_escrows_created_at ON escrows(created_at);

-- =====================================================
-- PAYMENT SERVICE SCHEMA
-- =====================================================

CREATE TABLE payments (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    account_id VARCHAR(255) NOT NULL,
    provider VARCHAR(100) NOT NULL,
    method VARCHAR(100) NOT NULL,
    amount_value DECIMAL(20,8) NOT NULL,
    amount_currency VARCHAR(3) NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    external_ref VARCHAR(255),
    provider_ref VARCHAR(255),
    metadata JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    CONSTRAINT payments_status_check CHECK (status IN ('pending', 'processing', 'completed', 'failed', 'cancelled', 'refunded')),
    CONSTRAINT payments_amount_positive CHECK (amount_value > 0)
);

CREATE TABLE payment_providers (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(100) UNIQUE NOT NULL,
    methods TEXT[] NOT NULL,
    currencies TEXT[] NOT NULL,
    enabled BOOLEAN DEFAULT true,
    config JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_payments_account_id ON payments(account_id);
CREATE INDEX idx_payments_provider ON payments(provider);
CREATE INDEX idx_payments_status ON payments(status);

-- =====================================================
-- LEDGER SERVICE SCHEMA
-- =====================================================

CREATE TABLE accounts (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    account_number VARCHAR(50) UNIQUE NOT NULL,
    account_type VARCHAR(50) NOT NULL,
    owner_id VARCHAR(255) NOT NULL,
    currency VARCHAR(3) NOT NULL,
    balance_value DECIMAL(20,8) NOT NULL DEFAULT 0,
    available_balance_value DECIMAL(20,8) NOT NULL DEFAULT 0,
    status VARCHAR(50) NOT NULL DEFAULT 'active',
    metadata JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    CONSTRAINT accounts_type_check CHECK (account_type IN ('user', 'merchant', 'escrow', 'fee', 'reserve')),
    CONSTRAINT accounts_status_check CHECK (status IN ('active', 'suspended', 'closed'))
);

CREATE TABLE ledger_entries (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    transaction_id UUID NOT NULL,
    account_id UUID NOT NULL REFERENCES accounts(id),
    entry_type VARCHAR(20) NOT NULL,
    amount_value DECIMAL(20,8) NOT NULL,
    amount_currency VARCHAR(3) NOT NULL,
    balance_after DECIMAL(20,8) NOT NULL,
    description TEXT,
    reference_id VARCHAR(255),
    metadata JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    CONSTRAINT ledger_entries_type_check CHECK (entry_type IN ('debit', 'credit')),
    CONSTRAINT ledger_entries_amount_not_zero CHECK (amount_value != 0)
);

CREATE INDEX idx_accounts_owner_id ON accounts(owner_id);
CREATE INDEX idx_ledger_entries_account_id ON ledger_entries(account_id);
CREATE INDEX idx_ledger_entries_transaction_id ON ledger_entries(transaction_id);

-- =====================================================
-- JOURNAL SERVICE SCHEMA
-- =====================================================

CREATE TABLE journal_entries (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    transaction_id UUID NOT NULL,
    entry_date DATE NOT NULL,
    description TEXT NOT NULL,
    reference VARCHAR(255),
    total_amount DECIMAL(20,8) NOT NULL,
    currency VARCHAR(3) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'posted',
    metadata JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    CONSTRAINT journal_entries_status_check CHECK (status IN ('draft', 'posted', 'reversed'))
);

CREATE TABLE journal_line_items (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    journal_entry_id UUID NOT NULL REFERENCES journal_entries(id),
    account_code VARCHAR(50) NOT NULL,
    account_name VARCHAR(255) NOT NULL,
    debit_amount DECIMAL(20,8) DEFAULT 0,
    credit_amount DECIMAL(20,8) DEFAULT 0,
    currency VARCHAR(3) NOT NULL,
    description TEXT,
    
    CONSTRAINT journal_line_items_amount_check CHECK (
        (debit_amount > 0 AND credit_amount = 0) OR 
        (credit_amount > 0 AND debit_amount = 0)
    )
);

CREATE INDEX idx_journal_entries_transaction_id ON journal_entries(transaction_id);
CREATE INDEX idx_journal_line_items_journal_entry_id ON journal_line_items(journal_entry_id);
