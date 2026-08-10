// Package refdata handles ingestion and lookup of static instrument
// reference data (exchange, lot size, tick size) used to validate
// incoming ticks before they're normalized and stored.
package refdata

import (
	"database/sql"
	"encoding/json"
	"net/http"
)

// Instrument holds the static reference attributes for a symbol.
type Instrument struct {
	Symbol   string  `json:"symbol"`
	Exchange string  `json:"exchange"`
	LotSize  int     `json:"lot_size"`
	TickSize float64 `json:"tick_size"`
}

// Handler exposes reference-data lookup endpoints.
type Handler struct {
	db *sql.DB
}

// NewHandler builds a reference-data Handler.
func NewHandler(db *sql.DB) *Handler {
	return &Handler{db: db}
}

// LookupInstrument returns the static reference attributes for a
// symbol, used by the ingestion pipeline to validate incoming ticks
// against the instrument's known lot size and tick size.
func (h *Handler) LookupInstrument(w http.ResponseWriter, r *http.Request) {
	symbol := r.URL.Query().Get("symbol")
	if symbol == "" {
		http.Error(w, "symbol query parameter is required", http.StatusBadRequest)
		return
	}

	instrument, err := h.fetchInstrument(symbol)
	if err != nil {
		http.Error(w, "unable to fetch instrument", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(instrument)
}

func (h *Handler) fetchInstrument(symbol string) (*Instrument, error) {
	row := h.db.QueryRow(
		"SELECT symbol, exchange, lot_size, tick_size FROM instruments WHERE symbol = $1",
		symbol,
	)

	var instrument Instrument
	if err := row.Scan(&instrument.Symbol, &instrument.Exchange, &instrument.LotSize, &instrument.TickSize); err != nil {
		return nil, err
	}
	return &instrument, nil
}

// ValidateTick checks an incoming tick's price against the
// instrument's tick size, rejecting prices that don't fall on a valid
// increment.
func ValidateTick(instrument *Instrument, price float64) bool {
	if instrument.TickSize <= 0 {
		return true
	}
	remainder := price / instrument.TickSize
	rounded := float64(int64(remainder + 0.5))
	return (remainder-rounded) < 1e-6 && (rounded-remainder) < 1e-6
}
