package config

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	// Database
	PostgresURL string `mapstructure:"POSTGRES_URL"`
	MongoDBURL  string `mapstructure:"MONGODB_URL"`
	RedisURL    string `mapstructure:"REDIS_URL"`

	// JWT
	JWTSecret           string        `mapstructure:"JWT_SECRET"`
	JWTExpiry           time.Duration `mapstructure:"JWT_EXPIRY"`
	RefreshTokenExpiry  time.Duration `mapstructure:"REFRESH_TOKEN_EXPIRY"`

	// Service Discovery
	ConsulURL string `mapstructure:"CONSUL_URL"`
	VaultURL  string `mapstructure:"VAULT_URL"`
	VaultToken string `mapstructure:"VAULT_TOKEN"`

	// Monitoring
	PrometheusURL string `mapstructure:"PROMETHEUS_URL"`
	GrafanaURL    string `mapstructure:"GRAFANA_URL"`
	JaegerURL     string `mapstructure:"JAEGER_URL"`

	// API Configuration
	APIGatewayPort      int    `mapstructure:"API_GATEWAY_PORT"`
	RateLimitPerMinute  int    `mapstructure:"RATE_LIMIT_PER_MINUTE"`
	CORSAllowedOrigins  string `mapstructure:"CORS_ALLOWED_ORIGINS"`

	// Environment
	Environment string `mapstructure:"ENVIRONMENT"`
	LogLevel    string `mapstructure:"LOG_LEVEL"`
	Debug       bool   `mapstructure:"DEBUG"`
}

func Load() (*Config, error) {
	config := &Config{}

	viper.SetConfigFile(".env")
	viper.AutomaticEnv()

	// Set defaults
	viper.SetDefault("POSTGRES_URL", "postgres://localhost:5432/chengetopay")
	viper.SetDefault("MONGODB_URL", "mongodb://localhost:27017/chengetopay")
	viper.SetDefault("REDIS_URL", "redis://localhost:6379/0")
	viper.SetDefault("JWT_SECRET", "change-this-secret-in-production")
	viper.SetDefault("JWT_EXPIRY", "24h")
	viper.SetDefault("REFRESH_TOKEN_EXPIRY", "7d")
	viper.SetDefault("CONSUL_URL", "http://consul:8500")
	viper.SetDefault("VAULT_URL", "http://vault:8200")
	viper.SetDefault("PROMETHEUS_URL", "http://prometheus:9090")
	viper.SetDefault("GRAFANA_URL", "http://grafana:3000")
	viper.SetDefault("JAEGER_URL", "http://jaeger:14268")
	viper.SetDefault("API_GATEWAY_PORT", 8080)
	viper.SetDefault("RATE_LIMIT_PER_MINUTE", 100)
	viper.SetDefault("CORS_ALLOWED_ORIGINS", "*")
	viper.SetDefault("ENVIRONMENT", "development")
	viper.SetDefault("LOG_LEVEL", "info")
	viper.SetDefault("DEBUG", false)

	// Read config file if exists
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("error reading config file: %w", err)
		}
	}

	// Unmarshal config
	if err := viper.Unmarshal(config); err != nil {
		return nil, fmt.Errorf("unable to decode config: %w", err)
	}

	// Parse durations
	if jwtExpiry := viper.GetString("JWT_EXPIRY"); jwtExpiry != "" {
		duration, err := time.ParseDuration(jwtExpiry)
		if err == nil {
			config.JWTExpiry = duration
		}
	}

	if refreshExpiry := viper.GetString("REFRESH_TOKEN_EXPIRY"); refreshExpiry != "" {
		duration, err := time.ParseDuration(refreshExpiry)
		if err == nil {
			config.RefreshTokenExpiry = duration
		}
	}

	return config, nil
}

func MustLoad() *Config {
	config, err := Load()
	if err != nil {
		panic(fmt.Sprintf("Failed to load config: %v", err))
	}
	return config
}
