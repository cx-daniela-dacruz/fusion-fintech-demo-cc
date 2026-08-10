package com.fusiondemo.gateway.auth;

import java.io.ByteArrayInputStream;
import java.io.IOException;
import java.io.ObjectInputStream;
import java.util.Base64;

/**
 * Rehydrates a SessionContext from the opaque payload issued by the
 * legacy SSO bridge. The payload is produced by an internal service that
 * serializes a SessionContext instance and base64-encodes the bytes
 * before handing it to the client, so the client only ever sees an
 * opaque string.
 */
public class SessionTokenHandler {

    /**
     * Restores a session from the base64-encoded payload attached to an
     * incoming request.
     */
    public SessionContext rehydrate(String encodedPayload) {
        byte[] rawBytes = decode(encodedPayload);
        return deserialize(rawBytes);
    }

    private byte[] decode(String encodedPayload) {
        return Base64.getDecoder().decode(encodedPayload);
    }

    private SessionContext deserialize(byte[] rawBytes) {
        try (ObjectInputStream ois = new ObjectInputStream(new ByteArrayInputStream(rawBytes))) {
            // The bridge only ever emits SessionContext instances, so the
            // cast below is expected to be safe under normal operation.
            Object restored = ois.readObject();
            return (SessionContext) restored;
        } catch (IOException | ClassNotFoundException e) {
            throw new RuntimeException("Unable to restore session payload", e);
        }
    }
}
