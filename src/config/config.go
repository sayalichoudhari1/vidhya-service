// Package config loads Vidhya Service configuration from environment variables.
//
// Local development values are provided via devops/local/.env (loaded by
// docker-compose). In AWS, the same environment variable names are injected
// by the ECS task definition (see aws.cfn.app.yml), sourced from SSM
// Parameter Store / Secrets Manager.
package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// Configuration is the singleton application configuration object.
type Configuration struct {
	Server   ServerConfig
	Database DatabaseConfig
	JWT      JWTConfig
	Redis    RedisConfig
	Kafka    KafkaConfig
	Log      LogConfig
	Seed     SeedConfig
}

// ServerConfig controls the HTTP server.
type ServerConfig struct {
	Address           string        // e.g. ":8080"
	URLPrefix         string        // e.g. "/vidhyaservice"
	ReadHeaderTimeout time.Duration
	ReadTimeout       time.Duration
	WriteTimeout      time.Duration
	IdleTimeout       time.Duration
	AllowedOrigins    []string // for CORS, so the UI dev team can call the API from localhost
}

// DatabaseConfig configures the Postgres connection (works for local Postgres
// today, and for AWS Aurora PostgreSQL once DB_HOST/DB_* are pointed at it).
type DatabaseConfig struct {
	Host               string
	Port               int
	User               string
	Password           string
	DBName             string
	SSLMode            string
	MaxOpenConnections int
	MaxIdleConnections int
}

// JWTConfig configures locally issued access/refresh tokens.
type JWTConfig struct {
	Secret               string
	AccessTokenDuration  time.Duration
	RefreshTokenDuration time.Duration
	Issuer               string
}

// RedisConfig configures the optional cache connection. Not required for the
// service to start locally; kept ready for future use (sessions, caching).
type RedisConfig struct {
	Enabled  bool
	Address  string
	Password string
	DB       int
}

// KafkaConfig configures the optional messaging connection. Not required for
// the service to start locally; kept ready for future async integrations
// (e.g. notifications, cross-service events).
type KafkaConfig struct {
	Enabled bool
	Brokers []string
}

// LogConfig configures structured logging.
type LogConfig struct {
	Level  string // DEBUG, INFO, WARN, ERROR
	Format string // TEXT or JSON
}

// SeedConfig controls whether/how the service seeds a first admin user so
// that a brand-new local database is immediately usable via the API.
type SeedConfig struct {
	Enabled       bool
	AdminEmail    string
	AdminPassword string
	AdminPhone    string
}

// Read loads configuration from environment variables, applying sensible
// local-development defaults so the service can run with a minimal .env.
func Read() (*Configuration, error) {
	cfg := &Configuration{
		Server: ServerConfig{
			Address:           getEnv("APP_SERVER_ADDRESS", ":8080"),
			URLPrefix:         getEnv("APP_SERVER_URL_PREFIX", "/vidhyaservice"),
			ReadHeaderTimeout: getEnvDuration("APP_SERVER_READ_HEADER_TIMEOUT", 5*time.Second),
			ReadTimeout:       getEnvDuration("APP_SERVER_READ_TIMEOUT", 10*time.Second),
			WriteTimeout:      getEnvDuration("APP_SERVER_WRITE_TIMEOUT", 15*time.Second),
			IdleTimeout:       getEnvDuration("APP_SERVER_IDLE_TIMEOUT", 60*time.Second),
			AllowedOrigins:    getEnvList("APP_SERVER_ALLOWED_ORIGINS", []string{"*"}),
		},
		Database: DatabaseConfig{
			Host:               getEnv("DB_HOST", "localhost"),
			Port:               getEnvInt("DB_PORT", 5432),
			User:               getEnv("DB_USER", "appuser"),
			Password:           getEnv("DB_PASS", "appuser"),
			DBName:             getEnv("DB_NAME", "vidhyadb"),
			SSLMode:            getEnv("DB_SSLMODE", "disable"),
			MaxOpenConnections: getEnvInt("DB_MAX_OPEN_CONNECTIONS", 25),
			MaxIdleConnections: getEnvInt("DB_MAX_IDLE_CONNECTIONS", 10),
		},
		JWT: JWTConfig{
			Secret:               getEnv("JWT_SECRET", "please-change-this-secret-in-every-environment"),
			AccessTokenDuration:  getEnvDuration("JWT_ACCESS_TOKEN_DURATION", 2*time.Hour),
			RefreshTokenDuration: getEnvDuration("JWT_REFRESH_TOKEN_DURATION", 7*24*time.Hour),
			Issuer:               getEnv("JWT_ISSUER", "vidhya-service"),
		},
		Redis: RedisConfig{
			Enabled:  getEnvBool("REDIS_ENABLED", false),
			Address:  getEnv("REDIS_ADDRESS", "localhost:6379"),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       getEnvInt("REDIS_DB", 0),
		},
		Kafka: KafkaConfig{
			Enabled: getEnvBool("KAFKA_ENABLED", false),
			Brokers: getEnvList("KAFKA_BROKERS", []string{"localhost:9092"}),
		},
		Log: LogConfig{
			Level:  getEnv("LOG_LEVEL", "INFO"),
			Format: getEnv("LOG_FORMAT", "TEXT"),
		},
		Seed: SeedConfig{
			Enabled:       getEnvBool("SEED_ADMIN_ENABLED", true),
			AdminEmail:    getEnv("SEED_ADMIN_EMAIL", "admin@vidhya.local"),
			AdminPassword: getEnv("SEED_ADMIN_PASSWORD", "Admin@123"),
			AdminPhone:    getEnv("SEED_ADMIN_PHONE", "+10000000000"),
		},
	}

	if err := cfg.validate(); err != nil {
		return nil, fmt.Errorf("config.Read: %w", err)
	}

	return cfg, nil
}

func (c *Configuration) validate() error {
	if c.Database.Host == "" {
		return fmt.Errorf("DB_HOST is required")
	}
	if c.JWT.Secret == "" {
		return fmt.Errorf("JWT_SECRET is required")
	}
	return nil
}

// PostgresDSN builds the "key=value" DSN string used by the Postgres driver.
func (d DatabaseConfig) PostgresDSN() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		d.Host, d.Port, d.User, d.Password, d.DBName, d.SSLMode,
	)
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	i, err := strconv.Atoi(v)
	if err != nil {
		return fallback
	}
	return i
}

func getEnvBool(key string, fallback bool) bool {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	b, err := strconv.ParseBool(v)
	if err != nil {
		return fallback
	}
	return b
}

func getEnvDuration(key string, fallback time.Duration) time.Duration {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	d, err := time.ParseDuration(v)
	if err != nil {
		return fallback
	}
	return d
}

func getEnvList(key string, fallback []string) []string {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	parts := strings.Split(v, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	if len(out) == 0 {
		return fallback
	}
	return out
}
