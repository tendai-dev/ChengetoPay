#!/bin/bash

# Database Connection Fix Script
# Priority 1 - Immediate Actions

echo "========================================="
echo "DATABASE CONNECTION FIX UTILITY"
echo "========================================="
echo ""

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Current problematic connection strings
POSTGRES_URL="postgresql://chengetopay_owner:4ixY5mEgxoP0@ep-weathered-bonus-a5eosddm.us-east-2.aws.neon.tech/chengetopay?sslmode=require"
MONGODB_URL="mongodb+srv://tendaimukurusystemsadministrator:Mukuru%402024@chengetopay.jvjvz.mongodb.net/?retryWrites=true&w=majority&appName=ChengetoPay"
REDIS_URL="redis://default:AVNS_Ej1eMBBJAoJy5sTnJCm@redis-chengetopay-tendaimukurusystemsadministrator-f61f.g.aivencloud.com:24660"

echo "Testing current database connections..."
echo ""

# Test PostgreSQL
echo -e "${YELLOW}1. Testing PostgreSQL (NeonDB)...${NC}"
echo "   URL: $POSTGRES_URL"

# Try different connection methods for PostgreSQL
echo "   Attempting connection with pooler endpoint..."
POSTGRES_POOLER="postgresql://chengetopay_owner:4ixY5mEgxoP0@ep-weathered-bonus-a5eosddm-pooler.us-east-2.aws.neon.tech/chengetopay?sslmode=require"

if command -v psql &> /dev/null; then
    if psql "$POSTGRES_POOLER" -c "SELECT 1" &> /dev/null; then
        echo -e "   ${GREEN}✓ PostgreSQL connection successful with pooler!${NC}"
        echo "   Working URL: $POSTGRES_POOLER"
        POSTGRES_WORKING="$POSTGRES_POOLER"
    else
        echo -e "   ${RED}✗ PostgreSQL connection failed${NC}"
        echo "   Error: SCRAM-SHA-256 authentication issue"
        echo ""
        echo "   REQUIRED ACTION:"
        echo "   1. Log into https://console.neon.tech"
        echo "   2. Find your database 'chengetopay'"
        echo "   3. Copy the POOLED connection string (not direct)"
        echo "   4. Update the POSTGRES_URL in .env"
    fi
else
    echo "   psql not installed. Install with: brew install postgresql"
fi

echo ""

# Test MongoDB
echo -e "${YELLOW}2. Testing MongoDB Atlas...${NC}"
echo "   Checking DNS for MongoDB cluster..."

# Check if the MongoDB cluster exists
if nslookup "chengetopay.jvjvz.mongodb.net" &> /dev/null; then
    echo -e "   ${GREEN}✓ MongoDB cluster DNS resolves${NC}"
else
    echo -e "   ${RED}✗ MongoDB cluster not found (DNS NXDOMAIN)${NC}"
    echo ""
    echo "   REQUIRED ACTION:"
    echo "   1. Log into https://cloud.mongodb.com"
    echo "   2. Check if cluster 'ChengetoPay' exists and is RUNNING"
    echo "   3. If paused, resume it"
    echo "   4. If deleted, create a new cluster"
    echo "   5. Get the new connection string from Atlas"
    echo "   6. Update Network Access to allow your IP"
fi

echo ""

# Test Redis
echo -e "${YELLOW}3. Testing Redis (Aiven)...${NC}"
echo "   Checking DNS for Redis service..."

REDIS_HOST="redis-chengetopay-tendaimukurusystemsadministrator-f61f.g.aivencloud.com"
if nslookup "$REDIS_HOST" &> /dev/null; then
    echo -e "   ${GREEN}✓ Redis host DNS resolves${NC}"
    
    if command -v redis-cli &> /dev/null; then
        if redis-cli -u "$REDIS_URL" ping &> /dev/null; then
            echo -e "   ${GREEN}✓ Redis connection successful!${NC}"
        else
            echo -e "   ${RED}✗ Redis connection failed (auth/network issue)${NC}"
        fi
    else
        echo "   redis-cli not installed. Install with: brew install redis"
    fi
else
    echo -e "   ${RED}✗ Redis host not found (DNS failure)${NC}"
    echo ""
    echo "   REQUIRED ACTION:"
    echo "   1. Log into https://console.aiven.io"
    echo "   2. Check if Redis service exists and is RUNNING"
    echo "   3. Get the current connection URI from the console"
    echo "   4. Update REDIS_URL in .env"
fi

echo ""
echo "========================================="
echo "QUICK FIX - LOCAL DEVELOPMENT DATABASES"
echo "========================================="
echo ""
echo "To continue development while fixing cloud databases:"
echo ""
echo -e "${GREEN}Option 1: Use Docker for local databases${NC}"
echo "docker run -d --name postgres-local -p 5432:5432 -e POSTGRES_DB=chengetopay -e POSTGRES_USER=postgres -e POSTGRES_PASSWORD=postgres postgres:15"
echo "docker run -d --name mongodb-local -p 27017:27017 mongo:6"
echo "docker run -d --name redis-local -p 6379:6379 redis:7"
echo ""
echo "Then update your .env with:"
echo "POSTGRES_URL=postgresql://postgres:postgres@localhost:5432/chengetopay?sslmode=disable"
echo "MONGODB_URL=mongodb://localhost:27017/chengetopay"
echo "REDIS_URL=redis://localhost:6379"
echo ""
echo -e "${GREEN}Option 2: Create docker-compose for local DBs${NC}"
echo "Create a docker-compose.local-db.yml file and run:"
echo "docker-compose -f docker-compose.local-db.yml up -d"
echo ""
echo "========================================="
echo "NEXT STEPS"
echo "========================================="
echo ""
echo "1. Fix database connections using the actions above"
echo "2. Update .env file with working connection strings"
echo "3. Copy .env to microservices/.env"
echo "4. Restart Docker services"
echo "5. Test service health endpoints"
echo ""
echo "For detailed instructions, see: PRIORITY_1_DATABASE_FIX.md"
