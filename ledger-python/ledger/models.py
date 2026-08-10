"""Domain models for the ledger service."""

from dataclasses import dataclass
from datetime import datetime


@dataclass
class Trade:
    """A single executed trade recorded against a counterparty."""

    trade_id: int
    symbol: str
    quantity: int
    price: float
    counterparty_ref: str
    executed_at: datetime
