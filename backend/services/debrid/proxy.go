package debrid

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"novastream/config"
	"novastream/services/streaming"
)

// ProxyRequest encapsulates a debrid proxy request from the handler layer.
type ProxyRequest struct {
	Provider    string
	ResourceURL string
	Method      string
	RangeHeader string
}

// ProxyService forwards streaming requests to the configured debrid provider.
type ProxyService struct {
	cfg        *config.Manager
	httpClient *http.Client
}

// NewProxyService constructs a new proxy service with a default HTTP client.
func NewProxyService(cfg *config.Manager) *ProxyService {
	return &ProxyService{
		cfg:        cfg,
		httpClient: &http.Client{Timeout: 5 * time.Minute},
	}
}

// Proxy performs a basic authenticated passthrough to a debrid provider.
// This is a minimal implementation intended to be expanded with provider-specific logic.
func (s *ProxyService) Proxy(ctx context.Context, req ProxyRequest) (*streaming.Response, error) {
	if s == nil || s.cfg == nil {
		return nil, fmt.Errorf("debrid proxy not configured")
	}
	trimmedURL := strings.TrimSpace(req.ResourceURL)
	if trimmedURL == "" {
		return nil, fmt.Errorf("missing resource URL for debrid proxy")
	}

	settings, err := s.cfg.Load()
	if err != nil {
		return nil, fmt.Errorf("load settings: %w", err)
	}

	providerConfig, err := findProvider(settings.Streaming.DebridProviders, req.Provider)
	if err != nil {
		return nil, err
	}

	method := strings.ToUpper(strings.TrimSpace(req.Method))
	if method == "" {
		method = http.MethodGet
	}
	if method != http.MethodGet && method != http.MethodHead {
		return nil, fmt.Errorf("unsupported proxy method %q", method)
	}

	httpReq, err := http.NewRequestWithContext(ctx, method, trimmedURL, nil)
	if err != nil {
		return nil, fmt.Errorf("build debrid proxy request: %w", err)
	}

	if req.RangeHeader != "" {
		httpReq.Header.Set("Range", req.RangeHeader)
	}

	apiKey := strings.TrimSpace(providerConfig.APIKey)
	if apiKey != "" {
		// Use a generic bearer token. Provider-specific adapters can refine this later.
		httpReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", apiKey))
	}

	resp, err := s.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("debrid proxy request failed: %w", err)
	}

	response := &streaming.Response{
		Body:          resp.Body,
		Headers:       resp.Header.Clone(),
		Status:        resp.StatusCode,
		ContentLength: resp.ContentLength,
	}

	if response.Headers == nil {
		response.Headers = http.Header{}
	}

	return response, nil
}

func findProvider(providers []config.DebridProviderSettings, name string) (config.DebridProviderSettings, error) {
	normalized := strings.TrimSpace(strings.ToLower(name))
	for _, provider := range providers {
		if !provider.Enabled {
			continue
		}
		if normalized == "" {
			return provider, nil
		}
		if strings.ToLower(strings.TrimSpace(provider.Name)) == normalized {
			return provider, nil
		}
		if strings.ToLower(strings.TrimSpace(provider.Provider)) == normalized {
			return provider, nil
		}
	}
	if normalized == "" {
		return config.DebridProviderSettings{}, fmt.Errorf("no enabled debrid providers configured")
	}
	return config.DebridProviderSettings{}, fmt.Errorf("debrid provider %q not configured", name)
}
