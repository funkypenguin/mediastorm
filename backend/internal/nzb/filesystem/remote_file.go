package filesystem

import (
	"context"
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/spf13/afero"

	"novastream/internal/nzb/metadata"
	metapb "novastream/internal/nzb/metadata/proto"
	"novastream/internal/nzb/utils"
)

// RemoteFileConfig wires dependencies required to expose metadata-backed files.
type RemoteFileConfig struct {
	MetadataService *metadata.MetadataService
	HealthReporter  HealthReporter
	ReaderFactory   ReaderFactory
	Encryption      EncryptionAdapter
	StreamingChunk  int64
	Password        string
	Salt            string
}

const defaultStreamingChunkSize int64 = 32 * 1024 * 1024 // 32 MiB

// MetadataRemoteFile resolves virtual paths into afero.File handles.
type MetadataRemoteFile struct {
	cfg RemoteFileConfig
}

// FileDescriptor captures the identifying metadata for a virtual file.
type FileDescriptor struct {
	VirtualPath    string
	NormalizedPath string
	SourceNZB      string
	FileSize       int64
	SegmentCount   int
	CreatedAt      time.Time
}

// NewMetadataRemoteFile constructs a remote file helper with sane defaults.
func NewMetadataRemoteFile(cfg RemoteFileConfig) *MetadataRemoteFile {
	if cfg.StreamingChunk <= 0 {
		cfg.StreamingChunk = defaultStreamingChunkSize
	}
	return &MetadataRemoteFile{cfg: cfg}
}

// OpenFile returns an afero.File for the supplied path when it exists in metadata.
func (mrf *MetadataRemoteFile) OpenFile(ctx context.Context, name string, args utils.PathWithArgs) (bool, afero.File, error) {
	if mrf.cfg.MetadataService == nil {
		return false, nil, fmt.Errorf("metadata service not configured")
	}
	if mrf.cfg.ReaderFactory == nil {
		return false, nil, fmt.Errorf("reader factory not configured")
	}

	normalized := normalizePath(name)

	if mrf.cfg.MetadataService.DirectoryExists(normalized) {
		dir := &MetadataVirtualDirectory{
			name:            name,
			normalizedPath:  normalized,
			metadataService: mrf.cfg.MetadataService,
		}
		return true, dir, nil
	}

	if !mrf.cfg.MetadataService.FileExists(normalized) {
		if mrf.isValidEmptyDirectory(normalized) {
			dir := &MetadataVirtualDirectory{
				name:            name,
				normalizedPath:  normalized,
				metadataService: mrf.cfg.MetadataService,
			}
			return true, dir, nil
		}
		return false, nil, nil
	}

	meta, err := mrf.cfg.MetadataService.ReadFileMetadata(normalized)
	if err != nil {
		return false, nil, fmt.Errorf("read metadata: %w", err)
	}
	if meta == nil {
		return false, nil, nil
	}
	if meta.Status == metapb.FileStatus_FILE_STATUS_CORRUPTED {
		return false, nil, ErrFileIsCorrupted
	}

	vf := &MetadataVirtualFile{
		name:            name,
		fileMeta:        meta,
		metadataService: mrf.cfg.MetadataService,
		healthReporter:  mrf.cfg.HealthReporter,
		args:            args,
		ctx:             ctx,
		readerFactory:   mrf.cfg.ReaderFactory,
		encryption:      mrf.cfg.Encryption,
		password:        mrf.cfg.Password,
		salt:            mrf.cfg.Salt,
		streamingChunk:  mrf.cfg.StreamingChunk,
		readerCache:     make(map[string]io.ReadCloser),
	}

	return true, vf, nil
}

// DescribeFile returns a lightweight descriptor for the requested virtual file without opening a stream.
func (mrf *MetadataRemoteFile) DescribeFile(name string) (*FileDescriptor, error) {
	if mrf.cfg.MetadataService == nil {
		return nil, fmt.Errorf("metadata service not configured")
	}

	normalized := normalizePath(name)
	if !mrf.cfg.MetadataService.FileExists(normalized) {
		return nil, fs.ErrNotExist
	}

	meta, err := mrf.cfg.MetadataService.ReadFileMetadata(normalized)
	if err != nil {
		return nil, fmt.Errorf("read metadata: %w", err)
	}
	if meta == nil {
		return nil, fs.ErrNotExist
	}

	desc := FileDescriptor{
		VirtualPath:    name,
		NormalizedPath: normalized,
		FileSize:       meta.GetFileSize(),
		SourceNZB:      strings.TrimSpace(meta.GetSourceNzbPath()),
		SegmentCount:   len(meta.GetSegmentData()),
	}
	if created := meta.GetCreatedAt(); created > 0 {
		desc.CreatedAt = time.Unix(created, 0)
	}

	return &desc, nil
}

