#!/bin/bash

# Project X Deployment Script
# Usage: ./deploy.sh [environment] [service]
# Example: ./deploy.sh production all
#          ./deploy.sh staging escrow-service

set -e

ENVIRONMENT=${1:-staging}
SERVICE=${2:-all}
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Logging functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Configuration
REGISTRY="ghcr.io/project-x"
SERVICES=("escrow-service" "payment-service" "ledger-service" "risk-service" "api-gateway")

# Environment configurations
declare -A STAGING_CONFIG=(
    ["NAMESPACE"]="projectx-staging"
    ["REPLICAS"]="1"
    ["CPU_LIMIT"]="500m"
    ["MEMORY_LIMIT"]="512Mi"
    ["DATABASE_URL"]="postgres://staging_user:staging_pass@staging-db:5432/projectx_staging"
)

declare -A PRODUCTION_CONFIG=(
    ["NAMESPACE"]="projectx-production"
    ["REPLICAS"]="3"
    ["CPU_LIMIT"]="1000m"
    ["MEMORY_LIMIT"]="1Gi"
    ["DATABASE_URL"]="postgres://prod_user:prod_pass@prod-db:5432/projectx_production"
)

# Functions
check_prerequisites() {
    log_info "Checking prerequisites..."
    
    # Check if required tools are installed
    local tools=("docker" "kubectl" "helm")
    for tool in "${tools[@]}"; do
        if ! command -v "$tool" &> /dev/null; then
            log_error "$tool is not installed"
            exit 1
        fi
    done
    
    # Check if Docker is running
    if ! docker info &> /dev/null; then
        log_error "Docker is not running"
        exit 1
    fi
    
    # Check kubectl connection
    if ! kubectl cluster-info &> /dev/null; then
        log_error "Cannot connect to Kubernetes cluster"
        exit 1
    fi
    
    log_success "Prerequisites check passed"
}

build_images() {
    local service=$1
    log_info "Building Docker images for $service..."
    
    if [ "$service" = "all" ]; then
        for svc in "${SERVICES[@]}"; do
            build_single_service "$svc"
        done
    else
        build_single_service "$service"
    fi
}

build_single_service() {
    local service=$1
    local service_dir="$PROJECT_ROOT/microservices/$service"
    
    if [ ! -d "$service_dir" ]; then
        log_error "Service directory not found: $service_dir"
        return 1
    fi
    
    log_info "Building $service..."
    
    # Build Docker image
    docker build -t "$REGISTRY/$service:latest" "$service_dir"
    docker build -t "$REGISTRY/$service:$(git rev-parse --short HEAD)" "$service_dir"
    
    # Push images
    docker push "$REGISTRY/$service:latest"
    docker push "$REGISTRY/$service:$(git rev-parse --short HEAD)"
    
    log_success "Built and pushed $service"
}

deploy_to_kubernetes() {
    local environment=$1
    local service=$2
    
    log_info "Deploying to $environment environment..."
    
    # Set configuration based on environment
    local -n config="${environment^^}_CONFIG"
    local namespace="${config[NAMESPACE]}"
    local replicas="${config[REPLICAS]}"
    local cpu_limit="${config[CPU_LIMIT]}"
    local memory_limit="${config[MEMORY_LIMIT]}"
    local database_url="${config[DATABASE_URL]}"
    
    # Create namespace if it doesn't exist
    kubectl create namespace "$namespace" --dry-run=client -o yaml | kubectl apply -f -
    
    if [ "$service" = "all" ]; then
        for svc in "${SERVICES[@]}"; do
            deploy_single_service "$svc" "$namespace" "$replicas" "$cpu_limit" "$memory_limit" "$database_url"
        done
    else
        deploy_single_service "$service" "$namespace" "$replicas" "$cpu_limit" "$memory_limit" "$database_url"
    fi
}

