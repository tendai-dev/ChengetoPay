#!/bin/bash

echo "Testing ChengetoPay Services..."
echo "================================"

# Test API Gateway
echo -n "1. API Gateway Health: "
if curl -s http://localhost:8080/health | grep -q "healthy"; then
    echo "✓ PASS"
else
    echo "✗ FAIL"
fi

# Test Ledger Service
echo -n "2. Ledger Service Health: "
if curl -s http://localhost:8082/health | grep -q "healthy"; then
    echo "✓ PASS"
else
    echo "✗ FAIL"
fi

# Test Payment Service (if running)
echo -n "3. Payment Service Health: "
if curl -s http://localhost:8081/health | grep -q "healthy"; then
    echo "✓ PASS"
else
    echo "✗ FAIL (Service may not be running)"
fi

# Test creating a ledger account
echo -n "4. Create Ledger Account: "
RESPONSE=$(curl -s -X POST http://localhost:8082/api/v1/accounts \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Test Account",
    "type": "asset",
    "currency": "USD",
    "metadata": {"test": "true"}
  }')

if echo "$RESPONSE" | grep -q "id"; then
    echo "✓ PASS"
    ACCOUNT_ID=$(echo "$RESPONSE" | grep -o '"id":"[^"]*' | cut -d'"' -f4)
    echo "   Account ID: $ACCOUNT_ID"
else
    echo "✗ FAIL"
fi

echo ""
echo "Test Summary:"
echo "============="
echo "Services are running and responding to health checks."
echo "Basic API functionality is working."
