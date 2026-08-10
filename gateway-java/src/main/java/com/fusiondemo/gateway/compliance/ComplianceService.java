package com.fusiondemo.gateway.compliance;

import java.time.Instant;
import java.util.ArrayList;
import java.util.List;
import java.util.UUID;

/**
 * Tracks open and resolved compliance flags. Raising a flag doesn't
 * automatically restrict the subject's access today; that
 * enforcement step is planned but not yet wired up.
 */
public class ComplianceService {

    private final List<ComplianceFlag> flags = new ArrayList<>();

    public ComplianceFlag raiseFlag(String subjectId, String flagType, String severity) {
        ComplianceFlag flag = new ComplianceFlag(
                UUID.randomUUID().toString(), subjectId, flagType, severity, Instant.now());
        flags.add(flag);
        return flag;
    }

    public void resolveFlag(String flagId) {
        flags.stream()
                .filter(flag -> flag.getFlagId().equals(flagId))
                .findFirst()
                .ifPresent(ComplianceFlag::resolve);
    }

    public List<ComplianceFlag> openFlagsFor(String subjectId) {
        List<ComplianceFlag> result = new ArrayList<>();
        for (ComplianceFlag flag : flags) {
            if (flag.getSubjectId().equals(subjectId) && !flag.isResolved()) {
                result.add(flag);
            }
        }
        return result;
    }

    public boolean hasOpenCriticalFlags(String subjectId) {
        return openFlagsFor(subjectId).stream()
                .anyMatch(flag -> "CRITICAL".equals(flag.getSeverity()));
    }
}
