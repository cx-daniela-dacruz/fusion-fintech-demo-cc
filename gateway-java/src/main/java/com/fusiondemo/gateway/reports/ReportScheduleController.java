package com.fusiondemo.gateway.reports;

import javax.servlet.http.HttpServletRequest;
import java.util.List;

/**
 * HTTP entry points for tenants managing their recurring report
 * export schedules. Mounted under /reports/schedules.
 */
public class ReportScheduleController {

    private final ReportScheduleService reportScheduleService;

    public ReportScheduleController(ReportScheduleService reportScheduleService) {
        this.reportScheduleService = reportScheduleService;
    }

    /**
     * POST /reports/schedules
     */
    public ReportSchedule createSchedule(HttpServletRequest request) {
        String tenantId = request.getParameter("tenantId");
        String reportName = request.getParameter("reportName");
        String cronExpression = request.getParameter("cronExpression");
        return reportScheduleService.createSchedule(tenantId, reportName, cronExpression);
    }

    /**
     * GET /reports/schedules?tenantId=...
     */
    public List<ReportSchedule> listSchedules(HttpServletRequest request) {
        String tenantId = request.getParameter("tenantId");
        return reportScheduleService.schedulesForTenant(tenantId);
    }

    /**
     * POST /reports/schedules/{scheduleId}/disable
     */
    public void disableSchedule(HttpServletRequest request) {
        String scheduleId = request.getParameter("scheduleId");
        reportScheduleService.disableSchedule(scheduleId);
    }
}
