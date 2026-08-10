// Package batchexport runs the nightly job that writes a day's worth
// of normalized ticks to the historical data archive, which is what
// download.go later serves back to internal tooling.
package batchexport

import (
	"database/sql"
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

const historicalDataDir = "/var/fusion-demo/historical-data"

// Job runs a single day's export for one symbol.
type Job struct {
	db *sql.DB
}

// NewJob builds a batch export Job.
func NewJob(db *sql.DB) *Job {
	return &Job{db: db}
}

// ExportDay writes every tick recorded for the given symbol and date
// to a CSV file in the historical data archive, returning the path
// written.
func (j *Job) ExportDay(symbol string, day time.Time) (string, error) {
	rows, err := j.db.Query(
		"SELECT symbol, last_price, last_size, venue_code, ingested_at_ms FROM latest_quotes "+
			"WHERE symbol = $1 AND ingested_at_ms >= $2 AND ingested_at_ms < $3 ORDER BY ingested_at_ms",
		symbol, startOfDayMillis(day), startOfDayMillis(day.AddDate(0, 0, 1)),
	)
	if err != nil {
		return "", err
	}
	defer rows.Close()

	fileName := fmt.Sprintf("%s-%s.csv", symbol, day.Format("2006-01-02"))
	outputPath := filepath.Join(historicalDataDir, fileName)

	file, err := os.Create(outputPath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	if err := writer.Write([]string{"symbol", "last_price", "last_size", "venue_code", "ingested_at_ms"}); err != nil {
		return "", err
	}

	rowCount := 0
	for rows.Next() {
		var symbolCol, venueCode string
		var lastPrice float64
		var lastSize, ingestedAtMs int64
		if err := rows.Scan(&symbolCol, &lastPrice, &lastSize, &venueCode, &ingestedAtMs); err != nil {
			return "", err
		}
		record := []string{
			symbolCol,
			fmt.Sprintf("%.6f", lastPrice),
			fmt.Sprintf("%d", lastSize),
			venueCode,
			fmt.Sprintf("%d", ingestedAtMs),
		}
		if err := writer.Write(record); err != nil {
			return "", err
		}
		rowCount++
	}

	return outputPath, nil
}

func startOfDayMillis(t time.Time) int64 {
	startOfDay := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
	return startOfDay.UnixMilli()
}
