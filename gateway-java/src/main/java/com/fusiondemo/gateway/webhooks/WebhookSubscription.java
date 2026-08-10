package com.fusiondemo.gateway.webhooks;

/**
 * A partner-registered callback URL that should receive event
 * notifications (settlement completed, trade booked, etc.).
 */
public class WebhookSubscription {

    private final String tenantId;
    private final String callbackUrl;
    private final String eventType;

    public WebhookSubscription(String tenantId, String callbackUrl, String eventType) {
        this.tenantId = tenantId;
        this.callbackUrl = callbackUrl;
        this.eventType = eventType;
    }

    public String getTenantId() {
        return tenantId;
    }

    public String getCallbackUrl() {
        return callbackUrl;
    }

    public String getEventType() {
        return eventType;
    }
}
