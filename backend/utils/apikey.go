package utils

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
)

// GenerateAPIKey returns a URL-safe random API key with 256 bits of entropy.
func GenerateAPIKey() (string, error) {
	const numBytes = 32 // 256 bits
	buf := make([]byte, numBytes)
	if _, err := rand.Read(buf); err != nil {
		return "", fmt.Errorf("generate api key: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}
