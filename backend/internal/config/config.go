package config

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
)

// Config holds all application configuration
type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	JWT      JWTConfig
	App      AppConfig
}

type ServerConfig struct {
	Port            int
	Host            string
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	ShutdownTimeout time.Duration
}

type DatabaseConfig struct {
	Host            string
	Port            int
	User            string
	Password        string
	DBName          string
	SSLMode         string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}

type JWTConfig struct {
	AccessSecret      string
	RefreshSecret     string
	AccessExpiration  time.Duration
	RefreshExpiration time.Duration
}

type AppConfig struct {
	Environment string // development, production, test
	LogLevel    string // debug, info, warn, error
}

// Load reads configuration from environment variables
func Load() (*Config, error) {
	viper.SetConfigFile(".env")
	viper.AutomaticEnv()

	// Ignore error if .env file is not found, as we may be relying solely on env vars
	_ = viper.ReadInConfig()

	cfg := &Config{
		Server: ServerConfig{
			Port:            viper.GetInt("SERVER_PORT"),
			Host:            viper.GetString("SERVER_HOST"),
			ReadTimeout:     viper.GetDuration("SERVER_READ_TIMEOUT"),
			WriteTimeout:    viper.GetDuration("SERVER_WRITE_TIMEOUT"),
			ShutdownTimeout: viper.GetDuration("SERVER_SHUTDOWN_TIMEOUT"),
		},
		Database: DatabaseConfig{
			Host:            viper.GetString("DB_HOST"),
			Port:            viper.GetInt("DB_PORT"),
			User:            viper.GetString("DB_USER"),
			Password:        viper.GetString("DB_PASSWORD"),
			DBName:          viper.GetString("DB_NAME"),
			SSLMode:         viper.GetString("DB_SSLMODE"),
			MaxOpenConns:    viper.GetInt("DB_MAX_OPEN_CONNS"),
			MaxIdleConns:    viper.GetInt("DB_MAX_IDLE_CONNS"),
			ConnMaxLifetime: viper.GetDuration("DB_CONN_MAX_LIFETIME"),
		},
		JWT: JWTConfig{
			AccessSecret:      viper.GetString("JWT_ACCESS_SECRET"),
			RefreshSecret:     viper.GetString("JWT_REFRESH_SECRET"),
			AccessExpiration:  viper.GetDuration("JWT_ACCESS_EXPIRATION"),
			RefreshExpiration: viper.GetDuration("JWT_REFRESH_EXPIRATION"),
		},
		App: AppConfig{
			Environment: viper.GetString("APP_ENV"),
			LogLevel:    viper.GetString("LOG_LEVEL"),
		},
	}

	// Apply defaults if empty
	applyDefaults(cfg)

	if err := Validate(cfg); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return cfg, nil
}

func applyDefaults(cfg *Config) {
	if cfg.Server.Port == 0 {
		cfg.Server.Port = 8080
	}
	if cfg.Server.Host == "" {
		cfg.Server.Host = "0.0.0.0"
	}
	if cfg.Server.ReadTimeout == 0 {
		cfg.Server.ReadTimeout = 15 * time.Second
	}
	if cfg.Server.WriteTimeout == 0 {
		cfg.Server.WriteTimeout = 15 * time.Second
	}
	if cfg.Server.ShutdownTimeout == 0 {
		cfg.Server.ShutdownTimeout = 30 * time.Second
	}

	if cfg.Database.Host == "" {
		cfg.Database.Host = "localhost"
	}
	if cfg.Database.Port == 0 {
		cfg.Database.Port = 5432
	}
	if cfg.Database.User == "" {
		cfg.Database.User = "postgres"
	}
	if cfg.Database.DBName == "" {
		cfg.Database.DBName = "virallens"
	}
	if cfg.Database.SSLMode == "" {
		cfg.Database.SSLMode = "disable"
	}
	if cfg.Database.MaxOpenConns == 0 {
		cfg.Database.MaxOpenConns = 25
	}
	if cfg.Database.MaxIdleConns == 0 {
		cfg.Database.MaxIdleConns = 5
	}
	if cfg.Database.ConnMaxLifetime == 0 {
		cfg.Database.ConnMaxLifetime = 5 * time.Minute
	}

	if cfg.JWT.AccessExpiration == 0 {
		cfg.JWT.AccessExpiration = 15 * time.Minute
	}
	if cfg.JWT.RefreshExpiration == 0 {
		cfg.JWT.RefreshExpiration = 7 * 24 * time.Hour
	}

	if cfg.App.Environment == "" {
		cfg.App.Environment = "development"
	}
	if cfg.App.LogLevel == "" {
		cfg.App.LogLevel = "info"
	}
}

// ConnectionString returns PostgreSQL connection string
func (c *DatabaseConfig) ConnectionString() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Password, c.DBName, c.SSLMode,
	)
}
