package env

import (
	"os"
	"testing"
)

func TestGetFallback(t *testing.T) {
	fallback := "BAR"
	env := Get("FOO", fallback)

	if env != fallback {
		t.Errorf("Fallback env incorrect, expected '%s', got '%s'", fallback, env)
	}
}

func TestGet(t *testing.T) {
	key := "FOO"
	value := "BAZ"

	os.Setenv(key, value)

	env := Get(key, "")

	if env != value {
		t.Errorf("Env incorrect, expected '%s', got '%s'", value, env)
	}
}
