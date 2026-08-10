// Package webhook delivers feed-health notifications to a
// partner-registered callback URL when an ingestion pipeline recovers
// from an outage.
package webhook

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
)

// Handler exposes the webhook test-delivery endpoint.
type Handler struct{}

// NewHandler builds a webhook Handler.
func NewHandler() *Handler {
	return &Handler{}
}

// SendTestNotification delivers a synthetic test event to a partner's
// registered callback URL so they can confirm it's reachable before
// going live.
func (h *Handler) SendTestNotification(w http.ResponseWriter, r *http.Request) {
	callbackURL := r.URL.Query().Get("callbackUrl")
	if callbackURL == "" {
		http.Error(w, "callbackUrl query parameter is required", http.StatusBadRequest)
		return
	}

	status, err := deliverTestEvent(callbackURL)
	if err != nil {
		http.Error(w, "unable to deliver test notification", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]int{"delivered_status": status})
}

func deliverTestEvent(callbackURL string) (int, error) {
	payload := bytes.NewBufferString(`{"event":"feed.health.test"}`)
	response, err := http.Post(callbackURL, "application/json", payload)
	if err != nil {
		return 0, err
	}
	defer response.Body.Close()
	io.Copy(io.Discard, response.Body)
	return response.StatusCode, nil
}
