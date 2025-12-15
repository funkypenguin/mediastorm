package streaming

import (
	"context"
	"fmt"
	"log"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"novastream/internal/nzb/filesystem"
	"novastream/internal/nzb/utils"
)

// MetadataProvider streams files from the metadata-backed filesystem.
type MetadataProvider struct {
	remote     *filesystem.MetadataRemoteFile
	streamRoot string
}

func NewMetadataProvider(remote *filesystem.MetadataRemoteFile) *MetadataProvider {
	return &MetadataProvider{remote: remote}
}

func NewMetadataProviderWithRoot(remote *filesystem.MetadataRemoteFile, streamRoot string) *MetadataProvider {
	return &MetadataProvider{remote: remote, streamRoot: streamRoot}
}

func (p *MetadataProvider) Stream(ctx context.Context, req Request) (*Response, error) {
	if p.remote == nil {
		return nil, fmt.Errorf("metadata provider not configured")
	}

	cleanPath := strings.TrimPrefix(req.Path, "/")
	if strings.HasPrefix(cleanPath, "streams/") {
		cleanPath = strings.TrimPrefix(cleanPath, "streams/")
	}

	// Log range requests for seek debugging
	originalRange := strings.TrimSpace(req.RangeHeader)
	rangeHeader := originalRange

	if override, applied := p.maybeOverrideRange(cleanPath, originalRange, req.SeekSeconds, req.DurationHintSeconds); applied {
		rangeHeader = override
	}

	if rangeHeader != "" {
		log.Printf("[streaming] SEEK: provider opening with range=%q path=%q method=%s", rangeHeader, cleanPath, req.Method)
	}

	args := utils.NewPathWithArgs(cleanPath)
	if rangeHeader != "" {
		args.SetRange(rangeHeader)
	}

	ok, file, err := p.remote.OpenFile(ctx, cleanPath, args)
	if err != nil {
		log.Printf("[streaming] provider open error path=%q err=%v", req.Path, err)
		return nil, err
	}
	if !ok {
		log.Printf("[streaming] provider miss path=%q", req.Path)
		return nil, ErrNotFound
	}

	// Update access time for active stream to prevent cleanup
	p.touchStreamDirectory(cleanPath)

	info, err := file.Stat()
	if err != nil {
		_ = file.Close()
		log.Printf("[streaming] provider stat error path=%q err=%v", req.Path, err)
		return nil, err
	}

	trimmedRange := strings.TrimSpace(req.RangeHeader)
	rangeLabel := trimmedRange
	if rangeLabel == "" {
		rangeLabel = "full"
	}

	if descriptor, ok := file.(interface {
		Descriptor() filesystem.FileDescriptor
	}); ok {
		summary := descriptor.Descriptor()
		created := ""
		if !summary.CreatedAt.IsZero() {
			created = summary.CreatedAt.UTC().Format(time.RFC3339)
		}
		log.Printf(
			"[streaming] provider resolved request path=%q virtual=%q source=%q size=%d segments=%d created=%s range=%q method=%s",
			req.Path,
			cleanPath,
			summary.SourceNZB,
			summary.FileSize,
			summary.SegmentCount,
			created,
			rangeLabel,
			req.Method,
		)
	} else {
		log.Printf(
			"[streaming] provider resolved request path=%q virtual=%q size=%d range=%q method=%s",
			req.Path,
			cleanPath,
			info.Size(),
			rangeLabel,
			req.Method,
		)
	}

	headers := defaultHeaders(req.Path)
	if info.Size() >= 0 {
		headers.Set("Content-Length", strconv.FormatInt(info.Size(), 10))
	}

	status := http.StatusOK
	if ranged, ok := file.(interface {
		RangeInfo() (int64, int64, int64)
	}); ok {
		start, end, total := ranged.RangeInfo()
		log.Printf("[streaming] SEEK: file supports ranges - start=%d end=%d total=%d requested=%q", start, end, total, trimmedRange)
		if total > 0 && end >= start {
			if trimmedRange != "" || start > 0 || (end >= 0 && end < total-1) {
				contentRange := fmt.Sprintf("bytes %d-%d/%d", start, end, total)
				headers.Set("Content-Range", contentRange)
				status = http.StatusPartialContent
				rangeSize := end - start + 1
				log.Printf("[streaming] SEEK: provider range response path=%q requested=%q sending=%q content-length=%d range-size=%d", req.Path, trimmedRange, contentRange, info.Size(), rangeSize)
			}
		}
	}
	if status == http.StatusOK && trimmedRange != "" {
		status = http.StatusPartialContent
		log.Printf("[streaming] SEEK: upgraded status to 206 Partial Content for range request=%q", trimmedRange)
	}

	if req.Method == http.MethodHead {
		_ = file.Close()
		log.Printf("[streaming] provider head path=%q size=%d", req.Path, info.Size())
		return &Response{Headers: headers, Status: status, ContentLength: info.Size()}, nil
	}

	log.Printf("[streaming] provider streaming path=%q size=%d", req.Path, info.Size())
	return &Response{
		Body:          file,
		Status:        status,
		Headers:       headers,
		ContentLength: info.Size(),
	}, nil
}

