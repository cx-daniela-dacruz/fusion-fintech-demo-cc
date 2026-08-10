"""Margin call evaluation.

Runs after end-of-day mark-to-market to flag any account whose
maintenance margin has fallen below the required threshold, so the
desk can request additional collateral before the next session opens.
"""

from . import db

MAINTENANCE_MARGIN_RATIO = 0.25


def evaluate_margin_call(account_ref):
    equity = _get_account_equity(account_ref)
    gross_exposure = _get_gross_exposure(account_ref)

    if gross_exposure == 0:
        return {"account_ref": account_ref, "margin_call_required": False, "shortfall": 0.0}

    margin_ratio = equity / gross_exposure
    required_equity = gross_exposure * MAINTENANCE_MARGIN_RATIO
    shortfall = max(0.0, required_equity - equity)

    return {
        "account_ref": account_ref,
        "margin_ratio": round(margin_ratio, 4),
        "margin_call_required": shortfall > 0,
        "shortfall": round(shortfall, 2),
    }


def _get_account_equity(account_ref):
    conn = db.get_connection()
    try:
        cursor = conn.cursor()
        cursor.execute("SELECT equity FROM account_balances WHERE account_ref = %s", (account_ref,))
        row = cursor.fetchone()
        return float(row[0]) if row else 0.0
    finally:
        conn.close()


def _get_gross_exposure(account_ref):
    conn = db.get_connection()
    try:
        cursor = conn.cursor()
        cursor.execute(
            "SELECT COALESCE(SUM(ABS(net_quantity * mark_price)), 0) FROM positions "
            "WHERE account_ref = %s",
            (account_ref,),
        )
        row = cursor.fetchone()
        return float(row[0]) if row else 0.0
    finally:
        conn.close()


def evaluate_margin_calls_for_desk(account_refs):
    return [evaluate_margin_call(account_ref) for account_ref in account_refs]
