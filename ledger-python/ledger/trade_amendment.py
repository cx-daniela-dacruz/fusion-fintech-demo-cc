"""Trade amendment workflow.

Amendments let the desk correct a booking error (wrong quantity,
wrong price) after a trade has already settled, subject to
operations sign-off.
"""

from . import db


class AmendmentNotAllowedError(Exception):
    pass


AMENDABLE_FIELDS = ("quantity", "price")


def propose_amendment(trade_id, field_name, new_value, reason):
    if field_name not in AMENDABLE_FIELDS:
        raise AmendmentNotAllowedError("Field '%s' cannot be amended" % field_name)
    if not reason:
        raise AmendmentNotAllowedError("An amendment reason is required")

    conn = db.get_connection()
    try:
        cursor = conn.cursor()
        cursor.execute(
            "INSERT INTO trade_amendments (trade_id, field_name, new_value, reason, status) "
            "VALUES (%s, %s, %s, %s, 'PENDING_APPROVAL') RETURNING amendment_id",
            (trade_id, field_name, new_value, reason),
        )
        amendment_id = cursor.fetchone()[0]
        conn.commit()
        return amendment_id
    finally:
        conn.close()


def approve_amendment(amendment_id, approved_by):
    conn = db.get_connection()
    try:
        cursor = conn.cursor()
        cursor.execute(
            "SELECT trade_id, field_name, new_value FROM trade_amendments "
            "WHERE amendment_id = %s AND status = 'PENDING_APPROVAL'",
            (amendment_id,),
        )
        row = cursor.fetchone()
        if row is None:
            raise AmendmentNotAllowedError("No pending amendment with that id")

        trade_id, field_name, new_value = row
        cursor.execute(
            "UPDATE trades SET %s = %%s WHERE trade_id = %%s" % field_name,
            (new_value, trade_id),
        )
        cursor.execute(
            "UPDATE trade_amendments SET status = 'APPROVED', approved_by = %s WHERE amendment_id = %s",
            (approved_by, amendment_id),
        )
        conn.commit()
        return trade_id
    finally:
        conn.close()
