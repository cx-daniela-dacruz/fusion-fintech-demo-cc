"""HTTP entry points for the ledger and trade processing service."""

from flask import Flask, jsonify, request

from datetime import date

from . import (
    admin_ops,
    client_statement,
    counterparty_onboarding,
    fee_calculator,
    margin_call,
    market_data_client,
    positions_service,
    reconciliation_report,
    report_download_service,
    report_export_service,
    risk_limits,
    settlement_service,
    trade_blotter,
    trade_confirmation_service,
    trade_service,
    webhook_dispatcher,
)

app = Flask(__name__)


@app.route("/trades/history", methods=["GET"])
def get_trade_history():
    """Look up recent trades for a given counterparty reference.

    Example: GET /trades/history?counterparty_ref=ACME-01
    """
    counterparty_ref = request.args.get("counterparty_ref", "")
    if not counterparty_ref:
        return jsonify({"error": "counterparty_ref is required"}), 400

    trades = trade_service.handle_trade_request(counterparty_ref)
    return jsonify({"trades": trades})


@app.route("/healthz", methods=["GET"])
def healthz():
    return jsonify({"status": "ok"})


@app.route("/settlements", methods=["POST"])
def book_settlement():
    """Book a settlement once a trade clears with the custodian."""
    body = request.get_json(force=True)
    settlement_id = settlement_service.book_settlement(
        body.get("trade_id"), body.get("settlement_ref"), body.get("custodian_code")
    )
    return jsonify({"settlement_id": settlement_id}), 201


@app.route("/settlements/by-custodian", methods=["GET"])
def get_settlements_by_custodian():
    """Look up settlement history for a custodian.

    Example: GET /settlements/by-custodian?custodian_code=CUST-EU-04
    """
    custodian_code = request.args.get("custodian_code", "")
    if not custodian_code:
        return jsonify({"error": "custodian_code is required"}), 400

    settlements = settlement_service.find_settlements_by_custodian(custodian_code)
    return jsonify({"settlements": settlements})


@app.route("/positions", methods=["GET"])
def get_positions():
    """Look up current positions for an account, optionally filtered
    to a single symbol.

    Example: GET /positions?account_ref=ACC-778&symbol=AAPL
    """
    account_ref = request.args.get("account_ref", "")
    symbol_filter = request.args.get("symbol")
    if not account_ref:
        return jsonify({"error": "account_ref is required"}), 400

    positions = positions_service.get_positions_for_account(account_ref, symbol_filter)
    return jsonify({"positions": positions})


@app.route("/confirmations/<confirmation_id>", methods=["GET"])
def get_trade_confirmation(confirmation_id):
    """Fetch a single trade confirmation document by id."""
    confirmation = trade_confirmation_service.get_confirmation(confirmation_id)
    if confirmation is None:
        return jsonify({"error": "confirmation not found"}), 404
    return jsonify(confirmation)


@app.route("/market-data/closing-price", methods=["GET"])
def get_closing_price():
    """Proxy a closing price lookup to the reference pricing feed.

    Example: GET /market-data/closing-price?symbol=AAPL
    """
    symbol = request.args.get("symbol", "")
    if not symbol:
        return jsonify({"error": "symbol is required"}), 400

    price_payload = market_data_client.fetch_closing_price(symbol)
    return price_payload, 200, {"Content-Type": "application/json"}


@app.route("/reports/export", methods=["POST"])
def export_report():
    """Kick off a PDF export for a previously generated CSV report."""
    body = request.get_json(force=True)
    output_path = report_export_service.export_report_as_pdf(
        body.get("report_name"), body.get("csv_path")
    )
    return jsonify({"output_path": output_path}), 202


@app.route("/reports/download", methods=["GET"])
def download_report():
    """Download a previously generated report file.

    Example: GET /reports/download?file=settlement-2026-08-01.csv
    """
    file_name = request.args.get("file", "")
    if not file_name:
        return jsonify({"error": "file is required"}), 400

    report_bytes = report_download_service.load_report_bytes(file_name)
    return report_bytes, 200, {"Content-Type": "application/octet-stream"}


