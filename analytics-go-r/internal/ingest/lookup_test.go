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

// ----------------------------------------------------------------------------
// Minimal in-process SQL driver used only by these tests.
// It records the query string and args supplied to QueryRow so we can assert
// that the production code uses parameterized queries instead of interpolating
// user input directly into the SQL text.
// ----------------------------------------------------------------------------

type recordingDriver struct{}

func (recordingDriver) Open(name string) (driver.Conn, error) {
	return &recordingConn{}, nil
}

type recordingConn struct {
	lastQuery string
	lastArgs  []driver.Value
	// rows to return on the next query (nil == sql.ErrNoRows)
	rows [][]driver.Value
}

func (c *recordingConn) Prepare(query string) (driver.Stmt, error) {
	c.lastQuery = query
	return &recordingStmt{conn: c, query: query}, nil
}

func (c *recordingConn) Close() error  { return nil }
func (c *recordingConn) Begin() (driver.Tx, error) { return nil, errors.New("not supported") }

type recordingStmt struct {
	conn  *recordingConn
	query string
}

func (s *recordingStmt) Close() error { return nil }
func (s *recordingStmt) NumInput() int { return -1 } // -1 means the driver doesn't check

func (s *recordingStmt) Exec(args []driver.Value) (driver.Result, error) {
	return nil, errors.New("not supported")
}

func (s *recordingStmt) Query(args []driver.Value) (driver.Rows, error) {
	s.conn.lastArgs = args
	return &recordingRows{rows: s.conn.rows}, nil
}

type recordingRows struct {
	rows [][]driver.Value
	pos  int
}

func (r *recordingRows) Columns() []string {
	return []string{"symbol", "last_price", "last_size", "venue_code", "ingested_at_ms"}
}

func (r *recordingRows) Close() error { return nil }

func (r *recordingRows) Next(dest []driver.Value) error {
	if r.pos >= len(r.rows) {
		return io.EOF
	}
	copy(dest, r.rows[r.pos])
	r.pos++
	return nil
}

// openTestDB registers a fresh recordingDriver instance under a unique name
// and returns both the *sql.DB and the underlying *recordingConn so tests can
// inspect what was sent to the driver.
var driverSeq int

func openTestDB(t *testing.T, rows [][]driver.Value) (*sql.DB, *recordingConn) {
	t.Helper()
	driverSeq++
	name := strings.Repeat("x", driverSeq) // unique driver name per call
	conn := &recordingConn{rows: rows}
	drv := &staticDriver{conn: conn}
	sql.Register(name, drv)
	db, err := sql.Open(name, "")
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return db, conn
}

type staticDriver struct{ conn *recordingConn }

func (d *staticDriver) Open(_ string) (driver.Conn, error) { return d.conn, nil }

// ----------------------------------------------------------------------------
// normalizeSymbol
// ----------------------------------------------------------------------------

func TestNormalizeSymbol_UpperCasesInput(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"aapl", "AAPL"},
		{"Msft", "MSFT"},
		{"GOOG", "GOOG"},
		{"  tsla  ", "TSLA"},
		{"\tAmzn\n", "AMZN"},
		{"", ""},
	}
	for _, tc := range tests {
		got := normalizeSymbol(tc.input)
		if got != tc.want {
			t.Errorf("normalizeSymbol(%q) = %q; want %q", tc.input, got, tc.want)
		}
	}
}

// ----------------------------------------------------------------------------
// queryLatestQuote – parameterized query assertion (SQL-injection regression)
// ----------------------------------------------------------------------------

// TestQueryLatestQuote_UsesParameterizedQuery verifies that the SQL sent to the
// driver contains a placeholder ($1) instead of the literal symbol value, and
// that the symbol is passed as a separate bind argument.  This is the direct
// regression test for CVE-89 / CWE-89 SQL injection (the vulnerability was
// caused by fmt.Sprintf interpolating the symbol into the query string).
func TestQueryLatestQuote_UsesParameterizedQuery(t *testing.T) {
	symbol := "AAPL"
	wantRows := [][]driver.Value{
		{symbol, float64(150.0), int64(100), "XNYS", int64(1700000000000)},
	}
	db, conn := openTestDB(t, wantRows)

	_, _ = queryLatestQuote(db, symbol)

	// The query text must NOT contain the literal symbol value.
	if strings.Contains(conn.lastQuery, symbol) {
		t.Errorf("query contains literal symbol value %q – parameterized query not used; query: %s",
			symbol, conn.lastQuery)
	}

	// The query text must contain a placeholder token.
	if !strings.Contains(conn.lastQuery, "$1") {
		t.Errorf("query does not contain placeholder $1; query: %s", conn.lastQuery)
	}

	// The symbol must be passed as the first bind argument.
	if len(conn.lastArgs) == 0 {
		t.Fatal("no bind arguments were passed to the driver")
	}
	if got := conn.lastArgs[0]; got != symbol {
		t.Errorf("first bind argument = %v; want %q", got, symbol)
	}
}

