#!/bin/bash

echo "🔍 Testing Database Connections..."
echo "=================================="

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m'

# Load environment variables
source .env

# Test PostgreSQL
echo -n "Testing PostgreSQL... "
if command -v psql &> /dev/null; then
    if psql "$POSTGRES_URL" -c "SELECT 1" &> /dev/null; then
        echo -e "${GREEN}✅ Connected${NC}"
    else
        echo -e "${RED}❌ Failed${NC}"
    fi
else
    # Try with Python if psql not available
    python3 -c "
import psycopg2
import os
try:
    conn = psycopg2.connect(os.environ.get('POSTGRES_URL'))
    conn.close()
    print('✅ Connected')
except Exception as e:
    print('❌ Failed:', str(e))
" 2>/dev/null || echo -e "${RED}❌ psql not installed${NC}"
fi

# Test MongoDB
echo -n "Testing MongoDB... "
python3 -c "
from pymongo import MongoClient
import os
try:
    client = MongoClient(os.environ.get('MONGODB_URL'))
    client.server_info()
    print('✅ Connected')
except Exception as e:
    print('❌ Failed:', str(e))
" 2>/dev/null || echo -e "${RED}❌ pymongo not installed${NC}"

# Test Redis
echo -n "Testing Redis... "
python3 -c "
import redis
import os
url = os.environ.get('REDIS_URL')
try:
    r = redis.from_url(url)
    r.ping()
    print('✅ Connected')
except Exception as e:
    print('❌ Failed:', str(e))
" 2>/dev/null || echo -e "${RED}❌ redis-py not installed${NC}"

echo ""
echo "Connection test complete!"
