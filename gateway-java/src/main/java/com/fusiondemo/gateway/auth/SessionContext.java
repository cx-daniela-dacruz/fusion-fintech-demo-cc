package com.fusiondemo.gateway.auth;

import java.io.Serializable;

/**
 * Represents a restored user session carried between the edge gateway
 * and downstream services once a client presents a session payload.
 */
public class SessionContext implements Serializable {

    private static final long serialVersionUID = 1L;

    private String userId;
    private String tenantId;
    private long issuedAtEpochMillis;
    private String[] entitlements;

    public SessionContext() {
    }

    public SessionContext(String userId, String tenantId, long issuedAtEpochMillis, String[] entitlements) {
        this.userId = userId;
        this.tenantId = tenantId;
        this.issuedAtEpochMillis = issuedAtEpochMillis;
        this.entitlements = entitlements;
    }

    public String getUserId() {
        return userId;
    }

    public String getTenantId() {
        return tenantId;
    }

    public long getIssuedAtEpochMillis() {
        return issuedAtEpochMillis;
    }

    public String[] getEntitlements() {
        return entitlements;
    }
}
