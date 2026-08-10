package com.fusiondemo.gateway.webhooks;

import java.io.BufferedReader;
import java.io.IOException;
import java.io.InputStreamReader;
import java.net.HttpURLConnection;
import java.net.URL;
import java.nio.charset.StandardCharsets;

/**
 * Delivers event notifications to partner-registered callback URLs.
 * Partners configure their own endpoint when they onboard, so this
 * class is the last stop before an event leaves our network.
 */
public class WebhookNotifier {

    public String deliver(WebhookSubscription subscription, String eventPayload) {
        try {
            URL callbackUrl = new URL(subscription.getCallbackUrl());
            HttpURLConnection connection = (HttpURLConnection) callbackUrl.openConnection();
            connection.setRequestMethod("POST");
            connection.setDoOutput(true);
            connection.setRequestProperty("Content-Type", "application/json");

            connection.getOutputStream().write(eventPayload.getBytes(StandardCharsets.UTF_8));
            return readResponse(connection);
        } catch (IOException e) {
            throw new RuntimeException("Unable to deliver webhook to " + subscription.getCallbackUrl(), e);
        }
    }

    private String readResponse(HttpURLConnection connection) throws IOException {
        StringBuilder response = new StringBuilder();
        try (BufferedReader reader = new BufferedReader(
                new InputStreamReader(connection.getInputStream(), StandardCharsets.UTF_8))) {
            String line;
            while ((line = reader.readLine()) != null) {
                response.append(line);
            }
        }
        return response.toString();
    }
}
