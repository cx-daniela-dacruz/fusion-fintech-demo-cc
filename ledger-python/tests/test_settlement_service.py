"""Tests for settlement_service — focused on SQL injection prevention.

These tests verify that:
1. Normal custodian code lookups work correctly (functional regression).
2. SQL injection payloads are treated as literal parameter values and never
   interpolated into the query string (security regression).
3. The Flask endpoint wires the query parameter through correctly.
"""

import sys
import os
import unittest
from unittest.mock import MagicMock, patch, call

# Ensure the package root is importable without a full install.
sys.path.insert(0, os.path.join(os.path.dirname(__file__), ".."))

from ledger import settlement_service


# ---------------------------------------------------------------------------
# Helpers
# ---------------------------------------------------------------------------

def _make_db_mock(rows=None):
    """Return a mock db connection whose cursor returns *rows* from fetchall."""
    rows = rows or []
    cursor_mock = MagicMock()
    cursor_mock.fetchall.return_value = rows

    conn_mock = MagicMock()
    conn_mock.cursor.return_value = cursor_mock

    return conn_mock, cursor_mock


# ---------------------------------------------------------------------------
# Unit tests for _normalize_custodian_code
# ---------------------------------------------------------------------------

class TestNormalizeCustodianCode(unittest.TestCase):
    def test_strips_whitespace(self):
        result = settlement_service._normalize_custodian_code("  CUST-EU-04  ")
        self.assertEqual(result, "CUST-EU-04")

    def test_converts_to_uppercase(self):
        result = settlement_service._normalize_custodian_code("cust-eu-04")
        self.assertEqual(result, "CUST-EU-04")

    def test_strips_and_uppercases(self):
        result = settlement_service._normalize_custodian_code(" cust-eu-04 ")
        self.assertEqual(result, "CUST-EU-04")

    def test_already_normalised_is_unchanged(self):
        result = settlement_service._normalize_custodian_code("CUST-EU-04")
        self.assertEqual(result, "CUST-EU-04")


# ---------------------------------------------------------------------------
# Unit tests for _query_settlements_for_custodian — parameterized query check
# ---------------------------------------------------------------------------

class TestQuerySettlementsForCustodian(unittest.TestCase):
    """Verify that _query_settlements_for_custodian uses a parameterized query."""

    def setUp(self):
        self.sample_rows = [
            (101, 1001, "REF-001", "BOOKED"),
            (102, 1002, "REF-002", "BOOKED"),
        ]

    @patch("ledger.settlement_service.db")
    def test_returns_list_of_dicts_for_normal_code(self, mock_db):
        conn_mock, cursor_mock = _make_db_mock(self.sample_rows)
        mock_db.get_connection.return_value = conn_mock

        result = settlement_service._query_settlements_for_custodian("CUST-EU-04")

        self.assertEqual(len(result), 2)
        self.assertEqual(result[0]["settlement_id"], 101)
        self.assertEqual(result[0]["trade_id"], 1001)
        self.assertEqual(result[0]["settlement_ref"], "REF-001")
        self.assertEqual(result[0]["status"], "BOOKED")

    @patch("ledger.settlement_service.db")
    def test_returns_empty_list_when_no_rows(self, mock_db):
        conn_mock, cursor_mock = _make_db_mock([])
        mock_db.get_connection.return_value = conn_mock

        result = settlement_service._query_settlements_for_custodian("CUST-EU-99")

        self.assertEqual(result, [])

    @patch("ledger.settlement_service.db")
    def test_execute_called_with_parameterized_tuple(self, mock_db):
        """The SQL query must be passed with a separate parameter tuple.

        This is the core assertion for the SQL injection fix: cursor.execute
        must receive (query_string, params_tuple) so that psycopg2 handles
        escaping at the driver level, rather than the custodian_code being
        interpolated into the SQL string before the call.
        """
        conn_mock, cursor_mock = _make_db_mock([])
        mock_db.get_connection.return_value = conn_mock

        custodian = "CUST-EU-04"
        settlement_service._query_settlements_for_custodian(custodian)

        # cursor.execute must have been called exactly once.
        cursor_mock.execute.assert_called_once()
        args, kwargs = cursor_mock.execute.call_args

        # First argument is the SQL template string — must NOT contain the
        # custodian code embedded directly.
        sql_template = args[0]
        self.assertNotIn(custodian, sql_template,
                         "custodian_code must not be interpolated into the SQL string")

        # Second argument must be a tuple containing the custodian value.
        self.assertEqual(len(args), 2,
                         "cursor.execute must receive a params tuple as its second argument")
        params = args[1]
        self.assertIn(custodian, params,
                      "custodian_code must be passed as a bound parameter")

    @patch("ledger.settlement_service.db")
    def test_sql_injection_payload_passed_as_parameter_not_interpolated(self, mock_db):
        """Classic SQL injection payloads must NOT alter the query structure.

        If the code were still using % string formatting, the payload below
        would close the string literal and append a UNION or tautology.
        With parameterized queries the entire payload is bound as a literal
        value, so cursor.execute receives it unchanged as the second element
        of the params tuple.
        """
        conn_mock, cursor_mock = _make_db_mock([])
        mock_db.get_connection.return_value = conn_mock

        injection_payload = "' OR '1'='1"
        settlement_service._query_settlements_for_custodian(injection_payload)

        args, _ = cursor_mock.execute.call_args
        sql_template = args[0]
        params = args[1]

        # The payload must NOT appear in the SQL template.
        self.assertNotIn(injection_payload, sql_template,
                         "Injection payload must not be embedded in the SQL template")
        # The payload must appear verbatim in the params tuple.
        self.assertIn(injection_payload, params,
                      "Injection payload must be passed as a bound parameter value")

    @patch("ledger.settlement_service.db")
    def test_union_injection_payload_passed_as_parameter(self, mock_db):
        """UNION-based injection payload is treated as a literal value."""
        conn_mock, cursor_mock = _make_db_mock([])
        mock_db.get_connection.return_value = conn_mock

        injection_payload = "x' UNION SELECT 1,2,3,4--"
        settlement_service._query_settlements_for_custodian(injection_payload)

        args, _ = cursor_mock.execute.call_args
        sql_template = args[0]
        params = args[1]

        self.assertNotIn("UNION", sql_template)
        self.assertIn(injection_payload, params)

    @patch("ledger.settlement_service.db")
    def test_stacked_query_injection_payload_passed_as_parameter(self, mock_db):
        """Stacked query / statement terminator payload is treated as a literal value."""
        conn_mock, cursor_mock = _make_db_mock([])
        mock_db.get_connection.return_value = conn_mock

        injection_payload = "x'; DROP TABLE settlements;--"
        settlement_service._query_settlements_for_custodian(injection_payload)

        args, _ = cursor_mock.execute.call_args
        sql_template = args[0]
        params = args[1]

        self.assertNotIn("DROP", sql_template)
        self.assertIn(injection_payload, params)

    @patch("ledger.settlement_service.db")
    def test_connection_is_always_closed(self, mock_db):
        """The DB connection must be closed even if cursor.execute raises."""
        conn_mock = MagicMock()
        cursor_mock = MagicMock()
        cursor_mock.execute.side_effect = Exception("db error")
        conn_mock.cursor.return_value = cursor_mock
        mock_db.get_connection.return_value = conn_mock

        with self.assertRaises(Exception):
            settlement_service._query_settlements_for_custodian("CUST-EU-04")

        conn_mock.close.assert_called_once()


