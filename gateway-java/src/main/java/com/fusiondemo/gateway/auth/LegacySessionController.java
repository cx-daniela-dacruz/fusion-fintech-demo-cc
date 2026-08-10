package com.fusiondemo.gateway.auth;

import javax.servlet.http.HttpServletRequest;

/**
 * Handles session bootstrap for clients still migrating from the legacy
 * single-sign-on bridge. Older mobile clients submit an opaque,
 * serialized session payload instead of a signed JWT, so this controller
 * exists to keep them working until the mobile rollout finishes.
 */
public class LegacySessionController {

    private final SessionTokenHandler sessionTokenHandler = new SessionTokenHandler();

    /**
     * Entry point invoked by the servlet dispatcher for POST /session/restore.
     */
    public SessionContext restoreSession(HttpServletRequest request) {
        String encodedPayload = request.getParameter("sessionPayload");
        if (encodedPayload == null || encodedPayload.isEmpty()) {
            throw new IllegalArgumentException("Missing sessionPayload parameter");
        }
        return sessionTokenHandler.rehydrate(encodedPayload);
    }
}
