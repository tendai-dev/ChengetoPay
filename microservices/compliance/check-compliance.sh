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
