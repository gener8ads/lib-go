package hash

import (
	"crypto/sha256"
	"fmt"
)

// Sha256 of the provided string
func Sha256(word string) string {
	code := sha256.New()
	code.Write([]byte(word))
	return fmt.Sprintf("%x", code.Sum(nil))
}
