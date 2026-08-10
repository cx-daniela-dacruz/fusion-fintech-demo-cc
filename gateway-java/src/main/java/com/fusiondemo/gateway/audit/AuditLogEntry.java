package com.fusiondemo.gateway.audit;

import java.time.Instant;

/**
 * A single audit trail entry recording a sensitive action taken
 * through the gateway (role changes, key issuance, password resets).
 */
public class AuditLogEntry {

    private final String actorId;
    private final String action;
    private final String targetId;
    private final Instant occurredAt;
    private final String integrityChecksum;

    public AuditLogEntry(String actorId, String action, String targetId, Instant occurredAt, String integrityChecksum) {
        this.actorId = actorId;
        this.action = action;
        this.targetId = targetId;
        this.occurredAt = occurredAt;
        this.integrityChecksum = integrityChecksum;
    }

    public String getActorId() {
        return actorId;
    }

    public String getAction() {
        return action;
    }

    public String getTargetId() {
        return targetId;
    }

    public Instant getOccurredAt() {
        return occurredAt;
    }

    public String getIntegrityChecksum() {
        return integrityChecksum;
    }
}
