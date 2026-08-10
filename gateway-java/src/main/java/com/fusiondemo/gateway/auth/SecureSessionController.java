package com.fusiondemo.gateway.auth;

import javax.servlet.http.HttpServletRequest;

/**
 * Handles session bootstrap for clients using the hardened session
 * bridge. This replaces the legacy raw-serialization flow with a
 * class-allowlisted deserialization path while leaving the wire payload
 * format unchanged, so existing client integrations don't break.
 */
public class SecureSessionController {

    private final SecureSessionTokenHandler sessionTokenHandler = new SecureSessionTokenHandler();

    /**
     * Entry point invoked by the servlet dispatcher for POST /session/restore/v2.
     */
    public SessionContext restoreSession(HttpServletRequest request) {
        String encodedPayload = request.getParameter("sessionPayload");
        if (encodedPayload == null || encodedPayload.isEmpty()) {
            throw new IllegalArgumentException("Missing sessionPayload parameter");
        }
        return sessionTokenHandler.rehydrate(encodedPayload);
    }
}
