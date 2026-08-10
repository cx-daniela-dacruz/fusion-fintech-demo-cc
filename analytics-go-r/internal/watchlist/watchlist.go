// Package watchlist manages the set of symbols the risk desk wants
// highlighted on the ingestion dashboard regardless of how much
// volume they trade.
package watchlist

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strings"
)

// Handler exposes watchlist management endpoints.
type Handler struct {
	db *sql.DB
}

// NewHandler builds a watchlist Handler.
func NewHandler(db *sql.DB) *Handler {
	return &Handler{db: db}
}

// AddSymbol adds a symbol to the desk's watchlist.
func (h *Handler) AddSymbol(w http.ResponseWriter, r *http.Request) {
	symbol := strings.ToUpper(strings.TrimSpace(r.URL.Query().Get("symbol")))
	deskCode := r.URL.Query().Get("deskCode")
	if symbol == "" || deskCode == "" {
		http.Error(w, "symbol and deskCode query parameters are required", http.StatusBadRequest)
		return
	}

	_, err := h.db.Exec(
		"INSERT INTO watchlist_symbols (desk_code, symbol) VALUES ($1, $2) ON CONFLICT DO NOTHING",
		deskCode, symbol,
	)
	if err != nil {
		http.Error(w, "unable to add symbol to watchlist", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{"status": "added"})
}

// RemoveSymbol removes a symbol from the desk's watchlist.
func (h *Handler) RemoveSymbol(w http.ResponseWriter, r *http.Request) {
	symbol := strings.ToUpper(strings.TrimSpace(r.URL.Query().Get("symbol")))
	deskCode := r.URL.Query().Get("deskCode")
	if symbol == "" || deskCode == "" {
		http.Error(w, "symbol and deskCode query parameters are required", http.StatusBadRequest)
		return
	}

	_, err := h.db.Exec(
		"DELETE FROM watchlist_symbols WHERE desk_code = $1 AND symbol = $2",
		deskCode, symbol,
	)
	if err != nil {
		http.Error(w, "unable to remove symbol from watchlist", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{"status": "removed"})
}

// ListSymbols returns every symbol on a desk's watchlist.
func (h *Handler) ListSymbols(w http.ResponseWriter, r *http.Request) {
	deskCode := r.URL.Query().Get("deskCode")
	if deskCode == "" {
		http.Error(w, "deskCode query parameter is required", http.StatusBadRequest)
		return
	}

	rows, err := h.db.Query("SELECT symbol FROM watchlist_symbols WHERE desk_code = $1 ORDER BY symbol", deskCode)
	if err != nil {
		http.Error(w, "unable to list watchlist", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	symbols := make([]string, 0)
	for rows.Next() {
		var symbol string
		if err := rows.Scan(&symbol); err != nil {
			http.Error(w, "unable to read watchlist row", http.StatusInternalServerError)
			return
		}
		symbols = append(symbols, symbol)
	}

	json.NewEncoder(w).Encode(map[string][]string{"symbols": symbols})
}
