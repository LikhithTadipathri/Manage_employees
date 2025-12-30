package handlers

import (
	"net/http"
	"strconv"
	"time"

	"employee-service/http/middlewares"
	"employee-service/http/response"
	employeeService "employee-service/services/employee"
	userService "employee-service/services/user"
)

// DashboardHandler handles dashboard-related requests
type DashboardHandler struct {
	employeeService *employeeService.Service
	userService     *userService.UserService
}

// NewDashboardHandler creates a new dashboard handler
func NewDashboardHandler(empService *employeeService.Service, usrService *userService.UserService) *DashboardHandler {
	return &DashboardHandler{
		employeeService: empService,
		userService:     usrService,
	}
}

// GetUserRecords handles GET /user/records (returns employee records for authenticated users)
func (h *DashboardHandler) GetUserRecords(w http.ResponseWriter, r *http.Request) {
	// Verify user is authenticated
	_, err := middlewares.GetUserFromContext(r)
	if err != nil {
		response.Error(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	limit := 10
	offset := 0

	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil {
			limit = l
		}
	}

	if offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil {
			offset = o
		}
	}

	employees, total, err := h.employeeService.ListEmployees(limit, offset)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Failed to fetch employee records")
		return
	}

	data := map[string]interface{}{
		"records": employees,
		"total":   total,
	}

	response.Success(w, http.StatusOK, data, "Employee records retrieved successfully")
}


// GetUserOverview handles GET /user/overview (returns employee overview for authenticated users)
func (h *DashboardHandler) GetUserOverview(w http.ResponseWriter, r *http.Request) {
	// Verify user is authenticated
	_, err := middlewares.GetUserFromContext(r)
	if err != nil {
		response.Error(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	_, total, err := h.employeeService.ListEmployees(1, 0)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Failed to fetch overview")
		return
	}

	overview := map[string]interface{}{
		"total_employee_records": total,
		"timestamp":              time.Now().Format(time.RFC3339),
		"status":                 "active",
	}

	response.Success(w, http.StatusOK, overview, "Employee overview retrieved successfully")
}

// GetAdminRecords handles GET /admin/records (returns admin user records)
func (h *DashboardHandler) GetAdminRecords(w http.ResponseWriter, r *http.Request) {
	// Get user from JWT context
	claims, err := middlewares.GetUserFromContext(r)
	if err != nil || claims.Role != "admin" {
		response.Error(w, http.StatusForbidden, "Admin access required")
		return
	}

	users, err := h.userService.GetAllUsers()
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Failed to fetch admin records")
		return
	}

	data := map[string]interface{}{
		"records": users,
		"total":   len(users),
	}

	response.Success(w, http.StatusOK, data, "Admin records retrieved successfully")
}


// GetAdminOverview handles GET /admin/overview (returns admin user count)
func (h *DashboardHandler) GetAdminOverview(w http.ResponseWriter, r *http.Request) {
	// Get user from JWT context
	claims, err := middlewares.GetUserFromContext(r)
	if err != nil || claims.Role != "admin" {
		response.Error(w, http.StatusForbidden, "Admin access required")
		return
	}

	users, err := h.userService.GetAllUsers()
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Failed to fetch overview")
		return
	}

	overview := map[string]interface{}{
		"total_admin_count": len(users),
		"timestamp":         time.Now().Format(time.RFC3339),
		"status":            "active",
	}

	response.Success(w, http.StatusOK, overview, "Admin overview retrieved successfully")
}

// GetAdminLogs handles GET /admin/logs
func (h *DashboardHandler) GetAdminLogs(w http.ResponseWriter, r *http.Request) {
	// Get user from JWT context
	claims, err := middlewares.GetUserFromContext(r)
	if err != nil || claims.Role != "admin" {
		response.Error(w, http.StatusForbidden, "Admin access required")
		return
	}

	logs := []map[string]interface{}{
		{
			"action":       "create",
			"record_id":    1,
			"performed_by": "admin",
			"timestamp":    time.Now().Format(time.RFC3339),
		},
		{
			"action":       "update",
			"record_id":    1,
			"performed_by": "admin",
			"timestamp":    time.Now().Format(time.RFC3339),
		},
	}

	response.Success(w, http.StatusOK, logs, "Logs retrieved successfully")
}


func (h *DashboardHandler) GetOverview(w http.ResponseWriter, r *http.Request) {
	_, total, err := h.employeeService.ListEmployees(10, 0)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Failed to fetch overview")
		return
	}

	overview := map[string]interface{}{
		"total_records": total,
		"timestamp":     time.Now().Format(time.RFC3339),
		"status":        "active",
		"uptime":        "99.9%",
	}

	response.Success(w, http.StatusOK, overview, "Overview retrieved successfully")
}
