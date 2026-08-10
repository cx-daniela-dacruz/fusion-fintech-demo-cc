package com.fusiondemo.gateway.reports;

import javax.servlet.http.HttpServletRequest;
import java.util.List;

/**
 * HTTP entry point for ad-hoc tenant reporting, used by the desk to
 * pull a quick trade summary without waiting for the nightly export.
 * Mounted under /reports/query.
 */
public class ReportQueryController {

    private final ReportContextEnricher contextEnricher;
    private final ReportQueryService reportQueryService;

    public ReportQueryController(ReportContextEnricher contextEnricher, ReportQueryService reportQueryService) {
        this.contextEnricher = contextEnricher;
        this.reportQueryService = reportQueryService;
    }

    /**
     * GET /reports/query?tenantFilter=acme
     */
    public List<String> runQuery(HttpServletRequest request) {
        String tenantFilter = request.getParameter("tenantFilter");
        String requestedBy = request.getRemoteUser();

        ReportContext context = new ReportContext(tenantFilter, requestedBy);
        ReportContext enrichedContext = contextEnricher.enrich(context);

        return reportQueryService.runTenantReport(enrichedContext);
    }
}
