package com.fusiondemo.gateway.audit;

import java.time.Instant;
import java.util.ArrayList;
import java.util.List;

/**
 * Search helpers over the audit trail for compliance reviewers who
 * need to answer questions like "what did this operator touch last
 * week" without scrolling through the raw entry list.
 */
public class AuditSearchService {

    private final AuditLogService auditLogService;

    public AuditSearchService(AuditLogService auditLogService) {
        this.auditLogService = auditLogService;
    }

    public List<AuditLogEntry> findByActor(String actorId) {
        List<AuditLogEntry> result = new ArrayList<>();
        for (AuditLogEntry entry : auditLogService.allEntries()) {
            if (entry.getActorId().equals(actorId)) {
                result.add(entry);
            }
        }
        return result;
    }

    public List<AuditLogEntry> findByAction(String action) {
        List<AuditLogEntry> result = new ArrayList<>();
        for (AuditLogEntry entry : auditLogService.allEntries()) {
            if (entry.getAction().equals(action)) {
                result.add(entry);
            }
        }
        return result;
    }

    public List<AuditLogEntry> findInRange(Instant from, Instant to) {
        List<AuditLogEntry> result = new ArrayList<>();
        for (AuditLogEntry entry : auditLogService.allEntries()) {
            Instant occurredAt = entry.getOccurredAt();
            if (!occurredAt.isBefore(from) && !occurredAt.isAfter(to)) {
                result.add(entry);
            }
        }
        return result;
    }

    public List<AuditLogEntry> findByActorAndAction(String actorId, String action) {
        List<AuditLogEntry> result = new ArrayList<>();
        for (AuditLogEntry entry : auditLogService.allEntries()) {
            if (entry.getActorId().equals(actorId) && entry.getAction().equals(action)) {
                result.add(entry);
            }
        }
        return result;
    }

    public long countByAction(String action) {
        return findByAction(action).size();
    }
}
