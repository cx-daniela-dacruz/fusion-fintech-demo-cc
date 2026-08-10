package com.fusiondemo.gateway.admin;

import com.fusiondemo.gateway.users.UserProfile;
import com.fusiondemo.gateway.users.UserProfileRepository;

/**
 * Back-office operations for support staff managing trading desk
 * accounts: deactivating a user, or promoting them to a new role
 * during an internal transfer.
 */
public class AdminUserService {

    private final UserProfileRepository repository;

    public AdminUserService(UserProfileRepository repository) {
        this.repository = repository;
    }

    public UserProfile changeRole(String userId, String newRole) {
        UserProfile profile = repository.findById(userId)
                .orElseThrow(() -> new IllegalArgumentException("Unknown user: " + userId));
        profile.setRole(newRole);
        repository.save(profile);
        return profile;
    }

    public void deactivate(String userId) {
        UserProfile profile = repository.findById(userId)
                .orElseThrow(() -> new IllegalArgumentException("Unknown user: " + userId));
        profile.setRole("deactivated");
        repository.save(profile);
    }
}
