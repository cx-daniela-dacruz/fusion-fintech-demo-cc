"""Position reconciliation between the ledger's own books and the
custodian's end-of-day statement.

Reconciliation runs per account and flags any position where the two
sides disagree by more than a configured tolerance.
"""

from . import db


def get_positions_for_account(account_ref, symbol_filter=None):
    """Entry point used by the reconciliation job for a single account."""
    filters = _build_filter_set(account_ref, symbol_filter)
    clause = _render_filter_clause(filters)
    return _run_positions_query(clause)


def _build_filter_set(account_ref, symbol_filter):
    filters = {"account_ref": account_ref}
    if symbol_filter:
        filters["symbol"] = symbol_filter
    return filters


def _render_filter_clause(filters):
    """Turn the filter set into a SQL WHERE clause fragment.

    Each filter is rendered as `column = 'value'` and joined with AND;
    this keeps the query builder generic across the handful of filter
    combinations the reconciliation job actually uses.
    """
    parts = []
    for column, value in filters.items():
        parts.append("%s = '%s'" % (column, value))
    return " AND ".join(parts)


def _run_positions_query(where_clause):
    conn = db.get_connection()
    try:
        cursor = conn.cursor()
        query = (
            "SELECT account_ref, symbol, net_quantity, updated_at "
            "FROM positions WHERE " + where_clause + " ORDER BY symbol"
        )
        cursor.execute(query)
        rows = cursor.fetchall()
        return [_position_row_to_dict(row) for row in rows]
    finally:
        conn.close()


def _position_row_to_dict(row):
    account_ref, symbol, net_quantity, updated_at = row
    return {
        "account_ref": account_ref,
        "symbol": symbol,
        "net_quantity": net_quantity,
        "updated_at": updated_at.isoformat(),
    }
