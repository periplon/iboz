package email

import (
	"crypto/sha256"
	"encoding/hex"
)

// SHA256Hasher hashes secrets using SHA256.
type SHA256Hasher struct{}

// NewSHA256Hasher creates a SHA256-backed hasher implementation.
func NewSHA256Hasher() SecretHasher {
	return SHA256Hasher{}
}

// Hash implements the SecretHasher interface.
func (SHA256Hasher) Hash(secret string) (string, error) {
	sum := sha256.Sum256([]byte(secret))
	return hex.EncodeToString(sum[:]), nil
}
