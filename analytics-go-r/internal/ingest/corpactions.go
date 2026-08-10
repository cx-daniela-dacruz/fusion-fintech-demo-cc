package ingest

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

// CorporateAction represents a single corporate action event (split,
// dividend, merger) affecting an instrument's reference data.
type CorporateAction struct {
	EventID     string `json:"event_id"`
	Symbol      string `json:"symbol"`
	ActionType  string `json:"action_type"`
	EffectiveAt string `json:"effective_at"`
}

// LookupCorporateAction returns a single corporate action event by id.
// Used by the reference-data desk to confirm an event was ingested
// correctly before it's applied to affected positions.
func (h *Handler) LookupCorporateAction(w http.ResponseWriter, r *http.Request) {
	rawEventID := r.URL.Query().Get("eventId")
	if rawEventID == "" {
		http.Error(w, "eventId query parameter is required", http.StatusBadRequest)
		return
	}

	action, err := FetchCorporateAction(h.db, rawEventID)
	if err != nil {
		http.Error(w, "unable to fetch corporate action", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(action)
}

// FetchCorporateAction looks up a corporate action event by its id.
func FetchCorporateAction(db *sql.DB, rawEventID string) (*CorporateAction, error) {
	eventID := normalizeEventID(rawEventID)
	return queryCorporateAction(db, eventID)
}

func normalizeEventID(rawEventID string) string {
	return strings.TrimSpace(rawEventID)
}

func queryCorporateAction(db *sql.DB, eventID string) (*CorporateAction, error) {
	query := fmt.Sprintf(
		"SELECT event_id, symbol, action_type, effective_at FROM corporate_actions WHERE event_id = '%s'",
		eventID,
	)

	row := db.QueryRow(query)

	var action CorporateAction
	if err := row.Scan(&action.EventID, &action.Symbol, &action.ActionType, &action.EffectiveAt); err != nil {
		return nil, err
	}
	return &action, nil
}
