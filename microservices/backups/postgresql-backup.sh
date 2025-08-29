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
