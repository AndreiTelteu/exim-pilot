package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// Config represents the complete application configuration
type Config struct {
	Server    ServerConfig    `yaml:"server" json:"server"`
	Database  DatabaseConfig  `yaml:"database" json:"database"`
	Exim      EximConfig      `yaml:"exim" json:"exim"`
	Logging   LoggingConfig   `yaml:"logging" json:"logging"`
	Retention RetentionConfig `yaml:"retention" json:"retention"`
	Security  SecurityConfig  `yaml:"security" json:"security"`
	Auth      AuthConfig      `yaml:"auth" json:"auth"`
}

// ServerConfig holds HTTP server configuration
type ServerConfig struct {
	Port           int      `yaml:"port" json:"port"`
	Host           string   `yaml:"host" json:"host"`
	ReadTimeout    int      `yaml:"read_timeout" json:"read_timeout"`   // seconds
	WriteTimeout   int      `yaml:"write_timeout" json:"write_timeout"` // seconds
	IdleTimeout    int      `yaml:"idle_timeout" json:"idle_timeout"`   // seconds
	AllowedOrigins []string `yaml:"allowed_origins" json:"allowed_origins"`
	LogRequests    bool     `yaml:"log_requests" json:"log_requests"`
	TLSEnabled     bool     `yaml:"tls_enabled" json:"tls_enabled"`
	TLSCertFile    string   `yaml:"tls_cert_file" json:"tls_cert_file"`
	TLSKeyFile     string   `yaml:"tls_key_file" json:"tls_key_file"`
}

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	Path            string `yaml:"path" json:"path"`
	MaxOpenConns    int    `yaml:"max_open_conns" json:"max_open_conns"`
	MaxIdleConns    int    `yaml:"max_idle_conns" json:"max_idle_conns"`
	ConnMaxLifetime int    `yaml:"conn_max_lifetime" json:"conn_max_lifetime"` // minutes
	BackupEnabled   bool   `yaml:"backup_enabled" json:"backup_enabled"`
	BackupInterval  int    `yaml:"backup_interval" json:"backup_interval"` // hours
	BackupPath      string `yaml:"backup_path" json:"backup_path"`
}

// EximConfig holds Exim-specific configuration
type EximConfig struct {
	LogPaths       []string `yaml:"log_paths" json:"log_paths"`
	SpoolDir       string   `yaml:"spool_dir" json:"spool_dir"`
	BinaryPath     string   `yaml:"binary_path" json:"binary_path"`
	ConfigFile     string   `yaml:"config_file" json:"config_file"`
	QueueRunUser   string   `yaml:"queue_run_user" json:"queue_run_user"`
	LogRotationDir string   `yaml:"log_rotation_dir" json:"log_rotation_dir"`
}

// LoggingConfig holds application logging configuration
type LoggingConfig struct {
	Level      string `yaml:"level" json:"level"`
	File       string `yaml:"file" json:"file"`
	MaxSize    int    `yaml:"max_size" json:"max_size"`       // MB
	MaxBackups int    `yaml:"max_backups" json:"max_backups"` // number of backup files
	MaxAge     int    `yaml:"max_age" json:"max_age"`         // days
	Compress   bool   `yaml:"compress" json:"compress"`
}

// RetentionConfig holds data retention policies
type RetentionConfig struct {
	LogEntriesDays      int `yaml:"log_entries_days" json:"log_entries_days"`
	AuditLogDays        int `yaml:"audit_log_days" json:"audit_log_days"`
	QueueSnapshotsDays  int `yaml:"queue_snapshots_days" json:"queue_snapshots_days"`
	DeliveryAttemptDays int `yaml:"delivery_attempt_days" json:"delivery_attempt_days"`
	CleanupInterval     int `yaml:"cleanup_interval" json:"cleanup_interval"` // hours
}

// SecurityConfig holds security-related configuration
type SecurityConfig struct {
	SessionTimeout   int      `yaml:"session_timeout" json:"session_timeout"`       // minutes
	MaxLoginAttempts int      `yaml:"max_login_attempts" json:"max_login_attempts"` // per IP
	LoginLockoutTime int      `yaml:"login_lockout_time" json:"login_lockout_time"` // minutes
	CSRFProtection   bool     `yaml:"csrf_protection" json:"csrf_protection"`
	SecureCookies    bool     `yaml:"secure_cookies" json:"secure_cookies"`
	ContentRedaction bool     `yaml:"content_redaction" json:"content_redaction"`
	AuditAllActions  bool     `yaml:"audit_all_actions" json:"audit_all_actions"`
	TrustedProxies   []string `yaml:"trusted_proxies" json:"trusted_proxies"`
}

