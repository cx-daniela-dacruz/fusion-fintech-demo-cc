package com.fusiondemo.gateway.sessionactivity;

import java.time.Instant;

/**
 * A single recorded session lifecycle event: created, refreshed, or
 * expired.
 */
public class SessionActivityEntry {

    private final String userId;
    private final String eventType;
    private final String sourceIp;
    private final Instant occurredAt;

    public SessionActivityEntry(String userId, String eventType, String sourceIp, Instant occurredAt) {
        this.userId = userId;
        this.eventType = eventType;
        this.sourceIp = sourceIp;
        this.occurredAt = occurredAt;
    }

    public String getUserId() {
        return userId;
    }

    public String getEventType() {
        return eventType;
    }

    public String getSourceIp() {
        return sourceIp;
    }

    public Instant getOccurredAt() {
        return occurredAt;
    }
}
