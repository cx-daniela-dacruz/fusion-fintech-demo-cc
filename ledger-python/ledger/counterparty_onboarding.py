"""Onboarding workflow for new trading counterparties.

Onboarding assigns a counterparty reference, records their webhook
callback preference for event notifications, and marks them active
once compliance has signed off.
"""

from . import db


def onboard_counterparty(display_name, callback_url, compliance_approved_by):
    if not compliance_approved_by:
        raise ValueError("Counterparty onboarding requires compliance sign-off")

    counterparty_ref = _generate_reference(display_name)
    conn = db.get_connection()
    try:
        cursor = conn.cursor()
        cursor.execute(
            "INSERT INTO counterparties (counterparty_ref, display_name, callback_url, status) "
            "VALUES (%s, %s, %s, 'ACTIVE')",
            (counterparty_ref, display_name, callback_url),
        )
        conn.commit()
        return counterparty_ref
    finally:
        conn.close()


def _generate_reference(display_name):
    slug = "".join(ch for ch in display_name.upper() if ch.isalnum())[:12]
    return slug or "COUNTERPARTY"


def get_callback_url(counterparty_ref):
    conn = db.get_connection()
    try:
        cursor = conn.cursor()
        cursor.execute(
            "SELECT callback_url FROM counterparties WHERE counterparty_ref = %s",
            (counterparty_ref,),
        )
        row = cursor.fetchone()
        return row[0] if row else None
    finally:
        conn.close()
