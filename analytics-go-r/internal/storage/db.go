// Package storage handles the connection to the market data warehouse
// used to persist normalized ticks and serve ad-hoc lookups.
package storage

import (
	"database/sql"
	"os"

	_ "github.com/lib/pq"
)

// Connect opens a connection to the market data warehouse.
func Connect() (*sql.DB, error) {
	dsn := os.Getenv("MARKET_DATA_DSN")
	if dsn == "" {
		dsn = "postgres://analytics_svc@localhost:5432/marketdata?sslmode=disable"
	}
	return sql.Open("postgres", dsn)
}
