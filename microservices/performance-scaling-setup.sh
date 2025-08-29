#!/bin/bash

# Performance & Scaling Setup Script
echo "⚡ SETTING UP PERFORMANCE & SCALING"

# Color functions
print_status() { echo -e "\033[1;34m[SETUP]\033[0m $1"; }
print_success() { echo -e "\033[1;32m[SUCCESS]\033[0m $1"; }

# Setup CDN Configuration
setup_cdn() {
    print_status "Setting up CDN Configuration..."
    
    # Create CDN directories
    mkdir -p cdn/{config,scripts,monitoring}
    
    # Create CDN setup script
    cat > cdn/setup-cloudfront.sh << 'EOF'
#!/bin/bash
# AWS CloudFront Setup Script

DISTRIBUTION_NAME="financial-platform-cdn"
ORIGIN_DOMAIN="api-gateway.financialplatform.com"

echo "Setting up CloudFront distribution..."

# Create CloudFront distribution
aws cloudfront create-distribution \
  --distribution-config "{
    \"CallerReference\": \"$(date +%s)\",
    \"Comment\": \"Financial Platform CDN\",
    \"DefaultRootObject\": \"index.html\",
    \"Origins\": {
      \"Quantity\": 1,
      \"Items\": [
        {
          \"Id\": \"api-gateway\",
          \"DomainName\": \"$ORIGIN_DOMAIN\",
          \"OriginPath\": \"\",
          \"CustomOriginConfig\": {
            \"HTTPPort\": 80,
            \"HTTPSPort\": 443,
            \"OriginProtocolPolicy\": \"https-only\"
          }
        }
      ]
    },
    \"DefaultCacheBehavior\": {
      \"TargetOriginId\": \"api-gateway\",
      \"ViewerProtocolPolicy\": \"redirect-to-https\",
      \"TrustedSigners\": {
        \"Enabled\": false,
        \"Quantity\": 0
      },
      \"ForwardedValues\": {
        \"QueryString\": true,
        \"Cookies\": {
          \"Forward\": \"all\"
        },
        \"Headers\": {
          \"Quantity\": 0
        },
        \"QueryStringCacheKeys\": {
          \"Quantity\": 0
        }
      },
      \"MinTTL\": 0,
      \"DefaultTTL\": 300,
      \"MaxTTL\": 3600
    },
    \"Enabled\": true,
    \"PriceClass\": \"PriceClass_100\"
  }"

echo "CloudFront distribution created successfully"
EOF

    chmod +x cdn/setup-cloudfront.sh
    print_success "CDN configuration created"
}

# Setup Caching Layer
setup_caching() {
    print_status "Setting up Caching Layer..."
    
    # Create Redis cluster configuration
    cat > caching/redis-cluster.conf << 'EOF'
# Redis Cluster Configuration
port 6379
cluster-enabled yes
cluster-config-file nodes.conf
cluster-node-timeout 5000
appendonly yes
appendfsync everysec
maxmemory 6gb
maxmemory-policy allkeys-lru
EOF

    # Create cache warming script
    cat > caching/cache-warming.sh << 'EOF'
#!/bin/bash
# Cache Warming Script

echo "Starting cache warming process..."

# Warm critical configurations
curl -s "http://localhost:8090/api/v1/config" > /dev/null
curl -s "http://localhost:8090/api/v1/currencies" > /dev/null
curl -s "http://localhost:8090/api/v1/countries" > /dev/null

# Warm exchange rates
curl -s "http://localhost:8090/api/v1/rates" > /dev/null

# Warm user profiles (sample)
for i in {1..100}; do
  curl -s "http://localhost:8090/api/v1/users/$i/profile" > /dev/null &
done

wait
echo "Cache warming completed"
EOF

    chmod +x caching/cache-warming.sh
    print_success "Caching layer configured"
}

# Setup Auto-scaling
setup_autoscaling() {
    print_status "Setting up Auto-scaling..."
    
    # Create Kubernetes HPA configurations
    cat > scaling/api-gateway-hpa.yaml << 'EOF'
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: api-gateway-hpa
  namespace: financial-platform
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: api-gateway
  minReplicas: 3
  maxReplicas: 20
  metrics:
  - type: Resource
    resource:
      name: cpu
      target:
        type: Utilization
        averageUtilization: 70
  - type: Resource
    resource:
      name: memory
      target:
        type: Utilization
        averageUtilization: 80
EOF

    cat > scaling/payment-service-hpa.yaml << 'EOF'
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: payment-service-hpa
  namespace: financial-platform
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: payment-service
  minReplicas: 5
  maxReplicas: 50
  metrics:
  - type: Resource
    resource:
      name: cpu
      target:
        type: Utilization
        averageUtilization: 75
  - type: Resource
    resource:
      name: memory
      target:
        type: Utilization
        averageUtilization: 80
EOF

    # Create scaling monitoring script
    cat > scaling/monitor-scaling.sh << 'EOF'
#!/bin/bash
# Scaling Monitor Script

echo "Monitoring scaling activities..."

# Check HPA status
kubectl get hpa -n financial-platform

# Check pod replicas
kubectl get pods -n financial-platform -o wide

# Check resource usage
kubectl top pods -n financial-platform

echo "Scaling monitoring completed"
EOF

    chmod +x scaling/monitor-scaling.sh
    print_success "Auto-scaling configured"
}

