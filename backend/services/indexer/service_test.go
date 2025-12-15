package indexer

import "testing"

func TestParseSize(t *testing.T) {
	if size := parseSize("1024", ""); size != 1024 {
		t.Fatalf("expected 1024, got %d", size)
	}
	if size := parseSize("", "2048"); size != 2048 {
		t.Fatalf("expected 2048, got %d", size)
	}
	if size := parseSize("abc", "xyz"); size != 0 {
		t.Fatalf("expected 0 for invalid inputs, got %d", size)
	}
}

func TestParsePubDate(t *testing.T) {
	sample := "Mon, 02 Jan 2006 15:04:05 -0700"
	parsed := parsePubDate(sample)
	if parsed.IsZero() {
		t.Fatal("expected parsed time")
	}
	if parsed.Year() != 2006 {
		t.Fatalf("expected year 2006, got %d", parsed.Year())
	}
	if !parsePubDate("invalid").IsZero() {
		t.Fatal("expected zero time for invalid date")
	}
}

func TestDedupe(t *testing.T) {
	got := dedupe([]string{"Action", "action", " Drama ", ""})
	if len(got) != 2 {
		t.Fatalf("expected 2 categories, got %d", len(got))
	}
	if got[0] != "Action" {
		t.Fatalf("expected first item to be Action, got %s", got[0])
	}
	if got[1] != "Drama" {
		t.Fatalf("expected second item to be Drama, got %s", got[1])
	}
}
