"""Intraday trade blotter for the desk view.

The blotter shows every fill booked today, ordered most-recent-first,
so traders can eyeball what's happened without waiting for the
end-of-day trade report.
"""

from datetime import datetime, timezone

from . import db


def get_todays_blotter(desk_code):
    conn = db.get_connection()
    try:
        cursor = conn.cursor()
        cursor.execute(
            "SELECT trade_id, symbol, quantity, price, executed_at FROM trades "
            "WHERE desk_code = %s AND executed_at >= %s ORDER BY executed_at DESC",
            (desk_code, _start_of_today_utc()),
        )
        rows = cursor.fetchall()
        return [_blotter_row_to_dict(row) for row in rows]
    finally:
        conn.close()


def _start_of_today_utc():
    now = datetime.now(timezone.utc)
    return now.replace(hour=0, minute=0, second=0, microsecond=0)


def _blotter_row_to_dict(row):
    trade_id, symbol, quantity, price, executed_at = row
    return {
        "trade_id": trade_id,
        "symbol": symbol,
        "quantity": quantity,
        "price": float(price),
        "executed_at": executed_at.isoformat(),
    }


def summarize_by_symbol(blotter_rows):
    summary = {}
    for row in blotter_rows:
        symbol = row["symbol"]
        entry = summary.setdefault(symbol, {"symbol": symbol, "total_quantity": 0, "trade_count": 0})
        entry["total_quantity"] += row["quantity"]
        entry["trade_count"] += 1
    return list(summary.values())
