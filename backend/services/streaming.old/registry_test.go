package streaming

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	streammeta "novastream/internal/nzb/metadata"
	metapb "novastream/internal/nzb/metadata/proto"
)

const sampleNZB = `<?xml version="1.0" encoding="utf-8"?>
<nzb>
  <file subject="example.mkv">
    <segments>
      <segment bytes="100">abc@example</segment>
    </segments>
  </file>
</nzb>`

const multiFileNZB = `<?xml version="1.0" encoding="utf-8"?>
<nzb>
  <file subject="sample.par2">
    <segments>
      <segment bytes="50">par@example</segment>
    </segments>
  </file>
  <file subject="movie.mkv">
    <segments>
      <segment bytes="75">mkv1@example</segment>
      <segment bytes="75">mkv2@example</segment>
    </segments>
  </file>
  <file subject="movie.mp4">
    <segments>
      <segment bytes="150">mp41@example</segment>
    </segments>
  </file>
</nzb>`

const parityHeavyNZB = `<?xml version="1.0" encoding="utf-8"?>
<nzb>
  <file subject="124_124_-_Superman.2025.1080p.Bluray.TrueHD.x265-LuCY.vol-08.par2_yEnc_438436804_1_612">
    <segments>
      <segment bytes="452429383">par@example</segment>
    </segments>
  </file>
  <file subject="Superman.2025.1080p.Bluray.TrueHD.x265-LuCY.mkv yEnc (1/612)">
    <segments>
      <segment bytes="500000000">video@example</segment>
    </segments>
  </file>
</nzb>`

func TestRegistryRegisterCreatesMetadataAndCleansOld(t *testing.T) {
	root := t.TempDir()
	metaSvc := streammeta.NewMetadataService(root)
	reg := NewRegistry(metaSvc, root, time.Hour)

	// Seed an old entry that should be removed.
	oldDir := filepath.Join(root, "old-entry")
	if err := os.MkdirAll(oldDir, 0o755); err != nil {
		t.Fatalf("mkdir old: %v", err)
	}
	meta := metaSvc.CreateFileMetadata(10, "old.nzb", metapb.FileStatus_FILE_STATUS_HEALTHY, nil, metapb.Encryption_NONE, "", "")
	if err := metaSvc.WriteFileMetadata(filepath.Join("old-entry", "stub.mkv"), meta); err != nil {
		t.Fatalf("write old metadata: %v", err)
	}
	oldTime := time.Now().Add(-2 * time.Hour)
	if err := os.Chtimes(oldDir, oldTime, oldTime); err != nil {
		t.Fatalf("chtimes: %v", err)
	}

	reg.maxAge = time.Minute

	registration, err := reg.Register(context.Background(), "sample.nzb", []byte(sampleNZB))
	if err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	// New metadata exists
	if registration == nil || registration.Path == "" {
		t.Fatalf("registration invalid: %+v", registration)
	}
	rel := strings.TrimPrefix(registration.Path, "streams/")
	if rel == registration.Path {
		t.Fatalf("registration path %q missing streams/ prefix", registration.Path)
	}
	registeredMeta, err := metaSvc.ReadFileMetadata(rel)
	if err != nil {
		t.Fatalf("metadata not written: %v", err)
	}
	if registeredMeta == nil {
		t.Fatalf("metadata missing for %q", rel)
	}
	if strings.TrimSpace(registeredMeta.SourceNzbPath) == "" {
		t.Fatalf("source NZB path not recorded")
	}
	if !strings.HasPrefix(registeredMeta.SourceNzbPath, root) {
		t.Fatalf("source NZB path %q not under root %q", registeredMeta.SourceNzbPath, root)
	}
	writtenNZB, err := os.ReadFile(registeredMeta.SourceNzbPath)
	if err != nil {
		t.Fatalf("nzb file not written: %v", err)
	}
	if !bytes.Equal(writtenNZB, []byte(sampleNZB)) {
		t.Fatalf("nzb file contents mismatch")
	}

	if _, err := os.Stat(oldDir); !os.IsNotExist(err) {
		t.Fatalf("expected old dir removed, err=%v", err)
	}
}

func TestParseNZBSelectsVideoFile(t *testing.T) {
	parsed, err := parseNZB([]byte(multiFileNZB))
	if err != nil {
		t.Fatalf("parseNZB error = %v", err)
	}
	if parsed.FileName != "movie.mp4" {
		t.Fatalf("FileName = %q, want movie.mp4", parsed.FileName)
	}
	if len(parsed.Segments) != 1 || parsed.Segments[0].Bytes != 150 {
		t.Fatalf("unexpected segments: %+v", parsed.Segments)
	}
}

func TestParseNZBSkipsParitySubjects(t *testing.T) {
	parsed, err := parseNZB([]byte(parityHeavyNZB))
	if err != nil {
		t.Fatalf("parseNZB error = %v", err)
	}
	if parsed.FileName != "Superman.2025.1080p.Bluray.TrueHD.x265-LuCY.mkv yEnc (1/612)" {
		t.Fatalf("FileName = %q, want mkv entry", parsed.FileName)
	}
	if len(parsed.Segments) != 1 || parsed.Segments[0].Bytes != 500000000 {
		t.Fatalf("unexpected segments: %+v", parsed.Segments)
	}
}

func TestSanitizeFileName(t *testing.T) {
	got := sanitizeFileName(" ../Some Movie!.mkv  ")
	if got != "Some_Movie_.mkv" {
		t.Fatalf("sanitizeFileName = %q", got)
	}
}

func TestSanitizeFileNameWithEmbeddedSlash(t *testing.T) {
	input := `[1/8] - "6qGG1n7jbSls1sVfuHHPZRsJOlsZCLQF.mkv" yEnc  5655451165 (1/7890)`
	want := "1_8_-_6qGG1n7jbSls1sVfuHHPZRsJOlsZCLQF.mkv_yEnc_5655451165_1_7890"
	if got := sanitizeFileName(input); got != want {
		t.Fatalf("sanitizeFileName(%q) = %q, want %q", input, got, want)
	}
}
