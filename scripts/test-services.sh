#!/bin/bash

# Test all services health endpoints
echo "🔍 Testing all microservices health endpoints..."

services=(
    "api-gateway:8080"
    "escrow-service:8081"
    "payment-service:8082"
    "ledger-service:8083"
    "journal-service:8084"
    "fees-service:8085"
    "refunds-service:8086"
    "transfers-service:8087"
    "payouts-service:8088"
    "reserves-service:8089"
    "reconciliation-service:8090"
    "treasury-service:8091"
    "risk-service:8092"
    "disputes-service:8093"
    "auth-service:8094"
    "compliance-service:8095"
)

failed=0
passed=0

for service in "${services[@]}"; do
    name="${service%:*}"
    port="${service#*:}"
    
    echo -n "Testing $name on port $port... "
    
    if curl -s -f "http://localhost:$port/health" > /dev/null 2>&1; then
        echo "✅ PASSED"
        ((passed++))
    else
        echo "❌ FAILED"
        ((failed++))
    fi
done

echo ""
echo "📊 Results: $passed passed, $failed failed"

if [ $failed -eq 0 ]; then
    echo "🎉 All services are healthy!"
    exit 0
else
    echo "⚠️  Some services are not responding"
    exit 1
fi
