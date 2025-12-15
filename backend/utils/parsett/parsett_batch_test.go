package parsett

import (
	"testing"
	"time"
)

func TestParseTitleBatch(t *testing.T) {
	titles := []string{
		"The.Matrix.1999.1080p.BluRay.x264-SPARKS",
		"Inception.2010.720p.BluRay.x264",
		"The.Simpsons.S01E01.1080p.BluRay.x265",
		"The.Dark.Knight.2008.1080p.BluRay.x264.DTS-HD.MA.5.1-RARBG",
	}

	start := time.Now()
	results, err := ParseTitleBatch(titles)
	elapsed := time.Since(start)

	if err != nil {
		t.Fatalf("ParseTitleBatch failed: %v", err)
	}

	if len(results) != len(titles) {
		t.Errorf("Expected %d results, got %d", len(titles), len(results))
	}

	// Check The Matrix
	if parsed := results["The.Matrix.1999.1080p.BluRay.x264-SPARKS"]; parsed == nil {
		t.Error("The Matrix result is nil")
	} else {
		if parsed.Title != "The Matrix" {
			t.Errorf("Expected title 'The Matrix', got '%s'", parsed.Title)
		}
		if parsed.Year != 1999 {
			t.Errorf("Expected year 1999, got %d", parsed.Year)
		}
	}

	// Check Inception
	if parsed := results["Inception.2010.720p.BluRay.x264"]; parsed == nil {
		t.Error("Inception result is nil")
	} else {
		if parsed.Title != "Inception" {
			t.Errorf("Expected title 'Inception', got '%s'", parsed.Title)
		}
		if parsed.Year != 2010 {
			t.Errorf("Expected year 2010, got %d", parsed.Year)
		}
	}

	t.Logf("Batch parsed %d titles in %v (avg: %v per title)", len(titles), elapsed, elapsed/time.Duration(len(titles)))
}

func BenchmarkParseTitleBatch_vs_Individual(b *testing.B) {
	titles := []string{
		"The.Matrix.1999.1080p.BluRay.x264-SPARKS",
		"Inception.2010.720p.BluRay.x264",
		"The.Simpsons.S01E01.1080p.BluRay.x265",
		"The.Dark.Knight.2008.1080p.BluRay.x264.DTS-HD.MA.5.1-RARBG",
		"Interstellar.2014.1080p.BluRay.x264",
		"The.Godfather.1972.1080p.BluRay.x264",
		"Pulp.Fiction.1994.1080p.BluRay.x264",
		"Fight.Club.1999.1080p.BluRay.x264",
		"Forrest.Gump.1994.1080p.BluRay.x264",
		"The.Shawshank.Redemption.1994.1080p.BluRay.x264",
	}

	b.Run("Batch", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, err := ParseTitleBatch(titles)
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("Individual", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			for _, title := range titles {
				_, err := ParseTitle(title)
				if err != nil {
					b.Fatal(err)
				}
			}
		}
	})
}
