package metadata_test

import (
	"path/filepath"
	"testing"
	"time"

	"novastream/internal/nzb/metadata"
	metapb "novastream/internal/nzb/metadata/proto"
)

func TestWriteAndReadMetadata(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	svc := metadata.NewMetadataService(root)

	original := svc.CreateFileMetadata(
		12345,
		"library/movie.nzb",
		metapb.FileStatus_FILE_STATUS_HEALTHY,
		[]*metapb.SegmentData{
			{
				SegmentSize: 2048,
				StartOffset: 0,
				EndOffset:   2047,
				Id:          "<segment@usenet>",
			},
		},
		metapb.Encryption_NONE,
		"",
		"",
	)

	const virtualPath = "movies/example.mkv"
	if err := svc.WriteFileMetadata(virtualPath, original); err != nil {
		t.Fatalf("WriteFileMetadata() error = %v", err)
	}

	read, err := svc.ReadFileMetadata(virtualPath)
	if err != nil {
		t.Fatalf("ReadFileMetadata() error = %v", err)
	}
	if read == nil {
		t.Fatalf("ReadFileMetadata() returned nil metadata")
	}

	if read.FileSize != original.FileSize || read.SourceNzbPath != original.SourceNzbPath {
		t.Fatalf("unexpected metadata contents: %+v", read)
	}

	if !svc.FileExists(virtualPath) {
		t.Fatalf("FileExists(%q) = false, want true", virtualPath)
	}

	files, err := svc.ListDirectory(filepath.Dir(virtualPath))
	if err != nil {
		t.Fatalf("ListDirectory() error = %v", err)
	}
	if len(files) != 1 || files[0] != filepath.Base(virtualPath) {
		t.Fatalf("ListDirectory() = %v, want [%q]", files, filepath.Base(virtualPath))
	}

	dirs, err := svc.ListSubdirectories("")
	if err != nil {
		t.Fatalf("ListSubdirectories() error = %v", err)
	}
	if len(dirs) != 1 || dirs[0] != filepath.Dir(virtualPath) {
		t.Fatalf("ListSubdirectories() = %v, want [%q]", dirs, filepath.Dir(virtualPath))
	}
}

func TestUpdateAndDeleteMetadata(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	svc := metadata.NewMetadataService(root)

	meta := svc.CreateFileMetadata(
		512,
		"shows/example.nzb",
		metapb.FileStatus_FILE_STATUS_HEALTHY,
		nil,
		metapb.Encryption_NONE,
		"",
		"",
	)

	const virtualPath = "shows/episode.mkv"
	if err := svc.WriteFileMetadata(virtualPath, meta); err != nil {
		t.Fatalf("WriteFileMetadata() error = %v", err)
	}

	before, err := svc.ReadFileMetadata(virtualPath)
	if err != nil {
		t.Fatalf("ReadFileMetadata() error = %v", err)
	}

	time.Sleep(1100 * time.Millisecond)

	if err := svc.UpdateFileMetadata(virtualPath, func(m *metapb.FileMetadata) {
		m.FileSize = 2048
		m.Status = metapb.FileStatus_FILE_STATUS_PARTIAL
	}); err != nil {
		t.Fatalf("UpdateFileMetadata() error = %v", err)
	}

	updated, err := svc.ReadFileMetadata(virtualPath)
	if err != nil {
		t.Fatalf("ReadFileMetadata() after update error = %v", err)
	}
	if updated.FileSize != 2048 {
		t.Fatalf("updated FileSize = %d, want 2048", updated.FileSize)
	}
	if updated.Status != metapb.FileStatus_FILE_STATUS_PARTIAL {
		t.Fatalf("updated Status = %v, want partial", updated.Status)
	}
	if updated.ModifiedAt <= before.ModifiedAt {
		t.Fatalf("ModifiedAt not bumped: before=%d after=%d", before.ModifiedAt, updated.ModifiedAt)
	}

	if err := svc.UpdateFileStatus(virtualPath, metapb.FileStatus_FILE_STATUS_CORRUPTED); err != nil {
		t.Fatalf("UpdateFileStatus() error = %v", err)
	}
	final, err := svc.ReadFileMetadata(virtualPath)
	if err != nil {
		t.Fatalf("ReadFileMetadata() final error = %v", err)
	}
	if final.Status != metapb.FileStatus_FILE_STATUS_CORRUPTED {
		t.Fatalf("UpdateFileStatus() status = %v, want corrupted", final.Status)
	}

	if err := svc.DeleteFileMetadata(virtualPath); err != nil {
		t.Fatalf("DeleteFileMetadata() error = %v", err)
	}
	if svc.FileExists(virtualPath) {
		t.Fatalf("FileExists(%q) = true, want false after delete", virtualPath)
	}
	if meta, err := svc.ReadFileMetadata(virtualPath); err != nil {
		t.Fatalf("ReadFileMetadata() after delete error = %v", err)
	} else if meta != nil {
		t.Fatalf("expected nil metadata after delete, got %+v", meta)
	}
}

func TestListDirectoryMissing(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	svc := metadata.NewMetadataService(root)

	time.Sleep(1 * time.Millisecond) // ensure no race with temp dir removal

	files, err := svc.ListDirectory("missing")
	if err != nil {
		t.Fatalf("ListDirectory() error = %v", err)
	}
	if len(files) != 0 {
		t.Fatalf("ListDirectory() = %v, want empty", files)
	}

	dirs, err := svc.ListSubdirectories("missing")
	if err != nil {
		t.Fatalf("ListSubdirectories() error = %v", err)
	}
	if len(dirs) != 0 {
		t.Fatalf("ListSubdirectories() = %v, want empty", dirs)
	}
}
