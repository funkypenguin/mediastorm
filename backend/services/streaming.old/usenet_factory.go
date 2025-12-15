package streaming

import (
	"context"
	"fmt"
	"io"
	"log"
	"strings"
	"time"

	"novastream/internal/nzb/filesystem"
	metapb "novastream/internal/nzb/metadata/proto"
	"novastream/internal/pool"
	"novastream/internal/usenet"
)

// UsenetReaderFactory creates range readers backed by NNTP providers.
type UsenetReaderFactory struct {
	pool       pool.Manager
	maxWorkers int
}

func NewUsenetReaderFactory(manager pool.Manager, maxWorkers int) *UsenetReaderFactory {
	if maxWorkers <= 0 {
		maxWorkers = 15
	}
	return &UsenetReaderFactory{pool: manager, maxWorkers: maxWorkers}
}

func (f *UsenetReaderFactory) NewReader(ctx context.Context, meta *metapb.FileMetadata, start, end int64) (io.ReadCloser, error) {
	if f.pool == nil {
		return nil, fmt.Errorf("usenet pool not configured")
	}
	cp, err := f.pool.GetPool()
	if err != nil {
		return nil, err
	}

	// Prebuffering disabled - it was causing "wrote more than declared Content-Length" errors
	// when clients made small range requests from byte 0. The prebuffering would expand the range
	// after HTTP headers were already sent with the original Content-Length.

	// CRITICAL FIX: Use limited segment range to prevent memory explosion
	// For streaming, use a larger window to reduce reader churn and improve performance
	maxSegments := f.maxWorkers * 8 // 8 segments per worker for better streaming
	if maxSegments > 100 {
		maxSegments = 100 // Cap at 100 segments max for streaming
	}
	if maxSegments < 20 {
		maxSegments = 20 // Minimum of 20 segments for streaming
	}

	// PERFORMANCE OPTIMIZATION: For streaming, use even larger windows to reduce churn
	// If this is a streaming request (large range), increase the window size
	requestedRangeSize := end - start + 1
	if requestedRangeSize > 50*1024*1024 { // 50MB+ requests are likely streaming
		maxSegments = f.maxWorkers * 12 // 12 segments per worker for streaming
		if maxSegments > 150 {
			maxSegments = 150 // Cap at 150 segments for streaming
		}
		if maxSegments < 30 {
			maxSegments = 30 // Minimum of 30 segments for streaming
		}
	}

	sr, err := usenet.BuildSegmentRangeWithLimit(meta, start, end, maxSegments)
	if err != nil {
		return nil, err
	}

	count := usenet.SegmentRangeCount(sr)
	createdAt := time.Unix(meta.GetCreatedAt(), 0)
	var age time.Duration
	if !createdAt.IsZero() {
		age = time.Since(createdAt)
	}

	log.Printf("[streaming] nntp reader start nzb=%q start=%d end=%d segments=%d size=%d groups=%s created=%s age=%s",
		strings.TrimSpace(meta.GetSourceNzbPath()),
		start,
		end,
		count,
		meta.GetFileSize(),
		usenet.SummarizeGroups(sr),
		createdAt.Format(time.RFC3339),
		age,
	)

	reader, err := usenet.NewUsenetReader(ctx, cp, sr, f.maxWorkers)
	if err != nil {
		log.Printf("[streaming] nntp reader error start=%d end=%d err=%v", start, end, err)
		return nil, err
	}
	return reader, nil
}

var _ filesystem.ReaderFactory = (*UsenetReaderFactory)(nil)
