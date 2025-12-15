package importer

import (
	"io"
	"io/fs"
	"os"
	"time"
)

// MemoryFileSystem implements fs.FS for reading RAR archives from pre-loaded memory
// This allows rarlist to perform fast sequential analysis on in-memory data
type MemoryFileSystem struct {
	files map[string][]byte // filename -> file content
}

// MemoryFile implements fs.File for reading from memory
type MemoryFile struct {
	name    string
	content []byte
	pos     int64
	closed  bool
}

// MemoryFileInfo implements fs.FileInfo for memory files
type MemoryFileInfo struct {
	name string
	size int64
}

// NewMemoryFileSystem creates a new memory-based filesystem
func NewMemoryFileSystem(files map[string][]byte) *MemoryFileSystem {
	return &MemoryFileSystem{
		files: files,
	}
}

// Open opens a file in the memory filesystem
func (mfs *MemoryFileSystem) Open(name string) (fs.File, error) {
	if content, exists := mfs.files[name]; exists {
		return &MemoryFile{
			name:    name,
			content: content,
			pos:     0,
			closed:  false,
		}, nil
	}

	return nil, &fs.PathError{
		Op:   "open",
		Path: name,
		Err:  fs.ErrNotExist,
	}
}

// Stat returns file information for a file in the memory filesystem
func (mfs *MemoryFileSystem) Stat(path string) (os.FileInfo, error) {
	if content, exists := mfs.files[path]; exists {
		return &MemoryFileInfo{
			name: path,
			size: int64(len(content)),
		}, nil
	}

	return nil, &fs.PathError{
		Op:   "stat",
		Path: path,
		Err:  fs.ErrNotExist,
	}
}

// MemoryFile methods implementing fs.File interface

func (mf *MemoryFile) Stat() (fs.FileInfo, error) {
	return &MemoryFileInfo{
		name: mf.name,
		size: int64(len(mf.content)),
	}, nil
}

func (mf *MemoryFile) Read(p []byte) (n int, err error) {
	if mf.closed {
		return 0, fs.ErrClosed
	}

	if mf.pos >= int64(len(mf.content)) {
		return 0, io.EOF
	}

	n = copy(p, mf.content[mf.pos:])
	mf.pos += int64(n)

	if mf.pos >= int64(len(mf.content)) {
		err = io.EOF
	}

	return n, err
}

func (mf *MemoryFile) Close() error {
	if mf.closed {
		return nil
	}
	mf.closed = true
	return nil
}

// Seek implements io.Seeker interface for efficient RAR part access
func (mf *MemoryFile) Seek(offset int64, whence int) (int64, error) {
	if mf.closed {
		return 0, fs.ErrClosed
	}

	var abs int64
	switch whence {
	case io.SeekStart:
		abs = offset
	case io.SeekCurrent:
		abs = mf.pos + offset
	case io.SeekEnd:
		abs = int64(len(mf.content)) + offset
	default:
		return 0, fs.ErrInvalid
	}

	if abs < 0 {
		return 0, fs.ErrInvalid
	}

	if abs > int64(len(mf.content)) {
		abs = int64(len(mf.content))
	}

	mf.pos = abs
	return abs, nil
}

// MemoryFileInfo methods implementing fs.FileInfo interface

func (mfi *MemoryFileInfo) Name() string       { return mfi.name }
func (mfi *MemoryFileInfo) Size() int64        { return mfi.size }
func (mfi *MemoryFileInfo) Mode() fs.FileMode  { return 0644 }
func (mfi *MemoryFileInfo) ModTime() time.Time { return time.Now() }
func (mfi *MemoryFileInfo) IsDir() bool        { return false }
func (mfi *MemoryFileInfo) Sys() interface{}   { return nil }