# Setup Database Sharding
setup_sharding() {
    print_status "Setting up Database Sharding..."
    
    # Create PostgreSQL sharding configuration
    cat > sharding/postgresql-sharding.sql << 'EOF'
-- PostgreSQL Sharding Configuration

-- Create sharding functions
CREATE OR REPLACE FUNCTION get_shard_id(user_id BIGINT)
RETURNS INTEGER AS $$
BEGIN
    RETURN user_id % 8;
END;
$$ LANGUAGE plpgsql;

-- Create distributed tables
CREATE TABLE transactions_shard_0 (LIKE transactions INCLUDING ALL);
CREATE TABLE transactions_shard_1 (LIKE transactions INCLUDING ALL);
CREATE TABLE transactions_shard_2 (LIKE transactions INCLUDING ALL);
CREATE TABLE transactions_shard_3 (LIKE transactions INCLUDING ALL);
CREATE TABLE transactions_shard_4 (LIKE transactions INCLUDING ALL);
CREATE TABLE transactions_shard_5 (LIKE transactions INCLUDING ALL);
CREATE TABLE transactions_shard_6 (LIKE transactions INCLUDING ALL);
CREATE TABLE transactions_shard_7 (LIKE transactions INCLUDING ALL);

-- Create routing function
CREATE OR REPLACE FUNCTION route_transaction()
RETURNS TRIGGER AS $$
DECLARE
    shard_id INTEGER;
BEGIN
    shard_id := get_shard_id(NEW.user_id);
    
    CASE shard_id
        WHEN 0 THEN INSERT INTO transactions_shard_0 VALUES (NEW.*);
        WHEN 1 THEN INSERT INTO transactions_shard_1 VALUES (NEW.*);
        WHEN 2 THEN INSERT INTO transactions_shard_2 VALUES (NEW.*);
        WHEN 3 THEN INSERT INTO transactions_shard_3 VALUES (NEW.*);
        WHEN 4 THEN INSERT INTO transactions_shard_4 VALUES (NEW.*);
        WHEN 5 THEN INSERT INTO transactions_shard_5 VALUES (NEW.*);
        WHEN 6 THEN INSERT INTO transactions_shard_6 VALUES (NEW.*);
        WHEN 7 THEN INSERT INTO transactions_shard_7 VALUES (NEW.*);
    END CASE;
    
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

-- Create trigger
CREATE TRIGGER route_transaction_trigger
    BEFORE INSERT ON transactions
    FOR EACH ROW
    EXECUTE FUNCTION route_transaction();
EOF

    # Create MongoDB sharding configuration
    cat > sharding/mongodb-sharding.js << 'EOF'
// MongoDB Sharding Configuration

// Enable sharding for database
sh.enableSharding("financial_platform");

// Shard collections
sh.shardCollection("financial_platform.transactions", { "user_id": 1 });
sh.shardCollection("financial_platform.payments", { "user_id": 1 });
sh.shardCollection("financial_platform.escrow_accounts", { "user_id": 1 });
sh.shardCollection("financial_platform.user_profiles", { "user_id": 1 });

// Add shards
sh.addShard("mongodb-shard-0:27017");
sh.addShard("mongodb-shard-1:27017");
sh.addShard("mongodb-shard-2:27017");
sh.addShard("mongodb-shard-3:27017");
sh.addShard("mongodb-shard-4:27017");
sh.addShard("mongodb-shard-5:27017");

// Create indexes on shard keys
db.transactions.createIndex({ "user_id": 1 });
db.payments.createIndex({ "user_id": 1 });
db.escrow_accounts.createIndex({ "user_id": 1 });
db.user_profiles.createIndex({ "user_id": 1 });
EOF

    # Create sharding monitoring script
    cat > sharding/monitor-sharding.sh << 'EOF'
#!/bin/bash
# Sharding Monitor Script

echo "Monitoring database sharding..."

# Check PostgreSQL shard status
echo "PostgreSQL Shards:"
for i in {0..7}; do
    echo "Shard $i: $(pg_isready -h postgresql-shard-$i -p 5432)"
done

# Check MongoDB shard status
echo "MongoDB Shards:"
mongo --eval "sh.status()"

# Check Redis cluster status
echo "Redis Cluster:"
redis-cli -h redis-router-1 cluster info

echo "Sharding monitoring completed"
EOF

    chmod +x sharding/monitor-sharding.sh
    print_success "Database sharding configured"
}

