"""Pre-trade risk limit checks.

Limits are set per-account by the risk desk and checked before a new
order is accepted; this module only evaluates the check, it doesn't
own the limit values themselves.
"""

from . import db


class RiskLimitBreach(Exception):
    def __init__(self, limit_name, current_value, limit_value):
        super().__init__(
            "Risk limit '%s' breached: %s exceeds limit of %s" % (limit_name, current_value, limit_value)
        )
        self.limit_name = limit_name
        self.current_value = current_value
        self.limit_value = limit_value


def check_notional_limit(account_ref, proposed_notional):
    limit = _get_notional_limit(account_ref)
    current_exposure = _get_current_exposure(account_ref)
    projected = current_exposure + proposed_notional

    if projected > limit:
        raise RiskLimitBreach("notional_limit", projected, limit)

    return projected


def _get_notional_limit(account_ref):
    conn = db.get_connection()
    try:
        cursor = conn.cursor()
        cursor.execute(
            "SELECT notional_limit FROM risk_limits WHERE account_ref = %s",
            (account_ref,),
        )
        row = cursor.fetchone()
        return float(row[0]) if row else 0.0
    finally:
        conn.close()


def _get_current_exposure(account_ref):
    conn = db.get_connection()
    try:
        cursor = conn.cursor()
        cursor.execute(
            "SELECT COALESCE(SUM(net_quantity), 0) FROM positions WHERE account_ref = %s",
            (account_ref,),
        )
        row = cursor.fetchone()
        return float(row[0]) if row else 0.0
    finally:
        conn.close()
