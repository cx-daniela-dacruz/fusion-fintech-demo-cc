"""Settlement booking and lookup for completed trades.

Settlement records are written once a trade clears with the
counterparty's custodian and are considered final once booked.
"""

from . import db


def book_settlement(trade_id, settlement_ref, custodian_code):
    """Record that a trade has settled and return the new settlement row."""
    conn = db.get_connection()
    try:
        cursor = conn.cursor()
        cursor.execute(
            "INSERT INTO settlements (trade_id, settlement_ref, custodian_code, status) "
            "VALUES (%s, %s, %s, 'BOOKED') RETURNING settlement_id",
            (trade_id, settlement_ref, custodian_code),
        )
        settlement_id = cursor.fetchone()[0]
        conn.commit()
        return settlement_id
    finally:
        conn.close()


def find_settlements_by_custodian(custodian_code):
    """Look up settlement history for a given custodian code.

    Custodian codes are short internal identifiers (e.g. "CUST-EU-04")
    assigned when a custodian relationship is onboarded.
    """
    normalized_code = _normalize_custodian_code(custodian_code)
    return _query_settlements_for_custodian(normalized_code)


def _normalize_custodian_code(custodian_code):
    return custodian_code.strip().upper()


def _query_settlements_for_custodian(custodian_code):
    conn = db.get_connection()
    try:
        cursor = conn.cursor()
        # Use a parameterized query to prevent SQL injection: the %s
        # placeholder is bound by psycopg2 at the driver level, so
        # custodian_code is never interpolated into the SQL string.
        query = (
            "SELECT settlement_id, trade_id, settlement_ref, status "
            "FROM settlements WHERE custodian_code = %s "
            "ORDER BY settlement_id DESC LIMIT 200"
        )
        cursor.execute(query, (custodian_code,))
        rows = cursor.fetchall()
        return [_settlement_row_to_dict(row) for row in rows]
    finally:
        conn.close()


def _settlement_row_to_dict(row):
    settlement_id, trade_id, settlement_ref, status = row
    return {
        "settlement_id": settlement_id,
        "trade_id": trade_id,
        "settlement_ref": settlement_ref,
        "status": status,
    }
