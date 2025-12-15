package debrid

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"novastream/config"
	"novastream/services/streaming"
)

// cachedURL represents a cached unrestricted URL with expiration.
type cachedURL struct {
	url       string
	filename  string
	expiresAt time.Time
}

// StreamingProvider implements streaming.Provider for debrid content.
type StreamingProvider struct {
	cfg      *config.Manager
	urlCache map[string]cachedURL
	cacheMux sync.RWMutex
	cacheTTL time.Duration
}

func parseDebridPath(path string) (provider, torrentID, fileID string, err error) {
	trimmed := strings.TrimSpace(path)
	if idx := strings.Index(trimmed, "?"); idx != -1 {
		trimmed = trimmed[:idx]
	}
	if idx := strings.Index(trimmed, "#"); idx != -1 {
		trimmed = trimmed[:idx]
	}
	if !strings.HasPrefix(trimmed, "/debrid/") {
		err = fmt.Errorf("invalid debrid path format: %s", path)
		return
	}

	segments := strings.Split(strings.TrimPrefix(trimmed, "/debrid/"), "/")
	if len(segments) < 2 {
		err = fmt.Errorf("invalid debrid path format: %s", path)
		return
	}

	provider = segments[0]
	torrentID = segments[1]

	if len(segments) >= 4 && segments[2] == "file" {
		fileID = segments[3]
	}

	return
}

func cacheKeyFor(torrentID, fileID string) string {
	if strings.TrimSpace(fileID) == "" {
		return torrentID
	}
	return fmt.Sprintf("%s:%s", torrentID, fileID)
}

// NewStreamingProvider creates a new debrid streaming provider.
func NewStreamingProvider(cfg *config.Manager) *StreamingProvider {
	return &StreamingProvider{
		cfg:      cfg,
		urlCache: make(map[string]cachedURL),
		cacheTTL: 10 * time.Minute, // Real-Debrid URLs are valid for longer, but 10 min is safe
	}
}

// getCachedURL retrieves a cached unrestricted URL if it exists and hasn't expired.
func (p *StreamingProvider) getCachedURL(cacheKey string) (url string, filename string, found bool) {
	p.cacheMux.RLock()
	defer p.cacheMux.RUnlock()

	cached, exists := p.urlCache[cacheKey]
	if !exists {
		return "", "", false
	}

	if time.Now().After(cached.expiresAt) {
		return "", "", false
	}

	return cached.url, cached.filename, true
}

// setCachedURL stores an unrestricted URL in the cache.
func (p *StreamingProvider) setCachedURL(cacheKey, url, filename string) {
	p.cacheMux.Lock()
	defer p.cacheMux.Unlock()

	p.urlCache[cacheKey] = cachedURL{
		url:       url,
		filename:  filename,
		expiresAt: time.Now().Add(p.cacheTTL),
	}

	// Clean up expired entries while we have the lock
	for id, cached := range p.urlCache {
		if time.Now().After(cached.expiresAt) {
			delete(p.urlCache, id)
		}
	}
}

