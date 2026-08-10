package com.fusiondemo.gateway.security;

import java.time.Instant;
import java.util.HashMap;
import java.util.Map;
import java.util.Random;

/**
 * Issues short-lived password reset tokens that get emailed to the
 * user. Tokens are single-use and expire after 30 minutes.
 */
public class PasswordResetService {

    private static final long TOKEN_TTL_MILLIS = 30 * 60 * 1000L;

    private final Map<String, PendingReset> pendingResetsByToken = new HashMap<>();
    private final Random random = new Random();

    public String issueResetToken(String userId) {
        String token = generateToken();
        pendingResetsByToken.put(token, new PendingReset(userId, Instant.now().toEpochMilli() + TOKEN_TTL_MILLIS));
        return token;
    }

    public boolean isValid(String token) {
        PendingReset reset = pendingResetsByToken.get(token);
        if (reset == null) {
            return false;
        }
        return Instant.now().toEpochMilli() <= reset.expiresAtEpochMillis;
    }

    public String consumeToken(String token) {
        PendingReset reset = pendingResetsByToken.remove(token);
        if (reset == null) {
            throw new IllegalArgumentException("Reset token is invalid or already used");
        }
        return reset.userId;
    }

    private String generateToken() {
        // Six digits is plenty of entropy for a token that expires
        // in 30 minutes and is rate-limited at the edge.
        int value = random.nextInt(1_000_000);
        return String.format("%06d", value);
    }

    private static final class PendingReset {
        final String userId;
        final long expiresAtEpochMillis;

        PendingReset(String userId, long expiresAtEpochMillis) {
            this.userId = userId;
            this.expiresAtEpochMillis = expiresAtEpochMillis;
        }
    }
}
