-- =====================================================
-- COMPLIANCE & RISK DATABASE SCHEMAS
-- =====================================================
-- Reconciliation, KYB, Disputes, Auth, Observability

-- =====================================================
-- RECONCILIATION SERVICE SCHEMA
-- =====================================================

CREATE TABLE reconciliation_batches (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    batch_date DATE NOT NULL,
    provider VARCHAR(100) NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    total_transactions INTEGER DEFAULT 0,
    matched_transactions INTEGER DEFAULT 0,
    unmatched_transactions INTEGER DEFAULT 0,
    total_amount DECIMAL(20,8) DEFAULT 0,
    currency VARCHAR(3) NOT NULL,
    started_at TIMESTAMP WITH TIME ZONE,
    completed_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    CONSTRAINT reconciliation_status_check CHECK (status IN ('pending', 'processing', 'completed', 'failed'))
);

CREATE TABLE reconciliation_items (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    batch_id UUID NOT NULL REFERENCES reconciliation_batches(id),
    internal_transaction_id UUID,
    external_transaction_id VARCHAR(255),
    amount_value DECIMAL(20,8) NOT NULL,
    currency VARCHAR(3) NOT NULL,
    transaction_date DATE NOT NULL,
    match_status VARCHAR(50) NOT NULL DEFAULT 'unmatched',
    discrepancy_amount DECIMAL(20,8) DEFAULT 0,
    notes TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    CONSTRAINT reconciliation_match_status_check CHECK (match_status IN ('matched', 'unmatched', 'disputed', 'resolved'))
);

-- =====================================================
-- KYB SERVICE SCHEMA
-- =====================================================

CREATE TABLE kyb_applications (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    business_id VARCHAR(255) NOT NULL,
    company_name VARCHAR(255) NOT NULL,
    registration_number VARCHAR(255),
    business_type VARCHAR(100),
    country VARCHAR(3) NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    risk_score INTEGER,
    documents JSONB,
    verification_data JSONB,
    submitted_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    reviewed_at TIMESTAMP WITH TIME ZONE,
    approved_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    CONSTRAINT kyb_status_check CHECK (status IN ('pending', 'under_review', 'approved', 'rejected', 'requires_info')),
    CONSTRAINT kyb_risk_score_range CHECK (risk_score >= 0 AND risk_score <= 100)
);

CREATE TABLE ubo_records (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    kyb_application_id UUID NOT NULL REFERENCES kyb_applications(id),
    full_name VARCHAR(255) NOT NULL,
    ownership_percentage DECIMAL(5,2) NOT NULL,
    nationality VARCHAR(3),
    date_of_birth DATE,
    identification_documents JSONB,
    verification_status VARCHAR(50) NOT NULL DEFAULT 'pending',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    CONSTRAINT ubo_ownership_range CHECK (ownership_percentage >= 0 AND ownership_percentage <= 100),
    CONSTRAINT ubo_verification_status_check CHECK (verification_status IN ('pending', 'verified', 'rejected'))
);

-- =====================================================
-- DISPUTES SERVICE SCHEMA
-- =====================================================

CREATE TABLE disputes (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    transaction_id UUID NOT NULL,
    dispute_type VARCHAR(50) NOT NULL,
    amount_value DECIMAL(20,8) NOT NULL,
    currency VARCHAR(3) NOT NULL,
    reason_code VARCHAR(50),
    description TEXT,
    status VARCHAR(50) NOT NULL DEFAULT 'open',
    priority VARCHAR(20) NOT NULL DEFAULT 'medium',
    due_date TIMESTAMP WITH TIME ZONE,
    evidence JSONB,
    external_ref VARCHAR(255),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    resolved_at TIMESTAMP WITH TIME ZONE,
    
    CONSTRAINT disputes_type_check CHECK (dispute_type IN ('chargeback', 'fraud', 'authorization', 'processing_error')),
    CONSTRAINT disputes_status_check CHECK (status IN ('open', 'under_review', 'accepted', 'rejected', 'closed')),
    CONSTRAINT disputes_priority_check CHECK (priority IN ('low', 'medium', 'high', 'critical')),
    CONSTRAINT disputes_amount_positive CHECK (amount_value > 0)
);

