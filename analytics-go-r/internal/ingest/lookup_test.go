package ingest

import (
	"database/sql"
	"database/sql/driver"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// ---------------------------------------------------------------------------
// Minimal in-process SQL driver used to capture query text and arguments
// without any external database. This lets us verify that queryLatestQuote
// passes the symbol as a driver.Value (parameterized) rather than
// interpolating it directly into the query string.
// ---------------------------------------------------------------------------

// capturedQuery records the last query text and its bound arguments.
type capturedQuery struct {
	query string
	args  []driver.NamedValue
}

var lastQuery capturedQuery

// testDriver implements driver.Driver.
type testDriver struct{}

func (testDriver) Open(name string) (driver.Conn, error) { return &testConn{}, nil }

// testConn implements driver.Conn.
type testConn struct{}

func (c *testConn) Prepare(query string) (driver.Stmt, error) {
	return &testStmt{query: query}, nil
}
func (c *testConn) Close() error          { return nil }
func (c *testConn) Begin() (driver.Tx, error) {
	return nil, fmt.Errorf("transactions not supported in test driver")
}

// testStmt implements driver.Stmt and driver.StmtQueryContext.
type testStmt struct {
	query string
}

func (s *testStmt) Close() error                                    { return nil }
func (s *testStmt) NumInput() int                                   { return -1 } // variadic
func (s *testStmt) Exec(args []driver.Value) (driver.Result, error) { return nil, nil }

// Query records the query+args and returns a single matching row or
// sql.ErrNoRows via an empty result set, depending on whether the first
// bound argument equals the "found" sentinel value.
func (s *testStmt) Query(args []driver.Value) (driver.Rows, error) {
	// Convert positional args to NamedValue for storage.
	named := make([]driver.NamedValue, len(args))
	for i, v := range args {
		named[i] = driver.NamedValue{Ordinal: i + 1, Value: v}
	}
	lastQuery = capturedQuery{query: s.query, args: named}

	// Return a row when the symbol is the known "AAPL" ticker; empty otherwise.
	sym := ""
	if len(args) > 0 {
		if s, ok := args[0].(string); ok {
			sym = s
		}
	}
	if sym == "AAPL" {
		return &testRows{
			cols: []string{"symbol", "last_price", "last_size", "venue_code", "ingested_at_ms"},
			data: [][]driver.Value{{"AAPL", float64(182.50), int64(100), "NYSE", int64(1700000000000)}},
		}, nil
	}
	// Empty result → Scan will return sql.ErrNoRows.
	return &testRows{
		cols: []string{"symbol", "last_price", "last_size", "venue_code", "ingested_at_ms"},
		data: nil,
	}, nil
}

// testRows implements driver.Rows.
type testRows struct {
	cols    []string
	data    [][]driver.Value
	current int
}

func (r *testRows) Columns() []string { return r.cols }
func (r *testRows) Close() error      { return nil }
func (r *testRows) Next(dest []driver.Value) error {
	if r.current >= len(r.data) {
		return io.EOF
	}
	copy(dest, r.data[r.current])
	r.current++
	return nil
}

func init() {
	sql.Register("testdb", testDriver{})
}

// openTestDB returns a *sql.DB backed by the in-process test driver.
func openTestDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("testdb", "")
	if err != nil {
		t.Fatalf("sql.Open: %v", err)
	}
	return db
}

// ---------------------------------------------------------------------------
// Tests for normalizeSymbol
// ---------------------------------------------------------------------------

func TestNormalizeSymbol_UpperCases(t *testing.T) {
	if got := normalizeSymbol("aapl"); got != "AAPL" {
		t.Errorf("normalizeSymbol(%q) = %q, want %q", "aapl", got, "AAPL")
	}
}

func TestNormalizeSymbol_TrimsWhitespace(t *testing.T) {
	if got := normalizeSymbol("  msft  "); got != "MSFT" {
		t.Errorf("normalizeSymbol(%q) = %q, want %q", "  msft  ", got, "MSFT")
	}
}

func TestNormalizeSymbol_AlreadyNormalized(t *testing.T) {
	if got := normalizeSymbol("TSLA"); got != "TSLA" {
		t.Errorf("normalizeSymbol(%q) = %q, want %q", "TSLA", got, "TSLA")
	}
}

// ---------------------------------------------------------------------------
// Tests that verify the SQL injection fix: the symbol must reach the
// database as a bound parameter, never embedded in the query string.
// ---------------------------------------------------------------------------

// TestQueryLatestQuote_UsesParameterizedQuery asserts that the query string
// sent to the driver contains a placeholder ($1) and NOT the literal symbol
// value. This is the primary regression guard for the SQL injection fix.
func TestQueryLatestQuote_UsesParameterizedQuery(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()

	_, _ = queryLatestQuote(db, "AAPL")

	// The query text must contain the placeholder, not the literal value.
	if !strings.Contains(lastQuery.query, "$1") {
		t.Errorf("query does not contain placeholder $1: %q", lastQuery.query)
	}
	if strings.Contains(lastQuery.query, "AAPL") {
		t.Errorf("query embeds literal symbol value (SQL injection risk): %q", lastQuery.query)
	}
}

// TestQueryLatestQuote_SymbolPassedAsArgument verifies that the symbol is
// provided as a separate bound argument (driver.NamedValue) and not baked
// into the query string.
func TestQueryLatestQuote_SymbolPassedAsArgument(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()

	_, _ = queryLatestQuote(db, "AAPL")

	if len(lastQuery.args) == 0 {
		t.Fatal("no bound arguments found; symbol must be passed as a parameter")
	}
	if got, ok := lastQuery.args[0].Value.(string); !ok || got != "AAPL" {
		t.Errorf("bound arg[0] = %v, want %q", lastQuery.args[0].Value, "AAPL")
	}
}

