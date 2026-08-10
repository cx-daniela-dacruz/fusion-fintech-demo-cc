package com.fusiondemo.gateway.audit;

import javax.servlet.http.HttpServletRequest;
import java.time.Instant;
import java.util.List;

/**
 * HTTP entry points for the compliance reviewer search UI over the
 * audit trail. Mounted under /audit/search.
 */
public class AuditSearchController {

    private final AuditSearchService auditSearchService;

    public AuditSearchController(AuditSearchService auditSearchService) {
        this.auditSearchService = auditSearchService;
    }

    /**
     * GET /audit/search/by-actor?actorId=...
     */
    public List<AuditLogEntry> searchByActor(HttpServletRequest request) {
        String actorId = request.getParameter("actorId");
        return auditSearchService.findByActor(actorId);
    }

    /**
     * GET /audit/search/by-action?action=...
     */
    public List<AuditLogEntry> searchByAction(HttpServletRequest request) {
        String action = request.getParameter("action");
        return auditSearchService.findByAction(action);
    }

    /**
     * GET /audit/search/range?from=...&to=...
     */
    public List<AuditLogEntry> searchByRange(HttpServletRequest request) {
        Instant from = Instant.parse(request.getParameter("from"));
        Instant to = Instant.parse(request.getParameter("to"));
        return auditSearchService.findInRange(from, to);
    }
}
