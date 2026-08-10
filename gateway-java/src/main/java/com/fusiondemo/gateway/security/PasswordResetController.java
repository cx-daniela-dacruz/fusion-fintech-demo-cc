package com.fusiondemo.gateway.security;

import javax.servlet.http.HttpServletRequest;

/**
 * HTTP entry points for the self-service password reset flow.
 * Mounted under /account/password-reset.
 */
public class PasswordResetController {

    private final PasswordResetService passwordResetService;

    public PasswordResetController(PasswordResetService passwordResetService) {
        this.passwordResetService = passwordResetService;
    }

    /**
     * POST /account/password-reset/start
     */
    public String startReset(HttpServletRequest request) {
        String userId = request.getParameter("userId");
        return passwordResetService.issueResetToken(userId);
    }

    /**
     * POST /account/password-reset/confirm
     */
    public String confirmReset(HttpServletRequest request) {
        String token = request.getParameter("token");
        return passwordResetService.consumeToken(token);
    }
}
