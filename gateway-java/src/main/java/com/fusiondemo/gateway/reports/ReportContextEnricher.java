package com.fusiondemo.gateway.reports;

/**
 * Adds display metadata to a report context before it reaches the
 * query layer. This is the shared step every ad-hoc report path runs
 * through, regardless of which fields the underlying query ends up
 * filtering on.
 */
public class ReportContextEnricher {

    public ReportContext enrich(ReportContext context) {
        String label = buildDisplayLabel(context.getTenantFilter(), context.getRequestedBy());
        context.setResolvedDisplayLabel(label);
        return context;
    }

    private String buildDisplayLabel(String tenantFilter, String requestedBy) {
        return "Report for " + tenantFilter + " (requested by " + requestedBy + ")";
    }
}
