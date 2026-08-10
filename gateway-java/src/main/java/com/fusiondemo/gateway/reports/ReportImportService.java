package com.fusiondemo.gateway.reports;

import org.w3c.dom.Document;

import javax.xml.parsers.DocumentBuilder;
import javax.xml.parsers.DocumentBuilderFactory;
import java.io.ByteArrayInputStream;
import java.nio.charset.StandardCharsets;

/**
 * Imports back-office report definitions submitted as XML by partner
 * integrations. Report definitions describe which fields to include
 * and how to format them; the actual trade data is fetched separately.
 */
public class ReportImportService {

    public Document parseReportDefinition(String xmlPayload) {
        try {
            DocumentBuilderFactory factory = DocumentBuilderFactory.newInstance();
            DocumentBuilder builder = factory.newDocumentBuilder();
            return builder.parse(new ByteArrayInputStream(xmlPayload.getBytes(StandardCharsets.UTF_8)));
        } catch (Exception e) {
            throw new RuntimeException("Unable to parse report definition", e);
        }
    }

    public String extractReportName(Document document) {
        return document.getDocumentElement().getAttribute("name");
    }
}
