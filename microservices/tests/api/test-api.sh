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
