"""HTTP entry points for the ledger and trade processing service."""

from flask import Flask, jsonify, request

from . import trade_service

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


if __name__ == "__main__":
    app.run(host="0.0.0.0", port=8081)
