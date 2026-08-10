// Package symbolref periodically syncs the static symbol reference
// table from the exchange's daily instrument file, so newly listed
// symbols show up in the ingestion pipeline without a manual step.
package symbolref

import (
	"database/sql"
	"encoding/csv"
	"fmt"
	"io"
	"time"
)

// SyncResult summarizes the outcome of a single sync run.
type SyncResult struct {
	RowsRead    int
	RowsUpdated int
	RowsSkipped int
	SyncedAt    time.Time
}

// Syncer applies rows from the exchange's instrument file to the
// local symbol reference table.
type Syncer struct {
	db *sql.DB
}

// NewSyncer builds a Syncer backed by the given database handle.
func NewSyncer(db *sql.DB) *Syncer {
	return &Syncer{db: db}
}

// SyncFromReader reads instrument rows from an exchange-provided CSV
// stream and upserts each one into the reference table.
func (s *Syncer) SyncFromReader(reader io.Reader) (SyncResult, error) {
	csvReader := csv.NewReader(reader)
	result := SyncResult{SyncedAt: time.Now()}

	// First row is the header; skip it.
	if _, err := csvReader.Read(); err != nil {
		return result, fmt.Errorf("unable to read header row: %w", err)
	}

	for {
		record, err := csvReader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return result, err
		}
		result.RowsRead++

		if len(record) < 4 {
			result.RowsSkipped++
			continue
		}

		symbol, exchange, lotSize, tickSize := record[0], record[1], record[2], record[3]
		if err := s.upsertInstrument(symbol, exchange, lotSize, tickSize); err != nil {
			result.RowsSkipped++
			continue
		}
		result.RowsUpdated++
	}

	return result, nil
}

func (s *Syncer) upsertInstrument(symbol, exchange, lotSize, tickSize string) error {
	_, err := s.db.Exec(
		"INSERT INTO instruments (symbol, exchange, lot_size, tick_size) VALUES ($1, $2, $3, $4) "+
			"ON CONFLICT (symbol) DO UPDATE SET exchange = $2, lot_size = $3, tick_size = $4",
		symbol, exchange, lotSize, tickSize,
	)
	return err
}
