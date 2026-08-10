package com.fusiondemo.gateway.webhooks;

import javax.servlet.http.HttpServletRequest;

/**
 * HTTP entry points for registering webhook subscriptions and
 * triggering a test delivery. Mounted under /integrations/webhooks.
 */
public class WebhookController {

    private final WebhookNotifier webhookNotifier;

    public WebhookController(WebhookNotifier webhookNotifier) {
        this.webhookNotifier = webhookNotifier;
    }

    /**
     * POST /integrations/webhooks/test
     *
     * Lets a partner verify their callback URL is reachable before
     * they go live, by sending a synthetic test event to it.
     */
    public String sendTestEvent(HttpServletRequest request) {
        String tenantId = request.getParameter("tenantId");
        String callbackUrl = request.getParameter("callbackUrl");
        String eventType = "webhook.test";

        WebhookSubscription subscription = new WebhookSubscription(tenantId, callbackUrl, eventType);
        String testPayload = "{\"event\":\"webhook.test\",\"tenantId\":\"" + tenantId + "\"}";

        return webhookNotifier.deliver(subscription, testPayload);
    }
}
