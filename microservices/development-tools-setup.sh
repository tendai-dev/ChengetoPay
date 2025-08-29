#!/bin/bash

# Development Tools Setup Script
echo "🛠️ SETTING UP DEVELOPMENT TOOLS"

# Color functions
print_status() { echo -e "\033[1;34m[SETUP]\033[0m $1"; }
print_success() { echo -e "\033[1;32m[SUCCESS]\033[0m $1"; }

# Setup API Documentation
setup_api_documentation() {
    print_status "Setting up API Documentation..."
    
    # Create documentation directories
    mkdir -p docs/{api,guides,examples,reference}
    
    # Create OpenAPI specification
    cat > docs/api/openapi.yaml << 'EOF'
openapi: 3.1.0
info:
  title: Financial Platform API
  description: Comprehensive API for Financial Platform with 37 microservices
  version: 1.0.0
  contact:
    name: Financial Platform API Team
    email: api@financialplatform.com

servers:
  - url: https://api.financialplatform.com
    description: Production API Server
  - url: http://localhost:8090
    description: Local Development Server

paths:
  /api/v1/payment/process:
    post:
      summary: Process Payment
      description: Process a new payment transaction
      tags: [Payments]
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/PaymentRequest'
      responses:
        '200':
          description: Payment processed successfully
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/PaymentResponse'

components:
  schemas:
    PaymentRequest:
      type: object
      required: [amount, currency, payment_method]
      properties:
        amount:
          type: number
          format: decimal
          minimum: 0.01
          description: Payment amount
        currency:
          type: string
          enum: [USD, EUR, GBP, JPY]
          description: Payment currency
        payment_method:
          type: string
          enum: [card, bank_transfer, crypto]
          description: Payment method

    PaymentResponse:
      type: object
      properties:
        transaction_id:
          type: string
          format: uuid
          description: Unique transaction ID
        status:
          type: string
          enum: [pending, processing, completed, failed]
          description: Payment status
        amount:
          type: number
          format: decimal
          description: Payment amount
        currency:
          type: string
          description: Payment currency
EOF

    # Create documentation generation script
    cat > docs/generate-docs.sh << 'EOF'
#!/bin/bash
# Documentation Generation Script

echo "Generating API documentation..."

# Generate OpenAPI documentation
echo "Generating OpenAPI documentation..."
swagger-codegen generate -i docs/api/openapi.yaml -l html2 -o docs/generated/html

# Generate Postman collection
echo "Generating Postman collection..."
swagger-codegen generate -i docs/api/openapi.yaml -l postman -o docs/generated/postman

echo "Documentation generation completed"
EOF

    chmod +x docs/generate-docs.sh
    print_success "API documentation configured"
}

