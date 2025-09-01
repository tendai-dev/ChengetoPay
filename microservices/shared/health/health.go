package health

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/mongo"
)

type Status string

const (
	StatusHealthy   Status = "healthy"
	StatusUnhealthy Status = "unhealthy"
	StatusDegraded  Status = "degraded"
)

type HealthCheck struct {
	Status      Status                 `json:"status"`
	Service     string                 `json:"service"`
	Version     string                 `json:"version"`
	Timestamp   time.Time              `json:"timestamp"`
	Checks      map[string]CheckResult `json:"checks"`
	Environment string                 `json:"environment"`
}

type CheckResult struct {
	Status  Status        `json:"status"`
	Message string        `json:"message,omitempty"`
	Latency time.Duration `json:"latency_ms"`
}

type Checker struct {
	ServiceName string
	Version     string
	PostgresDB  *sql.DB
	MongoDB     *mongo.Client
	RedisClient *redis.Client
}

func NewChecker(serviceName, version string) *Checker {
	return &Checker{
		ServiceName: serviceName,
		Version:     version,
	}
}

func (c *Checker) CheckPostgres(ctx context.Context) CheckResult {
	if c.PostgresDB == nil {
		return CheckResult{Status: StatusHealthy, Message: "Not configured"}
	}

	start := time.Now()
	err := c.PostgresDB.PingContext(ctx)
	latency := time.Since(start)

	if err != nil {
		return CheckResult{
			Status:  StatusUnhealthy,
			Message: err.Error(),
			Latency: latency,
		}
	}

	return CheckResult{
		Status:  StatusHealthy,
		Latency: latency,
	}
}

func (c *Checker) CheckMongoDB(ctx context.Context) CheckResult {
	if c.MongoDB == nil {
		return CheckResult{Status: StatusHealthy, Message: "Not configured"}
	}

	start := time.Now()
	err := c.MongoDB.Ping(ctx, nil)
	latency := time.Since(start)

	if err != nil {
		return CheckResult{
			Status:  StatusUnhealthy,
			Message: err.Error(),
			Latency: latency,
		}
	}

	return CheckResult{
		Status:  StatusHealthy,
		Latency: latency,
	}
}

func (c *Checker) CheckRedis(ctx context.Context) CheckResult {
	if c.RedisClient == nil {
		return CheckResult{Status: StatusHealthy, Message: "Not configured"}
	}

	start := time.Now()
	err := c.RedisClient.Ping(ctx).Err()
	latency := time.Since(start)

	if err != nil {
		return CheckResult{
			Status:  StatusUnhealthy,
			Message: err.Error(),
			Latency: latency,
		}
	}

	return CheckResult{
		Status:  StatusHealthy,
		Latency: latency,
	}
}

func (c *Checker) GetHealth(ctx context.Context) HealthCheck {
	checks := make(map[string]CheckResult)

	// Check all configured databases
	if c.PostgresDB != nil {
		checks["postgres"] = c.CheckPostgres(ctx)
	}
	if c.MongoDB != nil {
		checks["mongodb"] = c.CheckMongoDB(ctx)
	}
	if c.RedisClient != nil {
		checks["redis"] = c.CheckRedis(ctx)
	}

	// Determine overall status
	overallStatus := StatusHealthy
	for _, check := range checks {
		if check.Status == StatusUnhealthy {
			overallStatus = StatusUnhealthy
			break
		} else if check.Status == StatusDegraded {
			overallStatus = StatusDegraded
		}
	}

	return HealthCheck{
		Status:      overallStatus,
		Service:     c.ServiceName,
		Version:     c.Version,
		Timestamp:   time.Now(),
		Checks:      checks,
		Environment: getEnv("ENVIRONMENT", "development"),
	}
}

func (c *Checker) Handler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		health := c.GetHealth(ctx)

		w.Header().Set("Content-Type", "application/json")
		
		if health.Status == StatusUnhealthy {
			w.WriteHeader(http.StatusServiceUnavailable)
		} else {
			w.WriteHeader(http.StatusOK)
		}

		json.NewEncoder(w).Encode(health)
	}
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
