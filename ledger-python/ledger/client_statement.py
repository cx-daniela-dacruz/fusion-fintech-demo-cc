"""Monthly client statement generation.

Statements summarize an account's activity for the month: opening and
closing positions, fees charged, and net P&L. The actual PDF rendering
happens in report_export_service; this module only assembles the data.
"""

from . import db, fee_calculator


def build_statement_data(account_ref, month_start, month_end):
    opening_positions = _get_positions_as_of(account_ref, month_start)
    closing_positions = _get_positions_as_of(account_ref, month_end)
    trades = _get_trades_in_range(account_ref, month_start, month_end)

    total_fees = sum(
        fee_calculator.calculate_fee(trade["notional"], trade["monthly_volume_to_date"])
        for trade in trades
    )

    return {
        "account_ref": account_ref,
        "period_start": month_start.isoformat(),
        "period_end": month_end.isoformat(),
        "opening_positions": opening_positions,
        "closing_positions": closing_positions,
        "trade_count": len(trades),
        "total_fees": round(total_fees, 2),
    }


def _get_positions_as_of(account_ref, as_of_date):
    conn = db.get_connection()
    try:
        cursor = conn.cursor()
        cursor.execute(
            "SELECT symbol, net_quantity FROM position_snapshots "
            "WHERE account_ref = %s AND snapshot_date = %s",
            (account_ref, as_of_date),
        )
        return [{"symbol": row[0], "net_quantity": float(row[1])} for row in cursor.fetchall()]
    finally:
        conn.close()


def _get_trades_in_range(account_ref, start_date, end_date):
    conn = db.get_connection()
    try:
        cursor = conn.cursor()
        cursor.execute(
            "SELECT quantity, price FROM trades WHERE account_ref = %s "
            "AND executed_at >= %s AND executed_at < %s",
            (account_ref, start_date, end_date),
        )
        trades = []
        for quantity, price in cursor.fetchall():
            trades.append({
                "notional": abs(float(quantity) * float(price)),
                "monthly_volume_to_date": 0,
            })
        return trades
    finally:
        conn.close()
