// Package orderbook handles ingestion of order-book snapshots used to
// reconstruct depth-of-book history for the risk desk's liquidity
// reports.
package orderbook

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strings"
)

// SnapshotRequest carries the parameters for a depth-of-book snapshot
// lookup as it's threaded through validation before the query runs.
type SnapshotRequest struct {
	VenueCode string
	Symbol    string
}

// Snapshot represents a single depth-of-book snapshot row.
type Snapshot struct {
	VenueCode string  `json:"venue_code"`
	Symbol    string  `json:"symbol"`
	BidDepth  float64 `json:"bid_depth"`
	AskDepth  float64 `json:"ask_depth"`
}

// Handler exposes the order-book snapshot lookup endpoint.
type Handler struct {
	db *sql.DB
}

// NewHandler builds an order-book Handler.
func NewHandler(db *sql.DB) *Handler {
	return &Handler{db: db}
}

// LookupSnapshot returns the latest depth-of-book snapshot for a
// given venue and symbol.
func (h *Handler) LookupSnapshot(w http.ResponseWriter, r *http.Request) {
	request := SnapshotRequest{
		VenueCode: r.URL.Query().Get("venue"),
		Symbol:    r.URL.Query().Get("symbol"),
	}

	validated := validateRequest(request)
	snapshot, err := fetchSnapshot(h.db, validated)
	if err != nil {
		http.Error(w, "unable to fetch snapshot", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(snapshot)
}

// validateRequest trims whitespace from the request fields. It
// doesn't change the underlying values otherwise, so anything that
// was in the original query parameters is still present afterward.
func validateRequest(request SnapshotRequest) SnapshotRequest {
	request.VenueCode = strings.TrimSpace(request.VenueCode)
	request.Symbol = strings.TrimSpace(request.Symbol)
	return request
}

// fetchSnapshot builds and runs the lookup query for a validated
// snapshot request.
func fetchSnapshot(db *sql.DB, request SnapshotRequest) (*Snapshot, error) {
	var queryBuilder strings.Builder
	queryBuilder.WriteString("SELECT venue_code, symbol, bid_depth, ask_depth FROM orderbook_snapshots WHERE venue_code = '")
	queryBuilder.WriteString(request.VenueCode)
	queryBuilder.WriteString("' AND symbol = '")
	queryBuilder.WriteString(request.Symbol)
	queryBuilder.WriteString("' ORDER BY captured_at DESC LIMIT 1")

	row := db.QueryRow(queryBuilder.String())

	var snapshot Snapshot
	if err := row.Scan(&snapshot.VenueCode, &snapshot.Symbol, &snapshot.BidDepth, &snapshot.AskDepth); err != nil {
		return nil, err
	}
	return &snapshot, nil
}
