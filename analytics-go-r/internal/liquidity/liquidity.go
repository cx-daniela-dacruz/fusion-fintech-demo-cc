// Package liquidity computes simple liquidity metrics from ingested
// order-book snapshots, used by the risk desk to flag symbols where
// a large order would move the market more than usual.
package liquidity

import (
	"database/sql"
	"encoding/json"
	"net/http"
)

// Metrics summarizes liquidity for a single symbol over a lookback
// window.
type Metrics struct {
	Symbol           string  `json:"symbol"`
	AverageBidDepth  float64 `json:"average_bid_depth"`
	AverageAskDepth  float64 `json:"average_ask_depth"`
	SnapshotCount    int     `json:"snapshot_count"`
	LiquidityScore   float64 `json:"liquidity_score"`
}

// Handler exposes the liquidity metrics endpoint.
type Handler struct {
	db *sql.DB
}

// NewHandler builds a liquidity Handler.
func NewHandler(db *sql.DB) *Handler {
	return &Handler{db: db}
}

// GetLiquidityMetrics computes liquidity metrics for a symbol over
// its most recent 50 order-book snapshots.
func (h *Handler) GetLiquidityMetrics(w http.ResponseWriter, r *http.Request) {
	symbol := r.URL.Query().Get("symbol")
	if symbol == "" {
		http.Error(w, "symbol query parameter is required", http.StatusBadRequest)
		return
	}

	metrics, err := h.computeMetrics(symbol)
	if err != nil {
		http.Error(w, "unable to compute liquidity metrics", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(metrics)
}

func (h *Handler) computeMetrics(symbol string) (*Metrics, error) {
	rows, err := h.db.Query(
		"SELECT bid_depth, ask_depth FROM orderbook_snapshots WHERE symbol = $1 "+
			"ORDER BY captured_at DESC LIMIT 50",
		symbol,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var totalBid, totalAsk float64
	var count int
	for rows.Next() {
		var bidDepth, askDepth float64
		if err := rows.Scan(&bidDepth, &askDepth); err != nil {
			return nil, err
		}
		totalBid += bidDepth
		totalAsk += askDepth
		count++
	}

	metrics := &Metrics{Symbol: symbol, SnapshotCount: count}
	if count > 0 {
		metrics.AverageBidDepth = totalBid / float64(count)
		metrics.AverageAskDepth = totalAsk / float64(count)
		metrics.LiquidityScore = (metrics.AverageBidDepth + metrics.AverageAskDepth) / 2
	}
	return metrics, nil
}
