#!/bin/bash

# CI/CD Setup Script for Financial Platform
echo "🚀 SETTING UP CI/CD PIPELINE"
echo "============================"

# Color functions
print_status() {
    echo -e "\033[1;34m[SETUP]\033[0m $1"
}

print_success() {
    echo -e "\033[1;32m[SUCCESS]\033[0m $1"
}

print_error() {
    echo -e "\033[1;31m[ERROR]\033[0m $1"
}

print_warning() {
    echo -e "\033[1;33m[WARNING]\033[0m $1"
}

# Check if running as root
if [[ $EUID -eq 0 ]]; then
   print_error "This script should not be run as root"
   exit 1
fi

# Check prerequisites
check_prerequisites() {
    print_status "Checking prerequisites..."
    
    # Check Docker
    if ! command -v docker &> /dev/null; then
        print_error "Docker is not installed. Please install Docker first."
        exit 1
    fi
    
    # Check Docker Compose
    if ! command -v docker-compose &> /dev/null; then
        print_error "Docker Compose is not installed. Please install Docker Compose first."
        exit 1
    fi
    
    # Check kubectl
    if ! command -v kubectl &> /dev/null; then
        print_warning "kubectl is not installed. Kubernetes features will be disabled."
        KUBERNETES_ENABLED=false
    else
        KUBERNETES_ENABLED=true
    fi
    
    # Check Helm
    if ! command -v helm &> /dev/null; then
        print_warning "Helm is not installed. Helm features will be disabled."
        HELM_ENABLED=false
    else
        HELM_ENABLED=true
    fi
    
    # Check k6
    if ! command -v k6 &> /dev/null; then
        print_warning "k6 is not installed. Performance testing will be disabled."
        K6_ENABLED=false
    else
        K6_ENABLED=true
    fi
    
    print_success "Prerequisites check completed"
}

# Setup Docker Registry
setup_docker_registry() {
    print_status "Setting up Docker Registry..."
    
    if [ -f "registry-config.yml" ]; then
        docker-compose -f registry-config.yml up -d
        print_success "Docker Registry started"
        echo "Registry UI: http://localhost:8080"
        echo "Registry API: http://localhost:5000"
    else
        print_error "registry-config.yml not found"
        return 1
    fi
}

# Setup ArgoCD
setup_argocd() {
    if [ "$KUBERNETES_ENABLED" = true ]; then
        print_status "Setting up ArgoCD..."
        
        # Create namespace
        kubectl create namespace argocd --dry-run=client -o yaml | kubectl apply -f -
        
        # Install ArgoCD
        kubectl apply -n argocd -f https://raw.githubusercontent.com/argoproj/argo-cd/stable/manifests/install.yaml
        
        # Wait for ArgoCD to be ready
        print_status "Waiting for ArgoCD to be ready..."
        kubectl wait --for=condition=available --timeout=300s deployment/argocd-server -n argocd
        
        # Get initial admin password
        ARGOCD_PASSWORD=$(kubectl -n argocd get secret argocd-initial-admin-secret -o jsonpath="{.data.password}" | base64 -d)
        
        print_success "ArgoCD setup completed"
        echo "ArgoCD UI: http://localhost:8080"
        echo "Username: admin"
        echo "Password: $ARGOCD_PASSWORD"
    else
        print_warning "Skipping ArgoCD setup (Kubernetes not available)"
    fi
}

# Setup Jenkins
setup_jenkins() {
    print_status "Setting up Jenkins..."
    
    # Create Jenkins configuration
    cat > jenkins-config.yml << EOF
version: '3.8'
services:
  jenkins:
    image: jenkins/jenkins:lts-jdk17
    container_name: jenkins
    ports:
      - "8080:8080"
      - "50000:50000"
    volumes:
      - jenkins_home:/var/jenkins_home
      - /var/run/docker.sock:/var/run/docker.sock
    environment:
      - JENKINS_OPTS=--httpPort=8080
    networks:
      - jenkins_network

volumes:
  jenkins_home:

networks:
  jenkins_network:
    driver: bridge
EOF
    
    docker-compose -f jenkins-config.yml up -d
    
    print_success "Jenkins setup completed"
    echo "Jenkins UI: http://localhost:8080"
    echo "Initial admin password: Check docker logs jenkins"
}

