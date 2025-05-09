package config

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
)

// Environment represents the application environment
type Environment string

const (
	Development Environment = "development"
	Staging     Environment = "staging"
	Production  Environment = "production"
)

// ValidationError represents a configuration validation error
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

// Provider defines the interface for configuration management
type Provider interface {
	// GetString retrieves a string configuration value
	GetString(ctx context.Context, key string) (string, error)
	// GetInt retrieves an integer configuration value
	GetInt(ctx context.Context, key string) (int, error)
	// GetBool retrieves a boolean configuration value
	GetBool(ctx context.Context, key string) (bool, error)
	// GetSecret retrieves a secret value
	GetSecret(ctx context.Context, key string) (string, error)
	// GetEnvironment returns the current environment
	GetEnvironment() Environment
}

// EnvProvider implements Provider using environment variables
type EnvProvider struct {
	prefix      string
	environment Environment
}

// NewEnvProvider creates a new environment-based configuration provider
func NewEnvProvider(prefix string) Provider {
	env := os.Getenv("APP_ENV")
	if env == "" {
		env = string(Development)
	}
	return &EnvProvider{
		prefix:      prefix,
		environment: Environment(env),
	}
}

// GetEnvironment returns the current environment
func (p *EnvProvider) GetEnvironment() Environment {
	return p.environment
}

// GetString retrieves a string configuration value from environment variables
func (p *EnvProvider) GetString(ctx context.Context, key string) (string, error) {
	value := os.Getenv(p.prefix + key)
	if value == "" {
		return "", fmt.Errorf("environment variable %s%s not set", p.prefix, key)
	}
	return value, nil
}

// GetInt retrieves an integer configuration value from environment variables
func (p *EnvProvider) GetInt(ctx context.Context, key string) (int, error) {
	value, err := p.GetString(ctx, key)
	if err != nil {
		return 0, err
	}
	return strconv.Atoi(value)
}

// GetBool retrieves a boolean configuration value from environment variables
func (p *EnvProvider) GetBool(ctx context.Context, key string) (bool, error) {
	value, err := p.GetString(ctx, key)
	if err != nil {
		return false, err
	}
	return strconv.ParseBool(value)
}

// GetSecret retrieves a secret value from environment variables
func (p *EnvProvider) GetSecret(ctx context.Context, key string) (string, error) {
	return p.GetString(ctx, key)
}

// AWSSecretsProvider implements Provider using AWS Secrets Manager
type AWSSecretsProvider struct {
	client      *secretsmanager.Client
	secretName  string
	cache       map[string]string
	lastFetch   time.Time
	environment Environment
}

// NewAWSSecretsProvider creates a new AWS Secrets Manager based configuration provider
func NewAWSSecretsProvider(secretName string) (Provider, error) {
	// Load AWS configuration
	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	// Get environment from AWS Systems Manager Parameter Store or environment variable
	env := os.Getenv("APP_ENV")
	if env == "" {
		env = string(Development)
	}

	return &AWSSecretsProvider{
		client:      secretsmanager.NewFromConfig(cfg),
		secretName:  secretName,
		cache:       make(map[string]string),
		environment: Environment(env),
	}, nil
}

// GetEnvironment returns the current environment
func (p *AWSSecretsProvider) GetEnvironment() Environment {
	return p.environment
}

// GetString retrieves a string configuration value from AWS Secrets Manager
func (p *AWSSecretsProvider) GetString(ctx context.Context, key string) (string, error) {
	// Check cache first
	if value, ok := p.cache[key]; ok {
		return value, nil
	}

	// Fetch secret from AWS Secrets Manager
	secret, err := p.client.GetSecretValue(ctx, &secretsmanager.GetSecretValueInput{
		SecretId: aws.String(p.secretName),
	})
	if err != nil {
		return "", fmt.Errorf("failed to get secret: %w", err)
	}

	// Parse secret string as JSON
	var secretMap map[string]string
	if err := json.Unmarshal([]byte(*secret.SecretString), &secretMap); err != nil {
		return "", fmt.Errorf("failed to parse secret JSON: %w", err)
	}

	// Validate secret schema
	if err := validateSecretSchema(secretMap, p.environment); err != nil {
		return "", fmt.Errorf("invalid secret schema: %w", err)
	}

	// Update cache
	p.cache = secretMap
	p.lastFetch = time.Now()

	// Return requested value
	value, ok := secretMap[key]
	if !ok {
		return "", fmt.Errorf("secret key %s not found", key)
	}
	return value, nil
}

