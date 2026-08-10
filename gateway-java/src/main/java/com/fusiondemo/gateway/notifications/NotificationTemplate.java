package com.fusiondemo.gateway.notifications;

/**
 * A named message template used for account notifications (password
 * reset, new API key issued, role changed).
 */
public class NotificationTemplate {

    private final String templateId;
    private final String subject;
    private final String bodyPattern;

    public NotificationTemplate(String templateId, String subject, String bodyPattern) {
        this.templateId = templateId;
        this.subject = subject;
        this.bodyPattern = bodyPattern;
    }

    public String getTemplateId() {
        return templateId;
    }

    public String getSubject() {
        return subject;
    }

    public String render(String recipientName, String actionSummary) {
        return bodyPattern
                .replace("{{recipientName}}", recipientName)
                .replace("{{actionSummary}}", actionSummary);
    }
}
