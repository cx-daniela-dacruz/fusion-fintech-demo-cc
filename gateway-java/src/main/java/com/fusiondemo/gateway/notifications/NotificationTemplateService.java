package com.fusiondemo.gateway.notifications;

import java.util.HashMap;
import java.util.Map;
import java.util.Optional;

/**
 * Holds the built-in notification templates and renders one for a
 * given recipient and action. Real delivery (email/SMS) happens in
 * the shared notifications platform; this service only prepares the
 * message content.
 */
public class NotificationTemplateService {

    private final Map<String, NotificationTemplate> templatesById = new HashMap<>();

    public NotificationTemplateService() {
        register(new NotificationTemplate(
                "password-reset-requested",
                "Password reset requested",
                "Hi {{recipientName}}, a password reset was just requested. {{actionSummary}}"));
        register(new NotificationTemplate(
                "api-key-issued",
                "New API key issued",
                "Hi {{recipientName}}, a new API key was issued for your account. {{actionSummary}}"));
        register(new NotificationTemplate(
                "role-changed",
                "Your account role changed",
                "Hi {{recipientName}}, your account role was updated. {{actionSummary}}"));
    }

    private void register(NotificationTemplate template) {
        templatesById.put(template.getTemplateId(), template);
    }

    public Optional<String> render(String templateId, String recipientName, String actionSummary) {
        NotificationTemplate template = templatesById.get(templateId);
        if (template == null) {
            return Optional.empty();
        }
        return Optional.of(template.render(recipientName, actionSummary));
    }
}
