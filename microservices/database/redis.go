package database

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisConfig holds Redis configuration
type RedisConfig struct {
	URL      string
	Password string
	DB       int
	PoolSize int
	MinIdleConns int
	MaxRetries int
	DialTimeout time.Duration
	ReadTimeout time.Duration
	WriteTimeout time.Duration
}

// RedisDB represents the Redis connection
type RedisDB struct {
	client *redis.Client
}

// NewRedisDB creates a new Redis connection
func NewRedisDB(config RedisConfig) (*RedisDB, error) {
	opt, err := redis.ParseURL(config.URL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse Redis URL: %w", err)
	}

	// Override with config values
	if config.Password != "" {
		opt.Password = config.Password
	}
	if config.DB != 0 {
		opt.DB = config.DB
	}
	opt.PoolSize = config.PoolSize
	opt.MinIdleConns = config.MinIdleConns
	opt.MaxRetries = config.MaxRetries
	opt.DialTimeout = config.DialTimeout
	opt.ReadTimeout = config.ReadTimeout
	opt.WriteTimeout = config.WriteTimeout

	client := redis.NewClient(opt)

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to ping Redis: %w", err)
	}

	log.Println("✅ Redis connected successfully")

	return &RedisDB{client: client}, nil
}

// Close closes the Redis connection
func (r *RedisDB) Close() error {
	return r.client.Close()
}

// GetClient returns the Redis client
func (r *RedisDB) GetClient() *redis.Client {
	return r.client
}

// HealthCheck performs a health check on Redis
func (r *RedisDB) HealthCheck(ctx context.Context) error {
	return r.client.Ping(ctx).Err()
}

// Set sets a key-value pair with optional expiration
func (r *RedisDB) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return r.client.Set(ctx, key, value, expiration).Err()
}

// Get gets a value by key
func (r *RedisDB) Get(ctx context.Context, key string) (string, error) {
	return r.client.Get(ctx, key).Result()
}

// Del deletes one or more keys
func (r *RedisDB) Del(ctx context.Context, keys ...string) error {
	return r.client.Del(ctx, keys...).Err()
}

// Exists checks if a key exists
func (r *RedisDB) Exists(ctx context.Context, keys ...string) (int64, error) {
	return r.client.Exists(ctx, keys...).Result()
}

// Expire sets expiration for a key
func (r *RedisDB) Expire(ctx context.Context, key string, expiration time.Duration) error {
	return r.client.Expire(ctx, key, expiration).Err()
}

// TTL gets the time to live for a key
func (r *RedisDB) TTL(ctx context.Context, key string) (time.Duration, error) {
	return r.client.TTL(ctx, key).Result()
}

// Incr increments a key
func (r *RedisDB) Incr(ctx context.Context, key string) (int64, error) {
	return r.client.Incr(ctx, key).Result()
}

// IncrBy increments a key by a specific amount
func (r *RedisDB) IncrBy(ctx context.Context, key string, value int64) (int64, error) {
	return r.client.IncrBy(ctx, key, value).Result()
}

// HSet sets a hash field
func (r *RedisDB) HSet(ctx context.Context, key string, values ...interface{}) error {
	return r.client.HSet(ctx, key, values...).Err()
}

// HGet gets a hash field
func (r *RedisDB) HGet(ctx context.Context, key, field string) (string, error) {
	return r.client.HGet(ctx, key, field).Result()
}

// HGetAll gets all hash fields
func (r *RedisDB) HGetAll(ctx context.Context, key string) (map[string]string, error) {
	return r.client.HGetAll(ctx, key).Result()
}

// HDel deletes hash fields
func (r *RedisDB) HDel(ctx context.Context, key string, fields ...string) error {
	return r.client.HDel(ctx, key, fields...).Err()
}

// LPush pushes values to the left of a list
func (r *RedisDB) LPush(ctx context.Context, key string, values ...interface{}) error {
	return r.client.LPush(ctx, key, values...).Err()
}

// RPush pushes values to the right of a list
func (r *RedisDB) RPush(ctx context.Context, key string, values ...interface{}) error {
	return r.client.RPush(ctx, key, values...).Err()
}

// LPop pops a value from the left of a list
func (r *RedisDB) LPop(ctx context.Context, key string) (string, error) {
	return r.client.LPop(ctx, key).Result()
}

