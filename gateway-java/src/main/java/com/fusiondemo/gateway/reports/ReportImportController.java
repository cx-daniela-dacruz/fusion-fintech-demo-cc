package com.fusiondemo.gateway.reports;

import org.w3c.dom.Document;

import javax.servlet.http.HttpServletRequest;
import java.io.BufferedReader;
import java.io.IOException;

/**
 * HTTP entry point for partners uploading a custom report definition.
 * Mounted under /reports/definitions.
 */
public class ReportImportController {

    private final ReportImportService reportImportService;

    public ReportImportController(ReportImportService reportImportService) {
        this.reportImportService = reportImportService;
    }

    /**
     * POST /reports/definitions
     */
    public String importDefinition(HttpServletRequest request) throws IOException {
        String xmlPayload = readBody(request);
        Document document = reportImportService.parseReportDefinition(xmlPayload);
        return reportImportService.extractReportName(document);
    }

    private String readBody(HttpServletRequest request) throws IOException {
        StringBuilder body = new StringBuilder();
        try (BufferedReader reader = request.getReader()) {
            String line;
            while ((line = reader.readLine()) != null) {
                body.append(line);
            }
        }
        return body.toString();
    }
}
