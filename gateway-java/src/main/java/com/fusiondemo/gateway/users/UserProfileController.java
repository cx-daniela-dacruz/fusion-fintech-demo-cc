package com.fusiondemo.gateway.users;

import javax.servlet.http.HttpServletRequest;
import java.util.Optional;

/**
 * HTTP entry points for reading and updating the caller's own profile.
 * Mounted under /profile in the gateway routing table.
 */
public class UserProfileController {

    private final UserProfileService profileService;

    public UserProfileController(UserProfileService profileService) {
        this.profileService = profileService;
    }

    /**
     * GET /profile/{userId}
     */
    public Optional<UserProfile> getProfile(HttpServletRequest request) {
        String userId = request.getParameter("userId");
        if (userId == null || userId.isEmpty()) {
            throw new IllegalArgumentException("userId is required");
        }
        return profileService.getProfile(userId);
    }

    /**
     * POST /profile/{userId}/display-name
     */
    public UserProfile updateDisplayName(HttpServletRequest request) {
        String userId = request.getParameter("userId");
        String displayName = request.getParameter("displayName");
        return profileService.updateDisplayName(userId, displayName);
    }

    /**
     * POST /profile/{userId}/email
     */
    public UserProfile updateEmail(HttpServletRequest request) {
        String userId = request.getParameter("userId");
        String email = request.getParameter("email");
        return profileService.updateEmail(userId, email);
    }
}
