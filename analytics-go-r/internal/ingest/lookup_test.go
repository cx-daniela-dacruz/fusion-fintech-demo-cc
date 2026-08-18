package ingest

import (
	"database/sql"
	"database/sql/driver"
	"errors"
	"io"
	"testing"
)

// ---------------------------------------------------------------------------
// Minimal sql/driver mock so tests do not require a real database.
// ---------------------------------------------------------------------------

// mockDriver implements database/sql/driver.Driver and returns a
// pre-configured mockConn each time Open is called.
type mockDriver struct {
	conn *mockConn
}

func (d *mockDriver) Open(_ string) (driver.Conn, error) { return d.conn, nil }

// mockConn implements driver.Conn and captures every Query call so tests can
// inspect the SQL text and arguments that were actually sent to the driver.
type mockConn struct {
	lastQuery string
	lastArgs  []driver.NamedValue

	// rows to return from the next Query
	rows *mockRows
	// err to return from the next Query (simulates DB errors)
	queryErr error
}

func (c *mockConn) Prepare(query string) (driver.Stmt, error) {
	return &mockStmt{conn: c, query: query}, nil
}
func (c *mockConn) Close() error              { return nil }
func (c *mockConn) Begin() (driver.Tx, error) { return nil, errors.New("not supported") }

// mockStmt implements driver.Stmt and driver.StmtQueryContext.
type mockStmt struct {
	conn  *mockConn
	query string
}

