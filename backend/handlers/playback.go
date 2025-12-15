package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"novastream/models"
	playbacksvc "novastream/services/playback"
)

type playbackService interface {
	Resolve(ctx context.Context, candidate models.NZBResult) (*models.PlaybackResolution, error)
	QueueStatus(ctx context.Context, queueID int64) (*models.PlaybackResolution, error)
}

// PlaybackHandler resolves NZB candidates into playable streams via the local registry.
type PlaybackHandler struct {
	Service playbackService
}

var _ playbackService = (*playbacksvc.Service)(nil)

func NewPlaybackHandler(s playbackService) *PlaybackHandler {
	return &PlaybackHandler{Service: s}
}

// Resolve accepts an NZB indexer result and responds with a validated playback source.
func (h *PlaybackHandler) Resolve(w http.ResponseWriter, r *http.Request) {
	var request struct {
		Result models.NZBResult `json:"result"`
	}

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&request); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	log.Printf("[playback-handler] Received resolve request: Title=%q, GUID=%q, ServiceType=%q, titleId=%q, titleName=%q",
		request.Result.Title, request.Result.GUID, request.Result.ServiceType,
		request.Result.Attributes["titleId"], request.Result.Attributes["titleName"])

	resolution, err := h.Service.Resolve(r.Context(), request.Result)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resolution)
}

// QueueStatus reports the current resolution status for a previously queued playback request.
func (h *PlaybackHandler) QueueStatus(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	queueIDStr := vars["queueID"]
	queueID, err := strconv.ParseInt(queueIDStr, 10, 64)
	if err != nil || queueID <= 0 {
		http.Error(w, "invalid queue id", http.StatusBadRequest)
		return
	}

	status, err := h.Service.QueueStatus(r.Context(), queueID)
	if err != nil {
		switch {
		case errors.Is(err, playbacksvc.ErrQueueItemNotFound):
			http.Error(w, "queue item not found", http.StatusNotFound)
		case errors.Is(err, playbacksvc.ErrQueueItemFailed):
			http.Error(w, err.Error(), http.StatusBadGateway)
		default:
			http.Error(w, err.Error(), http.StatusBadGateway)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}
