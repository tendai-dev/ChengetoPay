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
