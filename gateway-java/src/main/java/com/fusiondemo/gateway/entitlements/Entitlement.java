package com.fusiondemo.gateway.entitlements;

/**
 * A single named capability a role can be granted, e.g. "view_reports"
 * or "manage_api_keys".
 */
public enum Entitlement {
    VIEW_REPORTS,
    EXPORT_REPORTS,
    MANAGE_API_KEYS,
    MANAGE_USERS,
    VIEW_AUDIT_LOG,
    RUN_DIAGNOSTICS,
    MANAGE_WEBHOOKS
}
