#!/bin/bash

# Compliance & Governance Setup Script
echo "🔒 SETTING UP COMPLIANCE & GOVERNANCE"

# Color functions
print_status() { echo -e "\033[1;34m[SETUP]\033[0m $1"; }
print_success() { echo -e "\033[1;32m[SUCCESS]\033[0m $1"; }

# Setup Audit Logging
setup_audit_logging() {
    print_status "Setting up Audit Logging..."
    
    # Create audit service directories
    mkdir -p audit/{config,scripts,reports}
    
    # Create audit service configuration
    cat > audit/config/audit-service.conf << 'EOF'
# Audit Service Configuration
[service]
name = "audit-service"
port = 8120
log_level = "INFO"

[storage]
type = "elasticsearch"
index_pattern = "audit-logs-{YYYY.MM.DD}"
retention_days = 2555
compression = true
encryption = true

[events]
authentication = true
authorization = true
financial_transactions = true
data_access = true
system_events = true

[compliance]
pci_dss = true
sox = true
gdpr = true
aml = true
EOF

    # Create audit event collector script
    cat > audit/scripts/collect-audit-events.sh << 'EOF'
#!/bin/bash
# Audit Event Collector Script

echo "Starting audit event collection..."

# Collect authentication events
curl -s "http://localhost:8090/api/v1/audit/authentication" | jq '.events[]' >> /audit/logs/auth-events.json

# Collect authorization events
curl -s "http://localhost:8090/api/v1/audit/authorization" | jq '.events[]' >> /audit/logs/authz-events.json

# Collect financial transaction events
curl -s "http://localhost:8090/api/v1/audit/financial" | jq '.events[]' >> /audit/logs/financial-events.json

# Collect data access events
curl -s "http://localhost:8090/api/v1/audit/data-access" | jq '.events[]' >> /audit/logs/data-access-events.json

echo "Audit event collection completed"
EOF

    chmod +x audit/scripts/collect-audit-events.sh
    print_success "Audit logging configured"
}

# Setup Data Encryption
setup_data_encryption() {
    print_status "Setting up Data Encryption..."
    
    # Create encryption configuration
    cat > encryption/encryption-config.yml << 'EOF'
# Data Encryption Configuration
encryption_standards:
  symmetric:
    aes_256_gcm: "AES-256-GCM"
    aes_256_cbc: "AES-256-CBC"
  
  asymmetric:
    rsa_4096: "RSA-4096"
    rsa_2048: "RSA-2048"

data_classification:
  levels:
    - name: "public"
      encryption_required: false
      retention: "1 year"
    - name: "internal"
      encryption_required: true
      algorithm: "aes_256_gcm"
      retention: "3 years"
    - name: "confidential"
      encryption_required: true
      algorithm: "aes_256_gcm"
      key_rotation: "90 days"
      retention: "5 years"
    - name: "restricted"
      encryption_required: true
      algorithm: "aes_256_gcm"
      key_rotation: "30 days"
      retention: "7 years"
    - name: "pci_data"
      encryption_required: true
      algorithm: "aes_256_gcm"
      key_rotation: "90 days"
      retention: "7 years"
      pci_compliant: true

encryption_at_rest:
  database_encryption:
    postgresql:
      enabled: true
      algorithm: "aes_256_gcm"
      key_management: "aws_kms"
    mongodb:
      enabled: true
      algorithm: "aes_256_gcm"
      key_management: "aws_kms"
    redis:
      enabled: true
      algorithm: "aes_256_gcm"
      key_management: "aws_kms"

encryption_in_transit:
  tls_configuration:
    tls_versions:
      tls_1_3: true
      tls_1_2: true
      tls_1_1: false
      tls_1_0: false
    cipher_suites:
      - "TLS_AES_256_GCM_SHA384"
      - "TLS_CHACHA20_POLY1305_SHA256"
      - "TLS_AES_128_GCM_SHA256"
EOF

    # Create key rotation script
    cat > encryption/rotate-keys.sh << 'EOF'
#!/bin/bash
# Key Rotation Script

echo "Starting key rotation process..."

# Rotate database encryption keys
echo "Rotating database encryption keys..."
aws kms schedule-key-deletion --key-id $(aws kms list-keys --query 'Keys[0].KeyId' --output text) --pending-window-in-days 7

# Create new encryption keys
echo "Creating new encryption keys..."
NEW_KEY_ID=$(aws kms create-key --description "Financial Platform Encryption Key $(date +%Y%m%d)" --query 'KeyMetadata.KeyId' --output text)

# Update key references
echo "Updating key references..."
aws kms update-alias --alias-name alias/financial-platform-encryption --target-key-id $NEW_KEY_ID

echo "Key rotation completed successfully"
EOF

    chmod +x encryption/rotate-keys.sh
    print_success "Data encryption configured"
}

