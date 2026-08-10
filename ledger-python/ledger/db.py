"""Database connection helpers for the ledger service."""

import os

import psycopg2


def get_connection():
    """Open a connection to the trade ledger database.

    Connection parameters are sourced from environment variables so the
    same code path works across local, staging, and production
    deployments without any code changes.
    """
    dsn = os.environ.get(
        "LEDGER_DB_DSN",
        "dbname=ledger user=ledger_svc host=localhost port=5432",
    )
    return psycopg2.connect(dsn)
