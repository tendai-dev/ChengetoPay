#!/bin/bash

# Advanced Monitoring Setup Script
echo "📊 SETTING UP ADVANCED MONITORING"

# Color functions
print_status() { echo -e "\033[1;34m[SETUP]\033[0m $1"; }
print_success() { echo -e "\033[1;32m[SUCCESS]\033[0m $1"; }

# Setup Alert Manager
setup_alertmanager() {
    print_status "Setting up Alert Manager..."
    mkdir -p alertmanager
    cat > alertmanager/alertmanager.yml << EOF
global:
  resolve_timeout: 5m
  slack_api_url: 'https://hooks.slack.com/services/YOUR_SLACK_WEBHOOK'

route:
  group_by: ['alertname', 'service', 'severity']
  group_wait: 10s
  group_interval: 10s
  repeat_interval: 1h
  receiver: 'slack-notifications'
  routes:
    - match:
        severity: critical
      receiver: 'pager-duty-critical'
      continue: true
      group_wait: 0s
      group_interval: 5s
      repeat_interval: 30m

receivers:
  - name: 'slack-notifications'
    slack_configs:
      - channel: '#alerts'
        title: '{{ template "slack.title" . }}'
        text: '{{ template "slack.text" . }}'
        send_resolved: true
        icon_emoji: ':warning:'

  - name: 'pager-duty-critical'
    pagerduty_configs:
      - routing_key: 'your-pagerduty-routing-key'
        description: '{{ template "pagerduty.description" . }}'
        severity: '{{ if eq .CommonLabels.severity "critical" }}critical{{ else }}warning{{ end }}'
EOF
    print_success "Alert Manager configured"
}

# Setup Distributed Tracing
setup_tracing() {
    print_status "Setting up Distributed Tracing..."
    mkdir -p tracing
    cat > tracing/docker-compose.yml << EOF
version: '3.8'
services:
  jaeger:
    image: jaegertracing/all-in-one:latest
    container_name: jaeger
    ports:
      - "16686:16686"
      - "14268:14268"
      - "14250:14250"
    environment:
      - COLLECTOR_OTLP_ENABLED=true
      - SPAN_STORAGE_TYPE=memory
    restart: unless-stopped
EOF
    print_success "Distributed Tracing configured"
}

# Setup Log Aggregation
setup_log_aggregation() {
    print_status "Setting up Log Aggregation..."
    mkdir -p elk/logstash/pipeline
    cat > elk/logstash/pipeline/logstash.conf << EOF
input {
  beats { port => 5044 }
  tcp { port => 5000; codec => json }
}

filter {
  if [service] { mutate { add_tag => ["financial-platform"] } }
  if [level] == "ERROR" { mutate { add_tag => ["error"] } }
  if [level] == "CRITICAL" { mutate { add_tag => ["critical"] } }
}

output {
  elasticsearch {
    hosts => ["elasticsearch:9200"]
    index => "financial-platform-%{+YYYY.MM.dd}"
  }
}
EOF
    print_success "Log Aggregation configured"
}

# Setup Monitoring Docker Compose
setup_monitoring_compose() {
    print_status "Setting up Monitoring Docker Compose..."
    cat > docker-compose-monitoring.yml << EOF
version: '3.8'
services:
  prometheus:
    image: prom/prometheus:latest
    container_name: prometheus
    ports: ["9090:9090"]
    volumes:
      - ./metrics/prometheus.yml:/etc/prometheus/prometheus.yml
      - prometheus_data:/prometheus
    restart: unless-stopped

  alertmanager:
    image: prom/alertmanager:latest
    container_name: alertmanager
    ports: ["9093:9093"]
    volumes:
      - ./alertmanager/alertmanager.yml:/etc/alertmanager/alertmanager.yml
      - alertmanager_data:/alertmanager
    restart: unless-stopped

  grafana:
    image: grafana/grafana:latest
    container_name: grafana
    ports: ["3000:3000"]
    environment:
      - GF_SECURITY_ADMIN_PASSWORD=admin123
    restart: unless-stopped

  jaeger:
    image: jaegertracing/all-in-one:latest
    container_name: jaeger
    ports: ["16686:16686"]
    environment:
      - COLLECTOR_OTLP_ENABLED=true
      - SPAN_STORAGE_TYPE=memory
    restart: unless-stopped

  elasticsearch:
    image: docker.elastic.co/elasticsearch/elasticsearch:8.8.0
    container_name: elasticsearch
    environment:
      - discovery.type=single-node
      - xpack.security.enabled=false
    ports: ["9200:9200"]
    restart: unless-stopped

  kibana:
    image: docker.elastic.co/kibana/kibana:8.8.0
    container_name: kibana
    ports: ["5601:5601"]
    environment:
      ELASTICSEARCH_HOSTS: http://elasticsearch:9200
    restart: unless-stopped

volumes:
  prometheus_data:
  alertmanager_data:
EOF
    print_success "Monitoring Docker Compose configured"
}

# Main setup
main() {
    echo "📊 ADVANCED MONITORING SETUP"
    echo "============================"
    
    setup_alertmanager
    setup_tracing
    setup_log_aggregation
    setup_monitoring_compose
    
    echo ""
    echo "✅ ADVANCED MONITORING SETUP COMPLETE"
    echo "📊 COMPONENTS: Alert Manager, Jaeger, ELK Stack, Grafana"
    echo "🌐 ACCESS: Prometheus:9090, Grafana:3000, Jaeger:16686, Kibana:5601"
    echo "🚀 START: docker-compose -f docker-compose-monitoring.yml up -d"
}

main "$@"