# Setup SonarQube
setup_sonarqube() {
    print_status "Setting up SonarQube..."
    
    # Create SonarQube configuration
    cat > sonarqube-config.yml << EOF
version: '3.8'
services:
  sonarqube:
    image: sonarqube:community
    container_name: sonarqube
    ports:
      - "9000:9000"
    environment:
      - SONAR_ES_BOOTSTRAP_CHECKS_DISABLE=true
    volumes:
      - sonarqube_data:/opt/sonarqube/data
      - sonarqube_extensions:/opt/sonarqube/extensions
      - sonarqube_logs:/opt/sonarqube/logs
    networks:
      - sonarqube_network

volumes:
  sonarqube_data:
  sonarqube_extensions:
  sonarqube_logs:

networks:
  sonarqube_network:
    driver: bridge
EOF
    
    docker-compose -f sonarqube-config.yml up -d
    
    print_success "SonarQube setup completed"
    echo "SonarQube UI: http://localhost:9000"
    echo "Default credentials: admin/admin"
}

# Setup Nexus Repository
setup_nexus() {
    print_status "Setting up Nexus Repository..."
    
    # Create Nexus configuration
    cat > nexus-config.yml << EOF
version: '3.8'
services:
  nexus:
    image: sonatype/nexus3:latest
    container_name: nexus
    ports:
      - "8081:8081"
    volumes:
      - nexus_data:/nexus-data
    environment:
      - NEXUS_SECURITY_RANDOMPASSWORD=false
    networks:
      - nexus_network

volumes:
  nexus_data:

networks:
  nexus_network:
    driver: bridge
EOF
    
    docker-compose -f nexus-config.yml up -d
    
    print_success "Nexus Repository setup completed"
    echo "Nexus UI: http://localhost:8081"
    echo "Default credentials: admin/admin123"
}

# Setup monitoring for CI/CD
setup_cicd_monitoring() {
    print_status "Setting up CI/CD monitoring..."
    
    # Create monitoring configuration
    cat > cicd-monitoring.yml << EOF
version: '3.8'
services:
  # Prometheus for CI/CD metrics
  prometheus-cicd:
    image: prom/prometheus:latest
    container_name: prometheus-cicd
    ports:
      - "9091:9090"
    volumes:
      - ./cicd-prometheus.yml:/etc/prometheus/prometheus.yml
      - prometheus_cicd_data:/prometheus
    command:
      - '--config.file=/etc/prometheus/prometheus.yml'
      - '--storage.tsdb.path=/prometheus'
    networks:
      - cicd_monitoring_network

  # Grafana for CI/CD dashboards
  grafana-cicd:
    image: grafana/grafana:latest
    container_name: grafana-cicd
    ports:
      - "3001:3000"
    environment:
      GF_SECURITY_ADMIN_PASSWORD: admin
      GF_USERS_ALLOW_SIGN_UP: false
    volumes:
      - grafana_cicd_data:/var/lib/grafana
    networks:
      - cicd_monitoring_network

volumes:
  prometheus_cicd_data:
  grafana_cicd_data:

networks:
  cicd_monitoring_network:
    driver: bridge
EOF
    
    # Create Prometheus configuration for CI/CD
    cat > cicd-prometheus.yml << EOF
global:
  scrape_interval: 15s

scrape_configs:
  - job_name: 'jenkins'
    static_configs:
      - targets: ['jenkins:8080']
    metrics_path: /prometheus

  - job_name: 'sonarqube'
    static_configs:
      - targets: ['sonarqube:9000']
    metrics_path: /api/metrics

  - job_name: 'nexus'
    static_configs:
      - targets: ['nexus:8081']
    metrics_path: /service/metrics/prometheus
EOF
    
    docker-compose -f cicd-monitoring.yml up -d
    
    print_success "CI/CD monitoring setup completed"
    echo "CI/CD Prometheus: http://localhost:9091"
    echo "CI/CD Grafana: http://localhost:3001 (admin/admin)"
}

