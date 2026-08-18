package ingest

import (
	"database/sql"
	"database/sql/driver"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// ---------------------------------------------------------------------------
// Minimal sql/driver mock – no external dependencies required
// ---------------------------------------------------------------------------

// mockDriver registers a fake "mockdb" driver that returns rows supplied at
// construction time. Each test installs a fresh mockDB so rows can be
// pre-loaded without touching a real database.
type mockDriver struct{}

func (mockDriver) Open(name string) (driver.Conn, error) {
	return mockConnRegistry[name], nil
}

// mockConnRegistry maps DSN strings to pre-built connections.
var mockConnRegistry = map[string]*mockConn{}

func init() {
	sql.Register("mockdb", mockDriver{})
}

// openMockDB registers a new mock connection under a unique DSN and returns an
// *sql.DB backed by it.
func openMockDB(name string, rows [][]driver.Value, cols []string, queryErr error) *sql.DB {
	mockConnRegistry[name] = &mockConn{
		rows:     rows,
		cols:     cols,
		queryErr: queryErr,
	}
	db, _ := sql.Open("mockdb", name)
	return db
}

// mockConn implements driver.Conn, driver.QueryerContext, and driver.Queryer.
type mockConn struct {
	rows     [][]driver.Value
	cols     []string
	queryErr error
	// lastQuery captures the literal SQL text sent to the driver; tests can
	// inspect this to confirm that no user-supplied values appear in the query
	// string (i.e., the query is parameterized).
	lastQuery string
	// lastArgs captures the bind arguments sent alongside the query.
	lastArgs []driver.Value
}

func (c *mockConn) Prepare(query string) (driver.Stmt, error) {
	c.lastQuery = query
	return &mockStmt{conn: c}, nil
}
func (c *mockConn) Close() error  { return nil }
func (c *mockConn) Begin() (driver.Tx, error) { return nil, errors.New("not supported") }

type mockStmt struct{ conn *mockConn }

func (s *mockStmt) Close() error { return nil }
func (s *mockStmt) NumInput() int { return -1 } // variadic: driver accepts any count

func (s *mockStmt) Exec(args []driver.Value) (driver.Result, error) {
	return nil, errors.New("not supported")
}

func (s *mockStmt) Query(args []driver.Value) (driver.Rows, error) {
	s.conn.lastArgs = args
	if s.conn.queryErr != nil {
		return nil, s.conn.queryErr
	}
	return &mockRows{
		cols: s.conn.cols,
		rows: s.conn.rows,
		pos:  0,
	}, nil
}

type mockRows struct {
	cols []string
	rows [][]driver.Value
	pos  int
}

func (r *mockRows) Columns() []string { return r.cols }
func (r *mockRows) Close() error      { return nil }

func (r *mockRows) Next(dest []driver.Value) error {
	if r.pos >= len(r.rows) {
		return io.EOF
	}
	copy(dest, r.rows[r.pos])
	r.pos++
	return nil
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
		{"TSLA", "TSLA"},
		{"  ", ""},
		{"", ""},
		{"gOoG", "GOOG"},
	}
	for _, tc := range cases {
		got := normalizeSymbol(tc.input)
		if got != tc.want {
			t.Errorf("normalizeSymbol(%q) = %q; want %q", tc.input, got, tc.want)
		}
	}
}

// ---------------------------------------------------------------------------
// queryLatestQuote – parameterized query tests
// ---------------------------------------------------------------------------

const lookupCols = "symbol,last_price,last_size,venue_code,ingested_at_ms"

// quoteColumns returns the slice expected by mockRows.
func quoteColumns() []string {
	return strings.Split(lookupCols, ",")
}

// TestQueryLatestQuote_ParameterizedQuery is the core security regression test.
// It verifies that the SQL query string sent to the driver NEVER contains the
// symbol value – the symbol must travel as a bind argument, not embedded text.
func TestQueryLatestQuote_ParameterizedQuery(t *testing.T) {
	// An SQL-injection payload that would break a naively concatenated query.
	maliciousPayload := "' OR '1'='1"

	rows := [][]driver.Value{
		{"SAFE", float64(100.0), int64(500), "NYSE", int64(1700000000000)},
	}
	conn := openMockDB("test-parameterized", rows, quoteColumns(), nil)
	defer conn.Close()

	_, _ = queryLatestQuote(conn, maliciousPayload)

	mc := mockConnRegistry["test-parameterized"]

	// The raw SQL string must NOT contain the tainted value.
	if strings.Contains(mc.lastQuery, maliciousPayload) {
		t.Errorf("SQL injection: user input found verbatim in query string: %q", mc.lastQuery)
	}

	// The bind argument list must contain exactly one entry with the payload.
	if len(mc.lastArgs) == 0 {
		t.Errorf("expected at least one bind argument; got none – query may be non-parameterized")
	}
	if len(mc.lastArgs) > 0 {
		got, ok := mc.lastArgs[0].(string)
		if !ok || got != maliciousPayload {
			t.Errorf("expected bind arg[0] = %q; got %v", maliciousPayload, mc.lastArgs[0])
		}
	}
}

// TestQueryLatestQuote_ReturnsQuote checks that a normal lookup returns the
// expected Quote fields when a matching row exists.
func TestQueryLatestQuote_ReturnsQuote(t *testing.T) {
	rows := [][]driver.Value{
		{"AAPL", float64(182.50), int64(1000), "NASDAQ", int64(1700000000001)},
	}
	db := openMockDB("test-returns-quote", rows, quoteColumns(), nil)
	defer db.Close()

	q, err := queryLatestQuote(db, "AAPL")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if q.Symbol != "AAPL" {
		t.Errorf("Symbol = %q; want %q", q.Symbol, "AAPL")
	}
	if q.LastPrice != 182.50 {
		t.Errorf("LastPrice = %v; want 182.50", q.LastPrice)
	}
	if q.LastSize != 1000 {
		t.Errorf("LastSize = %v; want 1000", q.LastSize)
	}
	if q.VenueCode != "NASDAQ" {
		t.Errorf("VenueCode = %q; want NASDAQ", q.VenueCode)
	}
	if q.IngestedAtMS != 1700000000001 {
		t.Errorf("IngestedAtMS = %v; want 1700000000001", q.IngestedAtMS)
	}
}

