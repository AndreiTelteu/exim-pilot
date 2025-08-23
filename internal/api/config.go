package api

import (
	"fmt"
	"os"
	"strconv"
)

// Config holds the API server configuration
type Config struct {
	Port           int
	Host           string
	ReadTimeout    int // seconds
	WriteTimeout   int // seconds
	IdleTimeout    int // seconds
	AllowedOrigins []string
	LogRequests    bool
}

// NewConfig creates a new configuration with defaults
func NewConfig() *Config {
	return &Config{
		Port:           8080,
		Host:           "0.0.0.0",
		ReadTimeout:    15,
		WriteTimeout:   15,
		IdleTimeout:    60,
		AllowedOrigins: []string{"*"}, // In production, specify exact origins
		LogRequests:    true,
	}
}

// LoadFromEnv loads configuration from environment variables
func (c *Config) LoadFromEnv() {
	if port := os.Getenv("API_PORT"); port != "" {
		if p, err := strconv.Atoi(port); err == nil {
			c.Port = p
		}
	}

	if host := os.Getenv("API_HOST"); host != "" {
		c.Host = host
	}

	if timeout := os.Getenv("API_READ_TIMEOUT"); timeout != "" {
		if t, err := strconv.Atoi(timeout); err == nil {
			c.ReadTimeout = t
		}
	}

	if timeout := os.Getenv("API_WRITE_TIMEOUT"); timeout != "" {
		if t, err := strconv.Atoi(timeout); err == nil {
			c.WriteTimeout = t
		}
	}

	if timeout := os.Getenv("API_IDLE_TIMEOUT"); timeout != "" {
		if t, err := strconv.Atoi(timeout); err == nil {
			c.IdleTimeout = t
		}
	}

	if logRequests := os.Getenv("API_LOG_REQUESTS"); logRequests != "" {
		c.LogRequests = logRequests == "true"
	}
}

// GetAddress returns the full address string for the server
func (c *Config) GetAddress() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}
