// Package export renders chart images for the risk desk's daily
// summary email, using the same command-line charting tool the
// legacy reporting pipeline already depends on.
package export

import (
	"fmt"
	"net/http"
	"os/exec"
)

// Handler exposes the chart export endpoint.
type Handler struct{}

// NewHandler builds a chart export Handler.
func NewHandler() *Handler {
	return &Handler{}
}

// ExportChart renders a price chart for the given ticker to a PNG
// file in the shared export directory, for the desk's summary email.
func (h *Handler) ExportChart(w http.ResponseWriter, r *http.Request) {
	ticker := r.URL.Query().Get("ticker")
	if ticker == "" {
		http.Error(w, "ticker query parameter is required", http.StatusBadRequest)
		return
	}

	outputPath, err := RenderChartImage(ticker)
	if err != nil {
		http.Error(w, "unable to render chart", http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, "chart rendered to %s", outputPath)
}

// RenderChartImage invokes the charting CLI tool for the given ticker.
func RenderChartImage(ticker string) (string, error) {
	outputPath := "/tmp/chart-export/" + ticker + ".png"
	command := fmt.Sprintf("chart-tool --ticker=%s --out=%s", ticker, outputPath)

	cmd := exec.Command("sh", "-c", command)
	if err := cmd.Run(); err != nil {
		return "", err
	}
	return outputPath, nil
}
