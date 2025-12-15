package filesystem_test

import (
	"errors"
	"testing"

	"novastream/internal/nzb/filesystem"
)

func TestErrorExports(t *testing.T) {
	checks := []error{
		filesystem.ErrInvalidWhence,
		filesystem.ErrSeekNegative,
		filesystem.ErrSeekTooFar,
		filesystem.ErrCannotRemoveRoot,
		filesystem.ErrCannotReadDirectory,
		filesystem.ErrNotDirectory,
		filesystem.ErrNoCipherConfig,
		filesystem.ErrNoEncryptionParams,
		filesystem.ErrFailedDecryptReader,
		filesystem.ErrFileIsCorrupted,
	}

	for i, err := range checks {
		if err == nil {
			t.Fatalf("error constant %d is nil", i)
		}
	}

	// Ensure errors can be matched via errors.Is so callers can wrap them.
	if !errors.Is(filesystem.ErrInvalidWhence, filesystem.ErrInvalidWhence) {
		t.Fatalf("errors.Is failed for ErrInvalidWhence")
	}
}

func TestRootPath(t *testing.T) {
	if filesystem.RootPath != "/" {
		t.Fatalf("RootPath = %q, want /", filesystem.RootPath)
	}
}