# ---------------------------------------------------------------------------
# Integration-style tests for find_settlements_by_custodian
# ---------------------------------------------------------------------------

class TestFindSettlementsByCustodian(unittest.TestCase):
    """Test the public entry point, which normalises the code first."""

    @patch("ledger.settlement_service.db")
    def test_normalises_code_before_query(self, mock_db):
        conn_mock, cursor_mock = _make_db_mock([])
        mock_db.get_connection.return_value = conn_mock

        settlement_service.find_settlements_by_custodian("  cust-eu-04  ")

        args, _ = cursor_mock.execute.call_args
        params = args[1]
        # Normalisation (strip + upper) must have been applied before binding.
        self.assertIn("CUST-EU-04", params)

    @patch("ledger.settlement_service.db")
    def test_returns_settlement_dicts(self, mock_db):
        rows = [(201, 2001, "REF-201", "BOOKED")]
        conn_mock, cursor_mock = _make_db_mock(rows)
        mock_db.get_connection.return_value = conn_mock

        result = settlement_service.find_settlements_by_custodian("CUST-EU-04")

        self.assertEqual(len(result), 1)
        self.assertEqual(result[0]["settlement_id"], 201)
        self.assertEqual(result[0]["status"], "BOOKED")


# ---------------------------------------------------------------------------
# Flask endpoint tests for GET /settlements/by-custodian
# ---------------------------------------------------------------------------

class TestGetSettlementsByCustodianEndpoint(unittest.TestCase):
    """Verify the Flask route hands the query param through without alteration."""

    def setUp(self):
        # Import here so patching of settlement_service.db above doesn't
        # interfere with module-level app creation.
        from ledger.app import app
        app.config["TESTING"] = True
        self.client = app.test_client()

    def test_missing_custodian_code_returns_400(self):
        resp = self.client.get("/settlements/by-custodian")
        self.assertEqual(resp.status_code, 400)
        data = resp.get_json()
        self.assertIn("error", data)

    @patch("ledger.settlement_service.db")
    def test_valid_custodian_code_returns_200(self, mock_db):
        rows = [(301, 3001, "REF-301", "BOOKED")]
        conn_mock, cursor_mock = _make_db_mock(rows)
        mock_db.get_connection.return_value = conn_mock

        resp = self.client.get("/settlements/by-custodian?custodian_code=CUST-EU-04")
        self.assertEqual(resp.status_code, 200)
        data = resp.get_json()
        self.assertIn("settlements", data)
        self.assertEqual(len(data["settlements"]), 1)

    @patch("ledger.settlement_service.db")
    def test_injection_payload_does_not_alter_sql_template(self, mock_db):
        """End-to-end: an injection string from the HTTP layer must be bound
        as a parameter and must never appear inside the SQL template string.
        """
        conn_mock, cursor_mock = _make_db_mock([])
        mock_db.get_connection.return_value = conn_mock

        # URL-encode the payload so Flask/Werkzeug handles decoding.
        payload = "' OR '1'='1"
        import urllib.parse
        encoded = urllib.parse.quote(payload)
        resp = self.client.get(f"/settlements/by-custodian?custodian_code={encoded}")

        # The endpoint should still return 200 — the payload is just treated
        # as a literal custodian_code value that happens to match no rows.
        self.assertEqual(resp.status_code, 200)

        args, _ = cursor_mock.execute.call_args
        sql_template = args[0]
        params = args[1]

        self.assertNotIn("OR", sql_template,
                         "Injection keyword must not appear in the SQL template")
        # After normalisation (strip+upper) the payload lands in params.
        self.assertTrue(any("OR" in str(p) for p in params),
                        "Injection payload must be present in the bound parameter tuple")


if __name__ == "__main__":
    unittest.main()
