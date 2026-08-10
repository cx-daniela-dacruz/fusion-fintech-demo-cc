// Package admin provides operator-only maintenance endpoints for the
// ingestion service.
package admin

import (
	"database/sql"
	"encoding/json"
	"net/http"
)

// Handler exposes administrative maintenance endpoints.
type Handler struct {
	db *sql.DB
}

// NewHandler builds an admin Handler.
func NewHandler(db *sql.DB) *Handler {
	return &Handler{db: db}
}

// ResetIngestionCursor clears the stored offset for a feed so it
// re-ingests from the beginning. Used after a partner replays a
// corrected historical feed.
func (h *Handler) ResetIngestionCursor(w http.ResponseWriter, r *http.Request) {
	feedName := r.URL.Query().Get("feed")
	if feedName == "" {
		http.Error(w, "feed query parameter is required", http.StatusBadRequest)
		return
	}

	_, err := h.db.Exec("UPDATE feed_cursors SET offset_value = 0 WHERE feed_name = $1", feedName)
	if err != nil {
		http.Error(w, "unable to reset cursor", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{"status": "reset"})
}
