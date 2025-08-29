#!/bin/bash

# Integration Test Suite for Financial Platform Microservices
echo "🧪 INTEGRATION TEST SUITE"
echo "========================"

# Color functions
print_status() {
    echo -e "\033[1;34m[TESTING]\033[0m $1"
}

print_success() {
    echo -e "\033[1;32m[PASS]\033[0m $1"
}

print_error() {
    echo -e "\033[1;31m[FAIL]\033[0m $1"
}

# Test configuration
BASE_URL="http://localhost"
TIMEOUT=30
RETRIES=3

# Service endpoints to test
declare -A services=(
    ["api-gateway"]="8090"
    ["database-service"]="8115"
    ["monitoring-service"]="8116"
    ["message-queue-service"]="8117"
    ["service-discovery"]="8118"
    ["vault-service"]="8119"
    ["escrow-service"]="8081"
    ["payment-service"]="8083"
    ["ledger-service"]="8084"
    ["risk-service"]="8085"
    ["treasury-service"]="8086"
    ["evidence-service"]="8087"
    ["compliance-service"]="8088"
    ["workflow-service"]="8089"
    ["journal-service"]="8091"
    ["fees-service"]="8092"
    ["refunds-service"]="8093"
    ["transfers-service"]="8094"
    ["fx-service"]="8095"
    ["payouts-service"]="8096"
    ["reserves-service"]="8097"
    ["reconciliation-service"]="8098"
    ["kyb-service"]="8099"
    ["sca-service"]="8100"
    ["disputes-service"]="8101"
    ["dx-service"]="8102"
    ["auth-service"]="8103"
    ["idempotency-service"]="8104"
    ["eventbus-service"]="8105"
    ["saga-service"]="8106"
    ["webhooks-service"]="8108"
    ["observability-service"]="8109"
    ["config-service"]="8110"
    ["workers-service"]="8111"
    ["portal-service"]="8112"
    ["data-platform-service"]="8113"
    ["compliance-ops-service"]="8114"
)

# Wait for service to be ready
wait_for_service() {
    local service=$1
    local port=$2
    local url="$BASE_URL:$port/health"
    
    print_status "Waiting for $service to be ready..."
    
    for i in $(seq 1 $RETRIES); do
        if curl -f -s "$url" > /dev/null 2>&1; then
            print_success "$service is ready"
            return 0
        fi
        
        echo "Attempt $i/$RETRIES: $service not ready, waiting..."
        sleep 5
    done
    
    print_error "$service failed to start"
    return 1
}

# Test health endpoint
test_health_endpoint() {
    local service=$1
    local port=$2
    local url="$BASE_URL:$port/health"
    
    print_status "Testing health endpoint for $service"
    
    if curl -f -s "$url" | jq -e '.status == "healthy"' > /dev/null 2>&1; then
        print_success "Health check passed for $service"
        return 0
    else
        print_error "Health check failed for $service"
        return 1
    fi
}

# Test metrics endpoint
test_metrics_endpoint() {
    local service=$1
    local port=$2
    local url="$BASE_URL:$port/metrics"
    
    print_status "Testing metrics endpoint for $service"
    
    if curl -f -s "$url" | grep -q "go_" > /dev/null 2>&1; then
        print_success "Metrics endpoint working for $service"
        return 0
    else
        print_error "Metrics endpoint failed for $service"
        return 1
    fi
}

# Test service-specific endpoints
test_service_endpoints() {
    local service=$1
    local port=$2
    
    case $service in
        "api-gateway")
            test_api_gateway_endpoints $port
            ;;
        "database-service")
            test_database_endpoints $port
            ;;
        "monitoring-service")
            test_monitoring_endpoints $port
            ;;
        "message-queue-service")
            test_message_queue_endpoints $port
            ;;
        "vault-service")
            test_vault_endpoints $port
            ;;
        *)
            test_generic_endpoints $service $port
            ;;
    esac
}

# Test API Gateway specific endpoints
test_api_gateway_endpoints() {
    local port=$1
    local base_url="$BASE_URL:$port"
    
    print_status "Testing API Gateway endpoints"
    
    # Test root endpoint
    if curl -f -s "$base_url/" | jq -e '.services' > /dev/null 2>&1; then
        print_success "API Gateway root endpoint working"
    else
        print_error "API Gateway root endpoint failed"
    fi
    
    # Test service routing
    if curl -f -s "$base_url/api/v1/health" > /dev/null 2>&1; then
        print_success "API Gateway service routing working"
    else
        print_error "API Gateway service routing failed"
    fi
}

# Test Database Service specific endpoints
test_database_endpoints() {
    local port=$1
    local base_url="$BASE_URL:$port"
    
    print_status "Testing Database Service endpoints"
    
    # Test database status
    if curl -f -s "$base_url/v1/status" | jq -e '.postgres' > /dev/null 2>&1; then
        print_success "Database status endpoint working"
    else
        print_error "Database status endpoint failed"
    fi
    
    # Test database health
    if curl -f -s "$base_url/v1/health" | jq -e '.status' > /dev/null 2>&1; then
        print_success "Database health endpoint working"
    else
        print_error "Database health endpoint failed"
    fi
}