func (mrf *MetadataRemoteFile) isValidEmptyDirectory(normalized string) bool {
	if normalized == RootPath {
		return true
	}

	parent := filepath.Dir(normalized)
	if parent == "." {
		parent = RootPath
	}

	if mrf.cfg.MetadataService.DirectoryExists(parent) {
		return true
	}

	if parent == normalized {
		return false
	}

	return mrf.isValidEmptyDirectory(parent)
}

// MetadataVirtualDirectory adapts metadata directories to afero.File.
type MetadataVirtualDirectory struct {
	name            string
	normalizedPath  string
	metadataService *metadata.MetadataService
}

func (mvd *MetadataVirtualDirectory) Close() error             { return nil }
func (mvd *MetadataVirtualDirectory) Name() string             { return mvd.name }
func (mvd *MetadataVirtualDirectory) Read([]byte) (int, error) { return 0, ErrCannotReadDirectory }
func (mvd *MetadataVirtualDirectory) ReadAt([]byte, int64) (int, error) {
	return 0, ErrCannotReadDirectory
}
func (mvd *MetadataVirtualDirectory) Seek(int64, int) (int64, error) {
	return 0, ErrCannotReadDirectory
}
func (mvd *MetadataVirtualDirectory) Write([]byte) (int, error)          { return 0, fs.ErrPermission }
func (mvd *MetadataVirtualDirectory) WriteAt([]byte, int64) (int, error) { return 0, fs.ErrPermission }
func (mvd *MetadataVirtualDirectory) WriteString(string) (int, error)    { return 0, fs.ErrPermission }
func (mvd *MetadataVirtualDirectory) Sync() error                        { return nil }
func (mvd *MetadataVirtualDirectory) Truncate(int64) error               { return fs.ErrPermission }

func (mvd *MetadataVirtualDirectory) Readdir(count int) ([]fs.FileInfo, error) {
	reader := metadata.NewMetadataReader(mvd.metadataService)
	dirs, _, err := reader.ListDirectoryContents(mvd.normalizedPath)
	if err != nil {
		return nil, err
	}

	var infos []fs.FileInfo
	for _, dir := range dirs {
		infos = append(infos, dir)
		if count > 0 && len(infos) >= count {
			return infos, nil
		}
	}

	files, err := mvd.metadataService.ListDirectory(mvd.normalizedPath)
	if err != nil {
		return nil, err
	}

	for _, file := range files {
		meta, err := mvd.metadataService.ReadFileMetadata(filepath.Join(mvd.normalizedPath, file))
		if err != nil || meta == nil {
			continue
		}
		info := &MetadataFileInfo{
			name:    file,
			size:    meta.FileSize,
			mode:    0o644,
			modTime: time.Unix(meta.ModifiedAt, 0),
			isDir:   false,
		}
		infos = append(infos, info)
		if count > 0 && len(infos) >= count {
			return infos, nil
		}
	}

	return infos, nil
}

func (mvd *MetadataVirtualDirectory) Readdirnames(n int) ([]string, error) {
	infos, err := mvd.Readdir(n)
	if err != nil {
		return nil, err
	}
	names := make([]string, len(infos))
	for i, info := range infos {
		names[i] = info.Name()
	}
	return names, nil
}

func (mvd *MetadataVirtualDirectory) Stat() (fs.FileInfo, error) {
	return &MetadataFileInfo{
		name:    filepath.Base(mvd.normalizedPath),
		size:    0,
		mode:    fs.ModeDir | 0o755,
		modTime: time.Now(),
		isDir:   true,
	}, nil
}

// MetadataFileInfo implements fs.FileInfo for metadata entries.
type MetadataFileInfo struct {
	name    string
	size    int64
	mode    fs.FileMode
	modTime time.Time
	isDir   bool
}

func (mfi *MetadataFileInfo) Name() string       { return mfi.name }
func (mfi *MetadataFileInfo) Size() int64        { return mfi.size }
func (mfi *MetadataFileInfo) Mode() fs.FileMode  { return mfi.mode }
func (mfi *MetadataFileInfo) ModTime() time.Time { return mfi.modTime }
func (mfi *MetadataFileInfo) IsDir() bool        { return mfi.isDir }
func (mfi *MetadataFileInfo) Sys() interface{}   { return nil }

