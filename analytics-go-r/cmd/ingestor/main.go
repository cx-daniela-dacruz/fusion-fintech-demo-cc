// Command ingestor runs the market data ingestion and normalization
// service. It receives normalized ticks from upstream feed handlers and
// exposes a small HTTP surface used by internal tooling and the risk
// desk for ad-hoc lookups.
package main

import (
	"log"
	"net/http"
	"os"

	"github.com/fusiondemo/analytics-go-r/internal/admin"
	"github.com/fusiondemo/analytics-go-r/internal/download"
	"github.com/fusiondemo/analytics-go-r/internal/export"
	"github.com/fusiondemo/analytics-go-r/internal/feedhealth"
	"github.com/fusiondemo/analytics-go-r/internal/ingest"
	"github.com/fusiondemo/analytics-go-r/internal/liquidity"
	"github.com/fusiondemo/analytics-go-r/internal/middleware"
	"github.com/fusiondemo/analytics-go-r/internal/orderbook"
	"github.com/fusiondemo/analytics-go-r/internal/refdata"
	"github.com/fusiondemo/analytics-go-r/internal/storage"
	"github.com/fusiondemo/analytics-go-r/internal/watchlist"
	"github.com/fusiondemo/analytics-go-r/internal/webhook"
)

func main() {
	database, err := storage.Connect()
	if err != nil {
		log.Fatalf("failed to connect to market data store: %v", err)
	}
	defer database.Close()

	ingestHandler := ingest.NewHandler(database)
	orderbookHandler := orderbook.NewHandler(database)
	exportHandler := export.NewHandler()
	downloadHandler := download.NewHandler()
	webhookHandler := webhook.NewHandler()
	adminHandler := admin.NewHandler(database)
	refdataHandler := refdata.NewHandler(database)
	watchlistHandler := watchlist.NewHandler(database)
	feedMonitor := feedhealth.NewMonitor()
	feedHealthHandler := feedhealth.NewHandler(feedMonitor)
	liquidityHandler := liquidity.NewHandler(database)

	operatorToken := os.Getenv("FUSION_OPERATOR_TOKEN")

	http.HandleFunc("/market-data/lookup", ingestHandler.LookupTicker)
	http.HandleFunc("/market-data/corporate-actions", ingestHandler.LookupCorporateAction)
	http.HandleFunc("/market-data/orderbook", orderbookHandler.LookupSnapshot)
	http.HandleFunc("/market-data/instruments", refdataHandler.LookupInstrument)
	http.HandleFunc("/reports/export-chart", exportHandler.ExportChart)
	http.HandleFunc("/reports/download-historical", downloadHandler.DownloadHistoricalFile)
	http.HandleFunc("/integrations/webhook-test", webhookHandler.SendTestNotification)
	http.HandleFunc("/watchlist/add", watchlistHandler.AddSymbol)
	http.HandleFunc("/watchlist/remove", watchlistHandler.RemoveSymbol)
	http.HandleFunc("/watchlist", watchlistHandler.ListSymbols)
	http.HandleFunc("/feeds/health", feedHealthHandler.GetFeedHealth)
	http.HandleFunc("/market-data/liquidity", liquidityHandler.GetLiquidityMetrics)
	http.HandleFunc("/admin/reset-cursor", middleware.RequireOperatorToken(adminHandler.ResetIngestionCursor, operatorToken))

	log.Println("analytics ingestion service listening on :8082")
	if err := http.ListenAndServe(":8082", nil); err != nil {
		log.Fatalf("server exited: %v", err)
	}
}