// RPop pops a value from the right of a list
func (r *RedisDB) RPop(ctx context.Context, key string) (string, error) {
	return r.client.RPop(ctx, key).Result()
}

// LLen gets the length of a list
func (r *RedisDB) LLen(ctx context.Context, key string) (int64, error) {
	return r.client.LLen(ctx, key).Result()
}

// SAdd adds members to a set
func (r *RedisDB) SAdd(ctx context.Context, key string, members ...interface{}) error {
	return r.client.SAdd(ctx, key, members...).Err()
}

// SRem removes members from a set
func (r *RedisDB) SRem(ctx context.Context, key string, members ...interface{}) error {
	return r.client.SRem(ctx, key, members...).Err()
}

// SMembers gets all members of a set
func (r *RedisDB) SMembers(ctx context.Context, key string) ([]string, error) {
	return r.client.SMembers(ctx, key).Result()
}

// SIsMember checks if a member exists in a set
func (r *RedisDB) SIsMember(ctx context.Context, key string, member interface{}) (bool, error) {
	return r.client.SIsMember(ctx, key, member).Result()
}

// ZAdd adds members to a sorted set
func (r *RedisDB) ZAdd(ctx context.Context, key string, members ...redis.Z) error {
	return r.client.ZAdd(ctx, key, members...).Err()
}

// ZRange gets members from a sorted set by rank
func (r *RedisDB) ZRange(ctx context.Context, key string, start, stop int64) ([]string, error) {
	return r.client.ZRange(ctx, key, start, stop).Result()
}

// ZRangeByScore gets members from a sorted set by score
func (r *RedisDB) ZRangeByScore(ctx context.Context, key string, opt *redis.ZRangeBy) ([]string, error) {
	return r.client.ZRangeByScore(ctx, key, opt).Result()
}

// ZRem removes members from a sorted set
func (r *RedisDB) ZRem(ctx context.Context, key string, members ...interface{}) error {
	return r.client.ZRem(ctx, key, members...).Err()
}

// Publish publishes a message to a channel
func (r *RedisDB) Publish(ctx context.Context, channel string, message interface{}) error {
	return r.client.Publish(ctx, channel, message).Err()
}

// Subscribe subscribes to channels
func (r *RedisDB) Subscribe(ctx context.Context, channels ...string) *redis.PubSub {
	return r.client.Subscribe(ctx, channels...)
}

// PSubscribe subscribes to patterns
func (r *RedisDB) PSubscribe(ctx context.Context, patterns ...string) *redis.PubSub {
	return r.client.PSubscribe(ctx, patterns...)
}

// Eval executes a Lua script
func (r *RedisDB) Eval(ctx context.Context, script string, keys []string, args ...interface{}) (interface{}, error) {
	return r.client.Eval(ctx, script, keys, args...).Result()
}

// EvalSha executes a Lua script by SHA
func (r *RedisDB) EvalSha(ctx context.Context, sha1 string, keys []string, args ...interface{}) (interface{}, error) {
	return r.client.EvalSha(ctx, sha1, keys, args...).Result()
}

// ScriptLoad loads a Lua script
func (r *RedisDB) ScriptLoad(ctx context.Context, script string) (string, error) {
	return r.client.ScriptLoad(ctx, script).Result()
}

// FlushDB flushes the current database
func (r *RedisDB) FlushDB(ctx context.Context) error {
	return r.client.FlushDB(ctx).Err()
}

// FlushAll flushes all databases
func (r *RedisDB) FlushAll(ctx context.Context) error {
	return r.client.FlushAll(ctx).Err()
}

// Info gets Redis server information
func (r *RedisDB) Info(ctx context.Context, section ...string) (string, error) {
	return r.client.Info(ctx, section...).Result()
}

// ConfigGet gets Redis configuration
func (r *RedisDB) ConfigGet(ctx context.Context, parameter string) (map[string]string, error) {
	return r.client.ConfigGet(ctx, parameter).Result()
}

// ConfigSet sets Redis configuration
func (r *RedisDB) ConfigSet(ctx context.Context, parameter, value string) error {
	return r.client.ConfigSet(ctx, parameter, value).Err()
}
