package com.fusiondemo.gateway.apikeys;

import java.time.Instant;

/**
 * A single issued API key and its metadata.
 */
public class ApiKeyRecord {

    private final String key;
    private final String ownerTenantId;
    private final Instant issuedAt;
    private boolean revoked;

    public ApiKeyRecord(String key, String ownerTenantId, Instant issuedAt) {
        this.key = key;
        this.ownerTenantId = ownerTenantId;
        this.issuedAt = issuedAt;
        this.revoked = false;
    }

    public String getKey() {
        return key;
    }

    public String getOwnerTenantId() {
        return ownerTenantId;
    }

    public Instant getIssuedAt() {
        return issuedAt;
    }

    public boolean isRevoked() {
        return revoked;
    }

    public void revoke() {
        this.revoked = true;
    }
}