# Setup Developer Portal
setup_developer_portal() {
    print_status "Setting up Developer Portal..."
    
    # Create portal directories
    mkdir -p portal/{config,assets,content}
    
    # Create portal configuration
    cat > portal/config/portal.conf << 'EOF'
# Developer Portal Configuration
[portal]
name = "Financial Platform Developer Portal"
description = "Comprehensive developer portal for Financial Platform APIs"
version = "1.0.0"
base_url = "https://developers.financialplatform.com"
contact_email = "developers@financialplatform.com"

[features]
api_explorer = true
code_samples = true
documentation = true
testing_tools = true
analytics = true
support = true

[api_management]
gateway_url = "https://api.financialplatform.com"
admin_url = "https://admin-api.financialplatform.com"

[api_versioning]
enabled = true
versioning_strategy = "url_path"
current_version = "v1"
supported_versions = ["v1", "v2"]
EOF

    # Create portal setup script
    cat > portal/setup-portal.sh << 'EOF'
#!/bin/bash
# Developer Portal Setup Script

echo "Setting up Developer Portal..."

# Install dependencies
echo "Installing dependencies..."
npm install -g @stoplight/elements
npm install -g @redocly/cli

# Generate portal content
echo "Generating portal content..."
mkdir -p portal/content/{docs,guides,examples}

# Copy API documentation
cp -r docs/api/* portal/content/docs/

echo "Developer Portal setup completed"
EOF

    chmod +x portal/setup-portal.sh
    print_success "Developer Portal configured"
}

# Setup Sandbox Environment
setup_sandbox_environment() {
    print_status "Setting up Sandbox Environment..."
    
    # Create sandbox directories
    mkdir -p sandbox/{config,data,mocks,tests}
    
    # Create sandbox configuration
    cat > sandbox/config/sandbox.conf << 'EOF'
# Sandbox Environment Configuration
[sandbox]
name = "Financial Platform Sandbox"
description = "Complete sandbox environment for development and testing"
version = "1.0.0"
base_url = "https://sandbox-api.financialplatform.com"
admin_url = "https://sandbox-admin.financialplatform.com"

[features]
isolated_data = true
mock_services = true
test_data_generation = true
api_testing = true
performance_testing = true
security_testing = true

[database]
postgresql_host = "sandbox-postgresql"
postgresql_port = 5432
postgresql_database = "financial_platform_sandbox"
postgresql_username = "sandbox_user"
postgresql_password = "sandbox_password"

mongodb_host = "sandbox-mongodb"
mongodb_port = 27017
mongodb_database = "financial_platform_sandbox"

redis_host = "sandbox-redis"
redis_port = 6379
redis_database = 0

[mock_services]
stripe_mock = true
paypal_mock = true
plaid_mock = true
jumio_mock = true

[test_data]
user_count = 1000
transaction_count = 5000
payment_count = 3000
escrow_count = 500
EOF

    # Create test data generation script
    cat > sandbox/generate-test-data.sh << 'EOF'
#!/bin/bash
# Test Data Generation Script

echo "Generating test data for sandbox environment..."

# Generate user data
echo "Generating user data..."
for i in {1..1000}; do
    echo "INSERT INTO users (user_id, email, first_name, last_name, phone, date_of_birth, created_at) VALUES ('$(uuidgen)', 'user$i@example.com', 'User$i', 'Test', '+1-555-0123', '1990-01-01', NOW());" >> sandbox/data/users.sql
done

# Generate transaction data
echo "Generating transaction data..."
for i in {1..5000}; do
    amount=$(echo "scale=2; $RANDOM / 32767 * 10000" | bc)
    echo "INSERT INTO transactions (transaction_id, user_id, amount, currency, status, created_at) VALUES ('$(uuidgen)', 'user$((i % 1000 + 1))', $amount, 'USD', 'completed', NOW());" >> sandbox/data/transactions.sql
done

echo "Test data generation completed"
EOF

    chmod +x sandbox/generate-test-data.sh
    print_success "Sandbox environment configured"
}

# Setup API Versioning
setup_api_versioning() {
    print_status "Setting up API Versioning..."
    
    # Create versioning directories
    mkdir -p versioning/{config,scripts,docs}
    
    # Create versioning configuration
    cat > versioning/config/versioning.conf << 'EOF'
# API Versioning Configuration
[versioning]
primary_strategy = "url_path"
fallback_strategy = "header"

[url_path_versioning]
enabled = true
pattern = "/api/v{version}"

[header_versioning]
enabled = true
header_name = "API-Version"
header_value = "v{version}"

[versions]
v1_status = "current"
v1_release_date = "2024-01-01"
v1_deprecation_date = "2025-01-01"
v1_sunset_date = "2026-01-01"

v2_status = "beta"
v2_release_date = "2024-06-01"
v2_deprecation_date = "2026-06-01"
v2_sunset_date = "2027-06-01"

[compatibility]
additive_changes_only = true
no_breaking_changes = true
deprecation_notice_period = "6 months"

[migration]
migration_guides = true
migration_tools = true
migration_support = true
EOF

    # Create version management script
    cat > versioning/manage-versions.sh << 'EOF'
#!/bin/bash
# API Version Management Script

echo "Managing API versions..."

# Check version usage
echo "Checking version usage..."
curl -s "https://api.financialplatform.com/api/v1/health" | jq '.version'
curl -s "https://api.financialplatform.com/api/v2/health" | jq '.version'

# Generate deprecation notices
echo "Generating deprecation notices..."
cat > versioning/deprecation-notices.md << 'NOTICE_EOF'
# API Version Deprecation Notices

## Version v1
- **Status**: Current
- **Release Date**: 2024-01-01
- **Deprecation Date**: 2025-01-01
- **Sunset Date**: 2026-01-01

## Version v2
- **Status**: Beta
- **Release Date**: 2024-06-01
- **Deprecation Date**: 2026-06-01
- **Sunset Date**: 2027-06-01

## Migration Timeline
- **v1 to v2**: Available now
- **v2 to v3**: Planning phase
NOTICE_EOF

echo "Version management completed"
EOF

    chmod +x versioning/manage-versions.sh
    print_success "API versioning configured"
}

# Setup Development Tools Docker Compose
setup_development_compose() {
    print_status "Setting up Development Tools Docker Compose..."
    
    cat > docker-compose-development.yml << EOF
version: '3.8'

services:
  # API Documentation
  swagger-ui:
    image: swaggerapi/swagger-ui:latest
    container_name: swagger-ui
    ports:
      - "8080:8080"
    environment:
      SWAGGER_JSON: /app/openapi.yaml
    volumes:
      - ./docs/api/openapi.yaml:/app/openapi.yaml
    networks:
      - development_network

  redoc:
    image: redocly/redoc:latest
    container_name: redoc
    ports:
      - "8081:80"
    environment:
      SPEC_URL: /app/openapi.yaml
    volumes:
      - ./docs/api/openapi.yaml:/app/openapi.yaml
    networks:
      - development_network

  # Developer Portal
  developer-portal:
    image: financial-platform/developer-portal:latest
    container_name: developer-portal
    ports:
      - "8082:3000"
    volumes:
      - ./portal/config:/app/config
      - ./portal/content:/app/content
    environment:
      - NODE_ENV=development
      - PORT=3000
    networks:
      - development_network

  # Sandbox Environment
  sandbox-api:
    image: financial-platform/sandbox-api:latest
    container_name: sandbox-api
    ports:
      - "8083:8090"
    volumes:
      - ./sandbox/config:/app/config
      - ./sandbox/data:/app/data
    environment:
      - ENVIRONMENT=sandbox
      - DATABASE_URL=postgresql://sandbox_user:sandbox_password@sandbox-postgresql:5432/financial_platform_sandbox
    networks:
      - development_network
    depends_on:
      - sandbox-postgresql
      - sandbox-mongodb
      - sandbox-redis

  sandbox-postgresql:
    image: postgres:15-alpine
    container_name: sandbox-postgresql
    ports:
      - "5433:5432"
    environment:
      - POSTGRES_DB=financial_platform_sandbox
      - POSTGRES_USER=sandbox_user
      - POSTGRES_PASSWORD=sandbox_password
    volumes:
      - sandbox_postgresql_data:/var/lib/postgresql/data
    networks:
      - development_network

  sandbox-mongodb:
    image: mongo:7.0
    container_name: sandbox-mongodb
    ports:
      - "27018:27017"
    environment:
      - MONGO_INITDB_DATABASE=financial_platform_sandbox
    volumes:
      - sandbox_mongodb_data:/data/db
    networks:
      - development_network

  sandbox-redis:
    image: redis:7.2-alpine
    container_name: sandbox-redis
    ports:
      - "6380:6379"
    volumes:
      - sandbox_redis_data:/data
    networks:
      - development_network

  # Mock Services
  mock-stripe:
    image: financial-platform/mock-stripe:latest
    container_name: mock-stripe
    ports:
      - "8084:3000"
    environment:
      - MOCK_MODE=true
    networks:
      - development_network

  mock-paypal:
    image: financial-platform/mock-paypal:latest
    container_name: mock-paypal
    ports:
      - "8085:3000"
    environment:
      - MOCK_MODE=true
    networks:
      - development_network

  # Testing Tools
  postman:
    image: postman/newman:latest
    container_name: postman
    volumes:
      - ./tests/postman:/etc/newman
    networks:
      - development_network

  k6:
    image: grafana/k6:latest
    container_name: k6
    ports:
      - "8086:6565"
    volumes:
      - ./tests/k6:/scripts
    networks:
      - development_network

  # Monitoring
  grafana:
    image: grafana/grafana:latest
    container_name: grafana
    ports:
      - "3000:3000"
    environment:
      - GF_SECURITY_ADMIN_PASSWORD=admin
    volumes:
      - grafana_data:/var/lib/grafana
    networks:
      - development_network

  prometheus:
    image: prom/prometheus:latest
    container_name: prometheus
    ports:
      - "9090:9090"
    volumes:
      - ./monitoring/prometheus.yml:/etc/prometheus/prometheus.yml
    networks:
      - development_network

volumes:
  sandbox_postgresql_data:
  sandbox_mongodb_data:
  sandbox_redis_data:
  grafana_data:

networks:
  development_network:
    driver: bridge
EOF

    print_success "Development Docker Compose configured"
}

# Setup Development Testing
setup_development_testing() {
    print_status "Setting up Development Testing..."
    
    # Create testing directories
    mkdir -p tests/{unit,integration,api,performance,security}
    
    # Create API test script
    cat > tests/api/test-api.sh << 'EOF'
#!/bin/bash
# API Testing Script

echo "Running API tests..."

# Test API Gateway
echo "Testing API Gateway..."
curl -s "http://localhost:8090/health" | jq '.status'

# Test Payment Service
echo "Testing Payment Service..."
curl -s "http://localhost:8083/health" | jq '.status'

# Test User Service
echo "Testing User Service..."
curl -s "http://localhost:8084/health" | jq '.status'

# Test Escrow Service
echo "Testing Escrow Service..."
curl -s "http://localhost:8081/health" | jq '.status'

# Test with Postman
echo "Running Postman tests..."
newman run tests/postman/collection.json --environment tests/postman/environment.json

# Test with k6
echo "Running k6 performance tests..."
k6 run tests/k6/load-test.js

echo "API testing completed"
EOF

    # Create k6 performance test
    cat > tests/k6/load-test.js << 'EOF'
import http from 'k6/http';
import { check, sleep } from 'k6';

export const options = {
  stages: [
    { duration: '2m', target: 100 }, // Ramp up
    { duration: '5m', target: 100 }, // Stay at 100 users
    { duration: '2m', target: 0 },   // Ramp down
  ],
  thresholds: {
    http_req_duration: ['p(95)<500'], // 95% of requests must complete below 500ms
    http_req_failed: ['rate<0.01'],   // Less than 1% of requests can fail
  },
};

const BASE_URL = 'http://localhost:8090';

export default function () {
  // Test health endpoint
  const healthResponse = http.get(`${BASE_URL}/health`);
  check(healthResponse, {
    'health status is 200': (r) => r.status === 200,
    'health response time < 200ms': (r) => r.timings.duration < 200,
  });

  // Test payment endpoint
  const paymentData = JSON.stringify({
    amount: 100.00,
    currency: 'USD',
    payment_method: 'card',
    description: 'Load test payment'
  });

  const paymentResponse = http.post(`${BASE_URL}/api/v1/payment/process`, paymentData, {
    headers: { 'Content-Type': 'application/json' },
  });

  check(paymentResponse, {
    'payment status is 200': (r) => r.status === 200,
    'payment response time < 1000ms': (r) => r.timings.duration < 1000,
  });

  sleep(1);
}
EOF

    chmod +x tests/api/test-api.sh
    print_success "Development testing configured"
}

# Main setup
main() {
    echo "🛠️ DEVELOPMENT TOOLS SETUP"
    echo "=========================="
    
    setup_api_documentation
    setup_developer_portal
    setup_sandbox_environment
    setup_api_versioning
    setup_development_compose
    setup_development_testing
    
    echo ""
    echo "✅ DEVELOPMENT TOOLS SETUP COMPLETE"
    echo "🛠️ COMPONENTS: API Documentation, Developer Portal, Sandbox Environment, API Versioning"
    echo "📚 FEATURES: OpenAPI/Swagger, Interactive Docs, Test Environment, Version Management"
    echo "🧪 TESTING: Unit Tests, Integration Tests, API Tests, Performance Tests"
    echo "🚀 START: docker-compose -f docker-compose-development.yml up -d"
}

main "$@"