// TestQueryLatestQuote_SQLInjectionPayloadNotInQuery ensures that a classic
// SQL injection payload is never interpolated into the query text.
func TestQueryLatestQuote_SQLInjectionPayloadNotInQuery(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()

	injectionPayloads := []string{
		"' OR '1'='1",
		"'; DROP TABLE latest_quotes; --",
		"' UNION SELECT 1,2,3,4,5 --",
		"1' AND sleep(5) --",
	}

	for _, payload := range injectionPayloads {
		lastQuery = capturedQuery{} // reset
		_, _ = queryLatestQuote(db, payload)

		if strings.Contains(lastQuery.query, payload) {
			t.Errorf("injection payload %q was interpolated into query: %q", payload, lastQuery.query)
		}
		// The query text must still contain the placeholder.
		if !strings.Contains(lastQuery.query, "$1") {
			t.Errorf("query lost its placeholder after payload %q: %q", payload, lastQuery.query)
		}
	}
}

// TestQueryLatestQuote_ReturnsQuoteOnHit verifies normal data retrieval.
func TestQueryLatestQuote_ReturnsQuoteOnHit(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()

	q, err := queryLatestQuote(db, "AAPL")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if q.Symbol != "AAPL" {
		t.Errorf("Symbol = %q, want %q", q.Symbol, "AAPL")
	}
	if q.LastPrice != 182.50 {
		t.Errorf("LastPrice = %v, want 182.50", q.LastPrice)
	}
}

// TestQueryLatestQuote_ReturnsErrorOnMiss verifies sql.ErrNoRows propagates.
func TestQueryLatestQuote_ReturnsErrorOnMiss(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()

	_, err := queryLatestQuote(db, "UNKNOWN")
	if err == nil {
		t.Error("expected an error for unknown symbol, got nil")
	}
}

// ---------------------------------------------------------------------------
// Tests for FetchLatestQuote (public API surface)
// ---------------------------------------------------------------------------

// TestFetchLatestQuote_NormalizesBeforeQuery verifies that FetchLatestQuote
// uppercases/trims the symbol before passing it to the query layer.
func TestFetchLatestQuote_NormalizesBeforeQuery(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()

	// "aapl" (lowercase) should be normalized to "AAPL" and hit the mock row.
	q, err := FetchLatestQuote(db, "aapl")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if q.Symbol != "AAPL" {
		t.Errorf("Symbol = %q, want %q", q.Symbol, "AAPL")
	}
	// Confirm normalized value was sent as a parameter.
	if got, ok := lastQuery.args[0].Value.(string); !ok || got != "AAPL" {
		t.Errorf("bound arg[0] = %v, want %q", lastQuery.args[0].Value, "AAPL")
	}
}

// ---------------------------------------------------------------------------
// Tests for the HTTP handler (LookupTicker)
// ---------------------------------------------------------------------------

// TestLookupTicker_MissingSymbol returns 400 when no symbol is provided.
func TestLookupTicker_MissingSymbol(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()

	h := NewHandler(db)
	req := httptest.NewRequest(http.MethodGet, "/lookup", nil)
	w := httptest.NewRecorder()

	h.LookupTicker(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

// TestLookupTicker_ValidSymbol returns 200 and JSON for a known ticker.
func TestLookupTicker_ValidSymbol(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()

	h := NewHandler(db)
	req := httptest.NewRequest(http.MethodGet, "/lookup?symbol=AAPL", nil)
	w := httptest.NewRecorder()

	h.LookupTicker(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}
	ct := w.Header().Get("Content-Type")
	if !strings.HasPrefix(ct, "application/json") {
		t.Errorf("Content-Type = %q, want application/json", ct)
	}
}

// TestLookupTicker_InjectionPayloadDoesNotReachQueryString ensures that an
// injection payload submitted via the HTTP query parameter is never embedded
// in the SQL query string sent to the database driver.
func TestLookupTicker_InjectionPayloadDoesNotReachQueryString(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()

	h := NewHandler(db)

	injectionPayloads := []string{
		"' OR '1'='1",
		"'; DROP TABLE latest_quotes; --",
		"' UNION SELECT 1,2,3,4,5 --",
	}

	for _, payload := range injectionPayloads {
		lastQuery = capturedQuery{} // reset
		encodedPayload := payload   // http.NewRequest URL-encodes query params automatically
		req := httptest.NewRequest(http.MethodGet, "/lookup?symbol="+encodedPayload, nil)
		w := httptest.NewRecorder()

		h.LookupTicker(w, req)

		// Regardless of HTTP status, the raw payload must NOT appear in the query text.
		if strings.Contains(lastQuery.query, payload) {
			t.Errorf("injection payload %q was interpolated into SQL query: %q", payload, lastQuery.query)
		}
	}
}

// TestLookupTicker_LowercaseSymbolNormalized checks that a lowercase ticker
// submitted via HTTP is normalized before reaching the database.
func TestLookupTicker_LowercaseSymbolNormalized(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()

	h := NewHandler(db)
	req := httptest.NewRequest(http.MethodGet, "/lookup?symbol=aapl", nil)
	w := httptest.NewRecorder()

	h.LookupTicker(w, req)

	if len(lastQuery.args) > 0 {
		if got, ok := lastQuery.args[0].Value.(string); ok && got != "AAPL" {
			t.Errorf("normalized symbol = %q, want %q", got, "AAPL")
		}
	}
}