// TestQueryLatestQuote_SQLInjectionPayloadNotInterpolated ensures that a classic
// SQL injection payload ends up as a bind argument rather than being spliced
// into the SQL text, making it harmless.
func TestQueryLatestQuote_SQLInjectionPayloadNotInterpolated(t *testing.T) {
	// Classic tautology injection that would bypass a WHERE clause if interpolated.
	payload := "' OR '1'='1"
	db, conn := openTestDB(t, nil) // no rows – we only care about the query shape

	_, _ = queryLatestQuote(db, payload)

	if strings.Contains(conn.lastQuery, payload) {
		t.Errorf("SQL injection payload was interpolated into the query string; query: %s", conn.lastQuery)
	}
	if len(conn.lastArgs) == 0 || conn.lastArgs[0] != payload {
		t.Errorf("payload was not passed as a bind argument; args: %v", conn.lastArgs)
	}
}

// TestQueryLatestQuote_DropTablePayloadNotInterpolated tests a destructive
// payload to confirm it is never concatenated into the SQL text.
func TestQueryLatestQuote_DropTablePayloadNotInterpolated(t *testing.T) {
	payload := "'; DROP TABLE latest_quotes; --"
	db, conn := openTestDB(t, nil)

	_, _ = queryLatestQuote(db, payload)

	if strings.Contains(conn.lastQuery, "DROP") {
		t.Errorf("destructive SQL was interpolated into the query string; query: %s", conn.lastQuery)
	}
	if len(conn.lastArgs) == 0 || conn.lastArgs[0] != payload {
		t.Errorf("payload was not passed as a bind argument; args: %v", conn.lastArgs)
	}
}

// TestQueryLatestQuote_ReturnsQuoteOnSuccess verifies that a well-formed row
// is correctly scanned into a Quote struct.
func TestQueryLatestQuote_ReturnsQuoteOnSuccess(t *testing.T) {
	wantRows := [][]driver.Value{
		{"AAPL", float64(182.50), int64(200), "XNAS", int64(1700000000001)},
	}
	db, _ := openTestDB(t, wantRows)

	quote, err := queryLatestQuote(db, "AAPL")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if quote.Symbol != "AAPL" {
		t.Errorf("Symbol = %q; want AAPL", quote.Symbol)
	}
	if quote.LastPrice != 182.50 {
		t.Errorf("LastPrice = %v; want 182.50", quote.LastPrice)
	}
	if quote.LastSize != 200 {
		t.Errorf("LastSize = %v; want 200", quote.LastSize)
	}
	if quote.VenueCode != "XNAS" {
		t.Errorf("VenueCode = %q; want XNAS", quote.VenueCode)
	}
	if quote.IngestedAtMS != 1700000000001 {
		t.Errorf("IngestedAtMS = %v; want 1700000000001", quote.IngestedAtMS)
	}
}

// TestQueryLatestQuote_ReturnsErrorWhenNoRows verifies that sql.ErrNoRows
// propagates back to the caller when the symbol is not found.
func TestQueryLatestQuote_ReturnsErrorWhenNoRows(t *testing.T) {
	db, _ := openTestDB(t, nil) // nil rows → driver returns io.EOF → sql.ErrNoRows

	_, err := queryLatestQuote(db, "UNKNOWN")
	if !errors.Is(err, sql.ErrNoRows) {
		t.Errorf("error = %v; want sql.ErrNoRows", err)
	}
}

// ----------------------------------------------------------------------------
// FetchLatestQuote – integration of normalizeSymbol + queryLatestQuote
// ----------------------------------------------------------------------------

