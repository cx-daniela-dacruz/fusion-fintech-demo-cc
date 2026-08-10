package com.fusiondemo.gateway.sessionactivity;

import java.time.Instant;
import java.util.ArrayList;
import java.util.List;

/**
 * Records session lifecycle events for later review, e.g. when a user
 * reports they were logged out unexpectedly and support needs to see
 * what happened.
 */
public class SessionActivityService {

    private final List<SessionActivityEntry> entries = new ArrayList<>();

    public SessionActivityEntry recordCreated(String userId, String sourceIp) {
        return record(userId, "SESSION_CREATED", sourceIp);
    }

    public SessionActivityEntry recordRefreshed(String userId, String sourceIp) {
        return record(userId, "SESSION_REFRESHED", sourceIp);
    }

    public SessionActivityEntry recordExpired(String userId, String sourceIp) {
        return record(userId, "SESSION_EXPIRED", sourceIp);
    }

    private SessionActivityEntry record(String userId, String eventType, String sourceIp) {
        SessionActivityEntry entry = new SessionActivityEntry(userId, eventType, sourceIp, Instant.now());
        entries.add(entry);
        return entry;
    }

    public List<SessionActivityEntry> historyForUser(String userId) {
        List<SessionActivityEntry> result = new ArrayList<>();
        for (SessionActivityEntry entry : entries) {
            if (entry.getUserId().equals(userId)) {
                result.add(entry);
            }
        }
        return result;
    }
}
