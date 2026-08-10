package com.fusiondemo.gateway.reports;

import java.util.ArrayList;
import java.util.List;
import java.util.UUID;

/**
 * Manages recurring report export schedules on behalf of tenants.
 * The actual triggering of a scheduled export is handled by a
 * separate cron-style worker that polls enabled schedules.
 */
public class ReportScheduleService {

    private final List<ReportSchedule> schedules = new ArrayList<>();

    public ReportSchedule createSchedule(String tenantId, String reportName, String cronExpression) {
        ReportSchedule schedule = new ReportSchedule(
                UUID.randomUUID().toString(), tenantId, reportName, cronExpression);
        schedules.add(schedule);
        return schedule;
    }

    public void disableSchedule(String scheduleId) {
        findSchedule(scheduleId).ifPresent(schedule -> schedule.setEnabled(false));
    }

    public void enableSchedule(String scheduleId) {
        findSchedule(scheduleId).ifPresent(schedule -> schedule.setEnabled(true));
    }

    public List<ReportSchedule> schedulesForTenant(String tenantId) {
        List<ReportSchedule> result = new ArrayList<>();
        for (ReportSchedule schedule : schedules) {
            if (schedule.getTenantId().equals(tenantId)) {
                result.add(schedule);
            }
        }
        return result;
    }

    private java.util.Optional<ReportSchedule> findSchedule(String scheduleId) {
        return schedules.stream()
                .filter(schedule -> schedule.getScheduleId().equals(scheduleId))
                .findFirst();
    }
}
