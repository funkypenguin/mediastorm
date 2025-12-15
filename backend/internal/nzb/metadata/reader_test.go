package metadata_test

import (
	"path/filepath"
	"testing"

	"novastream/internal/nzb/metadata"
	metapb "novastream/internal/nzb/metadata/proto"
)

func TestMetadataReader_ListDirectoryContents(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	svc := metadata.NewMetadataService(root)
	reader := metadata.NewMetadataReader(svc)

	baseMeta := svc.CreateFileMetadata(
		42,
		"library/sample.nzb",
		metapb.FileStatus_FILE_STATUS_HEALTHY,
		[]*metapb.SegmentData{
			{SegmentSize: 512, StartOffset: 0, EndOffset: 511, Id: "<seg@usenet>"},
		},
		metapb.Encryption_NONE,
		"",
		"",
	)

	if err := svc.WriteFileMetadata("movies/movie.mkv", baseMeta); err != nil {
		t.Fatalf("WriteFileMetadata() error = %v", err)
	}

	if err := svc.WriteFileMetadata("movies/extras/clip.mkv", baseMeta); err != nil {
		t.Fatalf("WriteFileMetadata() extras error = %v", err)
	}

	dirs, files, err := reader.ListDirectoryContents("movies")
	if err != nil {
		t.Fatalf("ListDirectoryContents() error = %v", err)
	}

	if len(files) != 1 {
		t.Fatalf("ListDirectoryContents() files len = %d, want 1", len(files))
	}
	if files[0].SourceNzbPath != baseMeta.SourceNzbPath {
		t.Fatalf("ListDirectoryContents() file metadata mismatch: %+v", files[0])
	}

	if len(dirs) != 1 {
		t.Fatalf("ListDirectoryContents() dirs len = %d, want 1", len(dirs))
	}
	if dirs[0].Name() != "extras" {
		t.Fatalf("ListDirectoryContents() dirs[0].Name() = %q, want extras", dirs[0].Name())
	}
}

func TestMetadataReader_PathChecks(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	svc := metadata.NewMetadataService(root)
	reader := metadata.NewMetadataReader(svc)

	meta := svc.CreateFileMetadata(
		10,
		"shows/episode.nzb",
		metapb.FileStatus_FILE_STATUS_HEALTHY,
		nil,
		metapb.Encryption_NONE,
		"",
		"",
	)

	if err := svc.WriteFileMetadata("shows/episode1.mkv", meta); err != nil {
		t.Fatalf("WriteFileMetadata() error = %v", err)
	}

	exists, err := reader.PathExists("shows")
	if err != nil {
		t.Fatalf("PathExists() error = %v", err)
	}
	if !exists {
		t.Fatalf("PathExists(%q) = false, want true", "shows")
	}

	isDir, err := reader.IsDirectory("shows")
	if err != nil {
		t.Fatalf("IsDirectory() error = %v", err)
	}
	if !isDir {
		t.Fatalf("IsDirectory(%q) = false, want true", "shows")
	}

	isDir, err = reader.IsDirectory("shows/episode1.mkv")
	if err != nil {
		t.Fatalf("IsDirectory(file) error = %v", err)
	}
	if isDir {
		t.Fatalf("IsDirectory(file) = true, want false")
	}

	segs, err := reader.GetFileSegments("shows/episode1.mkv")
	if err != nil {
		t.Fatalf("GetFileSegments() error = %v", err)
	}
	if len(segs) != 0 {
		t.Fatalf("GetFileSegments() len = %d, want 0", len(segs))
	}

	info, err := reader.GetDirectoryInfo("shows")
	if err != nil {
		t.Fatalf("GetDirectoryInfo() error = %v", err)
	}
	if !info.IsDir() {
		t.Fatalf("GetDirectoryInfo().IsDir = false, want true")
	}

	// Non-existent paths
	if _, err = reader.IsDirectory("missing"); err == nil {
		t.Fatalf("IsDirectory(%q) expected error", "missing")
	}

	if _, err = reader.GetFileSegments("missing/file.mkv"); err == nil {
		t.Fatalf("GetFileSegments missing expected error")
	}

	if exists, err = reader.PathExists("missing"); err != nil {
		t.Fatalf("PathExists missing error = %v", err)
	} else if exists {
		t.Fatalf("PathExists missing = true, want false")
	}
}

func TestMetadataReader_ListDirectoryContentsMissing(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	svc := metadata.NewMetadataService(root)
	reader := metadata.NewMetadataReader(svc)

	dirs, files, err := reader.ListDirectoryContents("missing")
	if err != nil {
		t.Fatalf("ListDirectoryContents missing error = %v", err)
	}
	if len(dirs) != 0 || len(files) != 0 {
		t.Fatalf("ListDirectoryContents missing should be empty, got dirs=%d files=%d", len(dirs), len(files))
	}

	// ListDirectory should align with ListDirectoryContents when no entries exist
	entries, err := svc.ListDirectory("missing")
	if err != nil {
		t.Fatalf("ListDirectory missing error = %v", err)
	}
	if len(entries) != 0 {
		t.Fatalf("ListDirectory missing expected empty, got %v", entries)
	}

	// Ensure directory listing surfaces nested directories once metadata is written
	meta := svc.CreateFileMetadata(1, "base.nzb", metapb.FileStatus_FILE_STATUS_HEALTHY, nil, metapb.Encryption_NONE, "", "")
	if err := svc.WriteFileMetadata(filepath.Join("missing", "new", "file.mkv"), meta); err != nil {
		t.Fatalf("WriteFileMetadata nested error = %v", err)
	}
	dirs, files, err = reader.ListDirectoryContents("missing")
	if err != nil {
		t.Fatalf("ListDirectoryContents existing error = %v", err)
	}
	if len(dirs) != 1 || dirs[0].Name() != "new" {
		t.Fatalf("ListDirectoryContents nested dirs mismatch: %#v", dirs)
	}
	if len(files) != 0 {
		t.Fatalf("ListDirectoryContents nested files len = %d, want 0", len(files))
	}
}
