package filesystem

import (
	"context"
	"io"
	"sync"

	metapb "novastream/internal/nzb/metadata/proto"
)

// SegmentFetcher streams a specific byte range from Usenet or cache.
type SegmentFetcher interface {
	Fetch(ctx context.Context, start, end int64) (io.ReadCloser, error)
}

// HealthReporter records corruption discovered during streaming.
type HealthReporter interface {
	MarkPartial(path string, err error)
	MarkCorrupted(path string, err error)
}

// RangePlanner decides the next byte range to request given a desired range and current reader state.
type RangePlanner interface {
	PlanRange(start, end int64) (int64, int64)
}

// ReaderFactory produces underlying range readers with optional encryption.
type ReaderFactory interface {
	NewReader(ctx context.Context, meta *metapb.FileMetadata, start, end int64) (io.ReadCloser, error)
}

// ReaderFactoryFunc adapts a function into a ReaderFactory implementation.
type ReaderFactoryFunc func(ctx context.Context, meta *metapb.FileMetadata, start, end int64) (io.ReadCloser, error)

func (f ReaderFactoryFunc) NewReader(ctx context.Context, meta *metapb.FileMetadata, start, end int64) (io.ReadCloser, error) {
	return f(ctx, meta, start, end)
}

// EncryptionAdapter wraps readers when metadata indicates encryption.
type EncryptionAdapter interface {
	Wrap(ctx context.Context, start, end int64, factory ReaderFactory) (io.ReadCloser, error)
}

// RangeTracker tracks the active range currently buffered in a MetadataVirtualFile.
type RangeTracker struct {
	mu sync.Mutex

	currentStart int64
	currentEnd   int64
}

func (rt *RangeTracker) Reset(start, end int64) {
	rt.mu.Lock()
	defer rt.mu.Unlock()
	rt.currentStart = start
	rt.currentEnd = end
}

func (rt *RangeTracker) Contains(offset int64) bool {
	rt.mu.Lock()
	defer rt.mu.Unlock()
	return offset >= rt.currentStart && (rt.currentEnd < 0 || offset <= rt.currentEnd)
}
