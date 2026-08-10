package com.fusiondemo.gateway.ratelimit;

import java.util.HashMap;
import java.util.Map;

/**
 * Tracks request counts per caller within the current fixed window
 * and decides whether a new request should be allowed. Backed by an
 * in-memory map for this demo; production uses the shared Redis
 * counter service so limits are enforced consistently across
 * gateway instances.
 */
public class RateLimitService {

    private final Map<String, WindowCounter> countersByCallerKey = new HashMap<>();

    public boolean tryConsume(String callerKey, RateLimitPolicy policy) {
        long nowSeconds = System.currentTimeMillis() / 1000;
        WindowCounter counter = countersByCallerKey.computeIfAbsent(callerKey, key -> new WindowCounter(nowSeconds));

        if (nowSeconds - counter.windowStartSeconds >= policy.getWindowSeconds()) {
            counter.windowStartSeconds = nowSeconds;
            counter.count = 0;
        }

        if (counter.count >= policy.getMaxRequests()) {
            return false;
        }

        counter.count++;
        return true;
    }

    private static final class WindowCounter {
        long windowStartSeconds;
        int count;

        WindowCounter(long windowStartSeconds) {
            this.windowStartSeconds = windowStartSeconds;
            this.count = 0;
        }
    }
}
