package utils

import (
	"os"
)

// GetEnvOrDefault retrieves an environment variable's value or returns a default value if not set
func GetEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
