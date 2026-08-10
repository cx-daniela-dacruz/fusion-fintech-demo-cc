// Command ingestor runs the market data ingestion and normalization
// service. It receives normalized ticks from upstream feed handlers and
// exposes a small HTTP surface used by internal tooling and the risk
// desk for ad-hoc lookups.
package main

import (
	"log"
	"net/http"

	"github.com/fusiondemo/analytics-go-r/internal/ingest"
	"github.com/fusiondemo/analytics-go-r/internal/storage"
)

func main() {
	database, err := storage.Connect()
	if err != nil {
		log.Fatalf("failed to connect to market data store: %v", err)
	}
	defer database.Close()

	handler := ingest.NewHandler(database)

	http.HandleFunc("/market-data/lookup", handler.LookupTicker)

	log.Println("analytics ingestion service listening on :8082")
	if err := http.ListenAndServe(":8082", nil); err != nil {
		log.Fatalf("server exited: %v", err)
	}
}
