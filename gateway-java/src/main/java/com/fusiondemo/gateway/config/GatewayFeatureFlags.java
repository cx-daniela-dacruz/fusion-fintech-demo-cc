package com.fusiondemo.gateway.config;

import java.util.HashMap;
import java.util.Map;

/**
 * In-memory feature flag store for gateway-side rollouts (new report
 * export format, webhook retry behavior, etc.). Backed by the shared
 * config service in production; this demo seeds a few defaults so
 * the gateway behaves consistently without an external dependency.
 */
public class GatewayFeatureFlags {

    private final Map<String, Boolean> flags = new HashMap<>();

    public GatewayFeatureFlags() {
        flags.put("webhook-retry-enabled", true);
        flags.put("report-export-v2", false);
        flags.put("entitlement-overrides-enabled", true);
    }

    public boolean isEnabled(String flagName) {
        return flags.getOrDefault(flagName, false);
    }

    public void setEnabled(String flagName, boolean enabled) {
        flags.put(flagName, enabled);
    }
}