func (s *mockStmt) Close() error                                    { return nil }
func (s *mockStmt) NumInput() int                                   { return -1 }
func (s *mockStmt) Exec(_ []driver.Value) (driver.Result, error)   { return nil, errors.New("not supported") }
func (s *mockStmt) Query(args []driver.Value) (driver.Rows, error) {
	// Convert []driver.Value to []driver.NamedValue for uniform capture.
	named := make([]driver.NamedValue, len(args))
	for i, v := range args {
		named[i] = driver.NamedValue{Ordinal: i + 1, Value: v}
	}
	s.conn.lastQuery = s.query
	s.conn.lastArgs = named
	if s.conn.queryErr != nil {
		return nil, s.conn.queryErr
	}
	if s.conn.rows == nil {
		return &mockRows{}, nil
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

// newTestDB registers a fresh mockDriver and returns the *sql.DB that uses it.
func newTestDB(t *testing.T, conn *mockConn) *sql.DB {
	t.Helper()
	driverName := "mock_" + t.Name()
	sql.Register(driverName, &mockDriver{conn: conn})
	db, err := sql.Open(driverName, "mock://")
	if err != nil {
		t.Fatalf("sql.Open: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return db
}

// ---------------------------------------------------------------------------
// Tests for normalizeSymbol
// ---------------------------------------------------------------------------

func TestNormalizeSymbol(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{name: "already uppercase", input: "AAPL", want: "AAPL"},
		{name: "lowercase", input: "aapl", want: "AAPL"},
		{name: "mixed case", input: "aApL", want: "AAPL"},
		{name: "leading spaces", input: "  AAPL", want: "AAPL"},
		{name: "trailing spaces", input: "AAPL  ", want: "AAPL"},
		{name: "surrounding spaces and lower", input: "  aapl  ", want: "AAPL"},
		{name: "empty string", input: "", want: ""},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := normalizeSymbol(tc.input)
			if got != tc.want {
				t.Errorf("normalizeSymbol(%q) = %q; want %q", tc.input, got, tc.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Tests for queryLatestQuote — parameterized query (SQL injection fix)
// ---------------------------------------------------------------------------

// quoteColumns lists the columns in the order row.Scan expects them.
var quoteColumns = []string{"symbol", "last_price", "last_size", "venue_code", "ingested_at_ms"}

// TestQueryLatestQuote_UsesParameterizedQuery verifies that the SQL sent to
// the driver contains a placeholder ($1) instead of the literal symbol value
// embedded in the query string. This is the core regression test for the SQL
// injection remediation.
func TestQueryLatestQuote_UsesParameterizedQuery(t *testing.T) {
	symbol := "AAPL"

	conn := &mockConn{
		rows: &mockRows{
			columns: quoteColumns,
			data: [][]driver.Value{
				{"AAPL", float64(150.25), int64(100), "XNAS", int64(1700000000000)},
			},
		},
	}
	db := newTestDB(t, conn)

	_, err := queryLatestQuote(db, symbol)
	if err != nil {
		t.Fatalf("queryLatestQuote returned unexpected error: %v", err)
	}

	// The query must NOT contain the literal symbol value.
	if contains(conn.lastQuery, symbol) {
		t.Errorf("SQL query contains the literal symbol %q — parameterization is broken.\nQuery: %s", symbol, conn.lastQuery)
	}

	// The query must contain a parameter placeholder ($1 for PostgreSQL).
	if !contains(conn.lastQuery, "$1") {
		t.Errorf("SQL query does not contain a parameter placeholder ($1).\nQuery: %s", conn.lastQuery)
	}

	// The symbol must be passed as a bind argument, not interpolated.
	if len(conn.lastArgs) == 0 {
		t.Fatalf("no bind arguments were passed to the driver; the symbol must be parameterized")
	}
	if conn.lastArgs[0].Value != symbol {
		t.Errorf("first bind argument = %v; want %q", conn.lastArgs[0].Value, symbol)
	}
}

// TestQueryLatestQuote_SQLInjectionPayloadIsNotExecuted verifies that a
// classic SQL injection payload is passed as a bind parameter rather than
// being concatenated into the query string.
func TestQueryLatestQuote_SQLInjectionPayloadIsNotExecuted(t *testing.T) {
	// Classic UNION-based SQL injection attempt.
	injectionPayloads := []string{
		"' OR '1'='1",
		"'; DROP TABLE latest_quotes; --",
		"AAPL' UNION SELECT username,password,1,1,1 FROM users --",
		"1' OR 1=1 --",
	}

	for _, payload := range injectionPayloads {
		t.Run("payload="+payload, func(t *testing.T) {
			conn := &mockConn{queryErr: sql.ErrNoRows}
			db := newTestDB(t, conn)

			// We expect an error (no rows) but not a panic or query manipulation.
			_, _ = queryLatestQuote(db, payload)

			// The raw payload must NOT appear verbatim in the constructed query.
			if contains(conn.lastQuery, payload) {
				t.Errorf("SQL injection payload was interpolated into the query string.\nPayload: %q\nQuery: %s", payload, conn.lastQuery)
			}

			// The payload must be passed as a bind argument.
			if len(conn.lastArgs) == 0 {
				t.Errorf("no bind arguments were recorded; payload was not parameterized")
				return
			}
			if conn.lastArgs[0].Value != payload {
				t.Errorf("first bind argument = %v; want %q", conn.lastArgs[0].Value, payload)
			}
		})
	}
}

// TestQueryLatestQuote_ReturnsCorrectQuote checks that a valid quote is
// decoded correctly from a mocked database row.
func TestQueryLatestQuote_ReturnsCorrectQuote(t *testing.T) {
	conn := &mockConn{
		rows: &mockRows{
			columns: quoteColumns,
			data: [][]driver.Value{
				{"MSFT", float64(310.50), int64(200), "XNAS", int64(1700000001234)},
			},
		},
	}
	db := newTestDB(t, conn)

	q, err := queryLatestQuote(db, "MSFT")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if q.Symbol != "MSFT" {
		t.Errorf("Symbol = %q; want %q", q.Symbol, "MSFT")
	}
	if q.LastPrice != 310.50 {
		t.Errorf("LastPrice = %v; want 310.50", q.LastPrice)
	}
	if q.LastSize != 200 {
		t.Errorf("LastSize = %v; want 200", q.LastSize)
	}
	if q.VenueCode != "XNAS" {
		t.Errorf("VenueCode = %q; want %q", q.VenueCode, "XNAS")
	}
	if q.IngestedAtMS != 1700000001234 {
		t.Errorf("IngestedAtMS = %v; want 1700000001234", q.IngestedAtMS)
	}
}

// TestQueryLatestQuote_NoRowsReturnsError verifies that sql.ErrNoRows (or any
// scan error) is propagated correctly to the caller.
func TestQueryLatestQuote_NoRowsReturnsError(t *testing.T) {
	conn := &mockConn{
		// Empty rows causes row.Scan to return sql.ErrNoRows.
		rows: &mockRows{columns: quoteColumns, data: nil},
	}
	db := newTestDB(t, conn)

	_, err := queryLatestQuote(db, "UNKN")
	if err == nil {
		t.Error("expected an error for an empty result set, got nil")
	}
}

// ---------------------------------------------------------------------------
// Tests for FetchLatestQuote — integration of normalizeSymbol + queryLatestQuote
// ---------------------------------------------------------------------------

// TestFetchLatestQuote_NormalizesSymbolBeforeQuery confirms that the symbol
// stored in the bind argument is the normalized (upper-cased, trimmed) version.
func TestFetchLatestQuote_NormalizesSymbolBeforeQuery(t *testing.T) {
	conn := &mockConn{
		rows: &mockRows{
			columns: quoteColumns,
			data: [][]driver.Value{
				{"AAPL", float64(150.0), int64(50), "XNAS", int64(1700000000000)},
			},
		},
	}
	db := newTestDB(t, conn)

	_, err := FetchLatestQuote(db, "  aapl  ")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(conn.lastArgs) == 0 {
		t.Fatal("no bind arguments were recorded")
	}
	gotArg, ok := conn.lastArgs[0].Value.(string)
	if !ok {
		t.Fatalf("bind argument is not a string: %T", conn.lastArgs[0].Value)
	}
	if gotArg != "AAPL" {
		t.Errorf("bind argument = %q; want %q (raw input was %q)", gotArg, "AAPL", "  aapl  ")
	}
}

// ---------------------------------------------------------------------------
// Helper
// ---------------------------------------------------------------------------

// contains reports whether substr appears in s.
func contains(s, substr string) bool {
	return len(substr) > 0 && len(s) >= len(substr) && indexOf(s, substr) >= 0
}

func indexOf(s, substr string) int {
	if len(substr) == 0 {
		return 0
	}
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
