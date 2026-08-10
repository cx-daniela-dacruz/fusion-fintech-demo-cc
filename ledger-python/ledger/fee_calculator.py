"""Trade fee calculation.

Fees are calculated per-fill using a simple tiered schedule based on
monthly notional volume; the actual rate card is reloaded nightly from
the pricing service, but the tiers themselves rarely change.
"""

TIERS = (
    (0, 1_000_000, 0.00050),
    (1_000_000, 10_000_000, 0.00035),
    (10_000_000, 50_000_000, 0.00020),
    (50_000_000, float("inf"), 0.00012),
)


def calculate_fee(notional, monthly_volume_to_date):
    rate = _rate_for_volume(monthly_volume_to_date)
    return round(notional * rate, 2)


def _rate_for_volume(monthly_volume_to_date):
    for lower, upper, rate in TIERS:
        if lower <= monthly_volume_to_date < upper:
            return rate
    return TIERS[-1][2]


def estimate_monthly_fees(daily_notionals, monthly_volume_to_date):
    running_volume = monthly_volume_to_date
    total_fees = 0.0
    for notional in daily_notionals:
        total_fees += calculate_fee(notional, running_volume)
        running_volume += notional
    return round(total_fees, 2)
