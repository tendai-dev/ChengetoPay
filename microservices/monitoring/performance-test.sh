#!/bin/bash
# Performance Testing Script

echo "Running performance tests..."

# Load testing with Apache Bench
echo "Testing API Gateway..."
ab -n 1000 -c 10 http://localhost:8090/health

echo "Testing Payment Service..."
ab -n 500 -c 5 http://localhost:8083/health

echo "Testing Escrow Service..."
ab -n 300 -c 3 http://localhost:8081/health

# Database performance test
echo "Testing database performance..."
pgbench -h localhost -U postgres -d financial_platform -c 10 -t 100

echo "Performance testing completed"
