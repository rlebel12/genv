package main

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/rlebel12/genv"
)

// Production-ready microservice configuration example
// This demonstrates real-world usage patterns with proper validation

// Domain types for a microservice
type (
	ServicePort      int
	DatabaseURL      string
	MicroEnvironment string
	FeatureFlag      bool
	MetricsInterval  time.Duration
	MicroLogLevel    string
	RetryPolicy      RetryConfig
)

type RetryConfig struct {
	MaxRetries int
	BackoffMs  int
}

type MicroserviceConfig struct {
	ServiceName     string
	Port            ServicePort
	Environment     MicroEnvironment
	DatabaseURL     DatabaseURL
	LogLevel        MicroLogLevel
	MetricsInterval MetricsInterval
	FeatureFlags    map[string]FeatureFlag
	ServiceTags     []string
	RetryPolicy     RetryPolicy
}

// NewProductionRegistry creates a registry with production-grade validation
func NewProductionRegistry() *genv.ParserRegistry {
	registry := genv.NewDefaultRegistry(
		// ServicePort with validation
		genv.WithParser(func(s string) (ServicePort, error) {
			port, err := strconv.Atoi(s)
			if err != nil {
				return ServicePort(0), fmt.Errorf("invalid port: %w", err)
			}
			if port < 1024 || port > 65535 {
				return ServicePort(0), errors.New("port must be between 1024-65535")
			}
			return ServicePort(port), nil
		}),
		// DatabaseURL with basic validation
		genv.WithParser(func(s string) (DatabaseURL, error) {
			if s == "" {
				return "", errors.New("database URL cannot be empty")
			}
			if !strings.HasPrefix(s, "postgres://") && !strings.HasPrefix(s, "mysql://") {
				return "", errors.New("database URL must start with postgres:// or mysql://")
			}
			return DatabaseURL(s), nil
		}),
		// Environment with strict validation
		genv.WithParser(func(s string) (MicroEnvironment, error) {
			env := strings.ToLower(s)
			switch env {
			case "development", "staging", "production":
				return MicroEnvironment(env), nil
			default:
				return "", fmt.Errorf("invalid environment: %s (must be development, staging, or production)", s)
			}
		}),
		// MetricsInterval with reasonable bounds
		genv.WithParser(func(s string) (MetricsInterval, error) {
			duration, err := time.ParseDuration(s)
			if err != nil {
				return MetricsInterval(0), fmt.Errorf("invalid duration: %w", err)
			}
			if duration < time.Second || duration > time.Hour {
				return MetricsInterval(0), errors.New("metrics interval must be between 1s and 1h")
			}
			return MetricsInterval(duration), nil
		}),
		// LogLevel with case-insensitive parsing
		genv.WithParser(func(s string) (MicroLogLevel, error) {
			level := strings.ToUpper(s)
			switch level {
			case "DEBUG", "INFO", "WARN", "ERROR", "FATAL":
				return MicroLogLevel(level), nil
			default:
				return "", fmt.Errorf("invalid log level: %s (must be DEBUG, INFO, WARN, ERROR, or FATAL)", s)
			}
		}),
	)

	return registry
}

// NewDevelopmentRegistry creates a registry with lenient validation for development
func NewDevelopmentRegistry() *genv.ParserRegistry {
	registry := genv.NewDefaultRegistry(
		// Lenient ServicePort (allows any port for dev)
		genv.WithParser(func(s string) (ServicePort, error) {
			port, err := strconv.Atoi(s)
			if err != nil {
				return ServicePort(8080), nil // Default for dev
			}
			return ServicePort(port), nil
		}),
		// Lenient DatabaseURL (allows local/test URLs)
		genv.WithParser(func(s string) (DatabaseURL, error) {
			if s == "" {
				return DatabaseURL("postgres://localhost:5432/testdb"), nil
			}
			return DatabaseURL(s), nil
		}),
		// Lenient Environment (defaults to development)
		genv.WithParser(func(s string) (MicroEnvironment, error) {
			if s == "" {
				return MicroEnvironment("development"), nil
			}
			return MicroEnvironment(strings.ToLower(s)), nil
		}),
		// Default MetricsInterval for dev
		genv.WithParser(func(s string) (MetricsInterval, error) {
			if s == "" {
				return MetricsInterval(30 * time.Second), nil
			}
			duration, err := time.ParseDuration(s)
			if err != nil {
				return MetricsInterval(30 * time.Second), nil
			}
			return MetricsInterval(duration), nil
		}),
		// Default LogLevel for dev
		genv.WithParser(func(s string) (MicroLogLevel, error) {
			if s == "" {
				return MicroLogLevel("DEBUG"), nil
			}
			return MicroLogLevel(strings.ToUpper(s)), nil
		}),
	)

	return registry
}

// LoadMicroserviceConfig demonstrates environment-specific configuration loading
func LoadMicroserviceConfig(environment string) (MicroserviceConfig, error) {
	var registry *genv.ParserRegistry
	var allowDefaults bool

	// Choose registry based on environment
	switch strings.ToLower(environment) {
	case "production":
		registry = NewProductionRegistry()
		allowDefaults = false // Strict in production
	case "development", "dev":
		registry = NewDevelopmentRegistry()
		allowDefaults = true // Lenient in development
	default:
		registry = NewProductionRegistry()
		allowDefaults = false // Default to strict
	}

	env := genv.New(
		genv.WithRegistry(registry),
		genv.WithAllowDefault(func(*genv.Genv) (bool, error) { return allowDefaults, nil }),
		genv.WithSplitKey(","),
	)

	var config MicroserviceConfig

	// Core service configuration
	env.Var("SERVICE_NAME").String(&config.ServiceName)
	genv.Type(env.Var("SERVICE_PORT").Default("8080"), &config.Port)
	genv.Type(env.Var("ENVIRONMENT").Default("development"), &config.Environment)
	genv.Type(env.Var("DATABASE_URL").Default("postgres://localhost:5432/devdb"), &config.DatabaseURL)
	genv.Type(env.Var("LOG_LEVEL").Default("INFO"), &config.LogLevel)
	genv.Type(env.Var("METRICS_INTERVAL").Default("30s"), &config.MetricsInterval)

	env.Var("SERVICE_TAGS").Default("api,backend").Optional().Strings(&config.ServiceTags)

	env.Var("RETRY_MAX_ATTEMPTS").Default("3").Int(&config.RetryPolicy.MaxRetries)
	env.Var("RETRY_BACKOFF_MS").Default("1000").Int(&config.RetryPolicy.BackoffMs)

	if err := env.Parse(); err != nil {
		return MicroserviceConfig{}, fmt.Errorf("load microservice config: %w", err)
	}

	return config, nil
}

// ValidateConfig performs additional business logic validation
func ValidateConfig(config MicroserviceConfig) error {
	if config.ServiceName == "" {
		return errors.New("service name is required")
	}

	if config.Environment == "production" {
		if config.LogLevel == MicroLogLevel("DEBUG") {
			return errors.New("DEBUG log level not allowed in production")
		}
		if config.MetricsInterval < MetricsInterval(10*time.Second) {
			return errors.New("metrics interval too aggressive for production")
		}
	}

	return nil
}