// MetadataVirtualFile is an afero.File backed by NZB metadata.
type MetadataVirtualFile struct {
	name            string
	fileMeta        *metapb.FileMetadata
	metadataService *metadata.MetadataService
	healthReporter  HealthReporter
	args            utils.PathWithArgs
	ctx             context.Context

	readerFactory ReaderFactory
	encryption    EncryptionAdapter

	password         string
	salt             string
	streamingChunk   int64
	originalRangeEnd int64

	mu sync.Mutex

	reader           io.ReadCloser
	closed           bool
	rangeInitialized bool
	startOffset      int64
	endOffset        int64
	currentOffset    int64

	// Reader cache to avoid recreating readers for overlapping ranges
	readerCache map[string]io.ReadCloser
}

type limitedReadCloser struct {
	reader io.Reader
	closer io.Closer
}

func newLimitedReadCloser(base io.ReadCloser, limit int64) io.ReadCloser {
	if limit <= 0 {
		return base
	}
	return &limitedReadCloser{
		reader: io.LimitReader(base, limit),
		closer: base,
	}
}

func (l *limitedReadCloser) Read(p []byte) (int, error) {
	return l.reader.Read(p)
}

func (l *limitedReadCloser) Close() error {
	return l.closer.Close()
}

func (mvf *MetadataVirtualFile) Name() string { return mvf.name }

// Descriptor exposes a snapshot of the underlying metadata for logging or diagnostics.
func (mvf *MetadataVirtualFile) Descriptor() FileDescriptor {
	desc := FileDescriptor{
		VirtualPath:    mvf.name,
		NormalizedPath: normalizePath(mvf.name),
	}
	if mvf.fileMeta != nil {
		desc.FileSize = mvf.fileMeta.GetFileSize()
		desc.SourceNZB = strings.TrimSpace(mvf.fileMeta.GetSourceNzbPath())
		desc.SegmentCount = len(mvf.fileMeta.GetSegmentData())
		if created := mvf.fileMeta.GetCreatedAt(); created > 0 {
			desc.CreatedAt = time.Unix(created, 0)
		}
	}
	return desc
}

func (mvf *MetadataVirtualFile) Read(p []byte) (int, error) {
	mvf.mu.Lock()
	defer mvf.mu.Unlock()

	if mvf.closed {
		return 0, io.EOF
	}

	mvf.initRangeLocked()

	if mvf.fileMeta.FileSize == 0 || (mvf.endOffset >= 0 && mvf.currentOffset > mvf.endOffset) {
		return 0, io.EOF
	}

	total := 0
	for total < len(p) {
		if err := mvf.ensureReader(); err != nil {
			if err == io.EOF {
				if total == 0 {
					return 0, io.EOF
				}
				return total, io.EOF
			}
			if total > 0 {
				return total, err
			}
			return 0, err
		}

		maxReadable := len(p) - total
		if maxReadable == 0 {
			break
		}

		if mvf.endOffset >= 0 {
			rangeRemaining := mvf.endOffset - mvf.currentOffset + 1
			if rangeRemaining <= 0 {
				if total == 0 {
					return 0, io.EOF
				}
				return total, io.EOF
			}
			if rangeRemaining < int64(maxReadable) {
				maxReadable = int(rangeRemaining)
			}
		}

		if maxReadable <= 0 {
			break
		}

		n, readErr := mvf.reader.Read(p[total : total+maxReadable])
		if n > 0 {
			mvf.currentOffset += int64(n)
			total += n
		}

		if readErr != nil {
			if readErr == io.EOF {
				mvf.releaseCurrentReaderLocked()

				if mvf.currentOffset >= mvf.fileMeta.FileSize || (mvf.endOffset >= 0 && mvf.currentOffset > mvf.endOffset) {
					if total == 0 {
						return 0, io.EOF
					}
					return total, io.EOF
				}

				if total == len(p) {
					return total, nil
				}
				continue
			}

			mvf.reportError(readErr, int64(n))
			return total, readErr
		}

		if mvf.endOffset >= 0 && mvf.currentOffset > mvf.endOffset {
			mvf.releaseCurrentReaderLocked()

			if total == 0 {
				return 0, io.EOF
			}
			return total, io.EOF
		}
		if mvf.currentOffset >= mvf.fileMeta.FileSize {
			mvf.releaseCurrentReaderLocked()

			if total == 0 {
				return 0, io.EOF
			}
			return total, io.EOF
		}
	}

	if total == 0 {
		return 0, nil
	}

	return total, nil
}

