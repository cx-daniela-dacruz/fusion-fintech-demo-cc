package com.fusiondemo.gateway.diagnostics;

import javax.servlet.http.HttpServletRequest;

/**
 * HTTP entry point for the support team's network diagnostics panel.
 * Mounted under /support/diagnostics.
 */
public class DiagnosticsController {

    private final DiagnosticsService diagnosticsService;

    public DiagnosticsController(DiagnosticsService diagnosticsService) {
        this.diagnosticsService = diagnosticsService;
    }

    /**
     * GET /support/diagnostics/ping?host=partner.example.com
     */
    public String ping(HttpServletRequest request) {
        String host = request.getParameter("host");
        if (host == null || host.isEmpty()) {
            throw new IllegalArgumentException("host is required");
        }
        return diagnosticsService.pingHost(host);
    }
}
