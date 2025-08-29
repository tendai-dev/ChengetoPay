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