// TestQueryLatestQuote_NotFound checks that sql.ErrNoRows propagates when the
// symbol has no matching row.
func TestQueryLatestQuote_NotFound(t *testing.T) {
	// No rows – the driver will immediately return io.EOF on Next().
	db := openMockDB("test-not-found", nil, quoteColumns(), nil)
	defer db.Close()

	_, err := queryLatestQuote(db, "UNKNOWN")
	if err == nil {
		t.Error("expected an error for missing row; got nil")
	}
}

// TestQueryLatestQuote_DBError checks that a driver-level error propagates.
func TestQueryLatestQuote_DBError(t *testing.T) {
	dbErr := errors.New("connection reset by peer")
	db := openMockDB("test-db-error", nil, quoteColumns(), dbErr)
	defer db.Close()

	_, err := queryLatestQuote(db, "AAPL")
	if err == nil {
		t.Error("expected an error; got nil")
	}
}

// ---------------------------------------------------------------------------
// FetchLatestQuote – integration of normalizeSymbol + queryLatestQuote
// ---------------------------------------------------------------------------

// TestFetchLatestQuote_NormalizesSymbol confirms that FetchLatestQuote
// upper-cases and trims the raw symbol before executing the query, and that
// the normalized form is used as the bind argument (not the raw form).
func TestFetchLatestQuote_NormalizesSymbol(t *testing.T) {
	rows := [][]driver.Value{
		{"MSFT", float64(300.0), int64(200), "NASDAQ", int64(1700000000002)},
	}
	db := openMockDB("test-normalizes", rows, quoteColumns(), nil)
	defer db.Close()

	q, err := FetchLatestQuote(db, "  msft  ")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if q.Symbol != "MSFT" {
		t.Errorf("Symbol = %q; want MSFT", q.Symbol)
	}

	mc := mockConnRegistry["test-normalizes"]
	// The bind argument must be the normalized form.
	if len(mc.lastArgs) > 0 {
		got, _ := mc.lastArgs[0].(string)
		if got != "MSFT" {
			t.Errorf("bind arg[0] = %q; want \"MSFT\" (normalized)", got)
		}
	}
}

// ---------------------------------------------------------------------------
// LookupTicker HTTP handler tests
// ---------------------------------------------------------------------------

// TestLookupTicker_MissingSymbol validates that the HTTP handler returns 400
// when no symbol query parameter is provided.
func TestLookupTicker_MissingSymbol(t *testing.T) {
	db := openMockDB("test-http-missing", nil, quoteColumns(), nil)
	defer db.Close()

	h := NewHandler(db)
	req := httptest.NewRequest(http.MethodGet, "/market-data/lookup", nil)
	w := httptest.NewRecorder()

	h.LookupTicker(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d; want %d", w.Code, http.StatusBadRequest)
	}
}

// TestLookupTicker_ValidSymbol validates that the handler returns 200 and a
// JSON body when the symbol resolves to a known quote.
func TestLookupTicker_ValidSymbol(t *testing.T) {
	rows := [][]driver.Value{
		{"GOOGL", float64(140.0), int64(300), "NASDAQ", int64(1700000000003)},
	}
	db := openMockDB("test-http-valid", rows, quoteColumns(), nil)
	defer db.Close()

	h := NewHandler(db)
	req := httptest.NewRequest(http.MethodGet, "/market-data/lookup?symbol=googl", nil)
	w := httptest.NewRecorder()

	h.LookupTicker(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d; want %d", w.Code, http.StatusOK)
	}
	ct := w.Header().Get("Content-Type")
	if !strings.HasPrefix(ct, "application/json") {
		t.Errorf("Content-Type = %q; want application/json", ct)
	}
	if !strings.Contains(w.Body.String(), "GOOGL") {
		t.Errorf("response body does not contain expected symbol; body: %s", w.Body.String())
	}
}

// TestLookupTicker_SQLInjectionPayload is the end-to-end security test.
// It sends a crafted SQL injection string as the ?symbol= parameter and
// confirms the query string reaching the driver does not contain the payload
// verbatim (i.e., the handler uses parameterized queries throughout).
func TestLookupTicker_SQLInjectionPayload(t *testing.T) {
	// Classic tautology injection and UNION-based exfiltration attempt.
	attackPayloads := []string{
		"' OR '1'='1",
		"'; DROP TABLE latest_quotes; --",
		"' UNION SELECT username, password, 1, 1, 1 FROM users --",
		"AAPL' AND SLEEP(5) --",
	}

	for _, payload := range attackPayloads {
		name := "test-sqli-" + payload[:4] // unique per test
		db := openMockDB(name, nil, quoteColumns(), nil)

		h := NewHandler(db)
		req := httptest.NewRequest(http.MethodGet, "/market-data/lookup?symbol="+payload, nil)
		w := httptest.NewRecorder()

		h.LookupTicker(w, req)

		mc := mockConnRegistry[name]
		if mc != nil && mc.lastQuery != "" {
			if strings.Contains(mc.lastQuery, payload) {
				t.Errorf("SQL injection payload found verbatim in query string for payload %q: %q",
					payload, mc.lastQuery)
			}
		}
		db.Close()
	}
}
