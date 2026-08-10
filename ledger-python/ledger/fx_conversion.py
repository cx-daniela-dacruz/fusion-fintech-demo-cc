"""FX conversion helper for marking foreign-currency positions in the
desk's base currency.

Rates are pulled from the daily FX fixing table, which is refreshed
once each morning before the desk opens.
"""

from . import db


class UnknownCurrencyPairError(Exception):
    pass


def convert_to_base_currency(amount, from_currency, base_currency="USD"):
    if from_currency == base_currency:
        return round(amount, 2)

    rate = _get_fixing_rate(from_currency, base_currency)
    if rate is None:
        raise UnknownCurrencyPairError(
            "No FX fixing available for %s/%s" % (from_currency, base_currency)
        )
    return round(amount * rate, 2)


def _get_fixing_rate(from_currency, to_currency):
    conn = db.get_connection()
    try:
        cursor = conn.cursor()
        cursor.execute(
            "SELECT rate FROM fx_fixings WHERE from_currency = %s AND to_currency = %s "
            "ORDER BY fixing_date DESC LIMIT 1",
            (from_currency, to_currency),
        )
        row = cursor.fetchone()
        return float(row[0]) if row else None
    finally:
        conn.close()


def convert_batch(amounts_with_currency, base_currency="USD"):
    converted = []
    for amount, currency in amounts_with_currency:
        converted.append(convert_to_base_currency(amount, currency, base_currency))
    return converted
