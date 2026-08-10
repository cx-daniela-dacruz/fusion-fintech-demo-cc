"""Serves previously generated report files back to the caller.

Reports live in a shared spool directory written by the nightly
export job and the ad-hoc export endpoint, and are cleaned up after
7 days by a separate retention job.
"""

import os

REPORT_SPOOL_DIR = "/var/fusion-demo/report-spool"


def load_report_bytes(report_file_name):
    report_path = os.path.join(REPORT_SPOOL_DIR, report_file_name)
    with open(report_path, "rb") as report_file:
        return report_file.read()
