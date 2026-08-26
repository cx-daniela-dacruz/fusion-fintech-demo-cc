package ingest

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
)

// ---------------------------------------------------------------------------
// Minimal in-process sql/driver mock (stdlib only, no external dependencies)
//
// The mock records the exact SQL string and positional arguments that were
// passed to db.QueryRow so tests can assert:
//   1. The query string contains a placeholder ($1) — not a literal symbol.
//   2. The symbol is passed as a separate bind argument.
//
// This is the canonical way to verify parameterized-query usage without
// adding packages beyond the Go standard library.
// ---------------------------------------------------------------------------

// recordedQuery holds the most recent query + args observed by the mock.
type recordedQuery struct {
	query string
	args  []driver.Value
}

// mockDriver implements driver.Driver. Each test registers a uniquely-named
// instance so parallel tests don't share state.
type mockDriver struct {
	mu       sync.Mutex
	rows     [][]driver.Value // rows to return (nil → io.EOF on first Next)
	last     recordedQuery
	queryErr error // optional error injected for error-path tests
}

func (d *mockDriver) Open(_ string) (driver.Conn, error) {
	return &mockConn{d: d}, nil
}

// mockConn implements driver.Conn.
type mockConn struct{ d *mockDriver }

func (c *mockConn) Prepare(query string) (driver.Stmt, error) {
	return &mockStmt{d: c.d, query: query}, nil
}
func (c *mockConn) Close() error                  { return nil }
func (c *mockConn) Begin() (driver.Tx, error)     { return nil, fmt.Errorf("not implemented") }

// mockStmt implements driver.Stmt.
type mockStmt struct {
	d     *mockDriver
	query string
}

func (s *mockStmt) Close() error { return nil }

// NumInput returns -1 so the sql package does not validate arg count.
func (s *mockStmt) NumInput() int { return -1 }

func (s *mockStmt) Exec(_ []driver.Value) (driver.Result, error) {
	return nil, fmt.Errorf("not implemented")
}

// Query records the query string and bound arguments, then returns mock rows.
func (s *mockStmt) Query(args []driver.Value) (driver.Rows, error) {
	s.d.mu.Lock()
	s.d.last = recordedQuery{query: s.query, args: append([]driver.Value(nil), args...)}
	rows := s.d.rows
	queryErr := s.d.queryErr
	s.d.mu.Unlock()

	if queryErr != nil {
		return nil, queryErr
	}
	return &mockRows{rows: rows}, nil
}

// mockRows implements driver.Rows.
type mockRows struct {
	rows    [][]driver.Value
	current int
}

var quoteColumns = []string{"symbol", "last_price", "last_size", "venue_code", "ingested_at_ms"}

func (r *mockRows) Columns() []string { return quoteColumns }
func (r *mockRows) Close() error      { return nil }
func (r *mockRows) Next(dest []driver.Value) error {
	if r.current >= len(r.rows) {
		return io.EOF
	}
	copy(dest, r.rows[r.current])
	r.current++
	return nil
}

// ---------------------------------------------------------------------------
// Driver registry — avoids re-registering the same name across test runs.
// ---------------------------------------------------------------------------

var (
	driversMu sync.Mutex
	registeredDrivers = map[string]*mockDriver{}
)

// openMockDB registers a uniquely-named mockDriver and returns an *sql.DB
// backed by it. Call db.Close() when done.
func openMockDB(t *testing.T, rows [][]driver.Value) (*sql.DB, *mockDriver) {
	t.Helper()

	// Use a name that is unique per test to avoid state leakage.
	name := "mock_" + strings.ReplaceAll(t.Name(), "/", "_")

	driversMu.Lock()
	d, exists := registeredDrivers[name]
	if !exists {
		d = &mockDriver{rows: rows}
		sql.Register(name, d)
		registeredDrivers[name] = d
	} else {
		// Driver already registered; update its state for the new sub-test.
		d.mu.Lock()
		d.rows = rows
		d.queryErr = nil
		d.last = recordedQuery{}
		d.mu.Unlock()
	}
	driversMu.Unlock()

	db, err := sql.Open(name, "")
	if err != nil {
		t.Fatalf("sql.Open: %v", err)
	}
	// Ensure a single connection so all calls hit the same mockConn/mockStmt.
	db.SetMaxOpenConns(1)
	return db, d
}

// ---------------------------------------------------------------------------
// queryLatestQuote — parameterized query regression tests
// ---------------------------------------------------------------------------

