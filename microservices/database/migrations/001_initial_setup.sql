-- =====================================================
-- MIGRATION 001: Initial Database Setup
-- =====================================================
-- Creates all core tables and indexes for the financial platform

-- Enable required extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- Run core schemas
\i core_schemas.sql
\i financial_schemas.sql
\i compliance_schemas.sql
\i infrastructure_schemas.sql

-- Insert default data
INSERT INTO payment_providers (name, methods, currencies, enabled, config) VALUES
('stripe', ARRAY['card', 'bank_transfer'], ARRAY['USD', 'EUR', 'GBP'], true, '{"api_version": "2023-10-16"}'),
('paypal', ARRAY['card', 'paypal'], ARRAY['USD', 'EUR'], true, '{"environment": "sandbox"}'),
('mpesa', ARRAY['mobile_money'], ARRAY['KES', 'USD'], true, '{"shortcode": "174379"}');

-- Insert default fee schedules
INSERT INTO fee_schedules (name, description, fee_type, calculation_method, percentage_rate, currency, enabled) VALUES
('Standard Transaction Fee', 'Default transaction processing fee', 'transaction', 'percentage', 2.9, 'USD', true),
('International Processing Fee', 'Fee for international transactions', 'processing', 'percentage', 3.4, 'USD', true),
('Withdrawal Fee', 'Fee for account withdrawals', 'withdrawal', 'fixed', NULL, 'USD', true);

UPDATE fee_schedules SET fixed_amount = 0.30 WHERE name = 'Withdrawal Fee';

-- Insert default organizations and roles
INSERT INTO organizations (name, slug, status) VALUES
('Default Organization', 'default', 'active'),
('Demo Company', 'demo', 'active');

-- Get organization IDs for role creation
DO $$
DECLARE
    default_org_id UUID;
    demo_org_id UUID;
BEGIN
    SELECT id INTO default_org_id FROM organizations WHERE slug = 'default';
    SELECT id INTO demo_org_id FROM organizations WHERE slug = 'demo';
    
    -- Insert default roles
    INSERT INTO roles (organization_id, name, description, permissions) VALUES
    (default_org_id, 'admin', 'Full system access', '{"*": ["*"]}'),
    (default_org_id, 'merchant', 'Merchant access', '{"payments": ["read", "create"], "escrows": ["read", "create"]}'),
    (default_org_id, 'support', 'Support team access', '{"disputes": ["read", "update"], "kyb": ["read", "update"]}'),
    (demo_org_id, 'admin', 'Demo admin access', '{"*": ["*"]}');
END $$;

-- Create default configuration entries
INSERT INTO configurations (service_name, config_key, config_value, config_type, description) VALUES
('payment-service', 'max_transaction_amount', '1000000', 'number', 'Maximum transaction amount in cents'),
('escrow-service', 'default_hold_period', '7', 'number', 'Default escrow hold period in days'),
('fx-service', 'rate_refresh_interval', '300', 'number', 'FX rate refresh interval in seconds'),
('reconciliation-service', 'batch_size', '1000', 'number', 'Reconciliation batch processing size');

-- Create default feature flags
INSERT INTO feature_flags (flag_name, enabled, rollout_percentage) VALUES
('enable_crypto_payments', false, 0),
('enable_advanced_fraud_detection', true, 100),
('enable_real_time_reconciliation', false, 25),
('enable_multi_currency_escrow', true, 100);

-- Create indexes for better performance
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_payments_created_at ON payments(created_at);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_escrows_updated_at ON escrows(updated_at);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_ledger_entries_created_at ON ledger_entries(created_at);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_journal_entries_created_at ON journal_entries(created_at);

-- Add comments for documentation
COMMENT ON TABLE escrows IS 'Core escrow transactions with buyer/seller relationships';
COMMENT ON TABLE payments IS 'Payment transactions across multiple providers';
COMMENT ON TABLE accounts IS 'Financial accounts with balance tracking';
COMMENT ON TABLE ledger_entries IS 'Double-entry bookkeeping ledger entries';
COMMENT ON TABLE journal_entries IS 'High-level journal entries for accounting';
COMMENT ON TABLE fee_schedules IS 'Configurable fee calculation rules';
COMMENT ON TABLE refunds IS 'Payment refund transactions with idempotency';
COMMENT ON TABLE fx_rates IS 'Foreign exchange rates with time validity';
COMMENT ON TABLE transfers IS 'Internal and external money transfers';
COMMENT ON TABLE payouts IS 'Scheduled payouts to external destinations';
COMMENT ON TABLE reserves IS 'Risk-based reserve holdings';
COMMENT ON TABLE reconciliation_batches IS 'Daily reconciliation processing batches';
COMMENT ON TABLE kyb_applications IS 'Know Your Business verification applications';
COMMENT ON TABLE disputes IS 'Transaction disputes and chargebacks';
COMMENT ON TABLE organizations IS 'Multi-tenant organization structure';
COMMENT ON TABLE users IS 'User accounts with role-based access';
COMMENT ON TABLE audit_logs IS 'Immutable audit trail for compliance';
COMMENT ON TABLE idempotency_keys IS 'Request deduplication and replay protection';
COMMENT ON TABLE events IS 'Event sourcing and domain events';
COMMENT ON TABLE webhook_endpoints IS 'Webhook delivery configuration';
COMMENT ON TABLE configurations IS 'Runtime service configuration';
COMMENT ON TABLE feature_flags IS 'Feature rollout and A/B testing';

-- Create views for common queries
CREATE VIEW active_escrows AS
SELECT * FROM escrows 
WHERE status IN ('pending', 'funded', 'delivered') 
AND created_at > NOW() - INTERVAL '90 days';

CREATE VIEW pending_reconciliation AS
SELECT * FROM reconciliation_items 
WHERE match_status = 'unmatched' 
ORDER BY created_at DESC;

CREATE VIEW high_value_transactions AS
SELECT p.*, e.buyer_id, e.seller_id 
FROM payments p
LEFT JOIN escrows e ON p.metadata->>'escrow_id' = e.id::text
WHERE p.amount_value > 10000
ORDER BY p.created_at DESC;

-- Grant permissions (adjust based on your user setup)
-- GRANT SELECT, INSERT, UPDATE ON ALL TABLES IN SCHEMA public TO financial_app_user;
-- GRANT USAGE ON ALL SEQUENCES IN SCHEMA public TO financial_app_user;
