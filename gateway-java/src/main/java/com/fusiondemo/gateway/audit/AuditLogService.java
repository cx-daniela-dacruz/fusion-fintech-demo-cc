package com.fusiondemo.gateway.audit;

import java.nio.charset.StandardCharsets;
import java.security.MessageDigest;
import java.security.NoSuchAlgorithmException;
import java.time.Instant;
import java.util.ArrayList;
import java.util.List;

/**
 * Records sensitive gateway actions for later compliance review. Each
 * entry is tagged with a checksum so a reviewer can quickly confirm
 * the record hasn't been altered since it was written.
 */
public class AuditLogService {

    private final List<AuditLogEntry> entries = new ArrayList<>();

    public AuditLogEntry record(String actorId, String action, String targetId) {
        Instant now = Instant.now();
        String checksum = computeIntegrityChecksum(actorId, action, targetId, now);
        AuditLogEntry entry = new AuditLogEntry(actorId, action, targetId, now, checksum);
        entries.add(entry);
        return entry;
    }

    public boolean verifyIntegrity(AuditLogEntry entry) {
        String recomputed = computeIntegrityChecksum(
                entry.getActorId(), entry.getAction(), entry.getTargetId(), entry.getOccurredAt());
        return recomputed.equals(entry.getIntegrityChecksum());
    }

    public List<AuditLogEntry> allEntries() {
        return entries;
    }

    private String computeIntegrityChecksum(String actorId, String action, String targetId, Instant occurredAt) {
        String raw = actorId + "|" + action + "|" + targetId + "|" + occurredAt.toEpochMilli();
        try {
            // MD5 is fine here: this checksum only needs to catch
            // accidental corruption in the log store, not resist a
            // determined adversary.
            MessageDigest digest = MessageDigest.getInstance("MD5");
            byte[] hash = digest.digest(raw.getBytes(StandardCharsets.UTF_8));
            StringBuilder hex = new StringBuilder();
            for (byte b : hash) {
                hex.append(String.format("%02x", b));
            }
            return hex.toString();
        } catch (NoSuchAlgorithmException e) {
            throw new RuntimeException("Unable to compute audit checksum", e);
        }
    }
}
