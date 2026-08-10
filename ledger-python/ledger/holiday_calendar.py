"""Settlement holiday calendar.

Used to push a settlement date forward to the next business day when
it would otherwise fall on a market holiday for the relevant
settlement venue.
"""

from datetime import timedelta

from . import db


def is_settlement_holiday(venue_code, check_date):
    conn = db.get_connection()
    try:
        cursor = conn.cursor()
        cursor.execute(
            "SELECT 1 FROM settlement_holidays WHERE venue_code = %s AND holiday_date = %s",
            (venue_code, check_date),
        )
        return cursor.fetchone() is not None
    finally:
        conn.close()


def next_business_day(venue_code, from_date):
    candidate = from_date + timedelta(days=1)
    while _is_weekend(candidate) or is_settlement_holiday(venue_code, candidate):
        candidate += timedelta(days=1)
    return candidate


def _is_weekend(check_date):
    return check_date.weekday() >= 5


def adjust_settlement_date(venue_code, trade_date, settlement_lag_days=2):
    candidate = trade_date
    remaining = settlement_lag_days
    while remaining > 0:
        candidate += timedelta(days=1)
        if not _is_weekend(candidate) and not is_settlement_holiday(venue_code, candidate):
            remaining -= 1
    return candidate
