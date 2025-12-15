package filesystem

import "errors"

var (
	ErrInvalidWhence       = errors.New("seek: invalid whence")
	ErrSeekNegative        = errors.New("seek: negative position")
	ErrSeekTooFar          = errors.New("seek: too far")
	ErrCannotRemoveRoot    = errors.New("cannot remove root directory")
	ErrCannotReadDirectory = errors.New("cannot read from directory")
	ErrNotDirectory        = errors.New("not a directory")
	ErrNoCipherConfig      = errors.New("no cipher configured for encryption")
	ErrNoEncryptionParams  = errors.New("no NZB data available for encryption parameters")
	ErrFailedDecryptReader = errors.New("failed to wrap reader with encryption")
	ErrFileIsCorrupted     = errors.New("file is corrupted, there are some missing segments")
)

const RootPath = "/"
