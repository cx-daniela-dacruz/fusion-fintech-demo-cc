package com.fusiondemo.gateway.compliance;

import java.time.Instant;

/**
 * A compliance flag raised against a tenant or user, e.g. pending KYC
 * refresh or a sanctions screening hit that needs manual review.
 */
public class ComplianceFlag {

    private final String flagId;
    private final String subjectId;
    private final String flagType;
    private final String severity;
    private final Instant raisedAt;
    private boolean resolved;

    public ComplianceFlag(String flagId, String subjectId, String flagType, String severity, Instant raisedAt) {
        this.flagId = flagId;
        this.subjectId = subjectId;
        this.flagType = flagType;
        this.severity = severity;
        this.raisedAt = raisedAt;
        this.resolved = false;
    }

    public String getFlagId() {
        return flagId;
    }

    public String getSubjectId() {
        return subjectId;
    }

    public String getFlagType() {
        return flagType;
    }

    public String getSeverity() {
        return severity;
    }

    public Instant getRaisedAt() {
        return raisedAt;
    }

    public boolean isResolved() {
        return resolved;
    }

    public void resolve() {
        this.resolved = true;
    }
}
