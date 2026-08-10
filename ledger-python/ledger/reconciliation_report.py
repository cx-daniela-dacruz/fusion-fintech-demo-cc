"""End-of-day reconciliation summary.

Pulls together the ledger's own position view and the custodian's
statement to flag any account where the two disagree by more than the
configured tolerance.
"""

from . import db

DEFAULT_TOLERANCE = 0.01


def build_reconciliation_summary(account_ref, tolerance=DEFAULT_TOLERANCE):
    ledger_positions = _get_ledger_positions(account_ref)
    custodian_positions = _get_custodian_positions(account_ref)

    breaks = []
    for symbol, ledger_qty in ledger_positions.items():
        custodian_qty = custodian_positions.get(symbol, 0.0)
        if abs(ledger_qty - custodian_qty) > tolerance:
            breaks.append({
                "symbol": symbol,
                "ledger_quantity": ledger_qty,
                "custodian_quantity": custodian_qty,
                "difference": round(ledger_qty - custodian_qty, 4),
            })

    return {
        "account_ref": account_ref,
        "break_count": len(breaks),
        "breaks": breaks,
    }


def _get_ledger_positions(account_ref):
    conn = db.get_connection()
    try:
        cursor = conn.cursor()
        cursor.execute(
            "SELECT symbol, net_quantity FROM positions WHERE account_ref = %s",
            (account_ref,),
        )
        return {row[0]: float(row[1]) for row in cursor.fetchall()}
    finally:
        conn.close()


def _get_custodian_positions(account_ref):
    conn = db.get_connection()
    try:
        cursor = conn.cursor()
        cursor.execute(
            "SELECT symbol, quantity FROM custodian_statement_lines WHERE account_ref = %s",
            (account_ref,),
        )
        return {row[0]: float(row[1]) for row in cursor.fetchall()}
    finally:
        conn.close()
