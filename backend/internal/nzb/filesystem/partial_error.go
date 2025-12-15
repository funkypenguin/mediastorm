package filesystem

import "fmt"

// PartialContentError represents a reader that returned data before encountering a failure.
type PartialContentError struct {
	BytesRead     int64
	TotalExpected int64
	UnderlyingErr error
}

func (e *PartialContentError) Error() string {
	return fmt.Sprintf("partial content: read %d/%d bytes, underlying error: %v", e.BytesRead, e.TotalExpected, e.UnderlyingErr)
}

func (e *PartialContentError) Unwrap() error {
	return e.UnderlyingErr
}

// CorruptedFileError represents a file that could not be read at all.
type CorruptedFileError struct {
	TotalExpected int64
	UnderlyingErr error
}

func (e *CorruptedFileError) Error() string {
	return fmt.Sprintf("corrupted file: no content available from %d expected bytes, underlying error: %v", e.TotalExpected, e.UnderlyingErr)
}

func (e *CorruptedFileError) Unwrap() error {
	return e.UnderlyingErr
}
