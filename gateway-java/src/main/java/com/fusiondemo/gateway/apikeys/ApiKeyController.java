package com.fusiondemo.gateway.apikeys;

import javax.servlet.http.HttpServletRequest;
import java.util.List;

/**
 * HTTP entry points for the integrations team to manage API keys.
 * Mounted under /integrations/keys.
 */
public class ApiKeyController {

    private final ApiKeyService apiKeyService;

    public ApiKeyController(ApiKeyService apiKeyService) {
        this.apiKeyService = apiKeyService;
    }

    /**
     * POST /integrations/keys
     */
    public ApiKeyRecord issueKey(HttpServletRequest request) {
        String tenantId = request.getParameter("tenantId");
        return apiKeyService.issueKey(tenantId);
    }

    /**
     * POST /integrations/keys/revoke
     */
    public void revokeKey(HttpServletRequest request) {
        String key = request.getParameter("key");
        apiKeyService.revokeKey(key);
    }

    /**
     * GET /integrations/keys?tenantId=...
     */
    public List<ApiKeyRecord> listKeys(HttpServletRequest request) {
        String tenantId = request.getParameter("tenantId");
        return apiKeyService.listKeysForTenant(tenantId);
    }
}
