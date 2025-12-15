package utils

import (
	"testing"
)

func TestGeneratePIN(t *testing.T) {
	pin, err := GeneratePIN()
	if err != nil {
		t.Fatalf("GeneratePIN() failed: %v", err)
	}

	if len(pin) != 6 {
		t.Errorf("Expected PIN length 6, got %d", len(pin))
	}

	// Check if all characters are digits
	for i, char := range pin {
		if char < '0' || char > '9' {
			t.Errorf("PIN character at position %d is not a digit: %c", i, char)
		}
	}

	// Check if PIN is within valid range (100000-999999)
	if pin < "100000" || pin > "999999" {
		t.Errorf("PIN %s is not within valid range (100000-999999)", pin)
	}
}

func TestValidatePIN(t *testing.T) {
	tests := []struct {
		pin      string
		expected bool
	}{
		{"123456", true},
		{"000000", true},
		{"999999", true},
		{"12345", false},   // too short
		{"1234567", false}, // too long
		{"12345a", false},  // contains non-digit
		{"", false},        // empty
		{"abc123", false},  // contains letters
	}

	for _, test := range tests {
		result := ValidatePIN(test.pin)
		if result != test.expected {
			t.Errorf("ValidatePIN(%q) = %v, expected %v", test.pin, result, test.expected)
		}
	}
}
