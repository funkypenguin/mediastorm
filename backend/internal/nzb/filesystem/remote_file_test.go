package filesystem_test

import (
	"bytes"
	"context"
	"errors"
	"io"
	"testing"

	"novastream/internal/nzb/filesystem"
	"novastream/internal/nzb/metadata"
	metapb "novastream/internal/nzb/metadata/proto"
	"novastream/internal/nzb/utils"
)

type stubHealthReporter struct {
	partial   []string
	corrupted []string
}

func (s *stubHealthReporter) MarkPartial(path string, err error) { s.partial = append(s.partial, path) }
func (s *stubHealthReporter) MarkCorrupted(path string, err error) {
	s.corrupted = append(s.corrupted, path)
}

func newMetadataService(t *testing.T) *metadata.MetadataService {
	t.Helper()
	return metadata.NewMetadataService(t.TempDir())
}

func writeMetadata(t *testing.T, svc *metadata.MetadataService, path string, size int64, status metapb.FileStatus) {
	t.Helper()
	meta := svc.CreateFileMetadata(size, "source.nzb", status, nil, metapb.Encryption_NONE, "", "")
	if err := svc.WriteFileMetadata(path, meta); err != nil {
		t.Fatalf("WriteFileMetadata() error = %v", err)
	}
}

func TestMetadataRemoteFile_OpenDirectory(t *testing.T) {
	svc := newMetadataService(t)
	writeMetadata(t, svc, "shows/season/file.mkv", 10, metapb.FileStatus_FILE_STATUS_HEALTHY)

	cfg := filesystem.RemoteFileConfig{
		MetadataService: svc,
		HealthReporter:  &stubHealthReporter{},
		ReaderFactory: filesystem.ReaderFactoryFunc(func(ctx context.Context, meta *metapb.FileMetadata, start, end int64) (io.ReadCloser, error) {
			return io.NopCloser(bytes.NewReader(nil)), nil
		}),
	}

	mrf := filesystem.NewMetadataRemoteFile(cfg)
	ok, file, err := mrf.OpenFile(context.Background(), "shows/season", utils.NewPathWithArgs("shows/season"))
	if err != nil {
		t.Fatalf("OpenFile() error = %v", err)
	}
	if !ok {
		t.Fatalf("OpenFile() ok = false, want true for directory")
	}
	if file == nil {
		t.Fatalf("OpenFile() returned nil file")
	}
}

func TestMetadataRemoteFile_OpenFile_ReadsBytes(t *testing.T) {
	svc := newMetadataService(t)
	writeMetadata(t, svc, "movies/title.mkv", 8, metapb.FileStatus_FILE_STATUS_HEALTHY)

	cfg := filesystem.RemoteFileConfig{
		MetadataService: svc,
		HealthReporter:  &stubHealthReporter{},
		ReaderFactory: filesystem.ReaderFactoryFunc(func(ctx context.Context, meta *metapb.FileMetadata, start, end int64) (io.ReadCloser, error) {
			return io.NopCloser(bytes.NewReader([]byte{1, 2, 3, 4, 5, 6, 7, 8})), nil
		}),
	}

	mrf := filesystem.NewMetadataRemoteFile(cfg)
	ok, file, err := mrf.OpenFile(context.Background(), "movies/title.mkv", utils.NewPathWithArgs("movies/title.mkv"))
	if err != nil {
		t.Fatalf("OpenFile() error = %v", err)
	}
	if !ok {
		t.Fatalf("OpenFile() ok = false, want true")
	}

	buf := make([]byte, 8)
	n, err := file.Read(buf)
	if err != nil && !errors.Is(err, io.EOF) {
		t.Fatalf("Read() error = %v", err)
	}
	if n == 0 {
		t.Fatalf("Read() read 0 bytes, want > 0")
	}
}

func TestMetadataRemoteFile_CorruptedMetadata(t *testing.T) {
	svc := newMetadataService(t)
	writeMetadata(t, svc, "movies/bad.mkv", 10, metapb.FileStatus_FILE_STATUS_CORRUPTED)

	cfg := filesystem.RemoteFileConfig{
		MetadataService: svc,
		HealthReporter:  &stubHealthReporter{},
		ReaderFactory: filesystem.ReaderFactoryFunc(func(ctx context.Context, meta *metapb.FileMetadata, start, end int64) (io.ReadCloser, error) {
			return io.NopCloser(bytes.NewReader(nil)), nil
		}),
	}

	mrf := filesystem.NewMetadataRemoteFile(cfg)
	ok, _, err := mrf.OpenFile(context.Background(), "movies/bad.mkv", utils.NewPathWithArgs("movies/bad.mkv"))
	if err != filesystem.ErrFileIsCorrupted {
		t.Fatalf("OpenFile() err = %v, want ErrFileIsCorrupted", err)
	}
	if ok {
		t.Fatalf("OpenFile() ok = true, want false for corrupted file")
	}
}

func TestMetadataRemoteFile_DescribeFile(t *testing.T) {
	svc := newMetadataService(t)
	writeMetadata(t, svc, "movies/title.mkv", 2048, metapb.FileStatus_FILE_STATUS_HEALTHY)

	mrf := filesystem.NewMetadataRemoteFile(filesystem.RemoteFileConfig{MetadataService: svc})

	desc, err := mrf.DescribeFile("movies/title.mkv")
	if err != nil {
		t.Fatalf("DescribeFile() error = %v", err)
	}
	if desc == nil {
		t.Fatalf("DescribeFile() returned nil descriptor")
	}
	if desc.FileSize != 2048 {
		t.Fatalf("DescribeFile() FileSize = %d, want 2048", desc.FileSize)
	}
	if desc.VirtualPath != "movies/title.mkv" {
		t.Fatalf("DescribeFile() VirtualPath = %q, want movies/title.mkv", desc.VirtualPath)
	}
}
