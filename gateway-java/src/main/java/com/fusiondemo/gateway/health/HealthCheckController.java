package com.fusiondemo.gateway.health;

import java.util.LinkedHashMap;
import java.util.Map;

/**
 * Liveness and readiness endpoints used by the load balancer and the
 * deployment pipeline to decide whether this instance should receive
 * traffic.
 */
public class HealthCheckController {

    private volatile boolean ready = false;

    /**
     * GET /healthz/live
     */
    public Map<String, String> live() {
        Map<String, String> body = new LinkedHashMap<>();
        body.put("status", "ok");
        return body;
    }

    /**
     * GET /healthz/ready
     */
    public Map<String, String> ready() {
        Map<String, String> body = new LinkedHashMap<>();
        body.put("status", ready ? "ready" : "not-ready");
        return body;
    }

    public void markReady() {
        this.ready = true;
    }

    public void markNotReady() {
        this.ready = false;
    }
}
