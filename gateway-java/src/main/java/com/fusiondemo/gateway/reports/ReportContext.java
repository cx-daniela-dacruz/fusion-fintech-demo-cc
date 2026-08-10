package com.fusiondemo.gateway.reports;

/**
 * Carries the parameters for an ad-hoc report query as it moves
 * through enrichment before reaching the query layer. Kept as a plain
 * mutable holder so additional fields can be layered on without
 * changing every method signature along the way.
 */
public class ReportContext {

    private String tenantFilter;
    private String requestedBy;
    private String resolvedDisplayLabel;

    public ReportContext(String tenantFilter, String requestedBy) {
        this.tenantFilter = tenantFilter;
        this.requestedBy = requestedBy;
    }

    public String getTenantFilter() {
        return tenantFilter;
    }

    public String getRequestedBy() {
        return requestedBy;
    }

    public String getResolvedDisplayLabel() {
        return resolvedDisplayLabel;
    }

    public void setResolvedDisplayLabel(String resolvedDisplayLabel) {
        this.resolvedDisplayLabel = resolvedDisplayLabel;
    }
}
