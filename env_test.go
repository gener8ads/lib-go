package lib

import (
	"os"
	"testing"
)

func TestGetenvFallback(t *testing.T) {
	fallback := "BAR"
	env := Getenv("FOO", fallback)

	if env != fallback {
		t.Errorf("Fallback env incorrect, expected '%s', got '%s'", fallback, env)
	}
}

func TestGetEnv(t *testing.T) {
	key := "FOO"
	value := "BAZ"

	os.Setenv(key, value)

	env := Getenv(key, "")

	if env != value {
		t.Errorf("Env incorrect, expected '%s', got '%s'", value, env)
	}
}
