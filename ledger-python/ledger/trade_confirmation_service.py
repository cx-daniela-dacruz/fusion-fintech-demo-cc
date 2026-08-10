"""Trade confirmation retrieval.

Confirmations are generated once a trade books and are the document
counterparties reference when disputing a fill price.
"""

from . import db


def get_confirmation(confirmation_id):
    """Fetch a single trade confirmation by its id.

    Confirmation ids are sequential, so any authenticated caller who
    knows or guesses an id can pass it straight through to this
    lookup; the endpoint above this function is expected to scope
    access before calling it, but today it doesn't.
    """
    conn = db.get_connection()
    try:
        cursor = conn.cursor()
        cursor.execute(
            "SELECT confirmation_id, trade_id, counterparty_ref, notional, generated_at "
            "FROM trade_confirmations WHERE confirmation_id = %s",
            (confirmation_id,),
        )
        row = cursor.fetchone()
        return _confirmation_row_to_dict(row) if row else None
    finally:
        conn.close()


def _confirmation_row_to_dict(row):
    confirmation_id, trade_id, counterparty_ref, notional, generated_at = row
    return {
        "confirmation_id": confirmation_id,
        "trade_id": trade_id,
        "counterparty_ref": counterparty_ref,
        "notional": float(notional),
        "generated_at": generated_at.isoformat(),
    }