// AuthConfig holds authentication configuration
type AuthConfig struct {
	DefaultUsername string `yaml:"default_username" json:"default_username"`
	DefaultPassword string `yaml:"default_password" json:"default_password"`
	PasswordMinLen  int    `yaml:"password_min_length" json:"password_min_length"`
	RequireStrongPw bool   `yaml:"require_strong_password" json:"require_strong_password"`
	SessionSecret   string `yaml:"session_secret" json:"session_secret"`
}

// DefaultConfig returns a configuration with sensible defaults
func DefaultConfig() *Config {
	return &Config{
		Server: ServerConfig{
			Port:           8080,
			Host:           "0.0.0.0",
			ReadTimeout:    15,
			WriteTimeout:   15,
			IdleTimeout:    60,
			AllowedOrigins: []string{"*"},
			LogRequests:    true,
			TLSEnabled:     false,
		},
		Database: DatabaseConfig{
			Path:            "/opt/exim-pilot/data/exim-pilot.db",
			MaxOpenConns:    25,
			MaxIdleConns:    5,
			ConnMaxLifetime: 5, // minutes
			BackupEnabled:   true,
			BackupInterval:  24, // hours
			BackupPath:      "/opt/exim-pilot/backups",
		},
		Exim: EximConfig{
			LogPaths: []string{
				"/var/log/exim4/mainlog",
				"/var/log/exim4/rejectlog",
				"/var/log/exim4/paniclog",
			},
			SpoolDir:       "/var/spool/exim4",
			BinaryPath:     "/usr/sbin/exim4",
			ConfigFile:     "/etc/exim4/exim4.conf",
			QueueRunUser:   "Debian-exim",
			LogRotationDir: "/var/log/exim4",
		},
		Logging: LoggingConfig{
			Level:      "info",
			File:       "/opt/exim-pilot/logs/exim-pilot.log",
			MaxSize:    100, // MB
			MaxBackups: 5,
			MaxAge:     30, // days
			Compress:   true,
		},
		Retention: RetentionConfig{
			LogEntriesDays:      90,
			AuditLogDays:        365,
			QueueSnapshotsDays:  30,
			DeliveryAttemptDays: 180,
			CleanupInterval:     24, // hours
		},
		Security: SecurityConfig{
			SessionTimeout:   60, // minutes
			MaxLoginAttempts: 5,
			LoginLockoutTime: 15, // minutes
			CSRFProtection:   true,
			SecureCookies:    true,
			ContentRedaction: true,
			AuditAllActions:  true,
			TrustedProxies:   []string{},
		},
		Auth: AuthConfig{
			DefaultUsername: "admin",
			DefaultPassword: "admin123",
			PasswordMinLen:  8,
			RequireStrongPw: true,
			SessionSecret:   "", // Will be generated if empty
		},
	}
}

// LoadFromFile loads configuration from a YAML file
func LoadFromFile(path string) (*Config, error) {
	config := DefaultConfig()

	// Check if file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return config, fmt.Errorf("configuration file not found: %s", path)
	}

	// Read file
	data, err := os.ReadFile(path)
	if err != nil {
		return config, fmt.Errorf("failed to read configuration file: %w", err)
	}

	// Parse YAML
	if err := yaml.Unmarshal(data, config); err != nil {
		return config, fmt.Errorf("failed to parse configuration file: %w", err)
	}

	// Load environment overrides
	config.LoadFromEnv()

	// Validate configuration
	if err := config.Validate(); err != nil {
		return config, fmt.Errorf("configuration validation failed: %w", err)
	}

	return config, nil
}

