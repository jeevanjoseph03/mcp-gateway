package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// Config holds all configuration for the MCP gateway
// This is loaded from config.yaml and environment variables
type Config struct {
	Gateway GatewayConfig  `yaml:"gateway"`
	Servers []ServerConfig `yaml:"servers"`
	Logging LoggingConfig  `yaml:"logging"`
}

// GatewayConfig configures the gateway itself
type GatewayConfig struct {
	Host         string        `yaml:"host"`          // What IP to bind to (0.0.0.0 for all interfaces)
	Port         int           `yaml:"port"`          // What port to listen on
	ReadTimeout  time.Duration `yaml:"read_timeout"`  // Max time to read request
	WriteTimeout time.Duration `yaml:"write_timeout"` // Max time to write response
	IdleTimeout  time.Duration `yaml:"idle_timeout"`  // Max time for keep-alive connections
}

// ServerConfig defines a backend MCP server this gateway can route to
type ServerConfig struct {
	ID          string            `yaml:"id"`          // Unique identifier for this server
	Name        string            `yaml:"name"`        // Human-readable name
	URL         string            `yaml:"url"`         // Where to find this server (http:// or ws://)
	Transport   string            `yaml:"transport"`   // "http" or "websocket"
	Timeout     time.Duration     `yaml:"timeout"`     // Max time to wait for response
	RetryCount  int               `yaml:"retry_count"` // How many times to retry on failure
	HealthCheck HealthCheckConfig `yaml:"health_check"`
	Tools       []string          `yaml:"tools,omitempty"` // Optional: list of tools this server provides
}

// HealthCheckConfig configures how we monitor server health
type HealthCheckConfig struct {
	Enabled          bool          `yaml:"enabled"`           // Should we check health?
	Interval         time.Duration `yaml:"interval"`          // How often to check
	Timeout          time.Duration `yaml:"timeout"`           // Max time for health check
	FailureThreshold int           `yaml:"failure_threshold"` // How many failures before marking unhealthy
	SuccessThreshold int           `yaml:"success_threshold"` // How many successes before marking healthy again
}

// LoggingConfig configures logging behavior
type LoggingConfig struct {
	Level  string `yaml:"level"`  // debug, info, warn, error
	Format string `yaml:"format"` // json or text
	Output string `yaml:"output"` // stdout, stderr, or file path
}

// Load reads and parses the configuration file
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Apply defaults for missing values
	cfg.applyDefaults()

	// Validate the configuration
	if err := cfg.validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return &cfg, nil
}

// applyDefaults fills in sensible default values
func (c *Config) applyDefaults() {
	if c.Gateway.Host == "" {
		c.Gateway.Host = "0.0.0.0"
	}
	if c.Gateway.Port == 0 {
		c.Gateway.Port = 8080
	}
	if c.Gateway.ReadTimeout == 0 {
		c.Gateway.ReadTimeout = 30 * time.Second
	}
	if c.Gateway.WriteTimeout == 0 {
		c.Gateway.WriteTimeout = 30 * time.Second
	}
	if c.Gateway.IdleTimeout == 0 {
		c.Gateway.IdleTimeout = 120 * time.Second
	}

	for i := range c.Servers {
		if c.Servers[i].Timeout == 0 {
			c.Servers[i].Timeout = 30 * time.Second
		}
		if c.Servers[i].RetryCount == 0 {
			c.Servers[i].RetryCount = 3
		}
		if c.Servers[i].Transport == "" {
			c.Servers[i].Transport = "http"
		}

		// Health check defaults
		if c.Servers[i].HealthCheck.Interval == 0 {
			c.Servers[i].HealthCheck.Interval = 30 * time.Second
		}
		if c.Servers[i].HealthCheck.Timeout == 0 {
			c.Servers[i].HealthCheck.Timeout = 5 * time.Second
		}
		if c.Servers[i].HealthCheck.FailureThreshold == 0 {
			c.Servers[i].HealthCheck.FailureThreshold = 3
		}
		if c.Servers[i].HealthCheck.SuccessThreshold == 0 {
			c.Servers[i].HealthCheck.SuccessThreshold = 2
		}
	}

	if c.Logging.Level == "" {
		c.Logging.Level = "info"
	}
	if c.Logging.Format == "" {
		c.Logging.Format = "json"
	}
	if c.Logging.Output == "" {
		c.Logging.Output = "stdout"
	}
}

// validate checks if the configuration is usable
func (c *Config) validate() error {
	if c.Gateway.Port <= 0 || c.Gateway.Port > 65535 {
		return fmt.Errorf("invalid port: %d", c.Gateway.Port)
	}

	for i, server := range c.Servers {
		if server.ID == "" {
			return fmt.Errorf("server %d has no id", i)
		}
		if server.URL == "" {
			return fmt.Errorf("server %s has no url", server.ID)
		}
		if server.Transport != "http" && server.Transport != "websocket" {
			return fmt.Errorf("server %s has invalid transport: %s", server.ID, server.Transport)
		}
	}

	return nil
}
