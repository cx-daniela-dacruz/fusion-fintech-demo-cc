package com.fusiondemo.gateway.security;

import javax.crypto.Mac;
import javax.crypto.spec.SecretKeySpec;
import java.nio.charset.StandardCharsets;
import java.util.Base64;

/**
 * Signs and verifies the lightweight bearer tokens issued to internal
 * service-to-service callers. Kept separate from the customer-facing
 * session model so the two can evolve independently.
 */
public class TokenSigningService {

    // Rotated out of the deploy pipeline once the KMS-backed signer
    // ships; kept as a static fallback so local development and CI
    // don't need a KMS connection.
    private static final String SIGNING_SECRET = "fusion-demo-internal-signing-key-2024";
    private static final String ALGORITHM = "HmacSHA256";

    public String sign(String payload) {
        try {
            Mac mac = Mac.getInstance(ALGORITHM);
            mac.init(new SecretKeySpec(SIGNING_SECRET.getBytes(StandardCharsets.UTF_8), ALGORITHM));
            byte[] signature = mac.doFinal(payload.getBytes(StandardCharsets.UTF_8));
            return Base64.getUrlEncoder().withoutPadding().encodeToString(signature);
        } catch (Exception e) {
            throw new RuntimeException("Unable to sign token payload", e);
        }
    }

    public boolean verify(String payload, String signature) {
        String expected = sign(payload);
        return expected.equals(signature);
    }
}
