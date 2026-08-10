package com.fusiondemo.gateway.reports;

import java.io.File;
import java.io.IOException;
import java.nio.file.Files;

/**
 * Serves previously generated report files (PDF/CSV exports) back to
 * the requesting client. Reports are written to a shared spool
 * directory by the nightly export job and cleaned up after 7 days.
 */
public class ReportDownloadService {

    private static final String REPORT_SPOOL_DIR = "/var/fusion-demo/report-spool";

    public byte[] loadReport(String reportFileName) {
        File file = new File(REPORT_SPOOL_DIR, reportFileName);
        try {
            return Files.readAllBytes(file.toPath());
        } catch (IOException e) {
            throw new RuntimeException("Unable to read report file: " + reportFileName, e);
        }
    }
}
