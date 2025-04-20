package common

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
)

// dir returns the absolute path of the given environment file (envFile) in the Go module's
// root directory. It searches for the 'go.mod' file from the current working directory upwards
// and appends the envFile to the directory containing 'go.mod'.
// It panics if it fails to find the 'go.mod' file.
// https://github.com/joho/godotenv/issues/126#issuecomment-1474645022
func dir(envFile string) string {
	currentDir, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	for {
		goModPath := filepath.Join(currentDir, "go.mod")
		if _, err := os.Stat(goModPath); err == nil {
			break
		}

		parent := filepath.Dir(currentDir)
		if parent == currentDir {
			panic(fmt.Errorf("go.mod not found"))
		}
		currentDir = parent
	}

	return filepath.Join(currentDir, envFile)
}

// LoadEnv loads the .env file.
// If GO_ENV value is "local", it's triggered.
// It'll panic if it fails to load .env file.
func LoadEnv() {
	env := os.Getenv("GO_ENV")
	if env == "" {
		env = "local"
		os.Setenv("GO_ENV", env)
	}

	if env == "local" {
		err := godotenv.Load(dir(".env"))
		if err != nil {
			panic(err)
		}
	}
}

// GetEnv returns the value of the environment variable.
// If the environment variable is not set, it returns the default value.
func GetEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// GetEssentialEnv returns the value of the environment variable.
// It'll panic if the environment variable is not set.
func GetEssentialEnv(key string) string {
	value := os.Getenv(key)
	if value == "" {
		panic(fmt.Errorf("required environment variable is not set: %s", key))
	}
	return value
}