func (mvf *MetadataVirtualFile) Close() error {
	mvf.mu.Lock()
	defer mvf.mu.Unlock()

	mvf.closed = true
	mvf.releaseCurrentReaderLocked()

	// Close all cached readers
	for key, reader := range mvf.readerCache {
		if reader != nil {
			_ = reader.Close()
		}
		delete(mvf.readerCache, key)
	}

	return nil
}

func (mvf *MetadataVirtualFile) Stat() (fs.FileInfo, error) {
	mvf.mu.Lock()
	defer mvf.mu.Unlock()

	mvf.initRangeLocked()

	size := int64(0)
	if mvf.fileMeta.FileSize > 0 {
		start := mvf.startOffset
		if start < 0 {
			start = 0
		}
		end := mvf.endOffset
		if end < 0 || end >= mvf.fileMeta.FileSize {
			end = mvf.fileMeta.FileSize - 1
		}
		if end >= start {
			size = end - start + 1
		}
	}

	return &MetadataFileInfo{
		name:    filepath.Base(mvf.name),
		size:    size,
		mode:    0o644,
		modTime: time.Unix(mvf.fileMeta.ModifiedAt, 0),
		isDir:   false,
	}, nil
}

func (mvf *MetadataVirtualFile) Seek(offset int64, whence int) (int64, error) {
	return 0, ErrInvalidWhence
}

func (mvf *MetadataVirtualFile) ReadAt([]byte, int64) (int, error)  { return 0, fs.ErrPermission }
func (mvf *MetadataVirtualFile) Readdir(int) ([]fs.FileInfo, error) { return nil, ErrNotDirectory }
func (mvf *MetadataVirtualFile) Readdirnames(int) ([]string, error) { return nil, ErrNotDirectory }
func (mvf *MetadataVirtualFile) Write([]byte) (int, error)          { return 0, fs.ErrPermission }
func (mvf *MetadataVirtualFile) WriteAt([]byte, int64) (int, error) { return 0, fs.ErrPermission }
func (mvf *MetadataVirtualFile) WriteString(string) (int, error)    { return 0, fs.ErrPermission }
func (mvf *MetadataVirtualFile) Sync() error                        { return nil }
func (mvf *MetadataVirtualFile) Truncate(int64) error               { return fs.ErrPermission }

func (mvf *MetadataVirtualFile) ensureReader() error {
	if mvf.reader != nil {
		return nil
	}

	mvf.initRangeLocked()

	if mvf.fileMeta.FileSize == 0 {
		return io.EOF
	}

	if mvf.currentOffset >= mvf.fileMeta.FileSize {
		return io.EOF
	}
	if mvf.endOffset >= 0 && mvf.currentOffset > mvf.endOffset {
		return io.EOF
	}

	start := mvf.currentOffset
	if start < 0 {
		start = 0
	}

	maxEnd := mvf.fileMeta.FileSize - 1
	end := mvf.endOffset
	if end < 0 || end > maxEnd {
		end = maxEnd
	}

	if end < start {
		end = start
	}

	if mvf.streamingChunk > 0 && mvf.originalRangeEnd == -1 {
		chunkEnd := start + mvf.streamingChunk - 1
		if chunkEnd < start {
			chunkEnd = start
		}
		if chunkEnd < end {
			end = chunkEnd
		}
	}

	// Create cache key for this range
	cacheKey := fmt.Sprintf("%d-%d", start, end)

	// Check cache first
	if cachedReader, exists := mvf.readerCache[cacheKey]; exists {
		// Cache hit - use existing reader (no logging needed for normal operation)
		mvf.reader = cachedReader
		return nil
	}

	// Cache miss - create new reader (only log in debug mode if needed)
	// fmt.Printf("[filesystem] creating new reader for range %s (cache has %d entries)\n", cacheKey, len(mvf.readerCache))

	// Create new reader
	reader, err := mvf.readerFactory.NewReader(mvf.ctx, mvf.fileMeta, start, end)
	if err != nil {
		return err
	}

	if mvf.fileMeta.Encryption != metapb.Encryption_NONE {
		if mvf.encryption == nil {
			_ = reader.Close()
			return ErrNoCipherConfig
		}
		wrapped, err := mvf.encryption.Wrap(mvf.ctx, start, end, ReaderFactoryFunc(func(ctx context.Context, _ *metapb.FileMetadata, s, e int64) (io.ReadCloser, error) {
			return reader, nil
		}))
		if err != nil {
			_ = reader.Close()
			return fmt.Errorf("wrap reader: %w", err)
		}
		reader = wrapped
	}

	rangeLength := end - start + 1
	if rangeLength > 0 {
		reader = newLimitedReadCloser(reader, rangeLength)
	}

	slog.Default().Info("[filesystem] SEEK: issuing range reader",
		"path", mvf.name,
		"range_start", start,
		"range_end", end,
		"range_length", rangeLength,
		"current_offset", mvf.currentOffset,
	)

	// Cache the reader
	mvf.readerCache[cacheKey] = reader
	mvf.reader = reader
	return nil
}

