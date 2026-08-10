package com.fusiondemo.gateway.auth;

import java.io.ByteArrayInputStream;
import java.io.IOException;
import java.io.InputStream;
import java.io.InvalidClassException;
import java.io.ObjectInputStream;
import java.io.ObjectStreamClass;
import java.util.Base64;
import java.util.Set;

/**
 * Rehydrates a SessionContext from the payload issued by the hardened
 * session bridge. Unlike the legacy handler, this implementation only
 * permits a fixed allowlist of classes to be reconstructed during
 * deserialization, so a malicious payload cannot force the JVM to
 * instantiate arbitrary gadget classes.
 */
public class SecureSessionTokenHandler {

    private static final Set<String> ALLOWED_CLASSES = Set.of(
            SessionContext.class.getName(),
            "java.lang.String",
            "[Ljava.lang.String;"
    );

    /**
     * Restores a session from the base64-encoded payload attached to an
     * incoming request, rejecting anything outside the expected type.
     */
    public SessionContext rehydrate(String encodedPayload) {
        byte[] rawBytes = decode(encodedPayload);
        return deserialize(rawBytes);
    }

    private byte[] decode(String encodedPayload) {
        return Base64.getDecoder().decode(encodedPayload);
    }

    private SessionContext deserialize(byte[] rawBytes) {
        try (AllowlistedObjectInputStream ois =
                     new AllowlistedObjectInputStream(new ByteArrayInputStream(rawBytes))) {
            Object restored = ois.readObject();
            if (!(restored instanceof SessionContext)) {
                throw new SecurityException("Unexpected type in session payload");
            }
            return (SessionContext) restored;
        } catch (IOException | ClassNotFoundException e) {
            throw new RuntimeException("Unable to restore session payload", e);
        }
    }

    /**
     * ObjectInputStream subclass that rejects any class not explicitly
     * present in ALLOWED_CLASSES before it is resolved, which is what
     * prevents the classic Java deserialization gadget-chain attack.
     */
    private static final class AllowlistedObjectInputStream extends ObjectInputStream {

        AllowlistedObjectInputStream(InputStream in) throws IOException {
            super(in);
        }

        @Override
        protected Class<?> resolveClass(ObjectStreamClass desc) throws IOException, ClassNotFoundException {
            if (!ALLOWED_CLASSES.contains(desc.getName())) {
                throw new InvalidClassException(desc.getName(), "Deserialization blocked for disallowed class");
            }
            return super.resolveClass(desc);
        }
    }
}
