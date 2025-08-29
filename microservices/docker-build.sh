#!/bin/bash

# Docker Build Script for Financial Platform Microservices
echo "🐳 BUILDING ALL DOCKER IMAGES"
echo "============================="

# Color functions
print_status() {
    echo -e "\033[1;34m[BUILDING]\033[0m $1"
}

print_success() {
    echo -e "\033[1;32m[SUCCESS]\033[0m $1"
}

print_error() {
    echo -e "\033[1;31m[ERROR]\033[0m $1"
}

# Check if Docker is running
if ! docker info > /dev/null 2>&1; then
    print_error "Docker is not running. Please start Docker and try again."
    exit 1
fi

# Build infrastructure services first
echo ""
echo "🔧 BUILDING INFRASTRUCTURE SERVICES"
echo "==================================="

# Build API Gateway
print_status "API Gateway"
if docker build -f api-gateway/Dockerfile -t financial-platform/api-gateway:latest .; then
    print_success "API Gateway built successfully"
else
    print_error "Failed to build API Gateway"
    exit 1
fi

# Build Database Service
print_status "Database Service"
if docker build -f database-service/Dockerfile -t financial-platform/database-service:latest .; then
    print_success "Database Service built successfully"
else
    print_error "Failed to build Database Service"
    exit 1
fi

# Build Monitoring Service
print_status "Monitoring Service"
if docker build -f monitoring-service/Dockerfile -t financial-platform/monitoring-service:latest .; then
    print_success "Monitoring Service built successfully"
else
    print_error "Failed to build Monitoring Service"
    exit 1
fi

# Build Message Queue Service
print_status "Message Queue Service"
if docker build -f message-queue-service/Dockerfile -t financial-platform/message-queue-service:latest .; then
    print_success "Message Queue Service built successfully"
else
    print_error "Failed to build Message Queue Service"
    exit 1
fi

# Build Service Discovery
print_status "Service Discovery"
if docker build -f service-discovery/Dockerfile -t financial-platform/service-discovery:latest .; then
    print_success "Service Discovery built successfully"
else
    print_error "Failed to build Service Discovery"
    exit 1
fi

# Build Vault Service
print_status "Vault Service"
if docker build -f vault-service/Dockerfile -t financial-platform/vault-service:latest .; then
    print_success "Vault Service built successfully"
else
    print_error "Failed to build Vault Service"
    exit 1
fi

echo ""
echo "💼 BUILDING BUSINESS SERVICES"
echo "============================="

# Build business services
services=(
    "escrow-service"
    "payment-service"
    "ledger-service"
    "risk-service"
    "treasury-service"
    "evidence-service"
    "compliance-service"
    "workflow-service"
    "journal-service"
    "fees-service"
    "refunds-service"
    "transfers-service"
    "fx-service"
    "payouts-service"
    "reserves-service"
    "reconciliation-service"
    "kyb-service"
    "sca-service"
    "disputes-service"
    "dx-service"
)

for service in "${services[@]}"; do
    print_status "$service"
    if docker build -f "$service/Dockerfile" -t "financial-platform/$service:latest" .; then
        print_success "$service built successfully"
    else
        print_error "Failed to build $service"
        exit 1
    fi
done

echo ""
echo "🔧 BUILDING CRITICAL INFRASTRUCTURE SERVICES"
echo "============================================"

# Build critical infrastructure services
critical_services=(
    "auth-service"
    "idempotency-service"
    "eventbus-service"
    "saga-service"
    "webhooks-service"
    "observability-service"
    "config-service"
    "workers-service"
    "portal-service"
    "data-platform-service"
    "compliance-ops-service"
)

for service in "${critical_services[@]}"; do
    print_status "$service"
    if docker build -f "$service/Dockerfile" -t "financial-platform/$service:latest" .; then
        print_success "$service built successfully"
    else
        print_error "Failed to build $service"
        exit 1
    fi
done

echo ""
echo "🎯 ALL DOCKER IMAGES BUILT SUCCESSFULLY!"
echo "========================================"
echo ""
echo "📊 Summary:"
echo "  • Infrastructure Services: 6"
echo "  • Business Services: 20"
echo "  • Critical Infrastructure: 11"
echo "  • Total Images: 37"
echo ""
echo "🚀 Next steps:"
echo "1. Run 'docker-compose up -d' to start all services"
echo "2. Run 'docker-compose ps' to check service status"
echo "3. Run 'docker-compose logs -f' to monitor logs"
echo ""
echo "🌐 Access Points:"
echo "  • API Gateway: http://localhost:8090"
echo "  • Grafana: http://localhost:3000 (admin/admin)"
echo "  • Prometheus: http://localhost:9090"
echo "  • RabbitMQ: http://localhost:15672 (guest/guest)"
echo "  • Consul: http://localhost:8500"
echo "  • Vault: http://localhost:8200"