// TestQueryLatestQuote_UsesParameterizedQuery is the primary regression test
// for the SQL injection fix. It asserts that:
//   - The SQL string does NOT contain the literal symbol value.
//   - A placeholder ($1) is present in the SQL string.
//   - The symbol is delivered as the first bound argument.
func TestQueryLatestQuote_UsesParameterizedQuery(t *testing.T) {
	rows := [][]driver.Value{
		{"AAPL", 182.5, int64(200), "XNAS", int64(1700000000000)},
	}
	db, drv := openMockDB(t, rows)
	defer db.Close()

	_, err := FetchLatestQuote(db, "aapl")
	if err != nil {
		t.Fatalf("FetchLatestQuote returned unexpected error: %v", err)
	}

	drv.mu.Lock()
	last := drv.last
	drv.mu.Unlock()

	// The literal symbol value must NOT appear in the query string.
	if strings.Contains(last.query, "AAPL") {
		t.Errorf("SQL injection risk: literal symbol embedded in query\nquery = %q", last.query)
	}

	// The query must use the $1 placeholder.
	if !strings.Contains(last.query, "$1") {
		t.Errorf("query does not contain parameterized placeholder $1\nquery = %q", last.query)
	}

	// Exactly one argument must be bound.
	if len(last.args) != 1 {
		t.Fatalf("expected 1 bound arg, got %d: %v", len(last.args), last.args)
	}
	if got, want := fmt.Sprintf("%v", last.args[0]), "AAPL"; got != want {
		t.Errorf("bound arg[0] = %q, want %q", got, want)
	}
}

// TestQueryLatestQuote_SQLInjectionPayloadsNotEmbedded verifies that classic
// SQL injection payloads are never spliced into the query string regardless of
// what the caller provides.
func TestQueryLatestQuote_SQLInjectionPayloadsNotEmbedded(t *testing.T) {
	rows := [][]driver.Value{
		{"AAPL", 182.5, int64(200), "XNAS", int64(1700000000000)},
	}

	payloads := []struct {
		name  string
		input string
	}{
		{"single-quote-or", "' OR '1'='1"},
		{"drop-table",      "'; DROP TABLE latest_quotes; --"},
		{"union-select",    "AAPL' UNION SELECT 1,2,3,4,5 --"},
		{"comment-bypass",  "AAPL'--"},
		{"semicolon-stacking", "AAPL'; SELECT * FROM users --"},
	}

	for _, p := range payloads {
		p := p
		t.Run(p.name, func(t *testing.T) {
			db, drv := openMockDB(t, rows)
			defer db.Close()

			// Outcome (error vs success) is irrelevant here; we only care about
			// what reaches the driver.
			FetchLatestQuote(db, p.input) //nolint:errcheck

			drv.mu.Lock()
			last := drv.last
			drv.mu.Unlock()

			// After normalisation the payload is upper-cased and trimmed.
			normalised := strings.ToUpper(strings.TrimSpace(p.input))
			if strings.Contains(last.query, normalised) {
				t.Errorf("injection payload found in SQL query string\npayload = %q\nquery   = %q",
					normalised, last.query)
			}
			// Also check the raw payload.
			if strings.Contains(last.query, p.input) {
				t.Errorf("raw injection payload found in SQL query string\npayload = %q\nquery   = %q",
					p.input, last.query)
			}
		})
	}
}

// TestQueryLatestQuote_ReturnsCorrectQuoteFields verifies that returned Quote
// fields are correctly mapped from the database row.
func TestQueryLatestQuote_ReturnsCorrectQuoteFields(t *testing.T) {
	rows := [][]driver.Value{
		{"MSFT", 415.2, int64(500), "XNAS", int64(1700001000000)},
	}
	db, _ := openMockDB(t, rows)
	defer db.Close()

	q, err := FetchLatestQuote(db, "msft")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if q.Symbol != "MSFT" {
		t.Errorf("Symbol = %q, want %q", q.Symbol, "MSFT")
	}
	if q.LastPrice != 415.2 {
		t.Errorf("LastPrice = %v, want 415.2", q.LastPrice)
	}
	if q.LastSize != int64(500) {
		t.Errorf("LastSize = %v, want 500", q.LastSize)
	}
	if q.VenueCode != "XNAS" {
		t.Errorf("VenueCode = %q, want %q", q.VenueCode, "XNAS")
	}
	if q.IngestedAtMS != int64(1700001000000) {
		t.Errorf("IngestedAtMS = %v, want 1700001000000", q.IngestedAtMS)
	}
}