// GetInt retrieves an integer configuration value from AWS Secrets Manager
func (p *AWSSecretsProvider) GetInt(ctx context.Context, key string) (int, error) {
	value, err := p.GetString(ctx, key)
	if err != nil {
		return 0, err
	}
	return strconv.Atoi(value)
}

// GetBool retrieves a boolean configuration value from AWS Secrets Manager
func (p *AWSSecretsProvider) GetBool(ctx context.Context, key string) (bool, error) {
	value, err := p.GetString(ctx, key)
	if err != nil {
		return false, err
	}
	return strconv.ParseBool(value)
}

// GetSecret retrieves a secret value from AWS Secrets Manager
func (p *AWSSecretsProvider) GetSecret(ctx context.Context, key string) (string, error) {
	return p.GetString(ctx, key)
}

// DatabaseConfig holds database connection configuration
type DatabaseConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
	SSLMode  string
}

// Validate checks if the database configuration is valid
func (c *DatabaseConfig) Validate(env Environment) error {
	if c.Host == "" {
		return &ValidationError{Field: "Host", Message: "host cannot be empty"}
	}

	// Validate host is a valid hostname or IP
	if host := net.ParseIP(c.Host); host == nil {
		if _, err := net.LookupHost(c.Host); err != nil {
			return &ValidationError{Field: "Host", Message: "invalid hostname or IP address"}
		}
	}

	if c.Port <= 0 || c.Port > 65535 {
		return &ValidationError{Field: "Port", Message: "port must be between 1 and 65535"}
	}

	if c.User == "" {
		return &ValidationError{Field: "User", Message: "user cannot be empty"}
	}

	if c.Password == "" {
		return &ValidationError{Field: "Password", Message: "password cannot be empty"}
	}

	// Stricter password validation for production
	if env == Production {
		if len(c.Password) < 12 {
			return &ValidationError{Field: "Password", Message: "password must be at least 12 characters long in production"}
		}
		if !regexp.MustCompile(`[A-Z]`).MatchString(c.Password) {
			return &ValidationError{Field: "Password", Message: "password must contain at least one uppercase letter in production"}
		}
		if !regexp.MustCompile(`[a-z]`).MatchString(c.Password) {
			return &ValidationError{Field: "Password", Message: "password must contain at least one lowercase letter in production"}
		}
		if !regexp.MustCompile(`[0-9]`).MatchString(c.Password) {
			return &ValidationError{Field: "Password", Message: "password must contain at least one number in production"}
		}
		if !regexp.MustCompile(`[^A-Za-z0-9]`).MatchString(c.Password) {
			return &ValidationError{Field: "Password", Message: "password must contain at least one special character in production"}
		}
	}

	if c.DBName == "" {
		return &ValidationError{Field: "DBName", Message: "database name cannot be empty"}
	}

	// Validate database name format
	if !regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9_]*$`).MatchString(c.DBName) {
		return &ValidationError{Field: "DBName", Message: "database name must start with a letter and contain only letters, numbers, and underscores"}
	}

	// Validate SSL mode
	validSSLModes := map[string]bool{
		"disable":     true,
		"require":     true,
		"verify-ca":   true,
		"verify-full": true,
	}
	if !validSSLModes[c.SSLMode] {
		return &ValidationError{Field: "SSLMode", Message: "invalid SSL mode"}
	}

	// Require SSL in production
	if env == Production && c.SSLMode == "disable" {
		return &ValidationError{Field: "SSLMode", Message: "SSL cannot be disabled in production"}
	}

	return nil
}

// validateSecretSchema validates the structure of secrets stored in AWS Secrets Manager
func validateSecretSchema(secrets map[string]string, env Environment) error {
	requiredKeys := []string{
		"DB_HOST",
		"DB_PORT",
		"DB_USER",
		"DB_PASSWORD",
		"DB_NAME",
		"DB_SSLMODE",
	}

	// Check for required keys
	for _, key := range requiredKeys {
		if _, ok := secrets[key]; !ok {
			return &ValidationError{
				Field:   key,
				Message: "required secret key not found",
			}
		}
	}

	// Validate port is a number
	if _, err := strconv.Atoi(secrets["DB_PORT"]); err != nil {
		return &ValidationError{
			Field:   "DB_PORT",
			Message: "port must be a valid number",
		}
	}

	// Validate SSL mode
	validSSLModes := map[string]bool{
		"disable":     true,
		"require":     true,
		"verify-ca":   true,
		"verify-full": true,
	}
	if !validSSLModes[secrets["DB_SSLMODE"]] {
		return &ValidationError{
			Field:   "DB_SSLMODE",
			Message: "invalid SSL mode",
		}
	}

	// Stricter validation for production
	if env == Production {
		// Validate host is not localhost in production
		if strings.ToLower(secrets["DB_HOST"]) == "localhost" {
			return &ValidationError{
				Field:   "DB_HOST",
				Message: "localhost is not allowed in production",
			}
		}

		// Validate SSL is enabled in production
		if secrets["DB_SSLMODE"] == "disable" {
			return &ValidationError{
				Field:   "DB_SSLMODE",
				Message: "SSL cannot be disabled in production",
			}
		}

		// Validate password complexity in production
		password := secrets["DB_PASSWORD"]
		if len(password) < 12 {
			return &ValidationError{
				Field:   "DB_PASSWORD",
				Message: "password must be at least 12 characters long in production",
			}
		}
		if !regexp.MustCompile(`[A-Z]`).MatchString(password) {
			return &ValidationError{
				Field:   "DB_PASSWORD",
				Message: "password must contain at least one uppercase letter in production",
			}
		}
		if !regexp.MustCompile(`[a-z]`).MatchString(password) {
			return &ValidationError{
				Field:   "DB_PASSWORD",
				Message: "password must contain at least one lowercase letter in production",
			}
		}
		if !regexp.MustCompile(`[0-9]`).MatchString(password) {
			return &ValidationError{
				Field:   "DB_PASSWORD",
				Message: "password must contain at least one number in production",
			}
		}
		if !regexp.MustCompile(`[^A-Za-z0-9]`).MatchString(password) {
			return &ValidationError{
				Field:   "DB_PASSWORD",
				Message: "password must contain at least one special character in production",
			}
		}
	}

	return nil
}

// GetDatabaseConfig retrieves database configuration using the provided config provider
func GetDatabaseConfig(ctx context.Context, provider Provider) (*DatabaseConfig, error) {
	host, err := provider.GetString(ctx, "DB_HOST")
	if err != nil {
		return nil, fmt.Errorf("failed to get DB_HOST: %w", err)
	}

	port, err := provider.GetInt(ctx, "DB_PORT")
	if err != nil {
		return nil, fmt.Errorf("failed to get DB_PORT: %w", err)
	}

	user, err := provider.GetString(ctx, "DB_USER")
	if err != nil {
		return nil, fmt.Errorf("failed to get DB_USER: %w", err)
	}

	password, err := provider.GetSecret(ctx, "DB_PASSWORD")
	if err != nil {
		return nil, fmt.Errorf("failed to get DB_PASSWORD: %w", err)
	}

	dbname, err := provider.GetString(ctx, "DB_NAME")
	if err != nil {
		return nil, fmt.Errorf("failed to get DB_NAME: %w", err)
	}

	sslmode, err := provider.GetString(ctx, "DB_SSLMODE")
	if err != nil {
		sslmode = "disable" // Default to disable if not set
	}

	cfg := &DatabaseConfig{
		Host:     host,
		Port:     port,
		User:     user,
		Password: password,
		DBName:   dbname,
		SSLMode:  sslmode,
	}

	// Validate configuration
	if err := cfg.Validate(provider.GetEnvironment()); err != nil {
		return nil, fmt.Errorf("invalid database configuration: %w", err)
	}

	return cfg, nil
}
