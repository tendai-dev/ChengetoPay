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