func (mvf *MetadataVirtualFile) releaseCurrentReaderLocked() {
	if mvf.reader == nil {
		return
	}
	current := mvf.reader
	_ = current.Close()
	mvf.reader = nil
	for cacheKey, cachedReader := range mvf.readerCache {
		if cachedReader == current {
			delete(mvf.readerCache, cacheKey)
			break
		}
	}
}

func (mvf *MetadataVirtualFile) initRangeLocked() {
	if mvf.rangeInitialized {
		return
	}

	mvf.rangeInitialized = true
	size := mvf.fileMeta.FileSize
	if size < 0 {
		size = 0
	}

	mvf.startOffset = 0
	if size > 0 {
		mvf.endOffset = size - 1
	} else {
		mvf.endOffset = -1
	}
	mvf.originalRangeEnd = -1

	if rh, err := mvf.args.Range(); err == nil && rh != nil {
		start := rh.Start
		end := rh.End
		originalEnd := end

		slog.Default().Info("[filesystem] SEEK: initializing range",
			"raw_start", start,
			"raw_end", end,
			"file_size", size,
			"path", mvf.name,
		)

		if start < 0 && end >= 0 {
			length := end
			if length > size {
				length = size
			}
			start = size - length
			if start < 0 {
				start = 0
			}
			end = size - 1
			slog.Default().Info("[filesystem] SEEK: adjusted suffix range",
				"adjusted_start", start,
				"adjusted_end", end,
			)
		}

		if start >= 0 {
			if size > 0 && start >= size {
				start = size - 1
			}
			mvf.startOffset = start
		}

		if end >= 0 {
			if size > 0 && end >= size {
				end = size - 1
			}
			mvf.endOffset = end
		} else {
			mvf.endOffset = size - 1
		}

		if originalEnd >= 0 {
			mvf.originalRangeEnd = originalEnd
		}

		if size == 0 {
			mvf.startOffset = 0
			mvf.endOffset = -1
		}

		if mvf.endOffset < mvf.startOffset {
			mvf.endOffset = mvf.startOffset
		}

		rangeSize := mvf.endOffset - mvf.startOffset + 1
		slog.Default().Info("[filesystem] SEEK: range initialized",
			"start_offset", mvf.startOffset,
			"end_offset", mvf.endOffset,
			"range_size", rangeSize,
			"file_size", size,
			"path", mvf.name,
		)
	}

	if mvf.startOffset < 0 {
		mvf.startOffset = 0
	}
	if size > 0 && mvf.endOffset >= size {
		mvf.endOffset = size - 1
	}

	mvf.currentOffset = mvf.startOffset
}

func (mvf *MetadataVirtualFile) RangeInfo() (int64, int64, int64) {
	mvf.mu.Lock()
	defer mvf.mu.Unlock()

	mvf.initRangeLocked()

	total := mvf.fileMeta.FileSize
	start := mvf.startOffset
	end := mvf.endOffset
	if start < 0 {
		start = 0
	}
	if total <= 0 {
		return start, -1, total
	}
	if end < 0 || end >= total {
		end = total - 1
	}
	if end < start {
		end = start - 1
	}
	return start, end, total
}

func (mvf *MetadataVirtualFile) reportError(err error, bytesRead int64) {
	if mvf.healthReporter == nil {
		return
	}
	if bytesRead > 0 {
		mvf.healthReporter.MarkPartial(mvf.name, err)
	} else {
		mvf.healthReporter.MarkCorrupted(mvf.name, err)
	}
}

func normalizePath(p string) string {
	cleaned := filepath.Clean(p)
	cleaned = strings.TrimPrefix(cleaned, "/")
	if cleaned == "." {
		return RootPath
	}
	return cleaned
}
