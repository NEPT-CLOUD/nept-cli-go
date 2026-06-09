package utils

import (
	"os"

	"github.com/joho/godotenv"
)

// NodeEnv holds the value of the NODE_ENV environment variable.
var NodeEnv string
var BackendUrl string = "https://server.nept.cloud"

func init() {
	// Load environment variables from .env file if it exists.
	// Ignoring error since .env file is optional in production.
	_ = godotenv.Load()
	NodeEnv = os.Getenv("NODE_ENV")
	BackendUrl = "https://server.nept.cloud"
	if NodeEnv == "dev" {
		BackendUrl = "http://localhost:8000"
	}

}

// GetNodeEnv returns the current value of the NODE_ENV environment variable.
// If the environment variable is not set, it returns a default value of "development".
func GetNodeEnv() string {
	// Re-read dynamically to allow runtime/test modifications of env variables.
	if env := os.Getenv("NODE_ENV"); env != "" {
		return env
	}
	return "pro"
}