// LoadFromEnv loads configuration overrides from environment variables
func (c *Config) LoadFromEnv() {
	// Server configuration
	if port := os.Getenv("EXIM_PILOT_PORT"); port != "" {
		if p, err := strconv.Atoi(port); err == nil {
			c.Server.Port = p
		}
	}

	if host := os.Getenv("EXIM_PILOT_HOST"); host != "" {
		c.Server.Host = host
	}

	if timeout := os.Getenv("EXIM_PILOT_READ_TIMEOUT"); timeout != "" {
		if t, err := strconv.Atoi(timeout); err == nil {
			c.Server.ReadTimeout = t
		}
	}

	if timeout := os.Getenv("EXIM_PILOT_WRITE_TIMEOUT"); timeout != "" {
		if t, err := strconv.Atoi(timeout); err == nil {
			c.Server.WriteTimeout = t
		}
	}

	if origins := os.Getenv("EXIM_PILOT_ALLOWED_ORIGINS"); origins != "" {
		c.Server.AllowedOrigins = strings.Split(origins, ",")
	}

	if tls := os.Getenv("EXIM_PILOT_TLS_ENABLED"); tls != "" {
		c.Server.TLSEnabled = tls == "true"
	}

	if cert := os.Getenv("EXIM_PILOT_TLS_CERT"); cert != "" {
		c.Server.TLSCertFile = cert
	}

	if key := os.Getenv("EXIM_PILOT_TLS_KEY"); key != "" {
		c.Server.TLSKeyFile = key
	}

	// Database configuration
	if dbPath := os.Getenv("EXIM_PILOT_DB_PATH"); dbPath != "" {
		c.Database.Path = dbPath
	}

	if maxConns := os.Getenv("EXIM_PILOT_DB_MAX_CONNS"); maxConns != "" {
		if m, err := strconv.Atoi(maxConns); err == nil {
			c.Database.MaxOpenConns = m
		}
	}

	// Exim configuration
	if logPaths := os.Getenv("EXIM_PILOT_LOG_PATHS"); logPaths != "" {
		c.Exim.LogPaths = strings.Split(logPaths, ",")
	}

	if spoolDir := os.Getenv("EXIM_PILOT_SPOOL_DIR"); spoolDir != "" {
		c.Exim.SpoolDir = spoolDir
	}

	if binaryPath := os.Getenv("EXIM_PILOT_BINARY_PATH"); binaryPath != "" {
		c.Exim.BinaryPath = binaryPath
	}

	// Logging configuration
	if logLevel := os.Getenv("EXIM_PILOT_LOG_LEVEL"); logLevel != "" {
		c.Logging.Level = logLevel
	}

	if logFile := os.Getenv("EXIM_PILOT_LOG_FILE"); logFile != "" {
		c.Logging.File = logFile
	}

	// Auth configuration
	if adminUser := os.Getenv("EXIM_PILOT_ADMIN_USER"); adminUser != "" {
		c.Auth.DefaultUsername = adminUser
	}

	if adminPass := os.Getenv("EXIM_PILOT_ADMIN_PASSWORD"); adminPass != "" {
		c.Auth.DefaultPassword = adminPass
	}

	if sessionSecret := os.Getenv("EXIM_PILOT_SESSION_SECRET"); sessionSecret != "" {
		c.Auth.SessionSecret = sessionSecret
	}

	// Security configuration
	if sessionTimeout := os.Getenv("EXIM_PILOT_SESSION_TIMEOUT"); sessionTimeout != "" {
		if t, err := strconv.Atoi(sessionTimeout); err == nil {
			c.Security.SessionTimeout = t
		}
	}

	if secureCookies := os.Getenv("EXIM_PILOT_SECURE_COOKIES"); secureCookies != "" {
		c.Security.SecureCookies = secureCookies == "true"
	}
}

