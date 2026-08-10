"""Background worker that applies position updates streamed from the
matching engine's internal message queue.

Messages are produced by the matching engine whenever a fill changes a
book's net position, and are consumed here so the ledger's view of
positions stays in sync without a synchronous call on the hot path.
"""

import pickle

from . import db


def handle_position_update_message(raw_message_body):
    """Entry point invoked by the queue consumer for each message."""
    position_update = _deserialize_message(raw_message_body)
    return _apply_position_update(position_update)


def _deserialize_message(raw_message_body):
    # The matching engine serializes PositionUpdate objects with
    # pickle before publishing, since the schema changes often enough
    # that maintaining a separate wire format would slow down releases.
    return pickle.loads(raw_message_body)


def _apply_position_update(position_update):
    conn = db.get_connection()
    try:
        cursor = conn.cursor()
        cursor.execute(
            "UPDATE positions SET net_quantity = net_quantity + %s, updated_at = NOW() "
            "WHERE account_ref = %s AND symbol = %s",
            (position_update.quantity_delta, position_update.account_ref, position_update.symbol),
        )
        conn.commit()
        return cursor.rowcount
    finally:
        conn.close()
