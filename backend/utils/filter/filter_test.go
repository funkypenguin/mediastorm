package filter

import (
	"testing"

	"novastream/models"
)

func TestResults_MovieFiltering(t *testing.T) {
	results := []models.NZBResult{
		{Title: "The.Matrix.1999.1080p.BluRay.x264-SPARKS"},        // Should match
		{Title: "The.Matrix.Reloaded.2003.1080p.BluRay.x264"},      // Wrong year, should be filtered
		{Title: "Inception.2010.1080p.BluRay.x264"},                // Wrong title, should be filtered
		{Title: "The.Matrix.1999.720p.WEB-DL.x264"},                // Should match
		{Title: "The.Matrix.2000.1080p.BluRay.x264"},               // Year within ±1, should match
		{Title: "The.Matrix.Resurrections.2021.1080p.BluRay.x264"}, // Wrong movie, should be filtered
	}

	opts := Options{
		ExpectedTitle: "The Matrix",
		ExpectedYear:  1999,
		IsMovie:       true,
	}

	filtered := Results(results, opts)

	// Should keep results at indices 0, 3, 4 (The Matrix 1999, 1999, 2000)
	expectedCount := 3
	if len(filtered) != expectedCount {
		t.Errorf("Expected %d results, got %d", expectedCount, len(filtered))
		for i, r := range filtered {
			t.Logf("  Result[%d]: %s", i, r.Title)
		}
	}

	// Verify we got the right ones
	expectedTitles := map[string]bool{
		"The.Matrix.1999.1080p.BluRay.x264-SPARKS": true,
		"The.Matrix.1999.720p.WEB-DL.x264":         true,
		"The.Matrix.2000.1080p.BluRay.x264":        true,
	}

	for _, r := range filtered {
		if !expectedTitles[r.Title] {
			t.Errorf("Unexpected result in filtered list: %s", r.Title)
		}
	}
}

func TestResults_TVShowFiltering(t *testing.T) {
	results := []models.NZBResult{
		{Title: "The.Simpsons.S01E01.1080p.BluRay.x265"},     // Should match
		{Title: "The.Simpsons.S01E02.720p.WEB-DL.x264"},      // Should match
		{Title: "Family.Guy.S01E01.1080p.BluRay.x264"},       // Wrong show, should be filtered
		{Title: "The.Simpsons.Movie.2007.1080p.BluRay.x264"}, // Will be filtered (title not similar enough - 66%)
	}

	opts := Options{
		ExpectedTitle: "The Simpsons",
		ExpectedYear:  0, // Year doesn't matter for TV shows
		IsMovie:       false,
	}

	filtered := Results(results, opts)

	// Should keep results at indices 0, 1 (The Simpsons episodes only)
	// Note: "The Simpsons Movie" is only 66.67% similar, below the 90% threshold
	expectedCount := 2
	if len(filtered) != expectedCount {
		t.Errorf("Expected %d results, got %d", expectedCount, len(filtered))
		for i, r := range filtered {
			t.Logf("  Result[%d]: %s", i, r.Title)
		}
	}
}

func TestResults_NoFiltering(t *testing.T) {
	results := []models.NZBResult{
		{Title: "Some.Random.Release.1080p.BluRay.x264"},
		{Title: "Another.Release.720p.WEB-DL.x264"},
	}

	opts := Options{
		ExpectedTitle: "",
		ExpectedYear:  0,
		IsMovie:       true,
	}

	// With no expected title, all results should be kept (parse errors)
	filtered := Results(results, opts)

	if len(filtered) != len(results) {
		t.Errorf("Expected all %d results to be kept, got %d", len(results), len(filtered))
	}
}

func TestResults_AlternateTitles(t *testing.T) {
	results := []models.NZBResult{
		{Title: "La.Casa.de.Papel.S01E01.1080p.NF.WEB-DL.x265"},
		{Title: "Random.Show.S01E01.1080p"},
	}

	opts := Options{
		ExpectedTitle:   "Money Heist",
		ExpectedYear:    0,
		IsMovie:         false,
		AlternateTitles: []string{"La Casa de Papel"},
	}

	filtered := Results(results, opts)
	if len(filtered) != 1 {
		t.Fatalf("expected 1 result to match alternate title, got %d", len(filtered))
	}
	if filtered[0].Title != results[0].Title {
		t.Fatalf("expected alternate title match for %q", results[0].Title)
	}
}

func TestResults_JapaneseRomanization(t *testing.T) {
	results := []models.NZBResult{
		{Title: "Ikusagami.S01E01.1080p.WEB-DL"},
		{Title: "Completely.Different.S01E01"},
	}

	opts := Options{
		ExpectedTitle:   "الساموراي الصامد الأخير",
		AlternateTitles: []string{"イクサガミ"},
	}

	filtered := Results(results, opts)
	if len(filtered) != 1 {
		t.Fatalf("expected romanized alternate to match, got %d results", len(filtered))
	}
	if filtered[0].Title != results[0].Title {
		t.Fatalf("expected Ikusagami release to survive filtering")
	}
}

func TestShouldFilter(t *testing.T) {
	tests := []struct {
		title    string
		expected bool
	}{
		{"The Matrix", true},
		{"", false},
		{"  ", false},
		{"The Simpsons S01E01", true},
	}

	for _, tt := range tests {
		result := ShouldFilter(tt.title)
		if result != tt.expected {
			t.Errorf("ShouldFilter(%q) = %v, want %v", tt.title, result, tt.expected)
		}
	}
}
