package env

import "os"

// Get returns an env var or fallback string
func Get(key string, fallback string) string {
	value := os.Getenv(key)

	if len(value) == 0 {
		return fallback
	}

	return value
}
