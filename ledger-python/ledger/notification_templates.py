"""Message templates for ledger-originated notifications (settlement
booked, margin call raised, statement ready).

Rendering only prepares the message text; actual delivery goes
through the shared notifications platform via webhook_dispatcher.
"""

TEMPLATES = {
    "settlement_booked": {
        "subject": "Settlement booked",
        "body": "Settlement {settlement_ref} for trade {trade_id} has been booked with {custodian_code}.",
    },
    "margin_call_raised": {
        "subject": "Margin call notice",
        "body": "Account {account_ref} has a margin shortfall of {shortfall}. Please post additional collateral.",
    },
    "statement_ready": {
        "subject": "Your statement is ready",
        "body": "Your statement for {period_start} through {period_end} is ready to download.",
    },
}


class UnknownTemplateError(Exception):
    pass


def render(template_name, **fields):
    template = TEMPLATES.get(template_name)
    if template is None:
        raise UnknownTemplateError("No template named '%s'" % template_name)

    return {
        "subject": template["subject"],
        "body": template["body"].format(**fields),
    }
