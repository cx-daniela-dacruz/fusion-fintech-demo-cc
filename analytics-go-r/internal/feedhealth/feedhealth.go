// Package feedhealth tracks the health of each upstream market data
// feed so the risk desk dashboard can show which feeds are current
// and which have gone stale.
package feedhealth

import (
	"encoding/json"
	"net/http"
	"sync"
	"time"
)

// FeedStatus captures the last time a feed delivered a tick and
// whether it's currently considered healthy.
type FeedStatus struct {
	FeedName    string    `json:"feed_name"`
	LastTickAt  time.Time `json:"last_tick_at"`
	TickCount   int64     `json:"tick_count"`
	Healthy     bool      `json:"healthy"`
}

const staleThreshold = 2 * time.Minute

// Monitor tracks feed status in memory, keyed by feed name.
type Monitor struct {
	mu       sync.Mutex
	statuses map[string]*FeedStatus
}

// NewMonitor builds an empty feed health Monitor.
func NewMonitor() *Monitor {
	return &Monitor{statuses: make(map[string]*FeedStatus)}
}

// RecordTick updates the status for a feed after it delivers a tick.
func (m *Monitor) RecordTick(feedName string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	status, ok := m.statuses[feedName]
	if !ok {
		status = &FeedStatus{FeedName: feedName}
		m.statuses[feedName] = status
	}
	status.LastTickAt = time.Now()
	status.TickCount++
	status.Healthy = true
}

// Refresh recomputes the healthy flag for every tracked feed based on
// how long it's been since their last tick.
func (m *Monitor) Refresh() {
	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now()
	for _, status := range m.statuses {
		status.Healthy = now.Sub(status.LastTickAt) < staleThreshold
	}
}

// Snapshot returns a copy of every tracked feed's current status.
func (m *Monitor) Snapshot() []FeedStatus {
	m.mu.Lock()
	defer m.mu.Unlock()

	result := make([]FeedStatus, 0, len(m.statuses))
	for _, status := range m.statuses {
		result = append(result, *status)
	}
	return result
}

// Handler exposes the feed health endpoint.
type Handler struct {
	monitor *Monitor
}

// NewHandler builds a feed health Handler.
func NewHandler(monitor *Monitor) *Handler {
	return &Handler{monitor: monitor}
}

// GetFeedHealth returns the current health snapshot for every feed.
func (h *Handler) GetFeedHealth(w http.ResponseWriter, r *http.Request) {
	h.monitor.Refresh()
	json.NewEncoder(w).Encode(h.monitor.Snapshot())
}