const seekOverrideThresholdBytes int64 = 1 << 20 // 1 MiB cushion to avoid overriding trivial differences

func (p *MetadataProvider) maybeOverrideRange(path, currentRange string, seekSeconds, durationSeconds float64) (string, bool) {
	if p.remote == nil {
		return currentRange, false
	}
	if !isPositiveFinite(seekSeconds) || !isPositiveFinite(durationSeconds) {
		return currentRange, false
	}

	desc, err := p.remote.DescribeFile(path)
	if err != nil || desc == nil || desc.FileSize <= 0 {
		return currentRange, false
	}

	approxStart := approximateByteOffset(desc.FileSize, seekSeconds, durationSeconds)
	if approxStart < 0 {
		return currentRange, false
	}

	originalStart := int64(-1)
	originalEnd := int64(-1)
	if rh, err := utils.ParseRangeHeader(currentRange); err == nil && rh != nil {
		if rh.Start >= 0 {
			originalStart = rh.Start
		}
		if rh.End >= 0 {
			originalEnd = rh.End
		}
	}

	if originalStart >= 0 && approxStart <= originalStart+seekOverrideThresholdBytes {
		return currentRange, false
	}

	var newRange string
	if originalEnd >= approxStart && originalEnd >= 0 {
		newRange = fmt.Sprintf("bytes=%d-%d", approxStart, originalEnd)
	} else {
		newRange = fmt.Sprintf("bytes=%d-", approxStart)
	}

	log.Printf(
		"[streaming] SEEK: overriding range using seekSeconds=%.3f duration=%.3f approxStart=%d originalStart=%d fileSize=%d path=%q",
		seekSeconds,
		durationSeconds,
		approxStart,
		originalStart,
		desc.FileSize,
		path,
	)

	return newRange, newRange != currentRange
}

func approximateByteOffset(size int64, seekSeconds, durationSeconds float64) int64 {
	if size <= 0 || !isPositiveFinite(durationSeconds) || seekSeconds < 0 || math.IsInf(seekSeconds, 0) || math.IsNaN(seekSeconds) {
		return -1
	}

	if seekSeconds >= durationSeconds {
		return size - 1
	}

	ratio := seekSeconds / durationSeconds
	if ratio <= 0 {
		return 0
	}

	approx := int64(math.Round(ratio * float64(size)))
	if approx >= size {
		approx = size - 1
	}
	if approx < 0 {
		approx = 0
	}
	return approx
}

func isPositiveFinite(value float64) bool {
	if value <= 0 || math.IsNaN(value) || math.IsInf(value, 0) {
		return false
	}
	return true
}

func defaultHeaders(path string) http.Header {
	headers := make(http.Header)
	headers.Set("Accept-Ranges", "bytes")
	headers.Set("Content-Type", guessContentType(path))
	return headers
}

var containerContentTypes = map[string]string{
	".mp4":  "video/mp4",
	".m4v":  "video/mp4",
	".webm": "video/webm",
	".mkv":  "video/x-matroska",
	".ts":   "video/mp2t",
	".m2ts": "video/mp2t",
	".mts":  "video/mp2t",
	".avi":  "video/x-msvideo",
	".mpg":  "video/mpeg",
	".mpeg": "video/mpeg",
	".m3u8": "application/vnd.apple.mpegurl",
}

func guessContentType(path string) string {
	ext := detectContainerExt(path)
	if ext == "" {
		return "application/octet-stream"
	}
	if ct, ok := containerContentTypes[ext]; ok {
		return ct
	}
	return "application/octet-stream"
}

func detectContainerExt(name string) string {
	lower := strings.ToLower(strings.TrimSpace(name))
	if lower == "" {
		return ""
	}

	ext := strings.ToLower(strings.TrimSpace(filepath.Ext(lower)))
	if _, ok := containerContentTypes[ext]; ok {
		return ext
	}

	for candidate := range containerContentTypes {
		if strings.HasSuffix(lower, candidate) ||
			strings.Contains(lower, candidate+"_") ||
			strings.Contains(lower, candidate+".") ||
			strings.Contains(lower, candidate+"-") {
			return candidate
		}
	}

	return ""
}

// touchStreamDirectory updates the modification time of the stream directory
// to prevent it from being cleaned up while actively in use.
func (p *MetadataProvider) touchStreamDirectory(cleanPath string) {
	if p.streamRoot == "" {
		return
	}

	// Extract stream ID from path like "46c1e06e-7843-45f5-87a2-84c42d8b4f36/file.mkv"
	parts := strings.Split(cleanPath, "/")
	if len(parts) == 0 {
		return
	}

	streamID := parts[0]
	streamDir := filepath.Join(p.streamRoot, streamID)

	// Update the directory's modification time
	now := time.Now()
	if err := os.Chtimes(streamDir, now, now); err != nil {
		// Log error but don't fail the stream request
		if !os.IsNotExist(err) {
			log.Printf("[streaming] failed to update access time for %s: %v", streamDir, err)
		}
	}
}