# Setup Compliance Monitoring
setup_compliance_monitoring() {
    print_status "Setting up Compliance Monitoring..."
    
    # Create compliance monitoring configuration
    cat > compliance/compliance-monitor.conf << 'EOF'
# Compliance Monitoring Configuration
[monitoring]
enabled = true
real_time = true

[regulations]
pci_dss = true
sox = true
gdpr = true
aml = true

[alerts]
critical_violations = true
compliance_warnings = true
data_breaches = true
suspicious_activities = true

[reporting]
daily_reports = true
weekly_reports = true
monthly_reports = true
EOF

    # Create compliance check script
    cat > compliance/check-compliance.sh << 'EOF'
#!/bin/bash
# Compliance Check Script

echo "Running compliance checks..."

# PCI DSS Compliance Check
echo "Checking PCI DSS compliance..."
curl -s "http://localhost:8090/api/v1/compliance/pci-dss" | jq '.status'

# SOX Compliance Check
echo "Checking SOX compliance..."
curl -s "http://localhost:8090/api/v1/compliance/sox" | jq '.status'

# GDPR Compliance Check
echo "Checking GDPR compliance..."
curl -s "http://localhost:8090/api/v1/compliance/gdpr" | jq '.status'

# AML Compliance Check
echo "Checking AML compliance..."
curl -s "http://localhost:8090/api/v1/compliance/aml" | jq '.status'

echo "Compliance checks completed"
EOF

    chmod +x compliance/check-compliance.sh
    print_success "Compliance monitoring configured"
}

# Setup Data Governance
setup_data_governance() {
    print_status "Setting up Data Governance..."
    
    # Create data governance configuration
    cat > governance/data-governance.conf << 'EOF'
# Data Governance Configuration
[governance]
enabled = true
framework = "comprehensive"

[classification]
levels = ["public", "internal", "confidential", "restricted", "pci_data"]
auto_classification = true

[quality]
dimensions = ["accuracy", "completeness", "consistency", "timeliness", "validity", "uniqueness"]
monitoring = true
thresholds = 90

[lineage]
tracking = true
visualization = true
tool = "apache_atlas"

[privacy]
controls = ["data_minimization", "purpose_limitation", "consent_management", "data_subject_rights"]
monitoring = true

[retention]
policies = true
automated_deletion = true
legal_hold = true

[access]
rbac = true
abac = true
jit_access = true
privileged_access = true
EOF

    # Create data quality check script
    cat > governance/check-data-quality.sh << 'EOF'
#!/bin/bash
# Data Quality Check Script

echo "Running data quality checks..."

# Check data accuracy
echo "Checking data accuracy..."
curl -s "http://localhost:8090/api/v1/data-quality/accuracy" | jq '.score'

# Check data completeness
echo "Checking data completeness..."
curl -s "http://localhost:8090/api/v1/data-quality/completeness" | jq '.score'

# Check data consistency
echo "Checking data consistency..."
curl -s "http://localhost:8090/api/v1/data-quality/consistency" | jq '.score'

# Check data timeliness
echo "Checking data timeliness..."
curl -s "http://localhost:8090/api/v1/data-quality/timeliness" | jq '.score'

echo "Data quality checks completed"
EOF

    chmod +x governance/check-data-quality.sh
    print_success "Data governance configured"
}

