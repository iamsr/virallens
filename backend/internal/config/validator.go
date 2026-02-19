package config

import "errors"

// Validate checks if configuration is valid
func Validate(cfg *Config) error {
	if err := validateServer(&cfg.Server); err != nil {
		return err
	}
	if err := validateDatabase(&cfg.Database); err != nil {
		return err
	}
	if err := validateJWT(&cfg.JWT); err != nil {
		return err
	}
	if err := validateApp(&cfg.App); err != nil {
		return err
	}
	return nil
}

func validateServer(cfg *ServerConfig) error {
	if cfg.Port < 1 || cfg.Port > 65535 {
		return errors.New("server port must be between 1 and 65535")
	}
	if cfg.Host == "" {
		return errors.New("server host cannot be empty")
	}
	if cfg.ReadTimeout <= 0 {
		return errors.New("server read timeout must be positive")
	}
	if cfg.WriteTimeout <= 0 {
		return errors.New("server write timeout must be positive")
	}
	return nil
}

func validateDatabase(cfg *DatabaseConfig) error {
	if cfg.Host == "" {
		return errors.New("database host cannot be empty")
	}
	if cfg.Port < 1 || cfg.Port > 65535 {
		return errors.New("database port must be between 1 and 65535")
	}
	if cfg.User == "" {
		return errors.New("database user cannot be empty")
	}
	if cfg.DBName == "" {
		return errors.New("database name cannot be empty")
	}
	if cfg.MaxOpenConns < 1 {
		return errors.New("database max open connections must be at least 1")
	}
	if cfg.MaxIdleConns < 0 {
		return errors.New("database max idle connections cannot be negative")
	}
	return nil
}

func validateJWT(cfg *JWTConfig) error {
	if cfg.AccessSecret == "" {
		return errors.New("JWT access secret cannot be empty")
	}
	if cfg.RefreshSecret == "" {
		return errors.New("JWT refresh secret cannot be empty")
	}
	if cfg.AccessExpiration <= 0 {
		return errors.New("JWT access expiration must be positive")
	}
	if cfg.RefreshExpiration <= 0 {
		return errors.New("JWT refresh expiration must be positive")
	}
	return nil
}

func validateApp(cfg *AppConfig) error {
	validEnvs := map[string]bool{
		"development": true,
		"production":  true,
		"test":        true,
	}
	if !validEnvs[cfg.Environment] {
		return errors.New("app environment must be development, production, or test")
	}

	validLogLevels := map[string]bool{
		"debug": true,
		"info":  true,
		"warn":  true,
		"error": true,
	}
	if !validLogLevels[cfg.LogLevel] {
		return errors.New("log level must be debug, info, warn, or error")
	}
	return nil
}
