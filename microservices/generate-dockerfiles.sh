#!/bin/bash

# Generate Dockerfiles for all microservices
echo "🐳 Generating Dockerfiles for all microservices..."

# Define all services with their ports
services=(
    "escrow-service:8081"
    "payment-service:8083"
    "ledger-service:8084"
    "risk-service:8085"
    "treasury-service:8086"
    "evidence-service:8087"
    "compliance-service:8088"
    "workflow-service:8089"
    "journal-service:8091"
    "fees-service:8092"
    "refunds-service:8093"
    "transfers-service:8094"
    "fx-service:8095"
    "payouts-service:8096"
    "reserves-service:8097"
    "reconciliation-service:8098"
    "kyb-service:8099"
    "sca-service:8100"
    "disputes-service:8101"
    "dx-service:8102"
    "auth-service:8103"
    "idempotency-service:8104"
    "eventbus-service:8105"
    "saga-service:8106"
    "vault-service:8107"
    "webhooks-service:8108"
    "observability-service:8109"
    "config-service:8110"
    "workers-service:8111"
    "portal-service:8112"
    "data-platform-service:8113"
    "compliance-ops-service:8114"
    "database-service:8115"
    "monitoring-service:8116"
    "message-queue-service:8117"
    "service-discovery:8118"
)

# Generate Dockerfile for each service
for service_entry in "${services[@]}"; do
    service=$(echo "$service_entry" | cut -d: -f1)
    port=$(echo "$service_entry" | cut -d: -f2)
    
    echo "Creating Dockerfile for $service (port $port)..."
    
    cat > "$service/Dockerfile" << EOF
# $service Dockerfile
FROM golang:1.24-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o $service ./$service

# Final stage
FROM alpine:latest

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates tzdata

# Create non-root user
RUN addgroup -g 1001 -S appgroup && \\
    adduser -u 1001 -S appuser -G appgroup

WORKDIR /app

# Copy binary from builder stage
COPY --from=builder /app/$service .

# Change ownership to non-root user
RUN chown -R appuser:appgroup /app

# Switch to non-root user
USER appuser

# Expose port
EXPOSE $port

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \\
    CMD wget --no-verbose --tries=1 --spider http://localhost:$port/health || exit 1

# Run the binary
CMD ["./$service", "-port=$port"]
EOF

    echo "✅ Created Dockerfile for $service"
done

echo ""
echo "🎯 All Dockerfiles generated successfully!"
echo "📁 Total Dockerfiles created: ${#services[@]}"
echo ""
echo "Next steps:"
echo "1. Run 'docker-compose build' to build all images"
echo "2. Run 'docker-compose up -d' to start all services"
echo "3. Run 'docker-compose ps' to check service status"
