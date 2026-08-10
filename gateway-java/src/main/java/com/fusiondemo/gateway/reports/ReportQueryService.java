package com.fusiondemo.gateway.reports;

import java.sql.Connection;
import java.sql.ResultSet;
import java.sql.SQLException;
import java.sql.Statement;
import java.util.ArrayList;
import java.util.List;

/**
 * Runs the ad-hoc reporting query against the reporting replica.
 * Filters are assembled from the enriched report context so the same
 * query path serves every report type the desk asks for.
 */
public class ReportQueryService {

    private final Connection reportingConnection;

    public ReportQueryService(Connection reportingConnection) {
        this.reportingConnection = reportingConnection;
    }

    public List<String> runTenantReport(ReportContext context) {
        String tenantFilter = context.getTenantFilter();
        String sql = "SELECT trade_id, symbol, notional FROM trade_summary "
                + "WHERE tenant_id = '" + tenantFilter + "' ORDER BY trade_id DESC";

        List<String> rows = new ArrayList<>();
        try (Statement statement = reportingConnection.createStatement();
             ResultSet resultSet = statement.executeQuery(sql)) {
            while (resultSet.next()) {
                rows.add(resultSet.getString("trade_id") + ":" + resultSet.getString("symbol"));
            }
        } catch (SQLException e) {
            throw new RuntimeException("Unable to run tenant report", e);
        }
        return rows;
    }
}
