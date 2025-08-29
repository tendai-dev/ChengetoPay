#!/bin/bash

# Project X Monitoring Setup Script
# Sets up Prometheus, Grafana, and Alertmanager for monitoring

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

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

# Create monitoring namespace
setup_namespace() {
    log_info "Setting up monitoring namespace..."
    
    kubectl create namespace monitoring --dry-run=client -o yaml | kubectl apply -f -
    
    log_success "Monitoring namespace created"
}

# Deploy Prometheus
deploy_prometheus() {
    log_info "Deploying Prometheus..."
    
    cat <<EOF | kubectl apply -f -
apiVersion: v1
kind: ConfigMap
metadata:
  name: prometheus-config
  namespace: monitoring
data:
  prometheus.yml: |
    global:
      scrape_interval: 15s
      evaluation_interval: 15s
    
    rule_files:
      - "alert_rules.yml"
    
    alerting:
      alertmanagers:
        - static_configs:
            - targets:
              - alertmanager:9093
    
    scrape_configs:
      - job_name: 'prometheus'
        static_configs:
          - targets: ['localhost:9090']
      
      - job_name: 'api-gateway'
        static_configs:
          - targets: ['api-gateway:8090']
        metrics_path: /metrics
      
      - job_name: 'escrow-service'
        static_configs:
          - targets: ['escrow-service:8081']
        metrics_path: /metrics
      
      - job_name: 'payment-service'
        static_configs:
          - targets: ['payment-service:8083']
        metrics_path: /metrics
      
      - job_name: 'ledger-service'
        static_configs:
          - targets: ['ledger-service:8084']
        metrics_path: /metrics
      
      - job_name: 'risk-service'
        static_configs:
          - targets: ['risk-service:8085']
        metrics_path: /metrics

  alert_rules.yml: |
    groups:
    - name: projectx_alerts
      rules:
      - alert: ServiceDown
        expr: up == 0
        for: 1m
        labels:
          severity: critical
        annotations:
          summary: "Service {{ \$labels.instance }} is down"
          description: "{{ \$labels.instance }} has been down for more than 1 minute."
      
      - alert: HighErrorRate
        expr: rate(http_requests_total{status=~"5.."}[5m]) > 0.1
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "High error rate on {{ \$labels.instance }}"
          description: "Error rate is {{ \$value }} errors per second."
      
      - alert: HighLatency
        expr: histogram_quantile(0.95, rate(http_request_duration_seconds_bucket[5m])) > 0.02
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "High latency on {{ \$labels.instance }}"
          description: "95th percentile latency is {{ \$value }} seconds."
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: prometheus
  namespace: monitoring
spec:
  replicas: 1
  selector:
    matchLabels:
      app: prometheus
  template:
    metadata:
      labels:
        app: prometheus
    spec:
      containers:
      - name: prometheus
        image: prom/prometheus:latest
        ports:
        - containerPort: 9090
        volumeMounts:
        - name: config
          mountPath: /etc/prometheus
        args:
          - '--config.file=/etc/prometheus/prometheus.yml'
          - '--storage.tsdb.path=/prometheus'
          - '--web.console.libraries=/etc/prometheus/console_libraries'
          - '--web.console.templates=/etc/prometheus/consoles'
          - '--storage.tsdb.retention.time=200h'
          - '--web.enable-lifecycle'
      volumes:
      - name: config
        configMap:
          name: prometheus-config
---
apiVersion: v1
kind: Service
metadata:
  name: prometheus
  namespace: monitoring
spec:
  selector:
    app: prometheus
  ports:
  - port: 9090
    targetPort: 9090
  type: LoadBalancer
EOF

    log_success "Prometheus deployed"
}

# Deploy Grafana
deploy_grafana() {
    log_info "Deploying Grafana..."
    
    cat <<EOF | kubectl apply -f -
apiVersion: v1
kind: ConfigMap
metadata:
  name: grafana-datasources
  namespace: monitoring
data:
  prometheus.yaml: |
    apiVersion: 1
    datasources:
    - name: Prometheus
      type: prometheus
      url: http://prometheus:9090
      access: proxy
      isDefault: true
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: grafana
  namespace: monitoring
spec:
  replicas: 1
  selector:
    matchLabels:
      app: grafana
  template:
    metadata:
      labels:
        app: grafana
    spec:
      containers:
      - name: grafana
        image: grafana/grafana:latest
        ports:
        - containerPort: 3000
        env:
        - name: GF_SECURITY_ADMIN_PASSWORD
          value: "admin123"
        volumeMounts:
        - name: datasources
          mountPath: /etc/grafana/provisioning/datasources
      volumes:
      - name: datasources
        configMap:
          name: grafana-datasources
---
apiVersion: v1
kind: Service
metadata:
  name: grafana
  namespace: monitoring
spec:
  selector:
    app: grafana
  ports:
  - port: 3000
    targetPort: 3000
  type: LoadBalancer
EOF

    log_success "Grafana deployed"
}

# Deploy Alertmanager
deploy_alertmanager() {
    log_info "Deploying Alertmanager..."
    
    cat <<EOF | kubectl apply -f -
apiVersion: v1
kind: ConfigMap
metadata:
  name: alertmanager-config
  namespace: monitoring
data:
  alertmanager.yml: |
    global:
      smtp_smarthost: 'localhost:587'
      smtp_from: 'alerts@projectx.com'
    
    route:
      group_by: ['alertname']
      group_wait: 10s
      group_interval: 10s
      repeat_interval: 1h
      receiver: 'web.hook'
    
    receivers:
    - name: 'web.hook'
      webhook_configs:
      - url: 'http://localhost:5001/'
        send_resolved: true
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: alertmanager
  namespace: monitoring
spec:
  replicas: 1
  selector:
    matchLabels:
      app: alertmanager
  template:
    metadata:
      labels:
        app: alertmanager
    spec:
      containers:
      - name: alertmanager
        image: prom/alertmanager:latest
        ports:
        - containerPort: 9093
        volumeMounts:
        - name: config
          mountPath: /etc/alertmanager
        args:
          - '--config.file=/etc/alertmanager/alertmanager.yml'
          - '--storage.path=/alertmanager'
      volumes:
      - name: config
        configMap:
          name: alertmanager-config
---
apiVersion: v1
kind: Service
metadata:
  name: alertmanager
  namespace: monitoring
spec:
  selector:
    app: alertmanager
  ports:
  - port: 9093
    targetPort: 9093
  type: ClusterIP
EOF

    log_success "Alertmanager deployed"
}

# Main setup function
main() {
    log_info "Setting up monitoring stack for Project X..."
    
    setup_namespace
    deploy_prometheus
    deploy_grafana
    deploy_alertmanager
    
    log_success "Monitoring stack deployed successfully!"
    log_info "Access Grafana at: kubectl port-forward -n monitoring service/grafana 3000:3000"
    log_info "Access Prometheus at: kubectl port-forward -n monitoring service/prometheus 9090:9090"
    log_info "Default Grafana credentials: admin/admin123"
}

main "$@"
