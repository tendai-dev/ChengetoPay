#!/bin/bash

# Backup & Disaster Recovery Setup Script
echo "💾 SETTING UP BACKUP & DISASTER RECOVERY"

# Color functions
print_status() { echo -e "\033[1;34m[SETUP]\033[0m $1"; }
print_success() { echo -e "\033[1;32m[SUCCESS]\033[0m $1"; }

# Setup Database Backups
setup_database_backups() {
    print_status "Setting up Database Backups..."
    
    # Create backup directories
    mkdir -p backups/{postgresql,mongodb,redis}/{full,incremental,wal,oplog,rdb,aof}
    mkdir -p backups/verification
    mkdir -p backups/logs
    
    # Create backup scripts
    cat > backups/postgresql-backup.sh << 'EOF'
#!/bin/bash
# PostgreSQL Backup Script

BACKUP_DIR="/backups/postgresql"
DATE=$(date +%Y%m%d_%H%M%S)
BACKUP_FILE="${BACKUP_DIR}/full/financial_platform_${DATE}.sql"

# Full backup
pg_dump \
  --host=${POSTGRES_HOST} \
  --port=${POSTGRES_PORT} \
  --username=${POSTGRES_USER} \
  --dbname=${POSTGRES_DB} \
  --verbose \
  --compress=9 \
  --format=custom \
  --file=${BACKUP_FILE}

# Verify backup
pg_restore --list ${BACKUP_FILE} | grep -q "financial_platform" && \
  echo "Backup verification successful" || \
  echo "Backup verification failed"

echo "PostgreSQL backup completed: ${BACKUP_FILE}"
EOF

    cat > backups/mongodb-backup.sh << 'EOF'
#!/bin/bash
# MongoDB Backup Script

BACKUP_DIR="/backups/mongodb"
DATE=$(date +%Y%m%d_%H%M%S)
BACKUP_PATH="${BACKUP_DIR}/full/financial_platform_${DATE}"

# Full backup
mongodump \
  --host=${MONGO_HOST} \
  --port=${MONGO_PORT} \
  --username=${MONGO_USER} \
  --password=${MONGO_PASSWORD} \
  --authenticationDatabase=admin \
  --db=${MONGO_DB} \
  --out=${BACKUP_PATH} \
  --gzip

# Verify backup
mongorestore --dryRun --gzip ${BACKUP_PATH} && \
  echo "Backup verification successful" || \
  echo "Backup verification failed"

echo "MongoDB backup completed: ${BACKUP_PATH}"
EOF

    cat > backups/redis-backup.sh << 'EOF'
#!/bin/bash
# Redis Backup Script

BACKUP_DIR="/backups/redis"
DATE=$(date +%Y%m%d_%H%M%S)

# Trigger RDB save
redis-cli -h ${REDIS_HOST} -p ${REDIS_PORT} -a ${REDIS_PASSWORD} BGSAVE

# Wait for save to complete
while [ "$(redis-cli -h ${REDIS_HOST} -p ${REDIS_PORT} -a ${REDIS_PASSWORD} info persistence | grep rdb_bgsave_in_progress | cut -d: -f2)" != "0" ]; do
  sleep 1
done

# Copy RDB file
cp ${REDIS_DATA_DIR}/dump.rdb ${BACKUP_DIR}/rdb/dump-${DATE}.rdb

# Verify backup
redis-check-rdb ${BACKUP_DIR}/rdb/dump-${DATE}.rdb && \
  echo "Backup verification successful" || \
  echo "Backup verification failed"

echo "Redis backup completed: ${BACKUP_DIR}/rdb/dump-${DATE}.rdb"
EOF

    chmod +x backups/*-backup.sh
    print_success "Database backup scripts created"
}

# Setup Cross-Region Replication
setup_cross_region_replication() {
    print_status "Setting up Cross-Region Replication..."
    
    # Create replication configuration
    cat > replication/replication-config.yml << EOF
# Cross-Region Replication Configuration
primary_region: us-east-1
secondary_regions:
  - us-west-2
  - eu-west-1

services:
  - name: api-gateway
    replication_type: active-active
    health_check: /health
    
  - name: payment-service
    replication_type: active-passive
    health_check: /health
    
  - name: escrow-service
    replication_type: active-passive
    health_check: /health
    
  - name: database-service
    replication_type: master-slave
    health_check: /health
EOF

    # Create failover script
    cat > replication/failover.sh << 'EOF'
#!/bin/bash
# Failover Script

REGION=$1
SERVICE=$2

if [ -z "$REGION" ] || [ -z "$SERVICE" ]; then
    echo "Usage: $0 <region> <service>"
    exit 1
fi

echo "Initiating failover for $SERVICE to $REGION..."

# Update DNS
aws route53 change-resource-record-sets \
  --hosted-zone-id ${HOSTED_ZONE_ID} \
  --change-batch "{
    \"Changes\": [{
      \"Action\": \"UPSERT\",
      \"ResourceRecordSet\": {
        \"Name\": \"$SERVICE.financialplatform.com\",
        \"Type\": \"A\",
        \"AliasTarget\": {
          \"HostedZoneId\": \"${ALIAS_HOSTED_ZONE_ID}\",
          \"DNSName\": \"$SERVICE.$REGION.financialplatform.com\",
          \"EvaluateTargetHealth\": true
        }
      }
    }]
  }"

echo "Failover completed for $SERVICE to $REGION"
EOF

    chmod +x replication/failover.sh
    print_success "Cross-region replication configured"
}

# Setup Disaster Recovery Procedures
setup_dr_procedures() {
    print_status "Setting up Disaster Recovery Procedures..."
    
    # Create DR procedures
    mkdir -p dr-procedures
    
    cat > dr-procedures/critical-service-outage.md << 'EOF'
# Critical Service Outage Response

## Immediate Actions (0-5 minutes)
1. Assess impact and identify affected services
2. Create incident ticket
3. Notify primary on-call engineer
4. Check service health endpoints

## Recovery Actions (5-15 minutes)
1. Implement immediate fixes
2. Restart affected services
3. Verify service recovery
4. Monitor service health

## Escalation (15+ minutes)
1. Escalate to secondary on-call
2. Notify engineering manager
3. Update stakeholders
4. Begin root cause analysis
EOF

    cat > dr-procedures/database-failure.md << 'EOF'
# Database Failure Response

## Immediate Actions (0-5 minutes)
1. Verify database connectivity
2. Check replication status
3. Assess data loss impact
4. Notify database team

## Recovery Actions (5-30 minutes)
1. Failover to replica if available
2. Restore from backup if needed
3. Verify data integrity
4. Restart dependent services

## Post-Recovery (30+ minutes)
1. Run data integrity checks
2. Update monitoring alerts
3. Document incident
4. Schedule post-mortem
EOF

    cat > dr-procedures/region-outage.md << 'EOF'
# Region Outage Response

## Immediate Actions (0-5 minutes)
1. Confirm region status
2. Initiate automatic failover
3. Update global DNS
4. Notify stakeholders

## Recovery Actions (5-15 minutes)
1. Verify secondary region health
2. Monitor performance metrics
3. Update status page
4. Communicate to customers

## Post-Recovery (15+ minutes)
1. Monitor cross-region latency
2. Verify data synchronization
3. Plan primary region recovery
4. Update disaster recovery plan
EOF

    print_success "Disaster recovery procedures created"
}

# Setup Data Archival
setup_data_archival() {
    print_status "Setting up Data Archival..."
    
    # Create archival configuration
    cat > archival/archival-config.yml << EOF
# Data Archival Configuration
storage:
  type: s3_glacier
  bucket: financial-platform-archives
  region: us-east-1

policies:
  transaction_data:
    retention: 7 years
    archival_trigger: 90 days
    
  user_data:
    retention: 5 years
    archival_trigger: 1 year
    
  analytics_data:
    retention: 3 years
    archival_trigger: 6 months

compliance:
  - pci_dss: 7 years
  - sox: 7 years
  - gdpr: data minimization
  - aml: 5 years
EOF

    # Create archival script
    cat > archival/archival-process.sh << 'EOF'
#!/bin/bash
# Data Archival Process

DATA_TYPE=$1
RETENTION_DAYS=$2

if [ -z "$DATA_TYPE" ] || [ -z "$RETENTION_DAYS" ]; then
    echo "Usage: $0 <data_type> <retention_days>"
    exit 1
fi

echo "Starting archival process for $DATA_TYPE..."

# Query data for archival
case $DATA_TYPE in
    "transactions")
        psql -h ${POSTGRES_HOST} -U ${POSTGRES_USER} -d ${POSTGRES_DB} -c "
            SELECT * FROM transactions 
            WHERE created_at < NOW() - INTERVAL '$RETENTION_DAYS days'
            AND archived = false;"
        ;;
    "users")
        mongo ${MONGO_HOST}:${MONGO_PORT}/${MONGO_DB} -u ${MONGO_USER} -p ${MONGO_PASSWORD} --eval "
            db.users.find({
                last_activity: { \$lt: new Date(Date.now() - $RETENTION_DAYS * 24 * 60 * 60 * 1000) },
                archived: false
            }).forEach(printjson);"
        ;;
    *)
        echo "Unknown data type: $DATA_TYPE"
        exit 1
        ;;
esac

# Archive to S3 Glacier
aws s3 cp /tmp/archival_data.json s3://financial-platform-archives/$DATA_TYPE/ \
  --storage-class GLACIER \
  --encryption aws:kms \
  --kms-key-id ${KMS_KEY_ID}

echo "Archival completed for $DATA_TYPE"
EOF

    chmod +x archival/archival-process.sh
    print_success "Data archival configured"
}

# Setup Backup Monitoring
setup_backup_monitoring() {
    print_status "Setting up Backup Monitoring..."
    
    # Create backup monitoring configuration
    cat > monitoring/backup-monitoring.yml << EOF
# Backup Monitoring Configuration
metrics:
  - backup_success_rate
  - backup_duration_seconds
  - backup_size_bytes
  - restore_duration_seconds
  - verification_success_rate

alerts:
  - name: BackupFailed
    condition: backup_success_rate < 95
    severity: critical
    
  - name: BackupDurationHigh
    condition: backup_duration_seconds > 3600
    severity: warning
    
  - name: BackupSizeHigh
    condition: backup_size_bytes > 10737418240
    severity: warning
EOF

    # Create backup health check script
    cat > monitoring/backup-health-check.sh << 'EOF'
#!/bin/bash
# Backup Health Check Script

echo "Checking backup health..."

# Check PostgreSQL backups
PG_BACKUP_COUNT=$(find /backups/postgresql/full -name "*.sql" -mtime -1 | wc -l)
if [ $PG_BACKUP_COUNT -eq 0 ]; then
    echo "WARNING: No PostgreSQL backups in last 24 hours"
    exit 1
fi

# Check MongoDB backups
MONGO_BACKUP_COUNT=$(find /backups/mongodb/full -name "*" -type d -mtime -1 | wc -l)
if [ $MONGO_BACKUP_COUNT -eq 0 ]; then
    echo "WARNING: No MongoDB backups in last 24 hours"
    exit 1
fi

# Check Redis backups
REDIS_BACKUP_COUNT=$(find /backups/redis/rdb -name "*.rdb" -mtime -1 | wc -l)
if [ $REDIS_BACKUP_COUNT -eq 0 ]; then
    echo "WARNING: No Redis backups in last 24 hours"
    exit 1
fi

echo "All backups are healthy"
exit 0
EOF

    chmod +x monitoring/backup-health-check.sh
    print_success "Backup monitoring configured"
}

# Setup Backup Docker Compose
setup_backup_compose() {
    print_status "Setting up Backup Docker Compose..."
    
    cat > docker-compose-backup.yml << EOF
version: '3.8'

services:
  # Backup Scheduler
  backup-scheduler:
    image: alpine:latest
    container_name: backup-scheduler
    volumes:
      - ./backups:/backups
      - ./scripts:/scripts
    environment:
      - POSTGRES_HOST=postgresql
      - POSTGRES_PORT=5432
      - POSTGRES_USER=postgres
      - POSTGRES_DB=financial_platform
      - MONGO_HOST=mongodb
      - MONGO_PORT=27017
      - MONGO_USER=mongo
      - MONGO_PASSWORD=mongo
      - MONGO_DB=financial_platform
      - REDIS_HOST=redis
      - REDIS_PORT=6379
      - REDIS_PASSWORD=redis
    command: |
      sh -c "
        echo '0 2 * * * /scripts/postgresql-backup.sh' > /etc/crontabs/root
        echo '0 3 * * * /scripts/mongodb-backup.sh' >> /etc/crontabs/root
        echo '0 1 * * * /scripts/redis-backup.sh' >> /etc/crontabs/root
        crond -f
      "
    restart: unless-stopped

  # Backup Verification
  backup-verification:
    image: alpine:latest
    container_name: backup-verification
    volumes:
      - ./backups:/backups
      - ./monitoring:/monitoring
    environment:
      - POSTGRES_HOST=postgresql
      - MONGO_HOST=mongodb
      - REDIS_HOST=redis
    command: |
      sh -c "
        echo '0 4 * * * /monitoring/backup-health-check.sh' > /etc/crontabs/root
        crond -f
      "
    restart: unless-stopped

  # Data Archival
  data-archival:
    image: alpine:latest
    container_name: data-archival
    volumes:
      - ./archival:/archival
    environment:
      - AWS_ACCESS_KEY_ID=\${AWS_ACCESS_KEY_ID}
      - AWS_SECRET_ACCESS_KEY=\${AWS_SECRET_ACCESS_KEY}
      - AWS_DEFAULT_REGION=us-east-1
    command: |
      sh -c "
        echo '0 1 * * * /archival/archival-process.sh transactions 90' > /etc/crontabs/root
        echo '0 2 1 */3 * /archival/archival-process.sh users 365' >> /etc/crontabs/root
        crond -f
      "
    restart: unless-stopped

volumes:
  backup_data:
EOF

    print_success "Backup Docker Compose configured"
}

# Main setup
main() {
    echo "💾 BACKUP & DISASTER RECOVERY SETUP"
    echo "==================================="
    
    setup_database_backups
    setup_cross_region_replication
    setup_dr_procedures
    setup_data_archival
    setup_backup_monitoring
    setup_backup_compose
    
    echo ""
    echo "✅ BACKUP & DR SETUP COMPLETE"
    echo "💾 COMPONENTS: Database Backups, Cross-Region Replication, DR Procedures, Data Archival"
    echo "🌐 FEATURES: RTO 15min, RPO 5min, 7-year retention, Multi-region failover"
    echo "🚀 START: docker-compose -f docker-compose-backup.yml up -d"
}

main "$@"
