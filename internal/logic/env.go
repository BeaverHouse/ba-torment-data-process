package logic

import (
	"os"
	"strconv"
)

// Retrieves an environment variable with a default string value
func GetEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// Retrieves an environment variable with a default int value
func GetIntEnv(key string, defaultValue int) int {
	valueStr := GetEnv(key, "")
	if valueStr == "" {
		return defaultValue
	}
	value, err := strconv.Atoi(valueStr)
	if err != nil {
		return defaultValue
	}
	return value
}

// Check whether the current environment is local
func IsLocalEnv() bool {
	return GetEnv("GO_ENV", "local") == "local"
}

// Check whether the current environment is development
func IsDevelopmentEnv() bool {
	return GetEnv("GO_ENV", "local") == "development"
}

// Check whether the current environment is production
func IsProductionEnv() bool {
	return GetEnv("GO_ENV", "local") == "production"
}
