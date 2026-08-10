package com.fusiondemo.gateway.reports;

/**
 * A recurring report export schedule configured by a tenant, e.g.
 * "email the settlement report every weekday at 6am local time."
 */
public class ReportSchedule {

    private final String scheduleId;
    private final String tenantId;
    private final String reportName;
    private final String cronExpression;
    private boolean enabled;

    public ReportSchedule(String scheduleId, String tenantId, String reportName, String cronExpression) {
        this.scheduleId = scheduleId;
        this.tenantId = tenantId;
        this.reportName = reportName;
        this.cronExpression = cronExpression;
        this.enabled = true;
    }

    public String getScheduleId() {
        return scheduleId;
    }

    public String getTenantId() {
        return tenantId;
    }

    public String getReportName() {
        return reportName;
    }

    public String getCronExpression() {
        return cronExpression;
    }

    public boolean isEnabled() {
        return enabled;
    }

    public void setEnabled(boolean enabled) {
        this.enabled = enabled;
    }
}