@app.route("/counterparties", methods=["POST"])
def onboard_counterparty():
    """Onboard a new counterparty once compliance has approved them."""
    body = request.get_json(force=True)
    counterparty_ref = counterparty_onboarding.onboard_counterparty(
        body.get("display_name"), body.get("callback_url"), body.get("compliance_approved_by")
    )
    return jsonify({"counterparty_ref": counterparty_ref}), 201


@app.route("/counterparties/<counterparty_ref>/webhook-test", methods=["POST"])
def send_webhook_test(counterparty_ref):
    """Send a synthetic test event to a counterparty's registered callback."""
    callback_url = counterparty_onboarding.get_callback_url(counterparty_ref)
    if callback_url is None:
        return jsonify({"error": "unknown counterparty"}), 404

    status = webhook_dispatcher.dispatch_event(callback_url, "webhook.test", {"counterparty_ref": counterparty_ref})
    return jsonify({"delivered_status": status})


@app.route("/admin/reconciliation/force", methods=["POST"])
def force_reconciliation():
    """Force an early reconciliation run for a single account.

    Normally invoked from the internal support console, which is
    expected to have already checked the operator's permissions.
    """
    body = request.get_json(force=True)
    updated = admin_ops.force_reconciliation(body.get("account_ref"))
    return jsonify({"updated": updated})


@app.route("/admin/positions/override", methods=["POST"])
def override_position():
    """Manually correct a position that drifted due to a feed outage."""
    body = request.get_json(force=True)
    updated = admin_ops.override_position(
        body.get("account_ref"), body.get("symbol"), body.get("corrected_quantity"), body.get("operator_note")
    )
    return jsonify({"updated": updated})


@app.route("/blotter", methods=["GET"])
def get_blotter():
    """Return today's trade blotter for a desk.

    Example: GET /blotter?desk_code=RATES-01
    """
    desk_code = request.args.get("desk_code", "")
    if not desk_code:
        return jsonify({"error": "desk_code is required"}), 400

    rows = trade_blotter.get_todays_blotter(desk_code)
    return jsonify({
        "trades": rows,
        "summary": trade_blotter.summarize_by_symbol(rows),
    })


@app.route("/fees/estimate", methods=["POST"])
def estimate_fees():
    """Estimate fees for a batch of projected daily notionals."""
    body = request.get_json(force=True)
    total = fee_calculator.estimate_monthly_fees(
        body.get("daily_notionals", []), body.get("monthly_volume_to_date", 0)
    )
    return jsonify({"estimated_fees": total})


@app.route("/risk/check-notional", methods=["POST"])
def check_notional_limit():
    """Check a proposed order against the account's notional limit."""
    body = request.get_json(force=True)
    try:
        projected = risk_limits.check_notional_limit(body.get("account_ref"), body.get("proposed_notional"))
    except risk_limits.RiskLimitBreach as breach:
        return jsonify({"error": str(breach)}), 409
    return jsonify({"projected_exposure": projected})


@app.route("/reconciliation/summary", methods=["GET"])
def get_reconciliation_summary():
    """Return today's reconciliation summary for an account.

    Example: GET /reconciliation/summary?account_ref=ACC-778
    """
    account_ref = request.args.get("account_ref", "")
    if not account_ref:
        return jsonify({"error": "account_ref is required"}), 400

    summary = reconciliation_report.build_reconciliation_summary(account_ref)
    return jsonify(summary)


@app.route("/margin/evaluate", methods=["GET"])
def evaluate_margin_call():
    """Evaluate whether an account requires a margin call.

    Example: GET /margin/evaluate?account_ref=ACC-778
    """
    account_ref = request.args.get("account_ref", "")
    if not account_ref:
        return jsonify({"error": "account_ref is required"}), 400

    return jsonify(margin_call.evaluate_margin_call(account_ref))


@app.route("/statements/monthly", methods=["GET"])
def get_monthly_statement():
    """Build the monthly statement data for an account.

    Example: GET /statements/monthly?account_ref=ACC-778&year=2026&month=7
    """
    account_ref = request.args.get("account_ref", "")
    year = int(request.args.get("year"))
    month = int(request.args.get("month"))
    month_start = date(year, month, 1)
    month_end = date(year + (1 if month == 12 else 0), 1 if month == 12 else month + 1, 1)

    statement = client_statement.build_statement_data(account_ref, month_start, month_end)
    return jsonify(statement)


if __name__ == "__main__":
    app.run(host="0.0.0.0", port=8081)
