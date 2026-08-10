package com.fusiondemo.gateway.entitlements;

import java.util.EnumSet;
import java.util.HashMap;
import java.util.Map;
import java.util.Set;

/**
 * Maps each role known to the gateway to the entitlements it carries.
 * Roles are coarse-grained on purpose; finer-grained overrides are
 * handled per-tenant by the entitlement service.
 */
public class EntitlementCatalog {

    private final Map<String, Set<Entitlement>> entitlementsByRole = new HashMap<>();

    public EntitlementCatalog() {
        entitlementsByRole.put("trader", EnumSet.of(
                Entitlement.VIEW_REPORTS));
        entitlementsByRole.put("risk_analyst", EnumSet.of(
                Entitlement.VIEW_REPORTS, Entitlement.EXPORT_REPORTS, Entitlement.VIEW_AUDIT_LOG));
        entitlementsByRole.put("admin", EnumSet.of(
                Entitlement.VIEW_REPORTS, Entitlement.EXPORT_REPORTS, Entitlement.MANAGE_API_KEYS,
                Entitlement.MANAGE_USERS, Entitlement.VIEW_AUDIT_LOG, Entitlement.RUN_DIAGNOSTICS,
                Entitlement.MANAGE_WEBHOOKS));
    }

    public Set<Entitlement> entitlementsFor(String role) {
        return entitlementsByRole.getOrDefault(role, EnumSet.noneOf(Entitlement.class));
    }

    public boolean roleHas(String role, Entitlement entitlement) {
        return entitlementsFor(role).contains(entitlement);
    }
}