// GetDirectURL returns the unrestricted HTTP download URL for the given debrid path.
// This URL can be used directly by FFmpeg for seekable input.
func (p *StreamingProvider) GetDirectURL(ctx context.Context, path string) (string, error) {
	provider, torrentID, fileID, err := parseDebridPath(path)
	if err != nil {
		return "", err
	}

	// Check cache first
	cacheKey := cacheKeyFor(torrentID, fileID)

	if cachedURL, _, found := p.getCachedURL(cacheKey); found {
		log.Printf("[debrid-stream] using cached URL for torrent %s file %s", torrentID, fileID)
		return cachedURL, nil
	}

	// Get provider config
	settings, err := p.cfg.Load()
	if err != nil {
		return "", fmt.Errorf("load settings: %w", err)
	}

	// Find the provider configuration
	var apiKey string
	for _, debridProvider := range settings.Streaming.DebridProviders {
		if strings.EqualFold(debridProvider.Provider, provider) && debridProvider.Enabled {
			apiKey = strings.TrimSpace(debridProvider.APIKey)
			break
		}
	}

	if apiKey == "" {
		return "", fmt.Errorf("provider %s not configured or not enabled", provider)
	}

	// Get provider from registry
	client, ok := GetProvider(provider, apiKey)
	if !ok {
		return "", fmt.Errorf("provider %s not registered", provider)
	}

	info, err := client.GetTorrentInfo(ctx, torrentID)
	if err != nil {
		return "", fmt.Errorf("get torrent info: %w", err)
	}

	restrictedLink, filename, _, matched := resolveRestrictedLink(info, fileID)
	if restrictedLink == "" {
		return "", fmt.Errorf("no download links available for torrent %s", torrentID)
	}
	if fileID != "" && !matched {
		log.Printf("[debrid-stream] requested file id %s not found for torrent %s; defaulting to first link", fileID, torrentID)
	}
	if filename != "" {
		log.Printf("[debrid-stream] resolved filename: %s", filename)
	}

	unrestricted, err := client.UnrestrictLink(ctx, restrictedLink)
	if err != nil {
		return "", fmt.Errorf("unrestrict link: %w", err)
	}

	downloadURL := unrestricted.DownloadURL
	if downloadURL == "" {
		return "", fmt.Errorf("no download URL returned from provider")
	}

	// Cache the URL and filename for future requests
	p.setCachedURL(cacheKey, downloadURL, filename)

	log.Printf("[debrid-stream] resolved direct URL for path %s: %s (filename: %s)", path, downloadURL, filename)
	return downloadURL, nil
}

// Stream handles /debrid/ paths by proxying to Real-Debrid download links.
func (p *StreamingProvider) Stream(ctx context.Context, req streaming.Request) (*streaming.Response, error) {
	// Only handle /debrid/ paths
	// Normalize path by removing leading slashes and webdav prefix
	cleanPath := strings.TrimPrefix(req.Path, "/")
	cleanPath = strings.TrimPrefix(cleanPath, "webdav/")

	if !strings.HasPrefix(cleanPath, "debrid/") {
		return nil, streaming.ErrNotFound
	}

	provider, torrentID, fileID, err := parseDebridPath("/" + cleanPath)
	if err != nil {
		return nil, err
	}

	log.Printf("[debrid-stream] streaming request: provider=%s torrentID=%s fileID=%s method=%s range=%q",
		provider, torrentID, fileID, req.Method, req.RangeHeader)

	// Get provider config
	settings, err := p.cfg.Load()
	if err != nil {
		return nil, fmt.Errorf("load settings: %w", err)
	}

	var providerConfig *config.DebridProviderSettings
	for i := range settings.Streaming.DebridProviders {
		pc := &settings.Streaming.DebridProviders[i]
		if !pc.Enabled {
			continue
		}
		if strings.EqualFold(pc.Provider, provider) {
			providerConfig = pc
			break
		}
	}

	if providerConfig == nil {
		return nil, fmt.Errorf("provider %q not configured", provider)
	}

	// Get provider from registry
	client, ok := GetProvider(strings.ToLower(providerConfig.Provider), providerConfig.APIKey)
	if !ok {
		return nil, fmt.Errorf("provider %q not registered", providerConfig.Provider)
	}

	return p.streamWithProvider(ctx, req, client, torrentID, fileID)
}

