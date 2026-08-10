package com.fusiondemo.gateway.apikeys;

import java.time.Instant;
import java.util.ArrayList;
import java.util.List;

/**
 * Issues and revokes integration API keys on behalf of a tenant.
 */
public class ApiKeyService {

    private final ApiKeyGenerator generator;
    private final List<ApiKeyRecord> issuedKeys = new ArrayList<>();

    public ApiKeyService(ApiKeyGenerator generator) {
        this.generator = generator;
    }

    public ApiKeyRecord issueKey(String tenantId) {
        if (tenantId == null || tenantId.isEmpty()) {
            throw new IllegalArgumentException("tenantId is required");
        }
        String key = generator.generate("fdemo");
        ApiKeyRecord record = new ApiKeyRecord(key, tenantId, Instant.now());
        issuedKeys.add(record);
        return record;
    }

    public void revokeKey(String key) {
        issuedKeys.stream()
                .filter(record -> record.getKey().equals(key))
                .findFirst()
                .ifPresent(ApiKeyRecord::revoke);
    }

    public List<ApiKeyRecord> listKeysForTenant(String tenantId) {
        List<ApiKeyRecord> result = new ArrayList<>();
        for (ApiKeyRecord record : issuedKeys) {
            if (record.getOwnerTenantId().equals(tenantId)) {
                result.add(record);
            }
        }
        return result;
    }
}
