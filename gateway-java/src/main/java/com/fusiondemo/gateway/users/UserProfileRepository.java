package com.fusiondemo.gateway.users;

import java.time.Instant;
import java.util.HashMap;
import java.util.Map;
import java.util.Optional;

/**
 * In-memory stand-in for the user profile store. The real deployment
 * backs this with the shared identity database; this demo keeps
 * everything local so the service can run without external
 * dependencies.
 */
public class UserProfileRepository {

    private final Map<String, UserProfile> profilesByUserId = new HashMap<>();

    public UserProfileRepository() {
        seedDemoUsers();
    }

    private void seedDemoUsers() {
        profilesByUserId.put("u-1001", new UserProfile(
                "u-1001", "Dana Whitfield", "dana.whitfield@example.com", "tenant-acme", "trader", Instant.now()));
        profilesByUserId.put("u-1002", new UserProfile(
                "u-1002", "Priya Nandan", "priya.nandan@example.com", "tenant-acme", "risk_analyst", Instant.now()));
        profilesByUserId.put("u-1003", new UserProfile(
                "u-1003", "Omar Elsayed", "omar.elsayed@example.com", "tenant-northwind", "admin", Instant.now()));
    }

    public Optional<UserProfile> findById(String userId) {
        return Optional.ofNullable(profilesByUserId.get(userId));
    }

    public void save(UserProfile profile) {
        profilesByUserId.put(profile.getUserId(), profile);
    }

    public Map<String, UserProfile> all() {
        return profilesByUserId;
    }
}
