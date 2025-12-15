package debrid

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"path"
	"strings"

	"novastream/config"
	"novastream/models"
)

// PlaybackService handles debrid playback resolution.
type PlaybackService struct {
	cfg           *config.Manager
	healthService *HealthService
}

// NewPlaybackService creates a new debrid playback service.
func NewPlaybackService(cfg *config.Manager, healthService *HealthService) *PlaybackService {
	if healthService == nil {
		healthService = NewHealthService(cfg)
	}
	return &PlaybackService{
		cfg:           cfg,
		healthService: healthService,
	}
}

// Resolve checks if a debrid item is cached and returns playback information.
// For debrid, we add the torrent, select files, and get the download link.
func (s *PlaybackService) Resolve(ctx context.Context, candidate models.NZBResult) (*models.PlaybackResolution, error) {
	log.Printf("[debrid-playback] resolve start title=%q link=%q", strings.TrimSpace(candidate.Title), strings.TrimSpace(candidate.Link))

	// Extract info hash from candidate
	infoHash := strings.TrimSpace(candidate.Attributes["infoHash"])
	if infoHash == "" {
		if strings.HasPrefix(strings.ToLower(candidate.Link), "magnet:") {
			infoHash = extractInfoHashFromMagnet(candidate.Link)
		}
		if infoHash == "" {
			return nil, fmt.Errorf("missing info hash")
		}
	}

	// Get provider config
	settings, err := s.cfg.Load()
	if err != nil {
		return nil, fmt.Errorf("load settings: %w", err)
	}

	// Determine provider - use attribute if specified, otherwise use first enabled provider
	provider := strings.TrimSpace(candidate.Attributes["provider"])

	var providerConfig *config.DebridProviderSettings
	for i := range settings.Streaming.DebridProviders {
		p := &settings.Streaming.DebridProviders[i]
		if !p.Enabled {
			continue
		}
		// If provider specified, match it; otherwise use first enabled
		if provider == "" || strings.EqualFold(p.Provider, provider) {
			providerConfig = p
			break
		}
	}

	if providerConfig == nil {
		if provider == "" {
			return nil, fmt.Errorf("no debrid provider configured or enabled")
		}
		return nil, fmt.Errorf("provider %q not configured or not enabled", provider)
	}

	// Get provider from registry
	client, ok := GetProvider(strings.ToLower(providerConfig.Provider), providerConfig.APIKey)
	if !ok {
		return nil, fmt.Errorf("provider %q not registered", providerConfig.Provider)
	}

	return s.resolveWithProvider(ctx, client, candidate, infoHash)
}

