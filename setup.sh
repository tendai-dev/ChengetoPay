#!/bin/bash

echo "🚀 ChengetoPay Platform Setup Script"
echo "======================================"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Check if .env file exists
if [ ! -f .env ]; then
    echo -e "${YELLOW}Creating .env file from template...${NC}"
    cp .env.example .env
    echo -e "${GREEN}✅ .env file created. Please update with your connection strings.${NC}"
else
    echo -e "${GREEN}✅ .env file exists${NC}"
fi

# Create necessary directories
echo -e "${YELLOW}Creating necessary directories...${NC}"
mkdir -p logs
mkdir -p data/postgres
mkdir -p data/mongodb
mkdir -p data/redis
echo -e "${GREEN}✅ Directories created${NC}"

# Install Go dependencies for shared module
echo -e "${YELLOW}Installing Go dependencies...${NC}"
cd microservices/shared
go mod download
cd ../..
echo -e "${GREEN}✅ Dependencies installed${NC}"

# Make scripts executable
echo -e "${YELLOW}Making scripts executable...${NC}"
chmod +x scripts/*.sh
echo -e "${GREEN}✅ Scripts are executable${NC}"

# Database setup reminder
echo ""
echo -e "${YELLOW}⚠️  IMPORTANT: Database Setup Required${NC}"
echo "======================================"
echo "Please update the .env file with your database connection strings:"
echo ""
echo "1. POSTGRES_URL - Your PostgreSQL connection string"
echo "2. MONGODB_URL - Your MongoDB connection string"
echo "3. REDIS_URL - Your Redis connection string"
echo ""
echo "Example:"
echo "POSTGRES_URL=postgres://user:pass@host:5432/chengetopay?sslmode=require"
echo "MONGODB_URL=mongodb+srv://user:pass@cluster.mongodb.net/chengetopay"
echo "REDIS_URL=redis://user:pass@host:6379/0"
echo ""
echo -e "${GREEN}Setup complete! Next steps:${NC}"
echo "1. Update .env with your connection strings"
echo "2. Run: docker-compose -f microservices/docker-compose.secure.yml up -d"
echo "3. Run: ./scripts/test-services.sh to verify all services"