CREATE TABLE dispute_messages (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    dispute_id UUID NOT NULL REFERENCES disputes(id),
    sender_type VARCHAR(50) NOT NULL,
    sender_id VARCHAR(255) NOT NULL,
    message TEXT NOT NULL,
    attachments JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    CONSTRAINT dispute_messages_sender_type_check CHECK (sender_type IN ('merchant', 'customer', 'admin', 'system'))
);

-- =====================================================
-- AUTH SERVICE SCHEMA
-- =====================================================

CREATE TABLE organizations (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL,
    slug VARCHAR(100) UNIQUE NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'active',
    settings JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    CONSTRAINT organizations_status_check CHECK (status IN ('active', 'suspended', 'deleted'))
);

CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    organization_id UUID NOT NULL REFERENCES organizations(id),
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255),
    first_name VARCHAR(255),
    last_name VARCHAR(255),
    status VARCHAR(50) NOT NULL DEFAULT 'active',
    email_verified BOOLEAN DEFAULT false,
    last_login_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    CONSTRAINT users_status_check CHECK (status IN ('active', 'suspended', 'deleted'))
);

CREATE TABLE roles (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    organization_id UUID NOT NULL REFERENCES organizations(id),
    name VARCHAR(100) NOT NULL,
    description TEXT,
    permissions JSONB NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    UNIQUE(organization_id, name)
);

CREATE TABLE user_roles (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id),
    role_id UUID NOT NULL REFERENCES roles(id),
    granted_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    granted_by UUID REFERENCES users(id),
    
    UNIQUE(user_id, role_id)
);

CREATE TABLE api_keys (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    organization_id UUID NOT NULL REFERENCES organizations(id),
    key_hash VARCHAR(255) UNIQUE NOT NULL,
    name VARCHAR(255) NOT NULL,
    permissions JSONB,
    last_used_at TIMESTAMP WITH TIME ZONE,
    expires_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    created_by UUID REFERENCES users(id)
);

-- =====================================================
-- OBSERVABILITY SERVICE SCHEMA
-- =====================================================

CREATE TABLE audit_logs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    organization_id UUID,
    user_id UUID,
    action VARCHAR(255) NOT NULL,
    resource_type VARCHAR(100) NOT NULL,
    resource_id VARCHAR(255),
    old_values JSONB,
    new_values JSONB,
    ip_address INET,
    user_agent TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE system_metrics (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    service_name VARCHAR(100) NOT NULL,
    metric_name VARCHAR(255) NOT NULL,
    metric_value DECIMAL(20,8) NOT NULL,
    metric_type VARCHAR(50) NOT NULL,
    tags JSONB,
    timestamp TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    CONSTRAINT system_metrics_type_check CHECK (metric_type IN ('counter', 'gauge', 'histogram', 'timer'))
);

-- Indexes for performance
CREATE INDEX idx_reconciliation_batches_date ON reconciliation_batches(batch_date);
CREATE INDEX idx_reconciliation_items_batch_id ON reconciliation_items(batch_id);
CREATE INDEX idx_kyb_applications_business_id ON kyb_applications(business_id);
CREATE INDEX idx_kyb_applications_status ON kyb_applications(status);
CREATE INDEX idx_disputes_transaction_id ON disputes(transaction_id);
CREATE INDEX idx_disputes_status ON disputes(status);
CREATE INDEX idx_users_organization_id ON users(organization_id);
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_api_keys_organization_id ON api_keys(organization_id);
CREATE INDEX idx_audit_logs_organization_id ON audit_logs(organization_id);
CREATE INDEX idx_audit_logs_created_at ON audit_logs(created_at);
CREATE INDEX idx_system_metrics_service_name ON system_metrics(service_name);
