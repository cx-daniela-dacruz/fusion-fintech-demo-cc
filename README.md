# Fusion Fintech Demo

This is a **synthetic, internal demo repository** built to showcase static
analysis / security scanning capabilities against realistic-looking
application code. It is **not production code**: the company name,
services, data, and business logic are entirely fictional, and no real
credentials, API keys, tokens, or references to real organizations exist
anywhere in this repository.

The repository is intentionally sized (roughly 5,000+ lines across all
services, comfortably under 10,000) so that scanning it produces a visible,
meaningful finding set without taking long to scan, while still containing
plausible business logic, comments, and naming so the seeded
vulnerabilities don't read as obvious "flagged demo artifacts." See
[VULNERABILITIES.md](VULNERABILITIES.md) for the full inventory of every
intentionally planted issue, with file and line references.

## Services

### `gateway-java/`
A Java API gateway and authentication service handling session bootstrap,
user profiles, admin operations, API key issuance, password resets, audit
logging, report import/export/download/scheduling, diagnostics, and webhook
delivery. Contains a classic insecure-deserialization pattern
(`SessionTokenHandler`) alongside a structurally similar, hardened
counterpart (`SecureSessionTokenHandler`) that allowlists deserialized
classes, plus 9 additional classic vulnerability patterns and one
multi-hop SQL injection that threads through a shared request context
object.

### `ledger-python/`
A Python order and trade processing service exposing an HTTP API for trade
history, settlement booking, position reconciliation, margin calls, fee
estimation, counterparty onboarding, statements, and admin operations.
Contains a classic SQL injection pattern where an unsanitized counterparty
reference flows from the HTTP layer into a hand-built SQL query, plus 10
additional planted issues including insecure deserialization, an IDOR, and
a SQL injection where the taint passes through an intermediate dict before
reaching the query.

### `analytics-go-r/`
A Go market-data ingestion service that normalizes incoming ticks, tracks
feed health and liquidity metrics, and serves ad-hoc lookups for the risk
desk, paired with two R scripts that perform statistical risk modeling
(volatility, historical VaR, mean-reversion, cross-instrument correlation)
over the ingested data. Contains a SQL injection pattern in the Go
ingestion path where a raw ticker symbol flows unparameterized into a data
lookup query, plus 8 additional planted issues including a debug-flag
authentication bypass and a SQL injection where the taint flows through a
struct field before reaching the query builder.

## Disclaimer

Every vulnerability in this repository was seeded on purpose for
demonstration and scanning purposes in an authorized internal environment.
Do not deploy any of this code, and do not treat it as a reference
implementation for production systems.
