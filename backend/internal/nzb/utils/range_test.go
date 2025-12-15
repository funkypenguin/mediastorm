package utils_test

import (
	"testing"

	"novastream/internal/nzb/utils"
)

func TestParseRangeHeader(t *testing.T) {
	t.Parallel()

	hdr, err := utils.ParseRangeHeader("bytes=0-1023")
	if err != nil {
		t.Fatalf("ParseRangeHeader() error = %v", err)
	}
	if hdr.Start != 0 || hdr.End != 1023 {
		t.Fatalf("ParseRangeHeader() = %#v, want start=0 end=1023", hdr)
	}

	if _, err := utils.ParseRangeHeader("items=0-1"); err == nil {
		t.Fatalf("expected error for invalid preamble")
	}

	if _, err := utils.ParseRangeHeader("bytes=0-1,2-3"); err == nil {
		t.Fatalf("expected error for multi-range header")
	}
}

func TestRangeDecodeFix(t *testing.T) {
	t.Parallel()

	hdr := &utils.RangeHeader{Start: 100, End: 199}
	offset, limit := hdr.Decode(1000)
	if offset != 100 || limit != 100 {
		t.Fatalf("Decode() = offset %d limit %d, want 100 100", offset, limit)
	}

	suffix := &utils.RangeHeader{Start: -1, End: 200}
	fixed := utils.FixRangeHeader(suffix, 1000)
	if fixed.Start != 800 || fixed.End != 999 {
		t.Fatalf("FixRangeHeader suffix = %#v, want start=800 end=999", fixed)
	}

	offset, limit = suffix.Decode(1000)
	if offset != 800 || limit != -1 {
		t.Fatalf("Decode suffix = %d %d, want 800 -1", offset, limit)
	}
}
