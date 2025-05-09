package config

import (
	"context"
	"fmt"
	"os"
)

// AWSConfigProvider implements Provider using AWS Secrets Manager
type AWSConfigProvider struct {
	secretsProvider Provider
}

// NewAWSConfigProvider creates a new AWS configuration provider
func NewAWSConfigProvider() (Provider, error) {
	// Get secret name from environment variable
	secretName := os.Getenv("AWS_SECRET_NAME")
	if secretName == "" {
		return nil, fmt.Errorf("AWS_SECRET_NAME environment variable not set")
	}

	// Create secrets provider
	secretsProvider, err := NewAWSSecretsProvider(secretName)
	if err != nil {
		return nil, fmt.Errorf("failed to create AWS secrets provider: %w", err)
	}

	return &AWSConfigProvider{
		secretsProvider: secretsProvider,
	}, nil
}

// GetEnvironment returns the current environment
func (p *AWSConfigProvider) GetEnvironment() Environment {
	return p.secretsProvider.GetEnvironment()
}

// GetString retrieves a string configuration value
func (p *AWSConfigProvider) GetString(ctx context.Context, key string) (string, error) {
	return p.secretsProvider.GetString(ctx, key)
}

// GetInt retrieves an integer configuration value
func (p *AWSConfigProvider) GetInt(ctx context.Context, key string) (int, error) {
	return p.secretsProvider.GetInt(ctx, key)
}

// GetBool retrieves a boolean configuration value
func (p *AWSConfigProvider) GetBool(ctx context.Context, key string) (bool, error) {
	return p.secretsProvider.GetBool(ctx, key)
}

// GetSecret retrieves a secret value
func (p *AWSConfigProvider) GetSecret(ctx context.Context, key string) (string, error) {
	return p.secretsProvider.GetSecret(ctx, key)
}
