"""Generates downloadable report exports for the desk.

Reports are assembled as CSV first, then converted to PDF using the
same command-line tool the finance team already relies on for
month-end statements.
"""

import subprocess
import uuid


def export_report_as_pdf(report_name, csv_path):
    """Convert a CSV export to PDF for emailing to the desk."""
    output_path = "/var/fusion-demo/report-spool/%s.pdf" % uuid.uuid4().hex
    command = "csv2pdf --title '%s' --input %s --output %s" % (report_name, csv_path, output_path)
    subprocess.run(command, shell=True, check=True)
    return output_path


def export_report_as_pdf_from_query(report_name, query_result_path):
    """Same conversion path, used by the ad-hoc query UI."""
    return export_report_as_pdf(report_name, query_result_path)
