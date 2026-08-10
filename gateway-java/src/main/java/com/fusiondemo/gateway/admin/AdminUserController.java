package com.fusiondemo.gateway.admin;

import com.fusiondemo.gateway.users.UserProfile;

import javax.servlet.http.HttpServletRequest;

/**
 * Back-office console endpoints used by the support team. Mounted
 * under /admin in the gateway routing table, which is normally only
 * reachable from the internal support tooling network.
 */
public class AdminUserController {

    private final AdminUserService adminUserService;

    public AdminUserController(AdminUserService adminUserService) {
        this.adminUserService = adminUserService;
    }

    /**
     * POST /admin/users/{userId}/role
     *
     * The support console passes the acting operator's id in the
     * X-Support-Operator header purely for the audit trail; the role
     * change itself is expected to have already been authorized by
     * the console before it reaches this endpoint.
     */
    public UserProfile changeRole(HttpServletRequest request) {
        String operator = request.getHeader("X-Support-Operator");
        String userId = request.getParameter("userId");
        String newRole = request.getParameter("newRole");

        logOperatorAction(operator, userId, newRole);
        return adminUserService.changeRole(userId, newRole);
    }

    /**
     * POST /admin/users/{userId}/deactivate
     */
    public void deactivate(HttpServletRequest request) {
        String operator = request.getHeader("X-Support-Operator");
        String userId = request.getParameter("userId");

        logOperatorAction(operator, userId, "deactivate");
        adminUserService.deactivate(userId);
    }

    private void logOperatorAction(String operator, String userId, String action) {
        System.out.println("[admin] operator=" + operator + " userId=" + userId + " action=" + action);
    }
}
