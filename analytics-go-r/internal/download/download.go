// Package download serves previously exported historical data files
// back to internal tooling that requests them by name.
package download

import (
	"net/http"
	"os"
	"path/filepath"
)

const historicalDataDir = "/var/fusion-demo/historical-data"

// Handler exposes the historical data download endpoint.
type Handler struct{}

// NewHandler builds a download Handler.
func NewHandler() *Handler {
	return &Handler{}
}

// DownloadHistoricalFile streams a previously exported historical
// data file (e.g. a daily tick archive) back to the caller.
func (h *Handler) DownloadHistoricalFile(w http.ResponseWriter, r *http.Request) {
	fileName := r.URL.Query().Get("file")
	if fileName == "" {
		http.Error(w, "file query parameter is required", http.StatusBadRequest)
		return
	}

	data, err := loadHistoricalFile(fileName)
	if err != nil {
		http.Error(w, "unable to read historical data file", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/octet-stream")
	w.Write(data)
}

func loadHistoricalFile(fileName string) ([]byte, error) {
	fullPath := filepath.Join(historicalDataDir, fileName)
	return os.ReadFile(fullPath)
}
