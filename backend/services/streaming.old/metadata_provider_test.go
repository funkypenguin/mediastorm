package streaming_test

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"strings"
	"testing"

	"novastream/internal/nzb/filesystem"
	"novastream/internal/nzb/metadata"
	metapb "novastream/internal/nzb/metadata/proto"
	streaming "novastream/services/streaming.old"
)

func TestMetadataProviderStream(t *testing.T) {
	svc := metadata.NewMetadataService(t.TempDir())
	meta := svc.CreateFileMetadata(5, "source.nzb", metapb.FileStatus_FILE_STATUS_HEALTHY, nil, metapb.Encryption_NONE, "", "")
	if err := svc.WriteFileMetadata("movies/title.mkv", meta); err != nil {
		t.Fatalf("WriteFileMetadata() error = %v", err)
	}

	data := []byte("hello")

	remote := filesystem.NewMetadataRemoteFile(filesystem.RemoteFileConfig{
		MetadataService: svc,
		ReaderFactory: filesystem.ReaderFactoryFunc(func(ctx context.Context, meta *metapb.FileMetadata, start, end int64) (io.ReadCloser, error) {
			if end < start {
				end = int64(len(data)) - 1
			}
			reader := bytes.NewReader(data)
			section := io.NewSectionReader(reader, start, end-start+1)
			return io.NopCloser(section), nil
		}),
	})

	provider := streaming.NewMetadataProvider(remote)

	resp, err := provider.Stream(context.Background(), streaming.Request{Path: "movies/title.mkv", Method: http.MethodGet})
	if err != nil {
		t.Fatalf("Stream() error = %v", err)
	}
	defer resp.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("ReadAll() error = %v", err)
	}
	if string(body) != string(data) {
		t.Fatalf("body = %q, want %q", body, data)
	}
}

func TestMetadataProviderStream_AppliesSeekOverride(t *testing.T) {
	svc := metadata.NewMetadataService(t.TempDir())
	meta := svc.CreateFileMetadata(1000, "source.nzb", metapb.FileStatus_FILE_STATUS_HEALTHY, nil, metapb.Encryption_NONE, "", "")
	if err := svc.WriteFileMetadata("movies/title.mkv", meta); err != nil {
		t.Fatalf("WriteFileMetadata() error = %v", err)
	}

	recordedStart := int64(-1)
	recordedEnd := int64(-1)
	remote := filesystem.NewMetadataRemoteFile(filesystem.RemoteFileConfig{
		MetadataService: svc,
		ReaderFactory: filesystem.ReaderFactoryFunc(func(ctx context.Context, meta *metapb.FileMetadata, start, end int64) (io.ReadCloser, error) {
			recordedStart = start
			recordedEnd = end
			return io.NopCloser(bytes.NewReader(make([]byte, meta.GetFileSize()))), nil
		}),
	})

	provider := streaming.NewMetadataProvider(remote)
	req := streaming.Request{
		Path:                "movies/title.mkv",
		Method:              http.MethodGet,
		SeekSeconds:         75,
		DurationHintSeconds: 100,
	}

	resp, err := provider.Stream(context.Background(), req)
	if err != nil {
		t.Fatalf("Stream() error = %v", err)
	}
	defer resp.Close()

	// Trigger the reader to be created
	buf := make([]byte, 16)
	_, _ = resp.Body.Read(buf)

	if recordedStart < 730 || recordedStart > 770 {
		t.Fatalf("expected start offset around 750, got %d", recordedStart)
	}
	if recordedEnd != meta.GetFileSize()-1 {
		t.Fatalf("expected end offset %d, got %d", meta.GetFileSize()-1, recordedEnd)
	}

	contentRange := resp.Headers.Get("Content-Range")
	if !strings.HasPrefix(contentRange, "bytes 750-") {
		t.Fatalf("Content-Range = %q, want prefix bytes 750-", contentRange)
	}
}