// TestQueryLatestQuote_NotFoundReturnsError ensures sql.ErrNoRows (or
// equivalent) propagates when the driver returns no rows.
func TestQueryLatestQuote_NotFoundReturnsError(t *testing.T) {
	db, _ := openMockDB(t, nil /* no rows */)
	defer db.Close()

	_, err := FetchLatestQuote(db, "UNKNOWN")
	if err == nil {
		t.Error("expected an error for no rows, got nil")
	}
}

// ---------------------------------------------------------------------------
// normalizeSymbol tests
// ---------------------------------------------------------------------------

func TestNormalizeSymbol(t *testing.T) {
	cases := []struct {
		input string
		want  string
	}{
		{"aapl", "AAPL"},
		{"  msft  ", "MSFT"},
		{"GOOG", "GOOG"},
		{" tsla", "TSLA"},
		{"amzn ", "AMZN"},
		{"", ""},
		{"MiXeD", "MIXED"},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(fmt.Sprintf("input=%q", tc.input), func(t *testing.T) {
			got := normalizeSymbol(tc.input)
			if got != tc.want {
				t.Errorf("normalizeSymbol(%q) = %q, want %q", tc.input, got, tc.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// HTTP handler (LookupTicker) integration tests
// ---------------------------------------------------------------------------

// TestLookupTicker_ValidSymbol exercises the full stack from HTTP request
// through handler → FetchLatestQuote → JSON response.
func TestLookupTicker_ValidSymbol(t *testing.T) {
	rows := [][]driver.Value{
		{"AAPL", 182.5, int64(200), "XNAS", int64(1700000000000)},
	}
	db, _ := openMockDB(t, rows)
	defer db.Close()

	h := NewHandler(db)
	req := httptest.NewRequest(http.MethodGet, "/lookup?symbol=aapl", nil)
	rec := httptest.NewRecorder()
	h.LookupTicker(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body = %s", rec.Code, http.StatusOK, rec.Body.String())
	}

	var q Quote
	if err := json.NewDecoder(rec.Body).Decode(&q); err != nil {
		t.Fatalf("decode response JSON: %v", err)
	}
	if q.Symbol != "AAPL" {
		t.Errorf("Symbol = %q, want AAPL", q.Symbol)
	}
}

// TestLookupTicker_MissingSymbol checks that omitting the ?symbol parameter
// returns 400 Bad Request.
func TestLookupTicker_MissingSymbol(t *testing.T) {
	db, _ := openMockDB(t, nil)
	defer db.Close()

	h := NewHandler(db)
	req := httptest.NewRequest(http.MethodGet, "/lookup", nil)
	rec := httptest.NewRecorder()
	h.LookupTicker(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

// TestLookupTicker_DBError verifies that a database error returns 500 and
// does not leak internal error details in the HTTP response body.
func TestLookupTicker_DBError(t *testing.T) {
	db, drv := openMockDB(t, nil)
	defer db.Close()

	drv.mu.Lock()
	drv.queryErr = fmt.Errorf("connection reset by peer")
	drv.mu.Unlock()

	h := NewHandler(db)
	req := httptest.NewRequest(http.MethodGet, "/lookup?symbol=AAPL", nil)
	rec := httptest.NewRecorder()
	h.LookupTicker(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusInternalServerError)
	}

	// The handler must not expose raw error messages to callers.
	body := rec.Body.String()
	if strings.Contains(body, "connection reset") {
		t.Errorf("HTTP response leaks internal error detail: %q", body)
	}
}

// TestLookupTicker_InjectionPayloadNotInQueryString verifies that a SQL
// injection payload submitted via the HTTP query string never appears in the
// SQL statement sent to the driver.
func TestLookupTicker_InjectionPayloadNotInQueryString(t *testing.T) {
	rows := [][]driver.Value{
		{"AAPL", 182.5, int64(200), "XNAS", int64(1700000000000)},
	}
	db, drv := openMockDB(t, rows)
	defer db.Close()

	// URL-encode the payload so the HTTP test request parses it correctly.
	payload := "' OR '1'='1"
	req := httptest.NewRequest(http.MethodGet, "/lookup?symbol="+payload, nil)
	rec := httptest.NewRecorder()
	h := NewHandler(db)
	h.LookupTicker(rec, req)

	drv.mu.Lock()
	last := drv.last
	drv.mu.Unlock()

	// The raw/normalised payload must not appear inside the SQL string.
	normalised := strings.ToUpper(strings.TrimSpace(payload))
	if strings.Contains(last.query, normalised) {
		t.Errorf("injection payload found in SQL query string\npayload = %q\nquery   = %q",
			normalised, last.query)
	}
}
