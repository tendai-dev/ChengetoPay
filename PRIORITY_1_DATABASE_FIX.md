# Priority 1: Database Connection Fixes - Action Plan

## 1. PostgreSQL (NeonDB) - SCRAM-SHA-256 Authentication Error

### Current Error:
```
SCRAM-SHA-256 error: server sent an invalid SCRAM-SHA-256 iteration count: "i=1"
```

### Immediate Actions:

#### Option A: Use Pooler Connection (Recommended)
1. Log into your NeonDB console at https://console.neon.tech
2. Find your database and look for the "Pooled connection" string
3. The pooled connection URL should look like:
   ```
   postgresql://chengetopay_owner:4ixY5mEgxoP0@ep-weathered-bonus-a5eosddm-pooler.us-east-2.aws.neon.tech/chengetopay?sslmode=require
   ```
   Note the `-pooler` suffix in the hostname

#### Option B: Update Connection Parameters
Try these alternative connection strings:
```bash
# With explicit SCRAM settings
postgresql://chengetopay_owner:4ixY5mEgxoP0@ep-weathered-bonus-a5eosddm.us-east-2.aws.neon.tech/chengetopay?sslmode=require&channel_binding=disable&options=scram_iterations%3D4096

# With different SSL mode
postgresql://chengetopay_owner:4ixY5mEgxoP0@ep-weathered-bonus-a5eosddm.us-east-2.aws.neon.tech/chengetopay?sslmode=prefer

# Direct connection without pooler
postgresql://chengetopay_owner:4ixY5mEgxoP0@ep-weathered-bonus-a5eosddm.us-east-2.aws.neon.tech:5432/chengetopay?sslmode=require
```

#### Option C: Contact Support
If above doesn't work, contact NeonDB support with this specific error. They may need to:
- Reset your database password
- Update server-side SCRAM configuration
- Provide alternative connection method

---

## 2. MongoDB Atlas - DNS Resolution Failure

### Current Error:
```
lookup _mongodb._tcp.chengetopay.jvjvz.mongodb.net: no such host
```

### Immediate Actions:

#### Step 1: Verify Cluster Status
1. Log into MongoDB Atlas at https://cloud.mongodb.com
2. Check if your cluster "ChengetoPay" is:
   - **Active** (not paused)
   - **Running** (not terminated)
   - Has the correct name

#### Step 2: Update Network Access
1. Go to Network Access in MongoDB Atlas
2. Add your current IP address or use `0.0.0.0/0` for testing (NOT for production)
3. Wait 1-2 minutes for changes to propagate

#### Step 3: Get Fresh Connection String
1. Click "Connect" on your cluster
2. Choose "Connect your application"
3. Copy the connection string
4. It should look like:
   ```
   mongodb+srv://tendaimukurusystemsadministrator:YOUR_PASSWORD@chengetopay.xxxxx.mongodb.net/?retryWrites=true&w=majority
   ```

#### Step 4: Use Standard Connection (if SRV fails)
If DNS SRV records aren't resolving, get the standard connection string:
1. In Atlas, choose "Connect with MongoDB Compass"
2. Toggle to "Standard connection string"
3. It will look like:
   ```
   mongodb://tendaimukurusystemsadministrator:YOUR_PASSWORD@cluster0-shard-00-00.xxxxx.mongodb.net:27017,cluster0-shard-00-01.xxxxx.mongodb.net:27017,cluster0-shard-00-02.xxxxx.mongodb.net:27017/?ssl=true&replicaSet=atlas-xxxxx
   ```

---

## 3. Redis (Aiven) - Hostname Not Resolving

### Current Error:
```
lookup redis-chengetopay-tendaimukurusystemsadministrator-f61f.g.aivencloud.com: no such host
```

### Immediate Actions:

#### Step 1: Verify Service Status
1. Log into Aiven Console at https://console.aiven.io
2. Check if your Redis service is:
   - **Running** (not powered off)
   - **Active** (not deleted)
   - Has correct name

#### Step 2: Get Current Connection Details
1. Click on your Redis service
2. Go to "Overview" tab
3. Copy the correct connection string
4. Should look like:
   ```
   redis://default:PASSWORD@redis-xxxxx.aivencloud.com:24660
   ```

#### Step 3: Common Issues
- Service may have been renamed
- Service may be in different region
- Service may need to be recreated if deleted

#### Alternative: Use Local Redis for Development
```bash
# Install local Redis
brew install redis

# Start Redis
redis-server

# Use local connection
REDIS_URL=redis://localhost:6379
```

---

## Quick Test Commands

After updating connection strings, test each database:

### Test PostgreSQL:
```bash
psql "YOUR_NEW_CONNECTION_STRING" -c "SELECT 1"
```

### Test MongoDB:
```bash
mongosh "YOUR_NEW_CONNECTION_STRING" --eval "db.adminCommand('ping')"
```

### Test Redis:
```bash
redis-cli -u "YOUR_NEW_CONNECTION_STRING" ping
```

---

## Update Environment Files

Once you have working connection strings, update:

1. `/microservices/.env`
2. `/microservices/docker-compose.yml`
3. `/microservices/docker-compose.secure.yml`

```env
POSTGRES_URL=YOUR_WORKING_POSTGRES_URL
MONGODB_URL=YOUR_WORKING_MONGODB_URL
REDIS_URL=YOUR_WORKING_REDIS_URL
```

---

## Verification Steps

1. Test connections individually
2. Restart Docker containers
3. Check service logs for connection success
4. Run health checks on services

---

## If All Else Fails - Local Development Setup

For immediate development while resolving cloud issues:

```bash
# Start local databases
docker-compose -f docker-compose.local-db.yml up -d

# This will start:
# - PostgreSQL on localhost:5432
# - MongoDB on localhost:27017
# - Redis on localhost:6379
```

---

## Support Contacts

- **NeonDB Support**: https://neon.tech/docs/introduction/support
- **MongoDB Atlas Support**: https://www.mongodb.com/support
- **Aiven Support**: https://help.aiven.io/

---

*Complete these fixes in order. PostgreSQL is most critical as it's the primary datastore.*
