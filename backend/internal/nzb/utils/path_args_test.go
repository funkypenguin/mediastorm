package utils_test

import (
	"testing"

	"novastream/internal/nzb/utils"
)

func TestNewPathWithArgsFromString(t *testing.T) {
	t.Parallel()

	p, err := utils.NewPathWithArgsFromString("/webdav/file?ARGS?webdav%20context%20key%20rangeKey=bytes%3D0-99")
	if err != nil {
		t.Fatalf("NewPathWithArgsFromString() error = %v", err)
	}

	if p.Path != "/webdav/file" {
		t.Fatalf("Path = %q, want /webdav/file", p.Path)
	}

	rng, err := p.Range()
	if err != nil {
		t.Fatalf("Range() error = %v", err)
	}
	if rng == nil || rng.Start != 0 || rng.End != 99 {
		t.Fatalf("Range() = %#v, want start=0 end=99", rng)
	}

	// Should round-trip using String()
	if got := p.String(); got != "/webdav/file?ARGS?webdav+context+key+rangeKey=bytes%3D0-99" && got != "/webdav/file?ARGS?webdav%20context%20key%20rangeKey=bytes%3D0-99" {
		t.Fatalf("String() = %q, want encoded args", got)
	}
}

func TestPathWithArgsMutators(t *testing.T) {
	t.Parallel()

	p := utils.NewPathWithArgs("/webdav/file")
	p.SetFileSize("123")
	p.SetRange("bytes=50-100")
	p.SetIsCopy()
	p.SetOrigin("unit-test")

	if p.Path != "/webdav/file" {
		t.Fatalf("Path = %q, want /webdav/file", p.Path)
	}

	if got, err := p.FileSize(); err != nil || got != 123 {
		t.Fatalf("FileSize() = %d, err=%v; want 123", got, err)
	}

	rng, err := p.Range()
	if err != nil {
		t.Fatalf("Range() error = %v", err)
	}
	if rng == nil || rng.Start != 50 || rng.End != 100 {
		t.Fatalf("Range() = %#v, want 50-100", rng)
	}

	if !p.IsCopy() {
		t.Fatalf("IsCopy() = false, want true")
	}

	if origin := p.Origin(); origin != "unit-test" {
		t.Fatalf("Origin() = %q, want unit-test", origin)
	}
}

func TestNewPathWithArgsFromStringInvalid(t *testing.T) {
	t.Parallel()

	if _, err := utils.NewPathWithArgsFromString("/foo?ARGS?not%%not%%query"); err == nil {
		t.Fatalf("expected error when query string malformed")
	}
}
