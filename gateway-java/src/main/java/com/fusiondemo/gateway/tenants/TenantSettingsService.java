package com.fusiondemo.gateway.tenants;

/**
 * Business logic for reading and updating tenant settings, including
 * the validation rules that keep values within sane operational
 * bounds regardless of which endpoint triggers the change.
 */
public class TenantSettingsService {

    private static final int MIN_SESSION_TIMEOUT_MINUTES = 5;
    private static final int MAX_SESSION_TIMEOUT_MINUTES = 480;
    private static final int MAX_ALLOWED_API_KEYS = 50;

    private final TenantSettingsRepository repository;

    public TenantSettingsService(TenantSettingsRepository repository) {
        this.repository = repository;
    }

    public TenantSettings getSettings(String tenantId) {
        return repository.getOrCreate(tenantId);
    }

    public TenantSettings updateSessionTimeout(String tenantId, int newTimeoutMinutes) {
        if (newTimeoutMinutes < MIN_SESSION_TIMEOUT_MINUTES || newTimeoutMinutes > MAX_SESSION_TIMEOUT_MINUTES) {
            throw new IllegalArgumentException(
                    "Session timeout must be between " + MIN_SESSION_TIMEOUT_MINUTES
                            + " and " + MAX_SESSION_TIMEOUT_MINUTES + " minutes");
        }
        TenantSettings settings = repository.getOrCreate(tenantId);
        settings.setSessionTimeoutMinutes(newTimeoutMinutes);
        repository.save(settings);
        return settings;
    }

    public TenantSettings updateMaxApiKeys(String tenantId, int newMax) {
        if (newMax < 0 || newMax > MAX_ALLOWED_API_KEYS) {
            throw new IllegalArgumentException("maxApiKeysPerTenant must be between 0 and " + MAX_ALLOWED_API_KEYS);
        }
        TenantSettings settings = repository.getOrCreate(tenantId);
        settings.setMaxApiKeysPerTenant(newMax);
        repository.save(settings);
        return settings;
    }

    public TenantSettings setWebhooksEnabled(String tenantId, boolean enabled) {
        TenantSettings settings = repository.getOrCreate(tenantId);
        settings.setWebhooksEnabled(enabled);
        repository.save(settings);
        return settings;
    }

    public TenantSettings setPdfExportEnabled(String tenantId, boolean enabled) {
        TenantSettings settings = repository.getOrCreate(tenantId);
        settings.setPdfExportEnabled(enabled);
        repository.save(settings);
        return settings;
    }
}
