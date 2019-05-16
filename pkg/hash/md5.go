package hash

import (
	"crypto/md5"
	"encoding/hex"
)

// MD5 creates an MD5 encoded string
func MD5(str string) string {
	hash := md5.Sum([]byte(str))

	return hex.EncodeToString(hash[:])
}