deploy_single_service() {
    local service=$1
    local namespace=$2
    local replicas=$3
    local cpu_limit=$4
    local memory_limit=$5
    local database_url=$6
    
    log_info "Deploying $service to namespace $namespace..."
    
    # Generate Kubernetes manifests
    cat <<EOF | kubectl apply -f -
apiVersion: apps/v1
kind: Deployment
metadata:
  name: $service
  namespace: $namespace
  labels:
    app: $service
    version: $(git rev-parse --short HEAD)
spec:
  replicas: $replicas
  selector:
    matchLabels:
      app: $service
  template:
    metadata:
      labels:
        app: $service
        version: $(git rev-parse --short HEAD)
    spec:
      containers:
      - name: $service
        image: $REGISTRY/$service:$(git rev-parse --short HEAD)
        ports:
        - containerPort: 8080
        env:
        - name: DATABASE_URL
          value: "$database_url"
        - name: ENVIRONMENT
          value: "$ENVIRONMENT"
        resources:
          limits:
            cpu: $cpu_limit
            memory: $memory_limit
          requests:
            cpu: $(echo "$cpu_limit" | sed 's/m$//' | awk '{print int($1/2)"m"}')
            memory: $(echo "$memory_limit" | sed 's/Mi$//' | awk '{print int($1/2)"Mi"}')
        livenessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 5
---
apiVersion: v1
kind: Service
metadata:
  name: $service
  namespace: $namespace
  labels:
    app: $service
spec:
  selector:
    app: $service
  ports:
  - port: 80
    targetPort: 8080
    protocol: TCP
  type: ClusterIP
EOF
    
    # Wait for deployment to be ready
    kubectl rollout status deployment/$service -n "$namespace" --timeout=300s
    
    log_success "Deployed $service successfully"
}

run_health_checks() {
    local namespace=$1
    log_info "Running health checks..."
    
    # Wait for all pods to be ready
    kubectl wait --for=condition=ready pod -l app -n "$namespace" --timeout=300s
    
    # Check service endpoints
    local services=$(kubectl get services -n "$namespace" -o jsonpath='{.items[*].metadata.name}')
    for service in $services; do
        log_info "Checking health of $service..."
        
        # Port forward and check health endpoint
        kubectl port-forward -n "$namespace" "service/$service" 8080:80 &
        local port_forward_pid=$!
        
        sleep 5
        
        if curl -f http://localhost:8080/health &> /dev/null; then
            log_success "$service is healthy"
        else
            log_error "$service health check failed"
        fi
        
        kill $port_forward_pid 2>/dev/null || true
    done
}

rollback_deployment() {
    local environment=$1
    local service=$2
    
    log_warning "Rolling back $service in $environment..."
    
    local -n config="${environment^^}_CONFIG"
    local namespace="${config[NAMESPACE]}"
    
    if [ "$service" = "all" ]; then
        for svc in "${SERVICES[@]}"; do
            kubectl rollout undo deployment/$svc -n "$namespace"
        done
    else
        kubectl rollout undo deployment/$service -n "$namespace"
    fi
    
    log_success "Rollback completed"
}

cleanup_old_images() {
    log_info "Cleaning up old Docker images..."
    
    # Remove images older than 7 days
    docker image prune -a --filter "until=168h" -f
    
    log_success "Cleanup completed"
}

# Main deployment flow
main() {
    log_info "Starting deployment of $SERVICE to $ENVIRONMENT environment"
    
    # Validate inputs
    if [[ ! " staging production " =~ " $ENVIRONMENT " ]]; then
        log_error "Invalid environment: $ENVIRONMENT. Use 'staging' or 'production'"
        exit 1
    fi
    
    if [[ "$SERVICE" != "all" ]] && [[ ! " ${SERVICES[*]} " =~ " $SERVICE " ]]; then
        log_error "Invalid service: $SERVICE. Use 'all' or one of: ${SERVICES[*]}"
        exit 1
    fi
    
    # Run deployment steps
    check_prerequisites
    build_images "$SERVICE"
    deploy_to_kubernetes "$ENVIRONMENT" "$SERVICE"
    run_health_checks "${ENVIRONMENT^^}_CONFIG[NAMESPACE]"
    cleanup_old_images
    
    log_success "Deployment completed successfully!"
    log_info "Services are available in namespace: ${ENVIRONMENT^^}_CONFIG[NAMESPACE]"
}

# Handle script arguments
case "${1:-}" in
    "rollback")
        rollback_deployment "$2" "$3"
        ;;
    "health")
        run_health_checks "${2^^}_CONFIG[NAMESPACE]"
        ;;
    "cleanup")
        cleanup_old_images
        ;;
    *)
        main
        ;;
esac
