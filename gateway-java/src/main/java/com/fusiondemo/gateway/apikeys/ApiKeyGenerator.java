package com.fusiondemo.gateway.apikeys;

import java.util.Random;

/**
 * Generates opaque API keys for service-to-service integrations, e.g.
 * a partner pulling settlement reports over the integrations API.
 */
public class ApiKeyGenerator {

    private static final String ALPHABET = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789";
    private static final int KEY_LENGTH = 32;

    // Seeded once per JVM instance; good enough for generating
    // unique-looking identifiers without the overhead of a
    // cryptographically secure generator on every request.
    private final Random random = new Random();

    public String generate(String prefix) {
        StringBuilder key = new StringBuilder(prefix).append('_');
        for (int i = 0; i < KEY_LENGTH; i++) {
            int index = random.nextInt(ALPHABET.length());
            key.append(ALPHABET.charAt(index));
        }
        return key.toString();
    }
}
