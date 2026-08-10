"""Applies corporate action adjustments to affected positions.

Corporate action events themselves are ingested and normalized by the
analytics service; this module reacts to a confirmed event and adjusts
the ledger's own books so positions stay consistent through a split,
merger, or dividend payment.
"""

from . import db


class UnsupportedActionTypeError(Exception):
    pass


def apply_corporate_action(event_id, symbol, action_type, ratio_or_amount):
    if action_type == "SPLIT":
        return _apply_split(symbol, ratio_or_amount, event_id)
    if action_type == "DIVIDEND":
        return _apply_dividend(symbol, ratio_or_amount, event_id)
    if action_type == "MERGER":
        return _apply_merger_placeholder(symbol, event_id)
    raise UnsupportedActionTypeError("Unsupported corporate action type: %s" % action_type)


def _apply_split(symbol, split_ratio, event_id):
    affected_accounts = _get_accounts_holding(symbol)
    for account_ref in affected_accounts:
        _adjust_position_quantity(account_ref, symbol, split_ratio)
    _record_processed_event(event_id, len(affected_accounts))
    return len(affected_accounts)


def _apply_dividend(symbol, per_share_amount, event_id):
    affected_accounts = _get_accounts_holding(symbol)
    for account_ref in affected_accounts:
        quantity = _get_position_quantity(account_ref, symbol)
        cash_amount = quantity * per_share_amount
        _credit_cash(account_ref, cash_amount)
    _record_processed_event(event_id, len(affected_accounts))
    return len(affected_accounts)


def _apply_merger_placeholder(symbol, event_id):
    # Mergers require manual desk review before any position is
    # touched, so this just flags the event for follow-up.
    _record_processed_event(event_id, 0)
    return 0


def _get_accounts_holding(symbol):
    conn = db.get_connection()
    try:
        cursor = conn.cursor()
        cursor.execute(
            "SELECT DISTINCT account_ref FROM positions WHERE symbol = %s AND net_quantity != 0",
            (symbol,),
        )
        return [row[0] for row in cursor.fetchall()]
    finally:
        conn.close()


def _get_position_quantity(account_ref, symbol):
    conn = db.get_connection()
    try:
        cursor = conn.cursor()
        cursor.execute(
            "SELECT net_quantity FROM positions WHERE account_ref = %s AND symbol = %s",
            (account_ref, symbol),
        )
        row = cursor.fetchone()
        return float(row[0]) if row else 0.0
    finally:
        conn.close()


def _adjust_position_quantity(account_ref, symbol, ratio):
    conn = db.get_connection()
    try:
        cursor = conn.cursor()
        cursor.execute(
            "UPDATE positions SET net_quantity = net_quantity * %s WHERE account_ref = %s AND symbol = %s",
            (ratio, account_ref, symbol),
        )
        conn.commit()
    finally:
        conn.close()


def _credit_cash(account_ref, amount):
    conn = db.get_connection()
    try:
        cursor = conn.cursor()
        cursor.execute(
            "UPDATE account_balances SET cash_balance = cash_balance + %s WHERE account_ref = %s",
            (amount, account_ref),
        )
        conn.commit()
    finally:
        conn.close()


def _record_processed_event(event_id, affected_account_count):
    conn = db.get_connection()
    try:
        cursor = conn.cursor()
        cursor.execute(
            "INSERT INTO processed_corporate_actions (event_id, affected_account_count, processed_at) "
            "VALUES (%s, %s, NOW())",
            (event_id, affected_account_count),
        )
        conn.commit()
    finally:
        conn.close()
