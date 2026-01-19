package config

import (
	"os"
	"strconv"
)

// Config holds application configuration
type Config struct {
	Port        int
	Environment string
	DatabaseURL string
	Debug       bool
}

// Load reads configuration from environment variables
func Load() *Config {
	port, _ := strconv.Atoi(getEnv("PORT", "8080"))

	return &Config{
		Port:        port,
		Environment: getEnv("GO_ENV", "development"),
		DatabaseURL: getEnv("DATABASE_URL", "postgres://localhost:5432/myapp?sslmode=disable"),
		Debug:       getEnv("DEBUG", "false") == "true",
	}
}

// getEnv gets an environment variable with a fallback default
func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

// IsDevelopment returns true if running in development mode
func (c *Config) IsDevelopment() bool {
	return c.Environment == "development"
}

// IsProduction returns true if running in production mode
func (c *Config) IsProduction() bool {
	return c.Environment == "production"
}