# Test Monitoring Service specific endpoints
test_monitoring_endpoints() {
    local port=$1
    local base_url="$BASE_URL:$port"
    
    print_status "Testing Monitoring Service endpoints"
    
    # Test service health aggregation
    if curl -f -s "$base_url/v1/health" | jq -e '.services' > /dev/null 2>&1; then
        print_success "Monitoring health aggregation working"
    else
        print_error "Monitoring health aggregation failed"
    fi
    
    # Test metrics aggregation
    if curl -f -s "$base_url/v1/metrics" | jq -e '.metrics' > /dev/null 2>&1; then
        print_success "Monitoring metrics aggregation working"
    else
        print_error "Monitoring metrics aggregation failed"
    fi
}

# Test Message Queue Service specific endpoints
test_message_queue_endpoints() {
    local port=$1
    local base_url="$BASE_URL:$port"
    
    print_status "Testing Message Queue Service endpoints"
    
    # Test queue status
    if curl -f -s "$base_url/v1/status" | jq -e '.rabbitmq' > /dev/null 2>&1; then
        print_success "Message queue status endpoint working"
    else
        print_error "Message queue status endpoint failed"
    fi
    
    # Test event store status
    if curl -f -s "$base_url/v1/events" | jq -e '.events' > /dev/null 2>&1; then
        print_success "Event store endpoint working"
    else
        print_error "Event store endpoint failed"
    fi
}

# Test Vault Service specific endpoints
test_vault_endpoints() {
    local port=$1
    local base_url="$BASE_URL:$port"
    
    print_status "Testing Vault Service endpoints"
    
    # Test vault status
    if curl -f -s "$base_url/v1/status" | jq -e '.initialized' > /dev/null 2>&1; then
        print_success "Vault status endpoint working"
    else
        print_error "Vault status endpoint failed"
    fi
    
    # Test secrets endpoint
    if curl -f -s "$base_url/v1/secrets" | jq -e '.secrets' > /dev/null 2>&1; then
        print_success "Vault secrets endpoint working"
    else
        print_error "Vault secrets endpoint failed"
    fi
}

# Test generic endpoints for business services
test_generic_endpoints() {
    local service=$1
    local port=$2
    local base_url="$BASE_URL:$port"
    
    print_status "Testing generic endpoints for $service"
    
    # Test service info endpoint
    if curl -f -s "$base_url/v1/info" | jq -e '.service' > /dev/null 2>&1; then
        print_success "$service info endpoint working"
    else
        print_error "$service info endpoint failed"
    fi
    
    # Test service version endpoint
    if curl -f -s "$base_url/v1/version" | jq -e '.version' > /dev/null 2>&1; then
        print_success "$service version endpoint working"
    else
        print_error "$service version endpoint failed"
    fi
}

# Run all tests
run_integration_tests() {
    echo ""
    echo "🚀 STARTING INTEGRATION TESTS"
    echo "============================="
    
    local total_tests=0
    local passed_tests=0
    local failed_tests=0
    
    for service in "${!services[@]}"; do
        local port="${services[$service]}"
        total_tests=$((total_tests + 3))  # health, metrics, service-specific
        
        echo ""
        echo "Testing $service (Port: $port)"
        echo "--------------------------------"
        
        # Wait for service to be ready
        if wait_for_service "$service" "$port"; then
            # Test health endpoint
            if test_health_endpoint "$service" "$port"; then
                passed_tests=$((passed_tests + 1))
            else
                failed_tests=$((failed_tests + 1))
            fi
            
            # Test metrics endpoint
            if test_metrics_endpoint "$service" "$port"; then
                passed_tests=$((passed_tests + 1))
            else
                failed_tests=$((failed_tests + 1))
            fi
            
            # Test service-specific endpoints
            if test_service_endpoints "$service" "$port"; then
                passed_tests=$((passed_tests + 1))
            else
                failed_tests=$((failed_tests + 1))
            fi
        else
            failed_tests=$((failed_tests + 3))
        fi
    done
    
    echo ""
    echo "📊 TEST RESULTS SUMMARY"
    echo "======================="
    echo "Total Tests: $total_tests"
    echo "Passed: $passed_tests"
    echo "Failed: $failed_tests"
    echo "Success Rate: $((passed_tests * 100 / total_tests))%"
    
    if [ $failed_tests -eq 0 ]; then
        echo ""
        echo "🎉 ALL INTEGRATION TESTS PASSED!"
        return 0
    else
        echo ""
        echo "❌ SOME TESTS FAILED!"
        return 1
    fi
}

# Main execution
main() {
    # Check if jq is installed
    if ! command -v jq &> /dev/null; then
        echo "Error: jq is required but not installed. Please install jq first."
        exit 1
    fi
    
    # Check if curl is installed
    if ! command -v curl &> /dev/null; then
        echo "Error: curl is required but not installed. Please install curl first."
        exit 1
    fi
    
    # Run integration tests
    if run_integration_tests; then
        exit 0
    else
        exit 1
    fi
}

# Run main function
main "$@"
