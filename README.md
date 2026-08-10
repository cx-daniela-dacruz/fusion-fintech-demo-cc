# Fusion Fintech Demo

This is a **synthetic, internal demo repository** built to showcase static
analysis / security scanning capabilities against realistic-looking
application code. It is **not production code**: the company name,
services, data, and business logic are entirely fictional, and no real
credentials, API keys, tokens, or references to real organizations exist
anywhere in this repository.

The repository is intentionally small (well under 10,000 lines across all
services) so it scans quickly during live demos, while still containing
plausible business logic, comments, and naming so the seeded
vulnerabilities don't read as obvious "flagged demo artifacts."

## Services

### `gateway-java/`
A Java API gateway and authentication service responsible for terminating
client sessions and bridging legacy clients onto the current session
model. Contains a classic insecure-deserialization pattern
(`SessionTokenHandler`) alongside a structurally similar, hardened
counterpart (`SecureSessionTokenHandler`) that allowlists deserialized
classes.

### `ledger-python/`
A Python order and trade processing service that exposes a small HTTP API
for looking up historical trades by counterparty reference. Contains a
classic SQL injection pattern where an unsanitized counterparty reference
flows from the HTTP layer into a hand-built SQL query.

### `analytics-go-r/`
A Go market-data ingestion service that normalizes incoming ticks and
serves ad-hoc lookups for the risk desk, paired with an R script that
performs statistical risk modeling (volatility, historical VaR, a simple
mean-reversion fit) over the ingested data. Contains a SQL injection
pattern in the Go ingestion path where a raw ticker symbol flows
unparameterized into a data lookup query.

## Disclaimer

Every vulnerability in this repository was seeded on purpose for
demonstration and scanning purposes in an authorized internal environment.
Do not deploy any of this code, and do not treat it as a reference
implementation for production systems.
