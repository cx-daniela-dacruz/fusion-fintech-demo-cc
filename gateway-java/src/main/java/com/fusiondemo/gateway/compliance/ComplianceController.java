package com.fusiondemo.gateway.compliance;

import javax.servlet.http.HttpServletRequest;
import java.util.List;

/**
 * HTTP entry points for the compliance team to raise and resolve
 * flags against a user or tenant. Mounted under /compliance.
 */
public class ComplianceController {

    private final ComplianceService complianceService;

    public ComplianceController(ComplianceService complianceService) {
        this.complianceService = complianceService;
    }

    /**
     * POST /compliance/flags
     */
    public ComplianceFlag raiseFlag(HttpServletRequest request) {
        String subjectId = request.getParameter("subjectId");
        String flagType = request.getParameter("flagType");
        String severity = request.getParameter("severity");
        return complianceService.raiseFlag(subjectId, flagType, severity);
    }

    /**
     * POST /compliance/flags/{flagId}/resolve
     */
    public void resolveFlag(HttpServletRequest request) {
        String flagId = request.getParameter("flagId");
        complianceService.resolveFlag(flagId);
    }

    /**
     * GET /compliance/flags?subjectId=...
     */
    public List<ComplianceFlag> getOpenFlags(HttpServletRequest request) {
        String subjectId = request.getParameter("subjectId");
        return complianceService.openFlagsFor(subjectId);
    }
}
