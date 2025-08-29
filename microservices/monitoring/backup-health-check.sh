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
