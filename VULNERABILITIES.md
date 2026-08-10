# Vulnerability Inventory

32 intentionally planted vulnerabilities across the three services. Line
numbers point at the sink; each also has a traceable source-to-sink path
through 2-4 function calls (see the file for the full chain) unless noted
as a static misconfiguration.

## gateway-java (12)

| # | CWE | Category | File | Line | Sink |
|---|---|---|---|---|---|
| 1 | CWE-502 | Insecure Deserialization | `gateway-java/.../auth/SessionTokenHandler.java` | 34 | `ois.readObject()` |
| 2 | CWE-502 (partial mitigation) | Insufficient Verification of Data Authenticity | `gateway-java/.../auth/SecureSessionTokenHandler.java` | 43 | `ois.readObject()` inside class-allowlisted stream |
| 3 | CWE-862/863 | Missing Authorization | `gateway-java/.../admin/AdminUserController.java` | 28 | `changeRole()` — no caller-role check before privilege change |
| 4 | CWE-338 | Weak PRNG | `gateway-java/.../apikeys/ApiKeyGenerator.java` | 17 | `new Random()` used for API key generation |
| 5 | CWE-330/338 | Predictable Token | `gateway-java/.../security/PasswordResetService.java` | 44 | `random.nextInt(1_000_000)` for reset token |
| 6 | CWE-798 | Hardcoded Credentials | `gateway-java/.../security/TokenSigningService.java` | 18 | `SIGNING_SECRET` literal |
| 7 | CWE-327 | Weak Hash (MD5) | `gateway-java/.../audit/AuditLogService.java` | 43 | `MessageDigest.getInstance("MD5")` |
| 8 | CWE-611 | XXE | `gateway-java/.../reports/ReportImportService.java` | 21 | `builder.parse()` with no external-entity hardening |
| 9 | CWE-22 | Path Traversal | `gateway-java/.../reports/ReportDownloadService.java` | 19 | `Files.readAllBytes(file.toPath())` |
| 10 | CWE-78 | Command Injection | `gateway-java/.../diagnostics/DiagnosticsService.java` | 17 | `Runtime.getRuntime().exec("ping -c 3 " + hostname)` |
| 11 | CWE-918 | SSRF | `gateway-java/.../webhooks/WebhookNotifier.java` | 20 | `callbackUrl.openConnection()` on unchecked partner URL |
| 12 | CWE-89 | SQL Injection (multi-hop) | `gateway-java/.../reports/ReportQueryService.java` | 30 | `statement.executeQuery(sql)` — taint threaded through `ReportContext` → `ReportContextEnricher` → here |

## ledger-python (11)

| # | CWE | Category | File | Line | Sink |
|---|---|---|---|---|---|
| 13 | CWE-89 | SQL Injection | `ledger-python/ledger/trade_service.py` | 32 | `cursor.execute(query)` |
| 14 | CWE-89 | SQL Injection | `ledger-python/ledger/settlement_service.py` | 48/50 | `%`-formatted query, `cursor.execute(query)` |
| 15 | CWE-502 | Insecure Deserialization | `ledger-python/ledger/position_sync_worker.py` | 24 | `pickle.loads(raw_message_body)` |
| 16 | CWE-502-adjacent | Unsafe YAML Load | `ledger-python/ledger/config_loader.py` | 13 | `yaml.load(config_file, Loader=yaml.Loader)` |
| 17 | CWE-78 | Command Injection | `ledger-python/ledger/report_export_service.py` | 16 | `subprocess.run(command, shell=True, check=True)` |
| 18 | CWE-798 | Hardcoded Credentials | `ledger-python/ledger/market_data_client.py` | 10 | `MARKET_DATA_API_KEY` literal (obviously-fake placeholder value) |
| 19 | CWE-22 | Path Traversal | `ledger-python/ledger/report_download_service.py` | 15 | `open(report_path, "rb")` |
| 20 | CWE-862 | Missing Authorization | `ledger-python/ledger/admin_ops.py` | 9 | `force_reconciliation()` — reachable via `/admin/reconciliation/force` with no auth check |
| 21 | CWE-918 | SSRF | `ledger-python/ledger/webhook_dispatcher.py` | 17 | `urllib.request.urlopen(request)` on unchecked callback URL |
| 22 | CWE-89 | SQL Injection (taint through dict) | `ledger-python/ledger/positions_service.py` | 46 | `cursor.execute(query)` — taint passed through a filter dict + `_render_filter_clause` before reaching here |
| 23 | CWE-639 | IDOR | `ledger-python/ledger/trade_confirmation_service.py` | 10 | `get_confirmation(confirmation_id)` — sequential id, no check that the caller owns this confirmation |

## analytics-go-r (9)

| # | CWE | Category | File | Line | Sink |
|---|---|---|---|---|---|
| 24 | CWE-89 | SQL Injection | `analytics-go-r/internal/ingest/lookup.go` | 39 | `db.QueryRow(query)` |
| 25 | CWE-89 | SQL Injection | `analytics-go-r/internal/ingest/corpactions.go` | 56 | `db.QueryRow(query)` |
| 26 | CWE-78 | Command Injection | `analytics-go-r/internal/export/chart.go` | 43 | `exec.Command("sh", "-c", command)` |
| 27 | CWE-295 | Improper Certificate Validation | `analytics-go-r/internal/httpclient/client.go` | 18 | `tls.Config{InsecureSkipVerify: true}` |
| 28 | CWE-338 | Weak PRNG | `analytics-go-r/internal/idgen/idgen.go` | 18 | `rand.Intn()` (math/rand) for correlation IDs |
| 29 | CWE-22 | Path Traversal | `analytics-go-r/internal/download/download.go` | 42 | `os.ReadFile(fullPath)` |
| 30 | CWE-918 | SSRF | `analytics-go-r/internal/webhook/webhook.go` | 42 | `http.Post(callbackURL, ...)` on unchecked URL |
| 31 | CWE-89 | SQL Injection (taint through struct field) | `analytics-go-r/internal/orderbook/orderbook.go` | 75 | `db.QueryRow(queryBuilder.String())` — taint carried via `SnapshotRequest` struct field, not a direct parameter |
| 32 | CWE-288 | Auth Bypass (debug flag) | `analytics-go-r/internal/middleware/debug.go` | 17 | `os.Getenv("FUSION_DEBUG_MODE") == "true"` skips the operator-token check entirely |

## Notes on realistic side-effects

Earlier scans against the smaller version of this repo also surfaced
findings that weren't explicitly planted but are legitimate consequences of
writing realistic surrounding code: SCA hits on pinned dependency versions,
missing security headers (HSTS, CSP), and secondary CWE classifications on
the same SQL injection sinks (e.g. `Parameter_Tampering` alongside
`SQL_Injection`). Expect a similar pattern here at a larger scale — the 31
above are the ones to check for specifically; anything beyond that is
either a real side effect of the added business logic or scanner-specific
noise.