// TestFetchLatestQuote_NormalizesSymbolBeforeQuery confirms that lower-case and
// whitespace-padded symbols are normalized to upper-case before hitting the DB.
func TestFetchLatestQuote_NormalizesSymbolBeforeQuery(t *testing.T) {
	wantRows := [][]driver.Value{
		{"AAPL", float64(150.0), int64(100), "XNYS", int64(1700000000000)},
	}
	db, conn := openTestDB(t, wantRows)

	_, _ = FetchLatestQuote(db, "  aapl  ")

	if len(conn.lastArgs) == 0 {
		t.Fatal("no bind arguments were passed to the driver")
	}
	if conn.lastArgs[0] != "AAPL" {
		t.Errorf("bind arg = %v; want \"AAPL\" (symbol should be normalized)", conn.lastArgs[0])
	}
}

// ----------------------------------------------------------------------------
// LookupTicker HTTP handler
// ----------------------------------------------------------------------------

// TestLookupTicker_MissingSymbolReturnsBadRequest exercises the guard that
// rejects requests without a symbol query parameter.
func TestLookupTicker_MissingSymbolReturnsBadRequest(t *testing.T) {
	db, _ := openTestDB(t, nil)
	h := NewHandler(db)

	req := httptest.NewRequest(http.MethodGet, "/lookup", nil)
	rr := httptest.NewRecorder()

	h.LookupTicker(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("status = %d; want %d", rr.Code, http.StatusBadRequest)
	}
}

// TestLookupTicker_SymbolNotFoundReturnsInternalServerError verifies that a
// missing quote results in a 500 rather than a panic or information leak.
func TestLookupTicker_SymbolNotFoundReturnsInternalServerError(t *testing.T) {
	db, _ := openTestDB(t, nil) // no rows
	h := NewHandler(db)

	req := httptest.NewRequest(http.MethodGet, "/lookup?symbol=UNKNOWN", nil)
	rr := httptest.NewRecorder()

	h.LookupTicker(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("status = %d; want %d", rr.Code, http.StatusInternalServerError)
	}
}

// TestLookupTicker_ValidSymbolReturnsJSON confirms the happy-path response
// shape and Content-Type.
func TestLookupTicker_ValidSymbolReturnsJSON(t *testing.T) {
	wantRows := [][]driver.Value{
		{"TSLA", float64(240.0), int64(50), "XNYS", int64(1700000000002)},
	}
	db, _ := openTestDB(t, wantRows)
	h := NewHandler(db)

	req := httptest.NewRequest(http.MethodGet, "/lookup?symbol=TSLA", nil)
	rr := httptest.NewRecorder()

	h.LookupTicker(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("status = %d; want %d", rr.Code, http.StatusOK)
	}
	ct := rr.Header().Get("Content-Type")
	if !strings.HasPrefix(ct, "application/json") {
		t.Errorf("Content-Type = %q; want application/json", ct)
	}
	body := rr.Body.String()
	if !strings.Contains(body, "TSLA") {
		t.Errorf("response body does not contain expected symbol; body: %s", body)
	}
}

// TestLookupTicker_SQLInjectionViaHTTPIsInert submits a SQL injection payload
// through the HTTP query parameter and asserts it is treated as data (bound
// argument), not code (interpolated SQL text).
func TestLookupTicker_SQLInjectionViaHTTPIsInert(t *testing.T) {
	db, conn := openTestDB(t, nil)
	h := NewHandler(db)

	// URL-encode the payload as a browser/curl would; net/http decodes it before
	// placing it in Query().
	req := httptest.NewRequest(http.MethodGet, "/lookup?symbol=%27+OR+%271%27%3D%271", nil)
	rr := httptest.NewRecorder()

	h.LookupTicker(rr, req)

	// We expect a 500 (no rows found for the garbage symbol) – not a 200 that
	// would indicate the injection bypassed the WHERE clause.
	if rr.Code == http.StatusOK {
		t.Error("injection payload returned HTTP 200 – possible SQL injection bypass")
	}

	// The query text must not contain the raw SQL fragments from the payload.
	if strings.Contains(conn.lastQuery, "OR") && strings.Contains(conn.lastQuery, "1=1") {
		t.Errorf("SQL injection fragments found in query text; query: %s", conn.lastQuery)
	}
}
