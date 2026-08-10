package com.fusiondemo.gateway.users;

import java.time.Instant;

/**
 * Basic profile information for an authenticated trading desk user.
 * Distinct from SessionContext, which only carries what's needed to
 * validate an active session.
 */
public class UserProfile {

    private String userId;
    private String displayName;
    private String email;
    private String tenantId;
    private String role;
    private Instant createdAt;

    public UserProfile() {
    }

    public UserProfile(String userId, String displayName, String email, String tenantId, String role, Instant createdAt) {
        this.userId = userId;
        this.displayName = displayName;
        this.email = email;
        this.tenantId = tenantId;
        this.role = role;
        this.createdAt = createdAt;
    }

    public String getUserId() {
        return userId;
    }

    public String getDisplayName() {
        return displayName;
    }

    public String getEmail() {
        return email;
    }

    public String getTenantId() {
        return tenantId;
    }

    public String getRole() {
        return role;
    }

    public Instant getCreatedAt() {
        return createdAt;
    }

    public void setDisplayName(String displayName) {
        this.displayName = displayName;
    }

    public void setEmail(String email) {
        this.email = email;
    }

    public void setRole(String role) {
        this.role = role;
    }
}
