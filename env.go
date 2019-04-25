package lib

import "os"

// Getenv returns an env var or fallback string
func Getenv(key string, fallback string) string {
	value := os.Getenv(key)

	if len(value) == 0 {
		return fallback
	}

	return value
}
