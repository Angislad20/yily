package config

import (
	"os"
)

// Config represents the configuration for the application.
type config struct {
	MasterKey string
}

// NewConfig creates a new Config instance by reading the necessary environment variables.
func NewConfig() *config {
	return &config{
		MasterKey: getEnv("YILY_MASTER_KEY", ""),
	}
}

// getEnv retrieves the value of the environment variable named by the key.
// If the variable is not present in the environment, it returns the defaultValue.
func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists && value != "" {
		return value
	}
	return defaultValue
}