func (s *PlaybackService) resolveWithProvider(ctx context.Context, client Provider, candidate models.NZBResult, infoHash string) (*models.PlaybackResolution, error) {
	providerName := client.Name()

	// Add the magnet to the provider
	log.Printf("[debrid-playback] adding magnet to %s", providerName)
	addResp, err := client.AddMagnet(ctx, candidate.Link)
	if err != nil {
		return nil, fmt.Errorf("add magnet: %w", err)
	}

	torrentID := addResp.ID
	log.Printf("[debrid-playback] torrent added with ID %s", torrentID)

	// Get torrent info to see available files
	info, err := client.GetTorrentInfo(ctx, torrentID)
	if err != nil {
		return nil, fmt.Errorf("get torrent info: %w", err)
	}

	// Select the most relevant media file (but send all files to trigger caching)
	selection := selectMediaFiles(info.Files, buildSelectionHints(candidate, info.Filename))
	if selection == nil || len(selection.OrderedIDs) == 0 {
		_ = client.DeleteTorrent(ctx, torrentID)
		return nil, fmt.Errorf("no media files found in torrent")
	}

	if selection.PreferredID != "" {
		log.Printf("[debrid-playback] primary file candidate: %q (reason: %s, id=%s)", selection.PreferredLabel, selection.PreferredReason, selection.PreferredID)
	}

	fileSelection := strings.Join(selection.OrderedIDs, ",")
	log.Printf("[debrid-playback] selecting %d media files for caching: %s", len(selection.OrderedIDs), fileSelection)
	logSelectedFileDetails(info.Files, selection)

	if err := client.SelectFiles(ctx, torrentID, fileSelection); err != nil {
		_ = client.DeleteTorrent(ctx, torrentID)
		return nil, fmt.Errorf("select files: %w", err)
	}

	// Get torrent info again to get download links
	info, err = client.GetTorrentInfo(ctx, torrentID)
	if err != nil {
		_ = client.DeleteTorrent(ctx, torrentID)
		return nil, fmt.Errorf("get torrent info after selection: %w", err)
	}

	// Check if cached
	isCached := strings.ToLower(info.Status) == "downloaded"
	log.Printf("[debrid-playback] torrent %s status=%s cached=%t links=%d", torrentID, info.Status, isCached, len(info.Links))

	if !isCached {
		// Torrent is not cached - it may be downloading. We must remove it from the account
		// to avoid leaving orphaned downloads (especially important for Torbox).
		log.Printf("[debrid-playback] torrent %s is not cached (status=%s), removing from %s account", torrentID, info.Status, providerName)
		if err := client.DeleteTorrent(ctx, torrentID); err != nil {
			log.Printf("[debrid-playback] warning: failed to delete non-cached torrent %s: %v", torrentID, err)
		}
		return nil, fmt.Errorf("torrent not cached (status: %s)", info.Status)
	}

	if len(info.Links) == 0 {
		_ = client.DeleteTorrent(ctx, torrentID)
		return nil, fmt.Errorf("no download links available")
	}

	restrictedLink, filename, preferredLinkIdx, matched := resolveRestrictedLink(info, selection.PreferredID)
	if !matched && selection.PreferredID != "" {
		log.Printf("[debrid-playback] preferred file id %s not found among %s links; defaulting to index %d", selection.PreferredID, providerName, preferredLinkIdx)
	}
	if filename != "" {
		log.Printf("[debrid-playback] resolved filename: %s", filename)
	}

	downloadURL := restrictedLink
	if selection.PreferredLabel != "" {
		log.Printf("[debrid-playback] using download link #%d for %q (reason: %s)", preferredLinkIdx, selection.PreferredLabel, selection.PreferredReason)
	} else {
		log.Printf("[debrid-playback] using download link #%d for selected file (id=%s)", preferredLinkIdx, selection.PreferredID)
	}

	// Keep torrent in provider for playback
	// Note: We don't delete the torrent here because we need it for streaming
	log.Printf("[debrid-playback] keeping torrent %s in %s for playback", torrentID, providerName)

	// Return webdavPath as a path that the streaming provider can recognize
	// Format: /debrid/{provider}/TORRENT_ID[/file/ID][/FILENAME]
	// This works with both web (/api/video/stream?path=...) and mobile (direct URL)
	// We append the filename so it can be displayed in the player UI
	webdavPath := fmt.Sprintf("/debrid/%s/%s", providerName, torrentID)
	if selection.PreferredID != "" {
		webdavPath = fmt.Sprintf("%s/file/%s", webdavPath, selection.PreferredID)
	}
	// Append filename for display purposes (will be ignored by streaming provider)
	if filename != "" {
		webdavPath = fmt.Sprintf("%s/%s", webdavPath, filename)
	}

	// If the link is an actual URL (not an internal reference like torrent_id:file_id),
	// verify it's accessible and check for archives
	isActualURL := strings.HasPrefix(downloadURL, "http://") || strings.HasPrefix(downloadURL, "https://")

	if isActualURL {
		// Check for unsupported archives
		if archiveExt := detectArchiveExtension(downloadURL); archiveExt != "" {
			_ = client.DeleteTorrent(ctx, torrentID)
			return nil, fmt.Errorf("download URL points to unsupported archive (%s)", archiveExt)
		}

		// Verify the download URL is accessible with a HEAD request
		log.Printf("[debrid-playback] verifying download URL is accessible: %s", downloadURL)
		headReq, err := http.NewRequestWithContext(ctx, http.MethodHead, downloadURL, nil)
		if err != nil {
			_ = client.DeleteTorrent(ctx, torrentID)
			return nil, fmt.Errorf("failed to create HEAD request: %w", err)
		}

		headResp, err := http.DefaultClient.Do(headReq)
		if err != nil {
			_ = client.DeleteTorrent(ctx, torrentID)
			return nil, fmt.Errorf("download URL not accessible: %w", err)
		}
		defer headResp.Body.Close()

		if headResp.StatusCode >= 400 {
			_ = client.DeleteTorrent(ctx, torrentID)
			return nil, fmt.Errorf("download URL returned error status: %d %s", headResp.StatusCode, headResp.Status)
		}

		log.Printf("[debrid-playback] download URL verified accessible (status: %d)", headResp.StatusCode)
	} else {
		// For providers like Torbox that use internal references (torrent_id:file_id),
		// the actual URL is resolved at stream time via UnrestrictLink
		log.Printf("[debrid-playback] download link is internal reference, will be resolved at stream time: %s", downloadURL)
	}

	resolution := &models.PlaybackResolution{
		QueueID:       0, // Debrid doesn't use queues
		WebDAVPath:    webdavPath,
		HealthStatus:  "cached",
		FileSize:      candidate.SizeBytes,
		SourceNZBPath: downloadURL, // Store the actual download URL here
	}

	log.Printf("[debrid-playback] resolution successful: webdavPath=%s downloadURL=%s", webdavPath, downloadURL)
	return resolution, nil
}

