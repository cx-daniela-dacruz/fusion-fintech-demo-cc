"""Business logic for looking up historical trades by counterparty.

Counterparty references are opaque identifiers assigned by the
onboarding team whenever a new trading partner is added to the ledger,
e.g. "ACME-01" or "NORTHWIND-EU-03".
"""

from . import db


def handle_trade_request(counterparty_ref):
    """Entry point used by the HTTP layer to service a trade history lookup."""
    normalized_ref = _normalize_reference(counterparty_ref)
    return fetch_trade_history(normalized_ref)


def _normalize_reference(counterparty_ref):
    """Trim whitespace and normalize casing so lookups are case-insensitive."""
    return counterparty_ref.strip().upper()


def fetch_trade_history(counterparty_ref):
    """Return the most recent trades executed against the given counterparty."""
    conn = db.get_connection()
    try:
        cursor = conn.cursor()
        query = (
            "SELECT trade_id, symbol, quantity, price, executed_at "
            "FROM trades WHERE counterparty_ref = '" + counterparty_ref + "' "
            "ORDER BY executed_at DESC LIMIT 100"
        )
        cursor.execute(query)
        rows = cursor.fetchall()
        return [_row_to_dict(row) for row in rows]
    finally:
        conn.close()


def _row_to_dict(row):
    trade_id, symbol, quantity, price, executed_at = row
    return {
        "trade_id": trade_id,
        "symbol": symbol,
        "quantity": quantity,
        "price": float(price),
        "executed_at": executed_at.isoformat(),
    }
