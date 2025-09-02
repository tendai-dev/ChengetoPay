#!/bin/bash

# Fix all Dockerfiles to use main.go instead of .

services=(
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

for service in "${services[@]}"; do
  if [ -f "$service/Dockerfile" ]; then
    # Fix the build command in Dockerfile
    sed -i '' "s|RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o $service \./$service|RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o $service .|g" "$service/Dockerfile"
    sed -i '' "s|RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o $service \.|RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o $service main.go|g" "$service/Dockerfile"
    echo "Fixed Dockerfile for $service"
  fi
done

echo "All Dockerfiles fixed!"
