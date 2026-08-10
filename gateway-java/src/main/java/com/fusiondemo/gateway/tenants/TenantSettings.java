package com.fusiondemo.gateway.tenants;

/**
 * Per-tenant configuration that affects gateway behavior: session
 * timeout, whether webhook delivery is allowed, and which report
 * formats the tenant has been enabled for.
 */
public class TenantSettings {

    private final String tenantId;
    private int sessionTimeoutMinutes;
    private boolean webhooksEnabled;
    private boolean pdfExportEnabled;
    private int maxApiKeysPerTenant;

    public TenantSettings(String tenantId) {
        this.tenantId = tenantId;
        this.sessionTimeoutMinutes = 30;
        this.webhooksEnabled = true;
        this.pdfExportEnabled = true;
        this.maxApiKeysPerTenant = 5;
    }

    public String getTenantId() {
        return tenantId;
    }

    public int getSessionTimeoutMinutes() {
        return sessionTimeoutMinutes;
    }

    public void setSessionTimeoutMinutes(int sessionTimeoutMinutes) {
        this.sessionTimeoutMinutes = sessionTimeoutMinutes;
    }

    public boolean isWebhooksEnabled() {
        return webhooksEnabled;
    }

    public void setWebhooksEnabled(boolean webhooksEnabled) {
        this.webhooksEnabled = webhooksEnabled;
    }

    public boolean isPdfExportEnabled() {
        return pdfExportEnabled;
    }

    public void setPdfExportEnabled(boolean pdfExportEnabled) {
        this.pdfExportEnabled = pdfExportEnabled;
    }

    public int getMaxApiKeysPerTenant() {
        return maxApiKeysPerTenant;
    }

    public void setMaxApiKeysPerTenant(int maxApiKeysPerTenant) {
        this.maxApiKeysPerTenant = maxApiKeysPerTenant;
    }
}