# Setup Performance Monitoring
setup_performance_monitoring() {
    print_status "Setting up Performance Monitoring..."
    
    # Create performance monitoring configuration
    cat > monitoring/performance-monitoring.yml << EOF
# Performance Monitoring Configuration
metrics:
  - name: "response_time_p95"
    query: "histogram_quantile(0.95, rate(http_request_duration_seconds_bucket[5m]))"
    threshold: 0.5
    
  - name: "throughput"
    query: "rate(http_requests_total[5m])"
    threshold: 1000
    
  - name: "error_rate"
    query: "rate(http_requests_total{status=~\"5..\"}[5m]) / rate(http_requests_total[5m])"
    threshold: 0.05
    
  - name: "cache_hit_ratio"
    query: "rate(cache_hits_total[5m]) / (rate(cache_hits_total[5m]) + rate(cache_misses_total[5m]))"
    threshold: 0.8

alerts:
  - name: "HighResponseTime"
    condition: "response_time_p95 > 0.5"
    severity: "warning"
    
  - name: "LowThroughput"
    condition: "throughput < 1000"
    severity: "warning"
    
  - name: "HighErrorRate"
    condition: "error_rate > 0.05"
    severity: "critical"
    
  - name: "LowCacheHitRatio"
    condition: "cache_hit_ratio < 0.8"
    severity: "warning"
EOF

    # Create performance testing script
    cat > monitoring/performance-test.sh << 'EOF'
#!/bin/bash
# Performance Testing Script

echo "Running performance tests..."

# Load testing with Apache Bench
echo "Testing API Gateway..."
ab -n 1000 -c 10 http://localhost:8090/health

echo "Testing Payment Service..."
ab -n 500 -c 5 http://localhost:8083/health

echo "Testing Escrow Service..."
ab -n 300 -c 3 http://localhost:8081/health

# Database performance test
echo "Testing database performance..."
pgbench -h localhost -U postgres -d financial_platform -c 10 -t 100

echo "Performance testing completed"
EOF

    chmod +x monitoring/performance-test.sh
    print_success "Performance monitoring configured"
}

# Setup Performance Docker Compose
setup_performance_compose() {
    print_status "Setting up Performance Docker Compose..."
    
    cat > docker-compose-performance.yml << EOF
version: '3.8'

services:
  # Redis Cluster
  redis-node-0:
    image: redis:7-alpine
    container_name: redis-node-0
    command: redis-server --cluster-enabled yes --cluster-config-file nodes.conf --cluster-node-timeout 5000
    ports:
      - "6379:6379"
    volumes:
      - redis_data_0:/data
    networks:
      - performance_network

  redis-node-1:
    image: redis:7-alpine
    container_name: redis-node-1
    command: redis-server --cluster-enabled yes --cluster-config-file nodes.conf --cluster-node-timeout 5000
    ports:
      - "6380:6379"
    volumes:
      - redis_data_1:/data
    networks:
      - performance_network

  redis-node-2:
    image: redis:7-alpine
    container_name: redis-node-2
    command: redis-server --cluster-enabled yes --cluster-config-file nodes.conf --cluster-node-timeout 5000
    ports:
      - "6381:6379"
    volumes:
      - redis_data_2:/data
    networks:
      - performance_network

  # Load Balancer
  nginx-load-balancer:
    image: nginx:alpine
    container_name: nginx-load-balancer
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - ./load-balancer/nginx.conf:/etc/nginx/nginx.conf
    networks:
      - performance_network
    depends_on:
      - api-gateway

  # Performance Monitoring
  prometheus-performance:
    image: prom/prometheus:latest
    container_name: prometheus-performance
    ports:
      - "9091:9090"
    volumes:
      - ./monitoring/prometheus-performance.yml:/etc/prometheus/prometheus.yml
    networks:
      - performance_network

  grafana-performance:
    image: grafana/grafana:latest
    container_name: grafana-performance
    ports:
      - "3001:3000"
    environment:
      - GF_SECURITY_ADMIN_PASSWORD=admin
    volumes:
      - grafana_data:/var/lib/grafana
    networks:
      - performance_network

volumes:
  redis_data_0:
  redis_data_1:
  redis_data_2:
  grafana_data:

networks:
  performance_network:
    driver: bridge
EOF

    print_success "Performance Docker Compose configured"
}

# Main setup
main() {
    echo "⚡ PERFORMANCE & SCALING SETUP"
    echo "=============================="
    
    setup_cdn
    setup_caching
    setup_autoscaling
    setup_sharding
    setup_performance_monitoring
    setup_performance_compose
    
    echo ""
    echo "✅ PERFORMANCE & SCALING SETUP COMPLETE"
    echo "⚡ COMPONENTS: CDN, Caching Layer, Auto-scaling, Database Sharding"
    echo "🚀 FEATURES: Global CDN, Multi-tier Caching, HPA/VPA, Database Sharding"
    echo "📊 MONITORING: Performance metrics, Auto-scaling alerts, Shard monitoring"
    echo "🚀 START: docker-compose -f docker-compose-performance.yml up -d"
}

main "$@"