func (p *StreamingProvider) streamWithProvider(ctx context.Context, req streaming.Request, client Provider, torrentID, fileID string) (*streaming.Response, error) {
	providerName := client.Name()
	var downloadURL string
	var filename string

	cacheKey := cacheKeyFor(torrentID, fileID)

	// Check cache first
	if cachedURL, cachedFilename, found := p.getCachedURL(cacheKey); found {
		log.Printf("[debrid-stream] using cached URL for torrent %s file %s", torrentID, fileID)
		downloadURL = cachedURL
		filename = cachedFilename
	} else {
		// Cache miss - need to unrestrict the link
		// Get fresh torrent info to get download links
		info, err := client.GetTorrentInfo(ctx, torrentID)
		if err != nil {
			return nil, fmt.Errorf("get torrent info: %w", err)
		}

		restrictedLink, resolvedFilename, _, matched := resolveRestrictedLink(info, fileID)
		if restrictedLink == "" {
			return nil, fmt.Errorf("no download links available for torrent %s", torrentID)
		}
		if fileID != "" && !matched {
			log.Printf("[debrid-stream] requested file id %s not available for torrent %s; defaulting to first link", fileID, torrentID)
		}
		if resolvedFilename != "" {
			log.Printf("[debrid-stream] resolved filename: %s", resolvedFilename)
		}

		log.Printf("[debrid-stream] unrestricting link: %s", restrictedLink)

		// Unrestrict the link to get the actual download URL
		unrestricted, err := client.UnrestrictLink(ctx, restrictedLink)
		if err != nil {
			return nil, fmt.Errorf("unrestrict link: %w", err)
		}

		// Use the direct download URL
		downloadURL = unrestricted.DownloadURL

		// Use filename from provider if available, otherwise use the resolved filename from torrent info
		if unrestricted.Filename != "" {
			filename = unrestricted.Filename
		} else {
			filename = resolvedFilename
		}

		// Cache the URL and filename for future requests
		p.setCachedURL(cacheKey, downloadURL, filename)

		log.Printf("[debrid-stream] proxying to unrestricted URL: %s", downloadURL)
	}

	// Create HTTP request to the provider
	httpReq, err := http.NewRequestWithContext(ctx, req.Method, downloadURL, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	// Forward range header if present
	if req.RangeHeader != "" {
		httpReq.Header.Set("Range", req.RangeHeader)
	}

	// Make the request
	httpClient := &http.Client{
		Timeout: 30 * time.Minute,
	}

	resp, err := httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	// For failed requests, close and return error
	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
		resp.Body.Close()
		return nil, fmt.Errorf("%s request failed: %s: %s", providerName, resp.Status, string(body))
	}

	log.Printf("[debrid-stream] %s response: status=%d content-length=%d range=%q",
		providerName, resp.StatusCode, resp.ContentLength, resp.Header.Get("Content-Range"))

	// Build streaming response
	headers := make(http.Header)
	for key, values := range resp.Header {
		for _, value := range values {
			headers.Add(key, value)
		}
	}

	// Ensure Accept-Ranges header is set
	if headers.Get("Accept-Ranges") == "" {
		headers.Set("Accept-Ranges", "bytes")
	}

	// Handle HEAD requests - close body immediately
	if req.Method == http.MethodHead {
		resp.Body.Close()
		return &streaming.Response{
			Status:        resp.StatusCode,
			Headers:       headers,
			ContentLength: resp.ContentLength,
			Body:          io.NopCloser(strings.NewReader("")),
			Filename:      filename,
		}, nil
	}

	return &streaming.Response{
		Status:        resp.StatusCode,
		Headers:       headers,
		ContentLength: resp.ContentLength,
		Body:          resp.Body,
		Filename:      filename,
	}, nil
}

// CompositeProvider combines multiple streaming providers.
type CompositeProvider struct {
	providers []streaming.Provider
}

// NewCompositeProvider creates a new composite provider.
func NewCompositeProvider(providers ...streaming.Provider) *CompositeProvider {
	return &CompositeProvider{providers: providers}
}

// GetDirectURL tries to get a direct URL from any provider that supports it.
func (c *CompositeProvider) GetDirectURL(ctx context.Context, path string) (string, error) {
	for _, provider := range c.providers {
		if provider == nil {
			continue
		}

		// Check if this provider supports direct URLs
		directProvider, ok := provider.(streaming.DirectURLProvider)
		if !ok {
			continue
		}

		url, err := directProvider.GetDirectURL(ctx, path)
		if err == nil && url != "" {
			return url, nil
		}

		// If not found, try next provider
		if err == streaming.ErrNotFound {
			continue
		}

		// For other errors, continue trying other providers
	}

	return "", streaming.ErrNotFound
}

// Stream tries each provider in order until one handles the request.
func (c *CompositeProvider) Stream(ctx context.Context, req streaming.Request) (*streaming.Response, error) {
	for _, provider := range c.providers {
		if provider == nil {
			continue
		}

		resp, err := provider.Stream(ctx, req)
		if err == nil {
			return resp, nil
		}

		// If not found, try next provider
		if err == streaming.ErrNotFound {
			continue
		}

		// For other errors, return immediately
		return nil, err
	}

	return nil, streaming.ErrNotFound
}
