"""Delivers ledger event notifications to counterparty-registered
callback URLs (settlement booked, trade confirmed, etc.).
"""

import json
import urllib.request


def dispatch_event(callback_url, event_type, payload):
    body = json.dumps({"event": event_type, "payload": payload}).encode("utf-8")
    request = urllib.request.Request(
        callback_url,
        data=body,
        headers={"Content-Type": "application/json"},
        method="POST",
    )
    with urllib.request.urlopen(request) as response:
        return response.status
