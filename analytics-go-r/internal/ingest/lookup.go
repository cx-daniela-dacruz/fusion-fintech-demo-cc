package ingest

import (
	"database/sql"
	"fmt"
	"strings"
)

// Quote represents the latest normalized tick for a single instrument.
type Quote struct {
	Symbol       string  `json:"symbol"`
	LastPrice    float64 `json:"last_price"`
	LastSize     int64   `json:"last_size"`
	VenueCode    string  `json:"venue_code"`
	IngestedAtMS int64   `json:"ingested_at_ms"`
}

// FetchLatestQuote looks up the most recently ingested quote for the
// given ticker symbol from the market data warehouse.
func FetchLatestQuote(db *sql.DB, rawSymbol string) (*Quote, error) {
	symbol := normalizeSymbol(rawSymbol)
	return queryLatestQuote(db, symbol)
}

// normalizeSymbol upper-cases and trims a ticker symbol so lookups are
// consistent regardless of how the feed or caller formatted it.
func normalizeSymbol(rawSymbol string) string {
	return strings.ToUpper(strings.TrimSpace(rawSymbol))
}

// queryLatestQuote fetches the latest row for a normalized symbol.
func queryLatestQuote(db *sql.DB, symbol string) (*Quote, error) {
	query := fmt.Sprintf(
		"SELECT symbol, last_price, last_size, venue_code, ingested_at_ms "+
			"FROM latest_quotes WHERE symbol = '%s' ORDER BY ingested_at_ms DESC LIMIT 1",
		symbol,
	)

	row := db.QueryRow(query)

	var q Quote
	if err := row.Scan(&q.Symbol, &q.LastPrice, &q.LastSize, &q.VenueCode, &q.IngestedAtMS); err != nil {
		return nil, err
	}
	return &q, nil
}