# Create CI/CD configuration files
create_cicd_configs() {
    print_status "Creating CI/CD configuration files..."
    
    # Create .gitignore for CI/CD
    cat > .gitignore << EOF
# CI/CD
*.log
*.tmp
*.cache
.env
.secrets

# Build artifacts
dist/
build/
*.tar.gz
*.zip

# Test results
coverage/
test-results/
performance-results/

# Docker
.dockerignore

# IDE
.vscode/
.idea/
*.swp
*.swo

# OS
.DS_Store
Thumbs.db
EOF
    
    # Create Docker Compose override for CI/CD
    cat > docker-compose.override.yml << EOF
version: '3.8'

services:
  # Override services for CI/CD environment
  api-gateway:
    environment:
      - ENVIRONMENT=ci
      - LOG_LEVEL=debug
    volumes:
      - ./logs:/app/logs

  database-service:
    environment:
      - ENVIRONMENT=ci
      - LOG_LEVEL=debug
    volumes:
      - ./logs:/app/logs

  monitoring-service:
    environment:
      - ENVIRONMENT=ci
      - LOG_LEVEL=debug
    volumes:
      - ./logs:/app/logs
EOF
    
    print_success "CI/CD configuration files created"
}

# Setup security scanning
setup_security_scanning() {
    print_status "Setting up security scanning..."
    
    # Create Trivy configuration
    cat > trivy-config.yml << EOF
version: '3.8'
services:
  trivy-server:
    image: aquasec/trivy:latest
    container_name: trivy-server
    ports:
      - "8082:8080"
    command: server --listen 0.0.0.0:8080
    networks:
      - security_network

networks:
  security_network:
    driver: bridge
EOF
    
    docker-compose -f trivy-config.yml up -d
    
    print_success "Security scanning setup completed"
    echo "Trivy Server: http://localhost:8082"
}

# Main setup function
main() {
    echo ""
    echo "🚀 FINANCIAL PLATFORM CI/CD SETUP"
    echo "================================="
    echo ""
    
    # Check prerequisites
    check_prerequisites
    
    echo ""
    echo "🔧 SETTING UP CI/CD COMPONENTS"
    echo "=============================="
    
    # Setup components
    setup_docker_registry
    setup_jenkins
    setup_sonarqube
    setup_nexus
    setup_cicd_monitoring
    setup_security_scanning
    
    if [ "$KUBERNETES_ENABLED" = true ]; then
        setup_argocd
    fi
    
    # Create configuration files
    create_cicd_configs
    
    echo ""
    echo "✅ CI/CD SETUP COMPLETE"
    echo "======================="
    echo ""
    echo "🌐 ACCESS POINTS:"
    echo "  • Jenkins: http://localhost:8080"
    echo "  • SonarQube: http://localhost:9000 (admin/admin)"
    echo "  • Nexus Repository: http://localhost:8081 (admin/admin123)"
    echo "  • Docker Registry UI: http://localhost:8080"
    echo "  • CI/CD Prometheus: http://localhost:9091"
    echo "  • CI/CD Grafana: http://localhost:3001 (admin/admin)"
    echo "  • Trivy Security Scanner: http://localhost:8082"
    if [ "$KUBERNETES_ENABLED" = true ]; then
        echo "  • ArgoCD: http://localhost:8080 (admin/[password from logs])"
    fi
    echo ""
    echo "📋 NEXT STEPS:"
    echo "1. Configure Jenkins pipelines"
    echo "2. Set up SonarQube quality gates"
    echo "3. Configure Nexus repositories"
    echo "4. Set up ArgoCD applications (if Kubernetes enabled)"
    echo "5. Configure security scanning policies"
    echo "6. Set up monitoring dashboards"
    echo ""
    echo "🎯 CI/CD PIPELINE READY FOR USE!"
}

# Run main function
main "$@"
