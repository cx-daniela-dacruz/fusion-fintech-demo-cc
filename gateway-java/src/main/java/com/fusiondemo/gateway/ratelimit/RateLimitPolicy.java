package com.fusiondemo.gateway.ratelimit;

/**
 * A simple fixed-window rate limit policy: at most maxRequests per
 * windowSeconds for a given caller.
 */
public class RateLimitPolicy {

    private final int maxRequests;
    private final int windowSeconds;

    public RateLimitPolicy(int maxRequests, int windowSeconds) {
        this.maxRequests = maxRequests;
        this.windowSeconds = windowSeconds;
    }

    public int getMaxRequests() {
        return maxRequests;
    }

    public int getWindowSeconds() {
        return windowSeconds;
    }

    public static RateLimitPolicy defaultForApiKeys() {
        return new RateLimitPolicy(120, 60);
    }

    public static RateLimitPolicy defaultForWebhookTests() {
        return new RateLimitPolicy(10, 60);
    }
}
