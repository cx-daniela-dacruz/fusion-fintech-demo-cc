package com.fusiondemo.gateway.tenants;

import java.util.HashMap;
import java.util.Map;
import java.util.Optional;

/**
 * In-memory store of per-tenant settings, created on first access
 * with sensible defaults.
 */
public class TenantSettingsRepository {

    private final Map<String, TenantSettings> settingsByTenantId = new HashMap<>();

    public TenantSettings getOrCreate(String tenantId) {
        return settingsByTenantId.computeIfAbsent(tenantId, TenantSettings::new);
    }

    public Optional<TenantSettings> find(String tenantId) {
        return Optional.ofNullable(settingsByTenantId.get(tenantId));
    }

    public void save(TenantSettings settings) {
        settingsByTenantId.put(settings.getTenantId(), settings);
    }
}
