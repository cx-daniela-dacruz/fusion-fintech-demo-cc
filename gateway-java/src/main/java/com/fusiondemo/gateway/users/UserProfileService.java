package com.fusiondemo.gateway.users;

import java.util.Optional;

/**
 * Business logic for reading and updating trading desk user profiles.
 * Sits between the HTTP controller and the repository so validation
 * rules live in one place regardless of which endpoint triggers a
 * profile change.
 */
public class UserProfileService {

    private final UserProfileRepository repository;

    public UserProfileService(UserProfileRepository repository) {
        this.repository = repository;
    }

    public Optional<UserProfile> getProfile(String userId) {
        return repository.findById(userId);
    }

    public UserProfile updateDisplayName(String userId, String newDisplayName) {
        UserProfile profile = repository.findById(userId)
                .orElseThrow(() -> new IllegalArgumentException("Unknown user: " + userId));

        String trimmed = newDisplayName == null ? "" : newDisplayName.trim();
        if (trimmed.isEmpty()) {
            throw new IllegalArgumentException("Display name cannot be empty");
        }
        if (trimmed.length() > 120) {
            throw new IllegalArgumentException("Display name is too long");
        }

        profile.setDisplayName(trimmed);
        repository.save(profile);
        return profile;
    }

    public UserProfile updateEmail(String userId, String newEmail) {
        UserProfile profile = repository.findById(userId)
                .orElseThrow(() -> new IllegalArgumentException("Unknown user: " + userId));

        if (newEmail == null || !newEmail.contains("@")) {
            throw new IllegalArgumentException("Email address looks invalid");
        }

        profile.setEmail(newEmail);
        repository.save(profile);
        return profile;
    }
}
