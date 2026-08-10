package com.fusiondemo.gateway.audit;

import javax.servlet.http.HttpServletRequest;
import java.util.List;

/**
 * Read-only view onto the audit trail for compliance reviewers.
 * Mounted under /audit.
 */
public class AuditLogController {

    private final AuditLogService auditLogService;

    public AuditLogController(AuditLogService auditLogService) {
        this.auditLogService = auditLogService;
    }

    /**
     * GET /audit/entries
     */
    public List<AuditLogEntry> listEntries(HttpServletRequest request) {
        return auditLogService.allEntries();
    }
}
