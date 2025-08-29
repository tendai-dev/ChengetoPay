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