func detectArchiveExtension(downloadURL string) string {
	if strings.TrimSpace(downloadURL) == "" {
		return ""
	}
	parsed, err := url.Parse(downloadURL)
	if err != nil {
		return ""
	}
	ext := strings.ToLower(path.Ext(parsed.Path))
	switch ext {
	case ".rar", ".zip", ".7z", ".tar", ".tar.gz", ".tgz":
		return ext
	default:
		return ""
	}
}

func logSelectedFileDetails(files []File, selection *mediaFileSelection) {
	if selection == nil {
		log.Printf("[debrid-playback] no media file selection available to log")
		return
	}

	if len(selection.OrderedIDs) == 0 {
		log.Printf("[debrid-playback] selection contained zero file IDs")
		return
	}

	fileLookup := make(map[string]File, len(files))
	for _, file := range files {
		fileLookup[fmt.Sprintf("%d", file.ID)] = file
	}

	log.Printf("[debrid-playback] detailed selected files (preferred id=%s):", selection.PreferredID)
	for idx, id := range selection.OrderedIDs {
		file, ok := fileLookup[id]
		preferred := selection.PreferredID == id
		if !ok {
			log.Printf("[debrid-playback]   #%d id=%s preferred=%t (details unavailable from provider)", idx+1, id, preferred)
			continue
		}

		sizeMB := float64(file.Bytes) / (1024 * 1024)
		log.Printf(
			"[debrid-playback]   #%d id=%s preferred=%t path=%q size=%d bytes (~%.2f MB) selected=%t",
			idx+1,
			id,
			preferred,
			file.Path,
			file.Bytes,
			sizeMB,
			file.Selected == 1,
		)
	}
}

// CheckHealthQuick performs a quick cached check without adding/removing torrents.
// This is useful for filtering search results or auto-selection.
func (s *PlaybackService) CheckHealthQuick(ctx context.Context, candidate models.NZBResult) (*DebridHealthCheck, error) {
	// Quick check - don't verify by adding
	return s.healthService.CheckHealth(ctx, candidate, false)
}

// FilterCachedResults filters a list of results to only include cached debrid items.
// This is useful for auto-selection or pre-filtering search results.
// Only checks the first 3 results to minimize API calls.
func (s *PlaybackService) FilterCachedResults(ctx context.Context, results []models.NZBResult) []models.NZBResult {
	var cached []models.NZBResult

	log.Printf("[debrid-playback] filtering %d results for cached items (checking first 3 only)", len(results))

	checked := 0
	for i, result := range results {
		// Only check debrid items
		if result.ServiceType != models.ServiceTypeDebrid {
			log.Printf("[debrid-playback] [%d/%d] skipping non-debrid result: %s", i+1, len(results), result.Title)
			continue
		}

		// Only check first 3 debrid results to minimize API calls
		if checked >= 3 {
			log.Printf("[debrid-playback] reached limit of 3 health checks, skipping remaining results")
			break
		}
		checked++

		health, err := s.CheckHealthQuick(ctx, result)
		if err != nil {
			log.Printf("[debrid-playback] [%d/%d] health check failed for %s: %v", i+1, len(results), result.Title, err)
			continue
		}

		if health == nil {
			log.Printf("[debrid-playback] [%d/%d] health check returned nil for %s", i+1, len(results), result.Title)
			continue
		}

		if health.Status == "error" && health.ErrorMessage != "" {
			log.Printf("[debrid-playback] [%d/%d] %s: healthy=%t cached=%t status=%s error=%q",
				i+1, len(results), result.Title, health.Healthy, health.Cached, health.Status, health.ErrorMessage)
		} else {
			log.Printf("[debrid-playback] [%d/%d] %s: healthy=%t cached=%t status=%s",
				i+1, len(results), result.Title, health.Healthy, health.Cached, health.Status)
		}

		if health.Healthy && health.Cached {
			cached = append(cached, result)
		}
	}

	log.Printf("[debrid-playback] filtered results: %d cached out of %d checked", len(cached), checked)
	return cached
}
