package config

import (
	"fmt"
	"os"
	"strconv"

	"gopkg.in/yaml.v3"
)

// Config represents the application configuration
type Config struct {
	Defaults  DefaultsConfig  `yaml:"defaults"`
	Providers ProvidersConfig `yaml:"providers"`
	Logging   LoggingConfig   `yaml:"logging"`
	Security  SecurityConfig  `yaml:"security"`
}

// DefaultsConfig contains default values
type DefaultsConfig struct {
	Region      string `yaml:"region"`
	Environment string `yaml:"environment"`
	Project     string `yaml:"project"`
}

// ProvidersConfig contains provider-specific configurations
type ProvidersConfig struct {
	AWS   AWSConfig   `yaml:"aws"`
	Azure AzureConfig `yaml:"azure"`
	GCP   GCPConfig   `yaml:"gcp"`
}

// AWSConfig contains AWS-specific settings
type AWSConfig struct {
	LocalStackURL string `yaml:"localstack_url"`
	DefaultRegion string `yaml:"default_region"`
}

// AzureConfig contains Azure-specific settings
type AzureConfig struct {
	DefaultRegion       string `yaml:"default_region"`
	DefaultSubscription string `yaml:"default_subscription"`
}

// GCPConfig contains GCP-specific settings
type GCPConfig struct {
	DefaultProject string `yaml:"default_project"`
	DefaultRegion  string `yaml:"default_region"`
}

// LoggingConfig contains logging settings
type LoggingConfig struct {
	Level  string `yaml:"level"`
	Format string `yaml:"format"`
	Output string `yaml:"output"`
}

// SecurityConfig contains security settings
type SecurityConfig struct {
	RequireConfirmation bool `yaml:"require_confirmation"`
	MaxRetries          int  `yaml:"max_retries"`
	TimeoutSeconds      int  `yaml:"timeout_seconds"`
}

// LoadConfig loads configuration from file and environment variables
func LoadConfig(configPath string) (*Config, error) {
	config := &Config{}

	// Load from file if exists
	if configPath != "" {
		if err := loadFromFile(configPath, config); err != nil {
			return nil, fmt.Errorf("failed to load config file: %w", err)
		}
	}

	// Override with environment variables
	loadFromEnvironment(config)

	// Set defaults if not specified
	setDefaults(config)

	return config, nil
}

// loadFromFile loads configuration from YAML file
func loadFromFile(path string, config *Config) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	return yaml.Unmarshal(data, config)
}

// loadFromEnvironment overrides config with environment variables
func loadFromEnvironment(config *Config) {
	// Defaults
	if env := os.Getenv("BOLT_DEFAULT_REGION"); env != "" {
		config.Defaults.Region = env
	}
	if env := os.Getenv("BOLT_DEFAULT_ENVIRONMENT"); env != "" {
		config.Defaults.Environment = env
	}
	if env := os.Getenv("BOLT_DEFAULT_PROJECT"); env != "" {
		config.Defaults.Project = env
	}

	// AWS
	if env := os.Getenv("BOLT_AWS_LOCALSTACK_URL"); env != "" {
		config.Providers.AWS.LocalStackURL = env
	}
	if env := os.Getenv("BOLT_AWS_DEFAULT_REGION"); env != "" {
		config.Providers.AWS.DefaultRegion = env
	}

	// Azure
	if env := os.Getenv("BOLT_AZURE_DEFAULT_REGION"); env != "" {
		config.Providers.Azure.DefaultRegion = env
	}
	if env := os.Getenv("BOLT_AZURE_DEFAULT_SUBSCRIPTION"); env != "" {
		config.Providers.Azure.DefaultSubscription = env
	}

	// GCP
	if env := os.Getenv("BOLT_GCP_DEFAULT_PROJECT"); env != "" {
		config.Providers.GCP.DefaultProject = env
	}
	if env := os.Getenv("BOLT_GCP_DEFAULT_REGION"); env != "" {
		config.Providers.GCP.DefaultRegion = env
	}

	// Logging
	if env := os.Getenv("BOLT_LOG_LEVEL"); env != "" {
		config.Logging.Level = env
	}
	if env := os.Getenv("BOLT_LOG_FORMAT"); env != "" {
		config.Logging.Format = env
	}

	// Security
	if env := os.Getenv("BOLT_REQUIRE_CONFIRMATION"); env != "" {
		if val, err := strconv.ParseBool(env); err == nil {
			config.Security.RequireConfirmation = val
		}
	}
	if env := os.Getenv("BOLT_MAX_RETRIES"); env != "" {
		if val, err := strconv.Atoi(env); err == nil {
			config.Security.MaxRetries = val
		}
	}
	if env := os.Getenv("BOLT_TIMEOUT_SECONDS"); env != "" {
		if val, err := strconv.Atoi(env); err == nil {
			config.Security.TimeoutSeconds = val
		}
	}
}

// setDefaults sets default values if not specified
func setDefaults(config *Config) {
	// Defaults
	if config.Defaults.Region == "" {
		config.Defaults.Region = "us-east-1"
	}
	if config.Defaults.Environment == "" {
		config.Defaults.Environment = "local"
	}

	// AWS
	if config.Providers.AWS.LocalStackURL == "" {
		config.Providers.AWS.LocalStackURL = "http://localhost:4566"
	}
	if config.Providers.AWS.DefaultRegion == "" {
		config.Providers.AWS.DefaultRegion = "us-east-1"
	}

	// Azure
	if config.Providers.Azure.DefaultRegion == "" {
		config.Providers.Azure.DefaultRegion = "eastus"
	}

	// GCP
	if config.Providers.GCP.DefaultRegion == "" {
		config.Providers.GCP.DefaultRegion = "us-central1"
	}

	// Logging
	if config.Logging.Level == "" {
		config.Logging.Level = "info"
	}
	if config.Logging.Format == "" {
		config.Logging.Format = "text"
	}
	if config.Logging.Output == "" {
		config.Logging.Output = "stdout"
	}

	// Security
	if config.Security.MaxRetries == 0 {
		config.Security.MaxRetries = 3
	}
	if config.Security.TimeoutSeconds == 0 {
		config.Security.TimeoutSeconds = 300
	}
	if !config.Security.RequireConfirmation {
		config.Security.RequireConfirmation = true
	}
}

// GetProviderConfig returns provider-specific configuration
func (c *Config) GetProviderConfig(providerType string) map[string]interface{} {
	switch providerType {
	case "aws":
		return map[string]interface{}{
			"localstack_url": c.Providers.AWS.LocalStackURL,
			"default_region": c.Providers.AWS.DefaultRegion,
		}
	case "azurerm":
		return map[string]interface{}{
			"default_region":       c.Providers.Azure.DefaultRegion,
			"default_subscription": c.Providers.Azure.DefaultSubscription,
		}
	case "google":
		return map[string]interface{}{
			"default_project": c.Providers.GCP.DefaultProject,
			"default_region":  c.Providers.GCP.DefaultRegion,
		}
	default:
		return map[string]interface{}{}
	}
}
