package com.fusiondemo.gateway.reports;

import javax.servlet.http.HttpServletRequest;

/**
 * HTTP entry point for downloading a previously generated report.
 * Mounted under /reports/download.
 */
public class ReportDownloadController {

    private final ReportDownloadService reportDownloadService;

    public ReportDownloadController(ReportDownloadService reportDownloadService) {
        this.reportDownloadService = reportDownloadService;
    }

    /**
     * GET /reports/download?file=settlement-2026-08-01.csv
     */
    public byte[] download(HttpServletRequest request) {
        String fileName = request.getParameter("file");
        if (fileName == null || fileName.isEmpty()) {
            throw new IllegalArgumentException("file is required");
        }
        return reportDownloadService.loadReport(fileName);
    }
}
