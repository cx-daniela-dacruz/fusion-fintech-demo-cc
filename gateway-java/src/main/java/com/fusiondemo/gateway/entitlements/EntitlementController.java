package com.fusiondemo.gateway.entitlements;

import javax.servlet.http.HttpServletRequest;
import java.util.Set;

/**
 * Read-only endpoint for the frontend to ask which entitlements the
 * current user has, so it can hide actions they're not allowed to
 * take. Mounted under /entitlements.
 */
public class EntitlementController {

    private final EntitlementService entitlementService;

    public EntitlementController(EntitlementService entitlementService) {
        this.entitlementService = entitlementService;
    }

    /**
     * GET /entitlements/{userId}
     */
    public Set<Entitlement> getEntitlements(HttpServletRequest request) {
        String userId = request.getParameter("userId");
        return entitlementService.entitlementsForUser(userId);
    }
}
