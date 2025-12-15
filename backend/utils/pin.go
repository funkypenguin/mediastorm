package utils

import (
	"crypto/rand"
	"fmt"
	"math/big"
)

// GeneratePIN returns a cryptographically secure 6-digit PIN.
func GeneratePIN() (string, error) {
	// Generate a random number between 100000 and 999999 (6 digits)
	max := big.NewInt(900000) // 999999 - 100000 + 1
	n, err := rand.Int(rand.Reader, max)
	if err != nil {
		return "", fmt.Errorf("generate pin: %w", err)
	}

	// Add 100000 to ensure it's always 6 digits
	pin := n.Int64() + 100000
	return fmt.Sprintf("%06d", pin), nil
}

// ValidatePIN checks if a string is a valid 6-digit PIN.
func ValidatePIN(pin string) bool {
	if len(pin) != 6 {
		return false
	}

	// Check if all characters are digits
	for _, char := range pin {
		if char < '0' || char > '9' {
			return false
		}
	}

	return true
}
