package filesystem_test

import (
	"errors"
	"fmt"
	"testing"

	"novastream/internal/nzb/filesystem"
)

func TestPartialContentError(t *testing.T) {
	wrapped := fmt.Errorf("downstream: %w", errors.New("missing segment"))
	err := &filesystem.PartialContentError{BytesRead: 512, TotalExpected: 1024, UnderlyingErr: wrapped}

	if got := err.Error(); got == "" {
		t.Fatalf("Error() returned empty string")
	}

	if !errors.Is(err, wrapped) {
		t.Fatalf("errors.Is did not unwrap to underlying error")
	}
}

func TestCorruptedFileError(t *testing.T) {
	wrapped := errors.New("article not found")
	err := &filesystem.CorruptedFileError{TotalExpected: 2048, UnderlyingErr: wrapped}

	if got := err.Error(); got == "" {
		t.Fatalf("Error() returned empty string")
	}

	if !errors.Is(err, wrapped) {
		t.Fatalf("errors.Is failed to unwrap underlying error")
	}
}
