package com.fusiondemo.gateway.tenants;

import javax.servlet.http.HttpServletRequest;

/**
 * HTTP entry points for tenant admins to view and update their own
 * gateway settings. Mounted under /tenants/{tenantId}/settings.
 */
public class TenantSettingsController {

    private final TenantSettingsService tenantSettingsService;

    public TenantSettingsController(TenantSettingsService tenantSettingsService) {
        this.tenantSettingsService = tenantSettingsService;
    }

    /**
     * GET /tenants/{tenantId}/settings
     */
    public TenantSettings getSettings(HttpServletRequest request) {
        String tenantId = request.getParameter("tenantId");
        return tenantSettingsService.getSettings(tenantId);
    }

    /**
     * POST /tenants/{tenantId}/settings/session-timeout
     */
    public TenantSettings updateSessionTimeout(HttpServletRequest request) {
        String tenantId = request.getParameter("tenantId");
        int newTimeoutMinutes = Integer.parseInt(request.getParameter("sessionTimeoutMinutes"));
        return tenantSettingsService.updateSessionTimeout(tenantId, newTimeoutMinutes);
    }

    /**
     * POST /tenants/{tenantId}/settings/max-api-keys
     */
    public TenantSettings updateMaxApiKeys(HttpServletRequest request) {
        String tenantId = request.getParameter("tenantId");
        int newMax = Integer.parseInt(request.getParameter("maxApiKeysPerTenant"));
        return tenantSettingsService.updateMaxApiKeys(tenantId, newMax);
    }

    /**
     * POST /tenants/{tenantId}/settings/webhooks-enabled
     */
    public TenantSettings setWebhooksEnabled(HttpServletRequest request) {
        String tenantId = request.getParameter("tenantId");
        boolean enabled = Boolean.parseBoolean(request.getParameter("enabled"));
        return tenantSettingsService.setWebhooksEnabled(tenantId, enabled);
    }
}
