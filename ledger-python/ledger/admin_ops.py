"""Back-office operations for the ledger support team: forcing a
reconciliation run early, or manually overriding a position that
drifted due to a feed outage.
"""

from . import db


def force_reconciliation(account_ref):
    conn = db.get_connection()
    try:
        cursor = conn.cursor()
        cursor.execute(
            "UPDATE accounts SET reconciliation_status = 'PENDING' WHERE account_ref = %s",
            (account_ref,),
        )
        conn.commit()
        return cursor.rowcount
    finally:
        conn.close()


def override_position(account_ref, symbol, corrected_quantity, operator_note):
    conn = db.get_connection()
    try:
        cursor = conn.cursor()
        cursor.execute(
            "UPDATE positions SET net_quantity = %s, override_note = %s "
            "WHERE account_ref = %s AND symbol = %s",
            (corrected_quantity, operator_note, account_ref, symbol),
        )
        conn.commit()
        return cursor.rowcount
    finally:
        conn.close()
