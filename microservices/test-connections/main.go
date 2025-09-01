package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/lib/pq"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	fmt.Println("🔍 Testing Database Connections...")
	fmt.Println("==================================")

	// Test PostgreSQL
	testPostgreSQL()

	// Test MongoDB
	testMongoDB()

	// Test Redis
	testRedis()

	fmt.Println("\n✅ All connections successful!")
}

func testPostgreSQL() {
	fmt.Print("Testing PostgreSQL... ")
	
	postgresURL := os.Getenv("POSTGRES_URL")
	if postgresURL == "" {
		postgresURL = "postgresql://neondb_owner:npg_A7n1FliTzPvk@ep-divine-scene-advwzvm6-pooler.c-2.us-east-1.aws.neon.tech/neondb?sslmode=require"
	}

	db, err := sql.Open("postgres", postgresURL)
	if err != nil {
		log.Fatalf("❌ Failed to connect: %v", err)
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		log.Fatalf("❌ Failed to ping: %v", err)
	}

	// Create schemas
	schemas := []string{
		"escrow", "payments", "ledger", "journal", "fees",
		"refunds", "transfers", "payouts", "reserves",
		"reconciliation", "treasury", "risk", "disputes",
		"auth", "compliance",
	}

	for _, schema := range schemas {
		_, err := db.Exec(fmt.Sprintf("CREATE SCHEMA IF NOT EXISTS %s", schema))
		if err != nil {
			fmt.Printf("\n  Warning: Could not create schema %s: %v", schema, err)
		}
	}

	fmt.Println("✅ Connected and schemas created")
}

func testMongoDB() {
	fmt.Print("Testing MongoDB... ")
	
	mongoURL := os.Getenv("MONGODB_URL")
	if mongoURL == "" {
		mongoURL = "mongodb+srv://tendai_db_user:UurJGh23sn9O9DhC@chengetopay.bafkgjv.mongodb.net/chengetopay?retryWrites=true&w=majority&appName=ChengetoPay"
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoURL))
	if err != nil {
		log.Fatalf("❌ Failed to connect: %v", err)
	}
	defer client.Disconnect(ctx)

	if err := client.Ping(ctx, nil); err != nil {
		log.Fatalf("❌ Failed to ping: %v", err)
	}

	fmt.Println("✅ Connected")
}

func testRedis() {
	fmt.Print("Testing Redis... ")
	
	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		redisURL = "redis://default:0OtBcZaZNou5kWMOzizSKvurera0BzDL@redis-13100.c323.us-east-1-2.ec2.redns.redis-cloud.com:13100"
	}

	opt, err := redis.ParseURL(redisURL)
	if err != nil {
		log.Fatalf("❌ Failed to parse URL: %v", err)
	}

	client := redis.NewClient(opt)
	defer client.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		log.Fatalf("❌ Failed to ping: %v", err)
	}

	fmt.Println("✅ Connected")
}
