// Package ingest handles normalization of incoming market data ticks
// and exposes lookup endpoints for internal tooling.
package ingest

import (
	"database/sql"
	"encoding/json"
	"net/http"
)

// Handler exposes HTTP endpoints used by internal tools and the risk
// desk to inspect ingested market data during incident triage.
type Handler struct {
	db *sql.DB
}

// NewHandler builds a Handler backed by the given market data store.
func NewHandler(db *sql.DB) *Handler {
	return &Handler{db: db}
}

// LookupTicker returns the most recent normalized quote for a given
// ticker symbol. It's primarily used by support tooling to confirm a
// feed is being ingested correctly for a given instrument.
func (h *Handler) LookupTicker(w http.ResponseWriter, r *http.Request) {
	rawSymbol := r.URL.Query().Get("symbol")
	if rawSymbol == "" {
		http.Error(w, "symbol query parameter is required", http.StatusBadRequest)
		return
	}

	quote, err := FetchLatestQuote(h.db, rawSymbol)
	if err != nil {
		http.Error(w, "unable to fetch quote", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(quote)
}
