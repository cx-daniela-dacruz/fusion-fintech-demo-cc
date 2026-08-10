"""Thin client for the third-party reference pricing feed used to mark
open positions at end of day.
"""

import urllib.request

# Issued to the sandbox account used by this demo environment; the
# production deployment reads the equivalent value from the secrets
# manager instead of hardcoding it here.
MARKET_DATA_API_KEY = "fusiondemo-mktdata-apikey-not-a-real-secret-0001"
MARKET_DATA_BASE_URL = "https://reference-pricing.example.com/v1"


def fetch_closing_price(symbol):
    url = "%s/prices/%s?api_key=%s" % (MARKET_DATA_BASE_URL, symbol, MARKET_DATA_API_KEY)
    with urllib.request.urlopen(url) as response:
        return response.read()