// Validate validates the configuration
func (c *Config) Validate() error {
	// Validate server configuration
	if c.Server.Port < 1 || c.Server.Port > 65535 {
		return fmt.Errorf("invalid server port: %d", c.Server.Port)
	}

	if c.Server.Host == "" {
		return fmt.Errorf("server host cannot be empty")
	}

	if c.Server.TLSEnabled {
		if c.Server.TLSCertFile == "" || c.Server.TLSKeyFile == "" {
			return fmt.Errorf("TLS enabled but cert or key file not specified")
		}

		if _, err := os.Stat(c.Server.TLSCertFile); os.IsNotExist(err) {
			return fmt.Errorf("TLS cert file not found: %s", c.Server.TLSCertFile)
		}

		if _, err := os.Stat(c.Server.TLSKeyFile); os.IsNotExist(err) {
			return fmt.Errorf("TLS key file not found: %s", c.Server.TLSKeyFile)
		}
	}

	// Validate database configuration
	if c.Database.Path == "" {
		return fmt.Errorf("database path cannot be empty")
	}

	// Ensure database directory exists or can be created
	dbDir := filepath.Dir(c.Database.Path)
	if err := os.MkdirAll(dbDir, 0755); err != nil {
		return fmt.Errorf("cannot create database directory %s: %w", dbDir, err)
	}

	if c.Database.MaxOpenConns < 1 {
		return fmt.Errorf("database max_open_conns must be at least 1")
	}

	if c.Database.MaxIdleConns < 0 {
		return fmt.Errorf("database max_idle_conns cannot be negative")
	}

	// Validate Exim configuration
	if len(c.Exim.LogPaths) == 0 {
		return fmt.Errorf("at least one Exim log path must be specified")
	}

	for _, logPath := range c.Exim.LogPaths {
		if logPath == "" {
			return fmt.Errorf("Exim log path cannot be empty")
		}
	}

	if c.Exim.SpoolDir == "" {
		return fmt.Errorf("Exim spool directory cannot be empty")
	}

	if c.Exim.BinaryPath == "" {
		return fmt.Errorf("Exim binary path cannot be empty")
	}

	// Check if Exim binary exists
	if _, err := os.Stat(c.Exim.BinaryPath); os.IsNotExist(err) {
		return fmt.Errorf("Exim binary not found: %s", c.Exim.BinaryPath)
	}

	// Validate logging configuration
	validLogLevels := map[string]bool{
		"debug": true, "info": true, "warn": true, "error": true, "fatal": true,
	}

	if !validLogLevels[c.Logging.Level] {
		return fmt.Errorf("invalid log level: %s", c.Logging.Level)
	}

	if c.Logging.File != "" {
		logDir := filepath.Dir(c.Logging.File)
		if err := os.MkdirAll(logDir, 0755); err != nil {
			return fmt.Errorf("cannot create log directory %s: %w", logDir, err)
		}
	}

	// Validate retention configuration
	if c.Retention.LogEntriesDays < 1 {
		return fmt.Errorf("log entries retention must be at least 1 day")
	}

	if c.Retention.AuditLogDays < 1 {
		return fmt.Errorf("audit log retention must be at least 1 day")
	}

	// Validate auth configuration
	if c.Auth.DefaultUsername == "" {
		return fmt.Errorf("default username cannot be empty")
	}

	if c.Auth.DefaultPassword == "" {
		return fmt.Errorf("default password cannot be empty")
	}

	if c.Auth.PasswordMinLen < 4 {
		return fmt.Errorf("minimum password length must be at least 4")
	}

	// Generate session secret if not provided
	if c.Auth.SessionSecret == "" {
		c.Auth.SessionSecret = generateSessionSecret()
	}

	// Validate security configuration
	if c.Security.SessionTimeout < 1 {
		return fmt.Errorf("session timeout must be at least 1 minute")
	}

	if c.Security.MaxLoginAttempts < 1 {
		return fmt.Errorf("max login attempts must be at least 1")
	}

	return nil
}

// SaveToFile saves the configuration to a YAML file
func (c *Config) SaveToFile(path string) error {
	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Marshal to YAML
	data, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("failed to marshal configuration: %w", err)
	}

	// Write to file
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write configuration file: %w", err)
	}

	return nil
}

// GetServerAddress returns the full server address
func (c *Config) GetServerAddress() string {
	return fmt.Sprintf("%s:%d", c.Server.Host, c.Server.Port)
}

// GetDatabaseConnMaxLifetime returns the connection max lifetime as a duration
func (c *Config) GetDatabaseConnMaxLifetime() time.Duration {
	return time.Duration(c.Database.ConnMaxLifetime) * time.Minute
}

// GetSessionTimeout returns the session timeout as a duration
func (c *Config) GetSessionTimeout() time.Duration {
	return time.Duration(c.Security.SessionTimeout) * time.Minute
}

// GetLoginLockoutTime returns the login lockout time as a duration
func (c *Config) GetLoginLockoutTime() time.Duration {
	return time.Duration(c.Security.LoginLockoutTime) * time.Minute
}

// GetCleanupInterval returns the cleanup interval as a duration
func (c *Config) GetCleanupInterval() time.Duration {
	return time.Duration(c.Retention.CleanupInterval) * time.Hour
}

// GetBackupInterval returns the backup interval as a duration
func (c *Config) GetBackupInterval() time.Duration {
	return time.Duration(c.Database.BackupInterval) * time.Hour
}

// generateSessionSecret generates a random session secret
func generateSessionSecret() string {
	// In a real implementation, this would generate a cryptographically secure random string
	// For now, we'll use a simple timestamp-based approach
	return fmt.Sprintf("exim-pilot-session-%d", time.Now().UnixNano())
}
