package hash

import (
	"crypto/sha256"
	"encoding/hex"
)

// SHA256 returns a hex-encoded SHA256 hash of the input string.
// Used for LLM prompt cache keys and idempotency key generation.
func SHA256(s string) string {
	h := sha256.Sum256([]byte(s))
	return hex.EncodeToString(h[:])
}

// SHA256Multi hashes multiple strings concatenated.
func SHA256Multi(parts ...string) string {
	h := sha256.New()
	for _, p := range parts {
		h.Write([]byte(p))
	}
	return hex.EncodeToString(h.Sum(nil))
}
