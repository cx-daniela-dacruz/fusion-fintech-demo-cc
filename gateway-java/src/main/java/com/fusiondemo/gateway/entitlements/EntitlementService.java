package com.fusiondemo.gateway.entitlements;

import com.fusiondemo.gateway.users.UserProfile;
import com.fusiondemo.gateway.users.UserProfileRepository;

import java.util.Set;

/**
 * Resolves what a specific user is entitled to do, combining their
 * role with the entitlement catalog. Consumed by controllers that
 * need to gate a specific action rather than a whole endpoint.
 */
public class EntitlementService {

    private final UserProfileRepository userProfileRepository;
    private final EntitlementCatalog entitlementCatalog;

    public EntitlementService(UserProfileRepository userProfileRepository, EntitlementCatalog entitlementCatalog) {
        this.userProfileRepository = userProfileRepository;
        this.entitlementCatalog = entitlementCatalog;
    }

    public Set<Entitlement> entitlementsForUser(String userId) {
        UserProfile profile = userProfileRepository.findById(userId)
                .orElseThrow(() -> new IllegalArgumentException("Unknown user: " + userId));
        return entitlementCatalog.entitlementsFor(profile.getRole());
    }

    public boolean userHasEntitlement(String userId, Entitlement entitlement) {
        return entitlementsForUser(userId).contains(entitlement);
    }

    public void requireEntitlement(String userId, Entitlement entitlement) {
        if (!userHasEntitlement(userId, entitlement)) {
            throw new SecurityException("User " + userId + " lacks entitlement " + entitlement);
        }
    }
}
