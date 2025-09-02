#!/bin/bash

# Build services in batches to avoid overwhelming the system

echo "Building batch 1: Core financial services..."
docker-compose -f docker-compose-complete.yml build \
  risk-service \
  treasury-service \
  evidence-service \
  compliance-service \
  workflow-service

echo "Building batch 2: Transaction services..."
docker-compose -f docker-compose-complete.yml build \
  journal-service \
  fees-service \
  refunds-service \
  transfers-service \
  fx-service

echo "Building batch 3: Operation services..."
docker-compose -f docker-compose-complete.yml build \
  payouts-service \
  reserves-service \
  reconciliation-service \
  kyb-service \
  sca-service

echo "Building batch 4: Support services..."
docker-compose -f docker-compose-complete.yml build \
  disputes-service \
  dx-service \
  auth-service \
  idempotency-service \
  eventbus-service

echo "Building batch 5: Infrastructure services..."
docker-compose -f docker-compose-complete.yml build \
  saga-service \
  webhooks-service \
  observability-service \
  config-service \
  workers-service

echo "Building batch 6: Platform services..."
docker-compose -f docker-compose-complete.yml build \
  portal-service \
  data-platform-service \
  compliance-ops-service

echo "All services built!"
