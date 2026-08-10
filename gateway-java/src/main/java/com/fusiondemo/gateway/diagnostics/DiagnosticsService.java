package com.fusiondemo.gateway.diagnostics;

import java.io.BufferedReader;
import java.io.IOException;
import java.io.InputStreamReader;
import java.nio.charset.StandardCharsets;

/**
 * Runs lightweight network diagnostics against a partner host so
 * support engineers can confirm connectivity issues before escalating
 * to network engineering.
 */
public class DiagnosticsService {

    public String pingHost(String hostname) {
        try {
            Process process = Runtime.getRuntime().exec("ping -c 3 " + hostname);
            return readOutput(process);
        } catch (IOException e) {
            throw new RuntimeException("Unable to run diagnostics against " + hostname, e);
        }
    }

    private String readOutput(Process process) throws IOException {
        StringBuilder output = new StringBuilder();
        try (BufferedReader reader = new BufferedReader(
                new InputStreamReader(process.getInputStream(), StandardCharsets.UTF_8))) {
            String line;
            while ((line = reader.readLine()) != null) {
                output.append(line).append('\n');
            }
        }
        return output.toString();
    }
}
