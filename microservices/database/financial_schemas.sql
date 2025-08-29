-- =====================================================
-- FINANCIAL SERVICES DATABASE SCHEMAS
-- =====================================================
-- Fees, Refunds, FX, Transfers, Payouts, Reserves

-- =====================================================
-- FEES SERVICE SCHEMA
-- =====================================================

CREATE TABLE fee_schedules (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    fee_type VARCHAR(50) NOT NULL,
    calculation_method VARCHAR(50) NOT NULL,
    percentage_rate DECIMAL(8,6),
    fixed_amount DECIMAL(20,8),
    min_amount DECIMAL(20,8),
    max_amount DECIMAL(20,8),
    currency VARCHAR(3) NOT NULL,
    enabled BOOLEAN DEFAULT true,
    effective_from TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    effective_until TIMESTAMP WITH TIME ZONE,
    metadata JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    CONSTRAINT fee_schedules_type_check CHECK (fee_type IN ('transaction', 'processing', 'withdrawal', 'currency_conversion')),
    CONSTRAINT fee_schedules_method_check CHECK (calculation_method IN ('percentage', 'fixed', 'tiered'))
);

CREATE TABLE fee_calculations (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    transaction_id UUID NOT NULL,
    fee_schedule_id UUID NOT NULL REFERENCES fee_schedules(id),
    base_amount DECIMAL(20,8) NOT NULL,
    calculated_fee DECIMAL(20,8) NOT NULL,
    currency VARCHAR(3) NOT NULL,
    calculation_details JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- =====================================================
-- REFUNDS SERVICE SCHEMA
-- =====================================================

CREATE TABLE refunds (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    original_payment_id UUID NOT NULL,
    refund_amount DECIMAL(20,8) NOT NULL,
    currency VARCHAR(3) NOT NULL,
    reason VARCHAR(255),
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    external_ref VARCHAR(255),
    idempotency_key VARCHAR(255) UNIQUE,
    metadata JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    processed_at TIMESTAMP WITH TIME ZONE,
    
    CONSTRAINT refunds_status_check CHECK (status IN ('pending', 'processing', 'completed', 'failed', 'cancelled')),
    CONSTRAINT refunds_amount_positive CHECK (refund_amount > 0)
);

-- =====================================================
-- FX SERVICE SCHEMA
-- =====================================================

CREATE TABLE fx_rates (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    base_currency VARCHAR(3) NOT NULL,
    target_currency VARCHAR(3) NOT NULL,
    rate DECIMAL(12,8) NOT NULL,
    provider VARCHAR(100) NOT NULL,
    valid_from TIMESTAMP WITH TIME ZONE NOT NULL,
    valid_until TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    CONSTRAINT fx_rates_rate_positive CHECK (rate > 0),
    CONSTRAINT fx_rates_currencies_different CHECK (base_currency != target_currency)
);

CREATE TABLE fx_conversions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    transaction_id UUID NOT NULL,
    from_currency VARCHAR(3) NOT NULL,
    to_currency VARCHAR(3) NOT NULL,
    from_amount DECIMAL(20,8) NOT NULL,
    to_amount DECIMAL(20,8) NOT NULL,
    exchange_rate DECIMAL(12,8) NOT NULL,
    rate_timestamp TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- =====================================================
-- TRANSFERS SERVICE SCHEMA
-- =====================================================

CREATE TABLE transfers (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    from_account_id UUID NOT NULL,
    to_account_id UUID NOT NULL,
    amount_value DECIMAL(20,8) NOT NULL,
    currency VARCHAR(3) NOT NULL,
    transfer_type VARCHAR(50) NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    reference VARCHAR(255),
    description TEXT,
    metadata JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    completed_at TIMESTAMP WITH TIME ZONE,
    
    CONSTRAINT transfers_type_check CHECK (transfer_type IN ('internal', 'external', 'split_payment', 'payout')),
    CONSTRAINT transfers_status_check CHECK (status IN ('pending', 'processing', 'completed', 'failed', 'cancelled')),
    CONSTRAINT transfers_amount_positive CHECK (amount_value > 0),
    CONSTRAINT transfers_different_accounts CHECK (from_account_id != to_account_id)
);

CREATE TABLE split_payments (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    parent_transfer_id UUID NOT NULL REFERENCES transfers(id),
    recipient_account_id UUID NOT NULL,
    amount_value DECIMAL(20,8) NOT NULL,
    percentage DECIMAL(5,4),
    currency VARCHAR(3) NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- =====================================================
-- PAYOUTS SERVICE SCHEMA
-- =====================================================

CREATE TABLE payouts (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    account_id UUID NOT NULL,
    amount_value DECIMAL(20,8) NOT NULL,
    currency VARCHAR(3) NOT NULL,
    destination_type VARCHAR(50) NOT NULL,
    destination_details JSONB NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    scheduled_for TIMESTAMP WITH TIME ZONE,
    external_ref VARCHAR(255),
    metadata JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    processed_at TIMESTAMP WITH TIME ZONE,
    
    CONSTRAINT payouts_destination_check CHECK (destination_type IN ('bank_account', 'card', 'mobile_money', 'crypto_wallet')),
    CONSTRAINT payouts_status_check CHECK (status IN ('pending', 'scheduled', 'processing', 'completed', 'failed', 'cancelled')),
    CONSTRAINT payouts_amount_positive CHECK (amount_value > 0)
);

-- =====================================================
-- RESERVES SERVICE SCHEMA
-- =====================================================

CREATE TABLE reserves (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    account_id UUID NOT NULL,
    reserve_type VARCHAR(50) NOT NULL,
    amount_value DECIMAL(20,8) NOT NULL,
    currency VARCHAR(3) NOT NULL,
    percentage DECIMAL(5,4),
    status VARCHAR(50) NOT NULL DEFAULT 'active',
    effective_from TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    effective_until TIMESTAMP WITH TIME ZONE,
    metadata JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    CONSTRAINT reserves_type_check CHECK (reserve_type IN ('rolling', 'fixed', 'risk_based')),
    CONSTRAINT reserves_status_check CHECK (status IN ('active', 'released', 'expired')),
    CONSTRAINT reserves_amount_positive CHECK (amount_value >= 0)
);

CREATE TABLE reserve_transactions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    reserve_id UUID NOT NULL REFERENCES reserves(id),
    transaction_type VARCHAR(50) NOT NULL,
    amount_value DECIMAL(20,8) NOT NULL,
    currency VARCHAR(3) NOT NULL,
    reference_id UUID,
    description TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    CONSTRAINT reserve_transactions_type_check CHECK (transaction_type IN ('hold', 'release', 'adjustment'))
);

-- Indexes for performance
CREATE INDEX idx_fee_schedules_fee_type ON fee_schedules(fee_type);
CREATE INDEX idx_fee_calculations_transaction_id ON fee_calculations(transaction_id);
CREATE INDEX idx_refunds_original_payment_id ON refunds(original_payment_id);
CREATE INDEX idx_refunds_status ON refunds(status);
CREATE INDEX idx_fx_rates_currencies ON fx_rates(base_currency, target_currency);
CREATE INDEX idx_fx_conversions_transaction_id ON fx_conversions(transaction_id);
CREATE INDEX idx_transfers_from_account ON transfers(from_account_id);
CREATE INDEX idx_transfers_to_account ON transfers(to_account_id);
CREATE INDEX idx_transfers_status ON transfers(status);
CREATE INDEX idx_payouts_account_id ON payouts(account_id);
CREATE INDEX idx_payouts_status ON payouts(status);
CREATE INDEX idx_reserves_account_id ON reserves(account_id);
CREATE INDEX idx_reserve_transactions_reserve_id ON reserve_transactions(reserve_id);
