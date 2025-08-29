#!/bin/bash

# =====================================================
# DATABASE MIGRATION SCRIPT
# =====================================================
# Runs database migrations for the financial platform

set -e

# Colors for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m'

print_status() {
    echo -e "${BLUE}[MIGRATION]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

# Database connection parameters
DB_HOST=${DB_HOST:-"ep-wispy-union-adi5og8a-pooler.c-2.us-east-1.aws.neon.tech"}
DB_PORT=${DB_PORT:-5432}
DB_NAME=${DB_NAME:-"neondb"}
DB_USER=${DB_USER:-"neondb_owner"}
DB_PASSWORD=${DB_PASSWORD:-"npg_6oAPnbj5zIKN"}
DB_SSL=${DB_SSL:-"require"}

# Construct connection string
CONNECTION_STRING="postgresql://${DB_USER}:${DB_PASSWORD}@${DB_HOST}:${DB_PORT}/${DB_NAME}?sslmode=${DB_SSL}"

print_status "Starting database migration..."
print_status "Target database: ${DB_HOST}:${DB_PORT}/${DB_NAME}"

# Check if psql is available
if ! command -v psql &> /dev/null; then
    print_error "psql is not installed. Please install PostgreSQL client tools."
    exit 1
fi

# Test database connection
print_status "Testing database connection..."
if ! psql "$CONNECTION_STRING" -c "SELECT version();" > /dev/null 2>&1; then
    print_error "Failed to connect to database. Please check your connection parameters."
    exit 1
fi
print_success "Database connection successful"

# Create migrations table if it doesn't exist
print_status "Creating migrations tracking table..."
psql "$CONNECTION_STRING" -c "
CREATE TABLE IF NOT EXISTS schema_migrations (
    version VARCHAR(255) PRIMARY KEY,
    applied_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);" > /dev/null

# Function to check if migration has been applied
migration_applied() {
    local version=$1
    local count=$(psql "$CONNECTION_STRING" -t -c "SELECT COUNT(*) FROM schema_migrations WHERE version = '$version';" | tr -d ' ')
    [ "$count" -gt 0 ]
}

# Function to apply migration
apply_migration() {
    local migration_file=$1
    local version=$(basename "$migration_file" .sql)
    
    if migration_applied "$version"; then
        print_warning "Migration $version already applied, skipping..."
        return 0
    fi
    
    print_status "Applying migration: $version"
    
    # Start transaction and apply migration
    if psql "$CONNECTION_STRING" -f "$migration_file" > /dev/null 2>&1; then
        # Record successful migration
        psql "$CONNECTION_STRING" -c "INSERT INTO schema_migrations (version) VALUES ('$version');" > /dev/null
        print_success "Migration $version applied successfully"
    else
        print_error "Failed to apply migration $version"
        return 1
    fi
}

# Change to migrations directory
cd "$(dirname "$0")/migrations"

# Apply migrations in order
print_status "Applying database migrations..."

# Apply initial setup migration
if [ -f "001_initial_setup.sql" ]; then
    apply_migration "001_initial_setup.sql"
else
    print_error "Initial setup migration not found!"
    exit 1
fi

# Apply any additional migrations
for migration_file in *.sql; do
    if [ "$migration_file" != "001_initial_setup.sql" ]; then
        apply_migration "$migration_file"
    fi
done

print_success "All migrations completed successfully!"

# Show migration status
print_status "Migration history:"
psql "$CONNECTION_STRING" -c "SELECT version, applied_at FROM schema_migrations ORDER BY applied_at;"

print_success "Database migration completed!"
