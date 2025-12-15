package debrid

import "testing"

func TestResolveRestrictedLinkMatchesPreferredID(t *testing.T) {
	info := &TorrentInfo{
		Files: []File{
			{ID: 1, Selected: 1, Path: "file1.mkv"},
			{ID: 2, Selected: 0, Path: "file2.mkv"},
			{ID: 3, Selected: 1, Path: "file3.mkv"},
		},
		Links: []string{"link-1", "link-3"},
	}

	link, filename, idx, matched := resolveRestrictedLink(info, "3")
	if !matched {
		t.Fatalf("expected match for file id 3")
	}
	if idx != 1 {
		t.Fatalf("expected index 1, got %d", idx)
	}
	if link != "link-3" {
		t.Fatalf("expected link link-3, got %s", link)
	}
	if filename != "file3.mkv" {
		t.Fatalf("expected filename file3.mkv, got %s", filename)
	}
}

func TestResolveRestrictedLinkFallsBackWhenMissing(t *testing.T) {
	info := &TorrentInfo{
		Files: []File{
			{ID: 1, Selected: 1, Path: "file1.mkv"},
		},
		Links: []string{"link-1"},
	}

	link, filename, idx, matched := resolveRestrictedLink(info, "99")
	if matched {
		t.Fatalf("expected no match")
	}
	if idx != 0 || link != "link-1" {
		t.Fatalf("expected fallback to first link")
	}
	if filename != "file1.mkv" {
		t.Fatalf("expected filename file1.mkv, got %s", filename)
	}
}

func TestParseDebridPathWithFileID(t *testing.T) {
	provider, torrentID, fileID, err := parseDebridPath("/debrid/realdebrid/ABC/file/5")
	if err != nil {
		t.Fatalf("parseDebridPath returned error: %v", err)
	}
	if provider != "realdebrid" || torrentID != "ABC" || fileID != "5" {
		t.Fatalf("unexpected parse result: %s %s %s", provider, torrentID, fileID)
	}
}