# Setup Compliance Docker Compose
setup_compliance_compose() {
    print_status "Setting up Compliance Docker Compose..."
    
    cat > docker-compose-compliance.yml << EOF
version: '3.8'

services:
  # Audit Service
  audit-service:
    image: financial-platform/audit-service:latest
    container_name: audit-service
    ports:
      - "8120:8120"
    volumes:
      - ./audit/config:/app/config
      - ./audit/logs:/app/logs
    environment:
      - ELASTICSEARCH_URL=http://elasticsearch:9200
      - LOG_LEVEL=INFO
    networks:
      - compliance_network

  # Compliance Monitor
  compliance-monitor:
    image: financial-platform/compliance-monitor:latest
    container_name: compliance-monitor
    ports:
      - "8121:8121"
    volumes:
      - ./compliance:/app/config
    environment:
      - PROMETHEUS_URL=http://prometheus:9090
      - ALERTMANAGER_URL=http://alertmanager:9093
    networks:
      - compliance_network

  # Data Governance Service
  data-governance:
    image: financial-platform/data-governance:latest
    container_name: data-governance
    ports:
      - "8122:8122"
    volumes:
      - ./governance:/app/config
    environment:
      - ELASTICSEARCH_URL=http://elasticsearch:9200
      - KAFKA_URL=kafka:9092
    networks:
      - compliance_network

  # Apache Atlas (Data Lineage)
  apache-atlas:
    image: apache/atlas:latest
    container_name: apache-atlas
    ports:
      - "21000:21000"
    environment:
      - ATLAS_SERVER_OPTS=-Xms1024m -Xmx2048m
    volumes:
      - atlas_data:/var/lib/atlas
    networks:
      - compliance_network

  # Elasticsearch (Audit Logs)
  elasticsearch:
    image: docker.elastic.co/elasticsearch/elasticsearch:8.11.0
    container_name: elasticsearch
    ports:
      - "9200:9200"
    environment:
      - discovery.type=single-node
      - xpack.security.enabled=false
    volumes:
      - elasticsearch_data:/usr/share/elasticsearch/data
    networks:
      - compliance_network

  # Kibana (Audit Logs Visualization)
  kibana:
    image: docker.elastic.co/kibana/kibana:8.11.0
    container_name: kibana
    ports:
      - "5601:5601"
    environment:
      - ELASTICSEARCH_HOSTS=http://elasticsearch:9200
    depends_on:
      - elasticsearch
    networks:
      - compliance_network

volumes:
  atlas_data:
  elasticsearch_data:

networks:
  compliance_network:
    driver: bridge
EOF

    print_success "Compliance Docker Compose configured"
}

# Setup Compliance Testing
setup_compliance_testing() {
    print_status "Setting up Compliance Testing..."
    
    # Create compliance test script
    cat > testing/run-compliance-tests.sh << 'EOF'
#!/bin/bash
# Compliance Testing Script

echo "Running comprehensive compliance tests..."

# Test audit logging
echo "Testing audit logging..."
curl -s "http://localhost:8120/health" | jq '.status'

# Test data encryption
echo "Testing data encryption..."
curl -s "http://localhost:8090/api/v1/encryption/status" | jq '.encryption_enabled'

# Test compliance monitoring
echo "Testing compliance monitoring..."
curl -s "http://localhost:8121/health" | jq '.status'

# Test data governance
echo "Testing data governance..."
curl -s "http://localhost:8122/health" | jq '.status'

# Test data lineage
echo "Testing data lineage..."
curl -s "http://localhost:21000/api/atlas/v2/types" | jq '.count'

# Test Elasticsearch
echo "Testing Elasticsearch..."
curl -s "http://localhost:9200/_cluster/health" | jq '.status'

# Test Kibana
echo "Testing Kibana..."
curl -s "http://localhost:5601/api/status" | jq '.status.overall.level'

echo "Compliance tests completed"
EOF

    chmod +x testing/run-compliance-tests.sh
    print_success "Compliance testing configured"
}

# Main setup
main() {
    echo "🔒 COMPLIANCE & GOVERNANCE SETUP"
    echo "================================"
    
    setup_audit_logging
    setup_data_encryption
    setup_compliance_monitoring
    setup_data_governance
    setup_compliance_compose
    setup_compliance_testing
    
    echo ""
    echo "✅ COMPLIANCE & GOVERNANCE SETUP COMPLETE"
    echo "🔒 COMPONENTS: Audit Logging, Data Encryption, Compliance Monitoring, Data Governance"
    echo "📋 FEATURES: PCI DSS, SOX, GDPR, AML Compliance, Data Lineage, Quality Management"
    echo "📊 MONITORING: Real-time compliance monitoring, Automated reporting, Alert management"
    echo "🚀 START: docker-compose -f docker-compose-compliance.yml up -d"
}

main "$@"
