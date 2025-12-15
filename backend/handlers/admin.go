package handlers

import (
	"encoding/json"
	"net/http"
	"path/filepath"
	"time"
)

// AdminHandler provides administrative endpoints for monitoring the server
type AdminHandler struct {
	hlsManager *HLSManager
}

// NewAdminHandler creates a new admin handler
func NewAdminHandler(hlsManager *HLSManager) *AdminHandler {
	return &AdminHandler{
		hlsManager: hlsManager,
	}
}

// StreamInfo represents information about an active stream
type StreamInfo struct {
	ID            string    `json:"id"`
	Type          string    `json:"type"` // "hls", "direct", or "debrid"
	Path          string    `json:"path"`
	OriginalPath  string    `json:"original_path,omitempty"`
	Filename      string    `json:"filename"`
	ClientIP      string    `json:"client_ip,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
	LastAccess    time.Time `json:"last_access"`
	Duration      float64   `json:"duration,omitempty"`
	BytesStreamed int64     `json:"bytes_streamed"`
	ContentLength int64     `json:"content_length,omitempty"`
	HasDV         bool      `json:"has_dv"`
	HasHDR        bool      `json:"has_hdr"`
	DVProfile     string    `json:"dv_profile,omitempty"`
	Segments      int       `json:"segments,omitempty"`
	UserAgent     string    `json:"user_agent,omitempty"`
}

// StreamsResponse is the response for the streams endpoint
type StreamsResponse struct {
	Streams []StreamInfo `json:"streams"`
	Count   int          `json:"count"`
	HLS     int          `json:"hls_count"`
	Direct  int          `json:"direct_count"`
}

// GetActiveStreams returns all active streams (both HLS and direct)
func (h *AdminHandler) GetActiveStreams(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	response := StreamsResponse{
		Streams: []StreamInfo{},
		Count:   0,
		HLS:     0,
		Direct:  0,
	}

	// Get HLS sessions
	if h.hlsManager != nil {
		h.hlsManager.mu.RLock()
		for _, session := range h.hlsManager.sessions {
			session.mu.RLock()

			// Extract filename from path
			filename := filepath.Base(session.Path)
			if filename == "" || filename == "." {
				filename = filepath.Base(session.OriginalPath)
			}

			info := StreamInfo{
				ID:            session.ID,
				Type:          "hls",
				Path:          session.Path,
				OriginalPath:  session.OriginalPath,
				Filename:      filename,
				CreatedAt:     session.CreatedAt,
				LastAccess:    session.LastAccess,
				Duration:      session.Duration,
				BytesStreamed: session.BytesStreamed,
				HasDV:         session.HasDV && !session.DVDisabled,
				HasHDR:        session.HasHDR,
				DVProfile:     session.DVProfile,
				Segments:      session.SegmentsCreated,
			}

			session.mu.RUnlock()
			response.Streams = append(response.Streams, info)
			response.HLS++
		}
		h.hlsManager.mu.RUnlock()
	}

	// Get direct streams from the global tracker
	tracker := GetStreamTracker()
	for _, stream := range tracker.GetActiveStreams() {
		info := StreamInfo{
			ID:            stream.ID,
			Type:          "direct",
			Path:          stream.Path,
			Filename:      stream.Filename,
			ClientIP:      stream.ClientIP,
			CreatedAt:     stream.StartTime,
			LastAccess:    stream.LastActivity,
			BytesStreamed: stream.BytesStreamed,
			ContentLength: stream.ContentLength,
			UserAgent:     stream.UserAgent,
		}
		response.Streams = append(response.Streams, info)
		response.Direct++
	}

	response.Count = len(response.Streams)
	json.NewEncoder(w).Encode(response)
}
