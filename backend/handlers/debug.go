package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

type DebugHandler struct {
	logger *log.Logger
}

type debugLogEntry struct {
	Level     string `json:"level"`
	Message   string `json:"message"`
	Timestamp string `json:"timestamp"`
}

type debugLogRequest struct {
	SessionID string          `json:"sessionId"`
	UserAgent string          `json:"userAgent"`
	Path      string          `json:"path"`
	Entries   []debugLogEntry `json:"entries"`
}

func NewDebugHandler(logger *log.Logger) *DebugHandler {
	h := &DebugHandler{logger: logger}
	if h.logger == nil {
		h.logger = log.New(os.Stdout, "", log.LstdFlags)
	}
	return h
}

func (h *DebugHandler) Capture(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var payload debugLogRequest
	decoder := json.NewDecoder(io.LimitReader(r.Body, 1<<20))
	if err := decoder.Decode(&payload); err != nil {
		http.Error(w, fmt.Sprintf("invalid payload: %v", err), http.StatusBadRequest)
		return
	}

	if len(payload.Entries) == 0 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]any{"status": "ignored", "reason": "no entries"})
		return
	}

	sessionID := strings.TrimSpace(payload.SessionID)
	if sessionID == "" {
		sessionID = fmt.Sprintf("anonymous-%d", time.Now().UnixNano())
	}

	ua := strings.TrimSpace(payload.UserAgent)
	sourcePath := strings.TrimSpace(payload.Path)
	remoteAddr := strings.TrimSpace(r.Header.Get("X-Forwarded-For"))
	if remoteAddr == "" {
		remoteAddr = strings.TrimSpace(r.RemoteAddr)
	}

	for _, entry := range payload.Entries {
		message := strings.TrimSpace(entry.Message)
		if message == "" {
			continue
		}
		level := strings.ToUpper(strings.TrimSpace(entry.Level))
		if level == "" {
			level = "LOG"
		}
		timestamp := strings.TrimSpace(entry.Timestamp)
		if timestamp == "" {
			timestamp = time.Now().UTC().Format(time.RFC3339)
		}

		logMessage := fmt.Sprintf(
			"[debug][session=%s][level=%s][remote=%s] ua=%q path=%q ts=%s message=%s",
			sessionID,
			level,
			remoteAddr,
			ua,
			sourcePath,
			timestamp,
			message,
		)
		h.logger.Println(logMessage)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]any{"status": "ok", "logged": len(payload.Entries)})
}
