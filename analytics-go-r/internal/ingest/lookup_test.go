package ingest

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
)

// ---------------------------------------------------------------------------
// Minimal in-process SQL driver used to verify parameterized query behaviour
// without requiring a real database connection.
// ---------------------------------------------------------------------------

// mockDriver implements database/sql/driver.Driver.
type mockDriver struct {
	mu   sync.Mutex
	conn *mockConn
}

func (d *mockDriver) Open(_ string) (driver.Conn, error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.conn, nil
}

// mockConn implements driver.Conn and captures every query that passes through it.
type mockConn struct {
	mu            sync.Mutex
	capturedQuery string
	capturedArgs  []driver.Value
	rows          *mockRows
	queryErr      error
}

func (c *mockConn) Prepare(query string) (driver.Stmt, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.capturedQuery = query
	return &mockStmt{conn: c, query: query}, nil
}
func (c *mockConn) Close() error  { return nil }
func (c *mockConn) Begin() (driver.Tx, error) { return nil, errors.New("not supported") }

// mockStmt implements driver.Stmt.
type mockStmt struct {
	conn  *mockConn
	query string
}

func (s *mockStmt) Close() error                                    { return nil }
func (s *mockStmt) NumInput() int                                   { return -1 } // variadic
func (s *mockStmt) Exec(_ []driver.Value) (driver.Result, error)   { return nil, errors.New("not supported") }
func (s *mockStmt) Query(args []driver.Value) (driver.Rows, error) {
	s.conn.mu.Lock()
	defer s.conn.mu.Unlock()
	s.conn.capturedArgs = args
	if s.conn.queryErr != nil {
		return nil, s.conn.queryErr
	}
	return s.conn.rows, nil
}

// mockRows implements driver.Rows.
type mockRows struct {
	columns []string
	data    [][]driver.Value
	pos     int
}

func (r *mockRows) Columns() []string { return r.columns }
func (r *mockRows) Close() error      { return nil }
func (r *mockRows) Next(dest []driver.Value) error {
	if r.pos >= len(r.data) {
		return io.EOF
	}
	copy(dest, r.data[r.pos])
	r.pos++
	return nil
}

// ---------------------------------------------------------------------------
// Test helpers
// ---------------------------------------------------------------------------

var (
	driverOnce sync.Once
	drv        *mockDriver
	dsnCounter int
	dsnMu      sync.Mutex
)

