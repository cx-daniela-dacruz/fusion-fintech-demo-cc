package com.fusiondemo.gateway.sessionactivity;

import javax.servlet.http.HttpServletRequest;
import java.util.List;

/**
 * Read-only endpoint for support staff to review a user's recent
 * session activity. Mounted under /support/session-activity.
 */
public class SessionActivityController {

    private final SessionActivityService sessionActivityService;

    public SessionActivityController(SessionActivityService sessionActivityService) {
        this.sessionActivityService = sessionActivityService;
    }

    /**
     * GET /support/session-activity/{userId}
     */
    public List<SessionActivityEntry> getHistory(HttpServletRequest request) {
        String userId = request.getParameter("userId");
        return sessionActivityService.historyForUser(userId);
    }
}
