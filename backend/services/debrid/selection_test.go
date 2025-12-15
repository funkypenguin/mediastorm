package debrid

import (
	"testing"

	"novastream/internal/mediaresolve"
)

func TestSelectMediaFilesPrefersEpisodeMatch(t *testing.T) {
	files := []File{
		{ID: 1, Path: "Series.S01E01.1080p.mkv"},
		{ID: 2, Path: "Series.S01E02.1080p.mkv"},
	}

	selection := selectMediaFiles(files, mediaresolve.SelectionHints{
		ReleaseTitle: "Series S01E02 1080p WEB-DL",
	})

	if selection == nil {
		t.Fatalf("expected selection, got nil")
	}
	if selection.PreferredID != "2" {
		t.Fatalf("expected preferred ID 2, got %s", selection.PreferredID)
	}
	if selection.OrderedIDs[0] != "2" {
		t.Fatalf("expected preferred ID to be first in ordered list, got %v", selection.OrderedIDs)
	}
	if selection.PreferredReason == "" {
		t.Fatalf("expected non-empty reason")
	}
}

func TestSelectMediaFilesPrefersTitleSimilarity(t *testing.T) {
	files := []File{
		{ID: 10, Path: "Sample.Extras.mkv"},
		{ID: 11, Path: "Movie.Title.2023.2160p.WEB-DL.x265.mkv"},
	}

	selection := selectMediaFiles(files, mediaresolve.SelectionHints{
		ReleaseTitle: "Movie.Title.2023.2160p.WEB-DL.x265-GROUP",
	})

	if selection == nil {
		t.Fatalf("expected selection, got nil")
	}
	if selection.PreferredID != "11" {
		t.Fatalf("expected preferred ID 11, got %s", selection.PreferredID)
	}
	if selection.OrderedIDs[0] != "11" {
		t.Fatalf("expected preferred ID to be first in ordered list, got %v", selection.OrderedIDs)
	}
	if selection.PreferredReason == "" {
		t.Fatalf("expected non-empty reason")
	}
}

func TestSelectMediaFilesUsesExplicitTargetEpisode(t *testing.T) {
	files := []File{
		{ID: 20, Path: "Show.S01E05.1080p.mkv"},
		{ID: 21, Path: "Show.S02E05.1080p.mkv"},
	}

	selection := selectMediaFiles(files, mediaresolve.SelectionHints{
		TargetSeason:  1,
		TargetEpisode: 5,
	})

	if selection == nil {
		t.Fatalf("expected selection, got nil")
	}
	if selection.PreferredID != "20" {
		t.Fatalf("expected preferred ID 20, got %s", selection.PreferredID)
	}
	if selection.PreferredReason == "" {
		t.Fatalf("expected reason to mention explicit target")
	}
}