// openMockDB registers the mock driver once and returns a *sql.DB backed by
// the given mockConn.
func openMockDB(t *testing.T, conn *mockConn) *sql.DB {
	t.Helper()
	driverOnce.Do(func() {
		drv = &mockDriver{}
		sql.Register("mockdb", drv)
	})

	dsnMu.Lock()
	dsnCounter++
	dsn := strings.Repeat("x", dsnCounter) // unique DSN to bypass sql.DB cache
	dsnMu.Unlock()

	drv.mu.Lock()
	drv.conn = conn
	drv.mu.Unlock()

	db, err := sql.Open("mockdb", dsn)
	if err != nil {
		t.Fatalf("sql.Open: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return db
}

// quoteRows returns a mockRows pre-loaded with a single Quote row.
func quoteRows() *mockRows {
	return &mockRows{
		columns: []string{"symbol", "last_price", "last_size", "venue_code", "ingested_at_ms"},
		data: [][]driver.Value{
			{"AAPL", float64(185.50), int64(100), "XNAS", int64(1700000000000)},
		},
	}
}

// ---------------------------------------------------------------------------
// normalizeSymbol tests
// ---------------------------------------------------------------------------

func TestNormalizeSymbol_UpperCasesInput(t *testing.T) {
	got := normalizeSymbol("aapl")
	if got != "AAPL" {
		t.Errorf("normalizeSymbol(%q) = %q; want %q", "aapl", got, "AAPL")
	}
}

func TestNormalizeSymbol_TrimsWhitespace(t *testing.T) {
	got := normalizeSymbol("  msft  ")
	if got != "MSFT" {
		t.Errorf("normalizeSymbol(%q) = %q; want %q", "  msft  ", got, "MSFT")
	}
}

func TestNormalizeSymbol_AlreadyNormalized(t *testing.T) {
	got := normalizeSymbol("GOOG")
	if got != "GOOG" {
		t.Errorf("normalizeSymbol(%q) = %q; want %q", "GOOG", got, "GOOG")
	}
}

// ---------------------------------------------------------------------------
// SQL-injection prevention tests
//
// These tests verify that the symbol value is passed as a *query parameter*
// ($1) and never concatenated into the SQL string.  They do this by
// capturing both the prepared-statement string and the driver-level argument
// list that the database/sql layer sends to the mock driver.
// ---------------------------------------------------------------------------

// TestQueryLatestQuote_UsesParameterizedQuery is the primary security regression
// test: it asserts that the query string sent to the driver contains the
// placeholder ($1) rather than the literal symbol value.  If the code ever
// reverts to fmt.Sprintf-style concatenation, the symbol will appear inside
// the query string and this test will fail.
func TestQueryLatestQuote_UsesParameterizedQuery(t *testing.T) {
	conn := &mockConn{rows: quoteRows()}
	db := openMockDB(t, conn)

	symbol := "AAPL"
	_, _ = queryLatestQuote(db, symbol)

	conn.mu.Lock()
	capturedQuery := conn.capturedQuery
	capturedArgs := conn.capturedArgs
	conn.mu.Unlock()

	// The literal symbol must NOT appear in the query string.
	if strings.Contains(capturedQuery, symbol) {
		t.Errorf("SQL injection risk: symbol %q found literally in query string: %q", symbol, capturedQuery)
	}

	// The query must contain a positional placeholder.
	if !strings.Contains(capturedQuery, "$1") {
		t.Errorf("expected parameterized placeholder $1 in query, got: %q", capturedQuery)
	}

	// The symbol must be passed as a driver argument, not embedded in SQL.
	if len(capturedArgs) == 0 {
		t.Fatalf("expected at least one driver argument; got none — symbol was probably concatenated into query string")
	}
	if capturedArgs[0] != driver.Value(symbol) {
		t.Errorf("driver arg[0] = %v; want %q", capturedArgs[0], symbol)
	}
}

// TestQueryLatestQuote_SQLInjectionPayload_DoesNotAlterQuery ensures a classic
// SQL-injection payload (' OR '1'='1) is transmitted to the driver as a plain
// string argument and does not mutate the SQL statement itself.
func TestQueryLatestQuote_SQLInjectionPayload_DoesNotAlterQuery(t *testing.T) {
	conn := &mockConn{rows: quoteRows()}
	db := openMockDB(t, conn)

	// Simulate a malicious value that a real attacker might supply via the
	// ?symbol= query parameter.
	malicious := "' OR '1'='1"
	_, _ = queryLatestQuote(db, malicious)

	conn.mu.Lock()
	capturedQuery := conn.capturedQuery
	capturedArgs := conn.capturedArgs
	conn.mu.Unlock()

	// The injected SQL fragment must NOT appear in the prepared statement.
	if strings.Contains(capturedQuery, "OR") {
		t.Errorf("SQL injection not prevented: injected OR clause found in query: %q", capturedQuery)
	}

	// The payload must still be present as a bound parameter (so the query
	// returns no rows rather than silently eating the input).
	if len(capturedArgs) == 0 || capturedArgs[0] != driver.Value(malicious) {
		t.Errorf("expected malicious payload as driver arg[0], got args=%v", capturedArgs)
	}
}

// TestQueryLatestQuote_UnionInjectionPayload checks a UNION-based exfiltration
// attempt (a common variant) is also treated as a plain argument.
func TestQueryLatestQuote_UnionInjectionPayload_DoesNotAlterQuery(t *testing.T) {
	conn := &mockConn{rows: quoteRows()}
	db := openMockDB(t, conn)

	malicious := "AAPL' UNION SELECT username,password,1,1,1 FROM users--"
	_, _ = queryLatestQuote(db, malicious)

	conn.mu.Lock()
	capturedQuery := conn.capturedQuery
	capturedArgs := conn.capturedArgs
	conn.mu.Unlock()

	if strings.Contains(strings.ToUpper(capturedQuery), "UNION") {
		t.Errorf("UNION injection not prevented: UNION keyword found in query: %q", capturedQuery)
	}
	if len(capturedArgs) == 0 || capturedArgs[0] != driver.Value(malicious) {
		t.Errorf("expected malicious payload as driver arg[0], got args=%v", capturedArgs)
	}
}

// ---------------------------------------------------------------------------
// FetchLatestQuote happy-path / error tests
// ---------------------------------------------------------------------------

func TestFetchLatestQuote_ReturnsQuote(t *testing.T) {
	conn := &mockConn{rows: quoteRows()}
	db := openMockDB(t, conn)

	q, err := FetchLatestQuote(db, "aapl") // intentionally lower-case
	if err != nil {
		t.Fatalf("FetchLatestQuote returned unexpected error: %v", err)
	}
	if q.Symbol != "AAPL" {
		t.Errorf("q.Symbol = %q; want %q", q.Symbol, "AAPL")
	}
	if q.LastPrice != 185.50 {
		t.Errorf("q.LastPrice = %v; want 185.50", q.LastPrice)
	}
	if q.LastSize != 100 {
		t.Errorf("q.LastSize = %v; want 100", q.LastSize)
	}
}

func TestFetchLatestQuote_PropagatesDBError(t *testing.T) {
	conn := &mockConn{queryErr: errors.New("connection reset")}
	db := openMockDB(t, conn)

	_, err := FetchLatestQuote(db, "AAPL")
	if err == nil {
		t.Error("expected an error when the DB returns an error; got nil")
	}
}

func TestFetchLatestQuote_NotFound(t *testing.T) {
	conn := &mockConn{
		// Empty rows → Scan returns sql.ErrNoRows.
		rows: &mockRows{
			columns: []string{"symbol", "last_price", "last_size", "venue_code", "ingested_at_ms"},
		},
	}
	db := openMockDB(t, conn)

	_, err := FetchLatestQuote(db, "UNKNOWN")
	if !errors.Is(err, sql.ErrNoRows) {
		t.Errorf("expected sql.ErrNoRows; got %v", err)
	}
}

// ---------------------------------------------------------------------------
// HTTP handler tests (LookupTicker)
// ---------------------------------------------------------------------------

func TestLookupTicker_MissingSymbolParameter(t *testing.T) {
	conn := &mockConn{rows: quoteRows()}
	db := openMockDB(t, conn)
	h := NewHandler(db)

	req := httptest.NewRequest(http.MethodGet, "/ticker", nil)
	rec := httptest.NewRecorder()
	h.LookupTicker(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400; got %d", rec.Code)
	}
}

func TestLookupTicker_ReturnsJSONQuote(t *testing.T) {
	conn := &mockConn{rows: quoteRows()}
	db := openMockDB(t, conn)
	h := NewHandler(db)

	req := httptest.NewRequest(http.MethodGet, "/ticker?symbol=AAPL", nil)
	rec := httptest.NewRecorder()
	h.LookupTicker(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200; got %d — body: %s", rec.Code, rec.Body.String())
	}

	var q Quote
	if err := json.NewDecoder(rec.Body).Decode(&q); err != nil {
		t.Fatalf("failed to decode response JSON: %v", err)
	}
	if q.Symbol != "AAPL" {
		t.Errorf("q.Symbol = %q; want AAPL", q.Symbol)
	}
}

func TestLookupTicker_DBError_Returns500(t *testing.T) {
	conn := &mockConn{queryErr: errors.New("db unavailable")}
	db := openMockDB(t, conn)
	h := NewHandler(db)

	req := httptest.NewRequest(http.MethodGet, "/ticker?symbol=AAPL", nil)
	rec := httptest.NewRecorder()
	h.LookupTicker(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("expected status 500; got %d", rec.Code)
	}
}

// TestLookupTicker_InjectionPayloadNotEmbedded is an end-to-end security test
// that starts from the HTTP layer.  It verifies that a malicious symbol value
// supplied via the query string is bound as a parameterized argument and never
// appears verbatim inside the prepared SQL statement.
func TestLookupTicker_InjectionPayloadNotEmbedded(t *testing.T) {
	conn := &mockConn{rows: quoteRows()}
	db := openMockDB(t, conn)
	h := NewHandler(db)

	payload := "' OR 1=1--"
	req := httptest.NewRequest(http.MethodGet, "/ticker?symbol="+payload, nil)
	rec := httptest.NewRecorder()
	h.LookupTicker(rec, req)

	conn.mu.Lock()
	capturedQuery := conn.capturedQuery
	conn.mu.Unlock()

	// The injected fragment must not appear in the SQL text.
	if strings.Contains(capturedQuery, "OR") {
		t.Errorf("HTTP injection not prevented: OR keyword found in prepared statement: %q", capturedQuery)
	}
	if strings.Contains(capturedQuery, "--") {
		t.Errorf("HTTP injection not prevented: comment marker found in prepared statement: %q", capturedQuery)
	}
}
