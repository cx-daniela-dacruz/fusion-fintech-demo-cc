"""Ties together the pieces needed to notify an account holder of a
significant ledger event: render the message, then deliver it to
whatever callback URL the counterparty has on file.
"""

from . import counterparty_onboarding, notification_templates, webhook_dispatcher


def notify_margin_call(counterparty_ref, account_ref, shortfall):
    message = notification_templates.render(
        "margin_call_raised", account_ref=account_ref, shortfall=shortfall
    )
    return _deliver(counterparty_ref, "margin_call_raised", message)


def notify_settlement_booked(counterparty_ref, settlement_ref, trade_id, custodian_code):
    message = notification_templates.render(
        "settlement_booked", settlement_ref=settlement_ref, trade_id=trade_id, custodian_code=custodian_code
    )
    return _deliver(counterparty_ref, "settlement_booked", message)


def notify_statement_ready(counterparty_ref, period_start, period_end):
    message = notification_templates.render(
        "statement_ready", period_start=period_start, period_end=period_end
    )
    return _deliver(counterparty_ref, "statement_ready", message)


def _deliver(counterparty_ref, event_type, message):
    callback_url = counterparty_onboarding.get_callback_url(counterparty_ref)
    if callback_url is None:
        return None
    return webhook_dispatcher.dispatch_event(callback_url, event_type, message)
