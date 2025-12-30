package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"employee-service/errors"
	"employee-service/http/middlewares"
	"employee-service/http/response"
	"employee-service/models/leave"
	"employee-service/models/user"
	leaveService "employee-service/services/leave"

	"github.com/go-chi/chi/v5"
)

// LeaveHandler handles HTTP requests for leave management
type LeaveHandler struct {
	service *leaveService.Service
}

// NewLeaveHandler creates a new leave handler
func NewLeaveHandler(service *leaveService.Service) *LeaveHandler {
	return &LeaveHandler{service: service}
}

// ApplyLeave handles POST /leave/apply
func (h *LeaveHandler) ApplyLeave(w http.ResponseWriter, r *http.Request) {
	// Get user from context
	userCtx, err := middlewares.GetUserFromContext(r)
	if err != nil {
		response.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	// ✔ Only employees can apply for leave
	if userCtx.Role != user.RoleEmployee {
		response.Error(w, http.StatusForbidden, "only employees can apply for leave")
		return
	}

	var req leave.ApplyLeaveRequest

	// Decode request body
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Apply leave - Use UserID which represents the employee for this system
	leaveRequest, err := h.service.ApplyLeave(userCtx.UserID, &req)
	if err != nil {
		// Log the actual error for debugging
		errors.LogError("ApplyLeave failed", err)
		
		// Check if it's a validation error
		if validationErr, ok := err.(*errors.ValidationError); ok {
			response.ErrorWithFields(w, http.StatusBadRequest, "Validation failed", validationErr.Fields)
			return
		}

		// Check if employee record doesn't exist
		if appErr, ok := err.(*errors.AppError); ok {
			if appErr.Code == 404 {
				response.Error(w, http.StatusNotFound, "employee record not found. Please contact HR to create your employee profile.")
				return
			}
		}

		response.Error(w, http.StatusInternalServerError, "failed to apply leave: " + err.Error())
		return
	}

	response.Success(w, http.StatusCreated, leaveRequest, "Leave request submitted successfully")
}

// GetMyLeaveRequests handles GET /leave/my-requests
func (h *LeaveHandler) GetMyLeaveRequests(w http.ResponseWriter, r *http.Request) {
	// Get user from context
	userCtx, err := middlewares.GetUserFromContext(r)
	if err != nil {
		response.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	// Only employees can view their own leave requests
	if userCtx.Role != user.RoleEmployee {
		response.Error(w, http.StatusForbidden, "only employees can view leave requests")
		return
	}

	// Get query parameters for filtering
	statusFilter := r.URL.Query().Get("status")

	// Get leave requests - Use UserID which represents the employee for this system
	leaveRequests, err := h.service.GetEmployeeLeaveRequests(userCtx.UserID)
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			if appErr.Code == 404 {
				response.Error(w, http.StatusNotFound, "employee record not found. Please contact HR to create your employee profile.")
				return
			}
		}
		response.Error(w, http.StatusInternalServerError, "failed to retrieve leave requests")
		return
	}

	// Filter by status if provided
	if statusFilter != "" {
		filteredRequests := []leave.LeaveRequest{}
		for _, lr := range leaveRequests {
			if lr.Status == leave.LeaveStatus(statusFilter) {
				filteredRequests = append(filteredRequests, lr)
			}
		}
		leaveRequests = filteredRequests
	}

	response.Success(w, http.StatusOK, map[string]interface{}{
		"count":          len(leaveRequests),
		"leave_requests": leaveRequests,
	}, "Leave requests retrieved successfully")
}

// CancelLeave handles DELETE /leave/cancel/:id
func (h *LeaveHandler) CancelLeave(w http.ResponseWriter, r *http.Request) {
	// Get user from context
	userCtx, err := middlewares.GetUserFromContext(r)
	if err != nil {
		response.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	// Only employees can cancel leave requests
	if userCtx.Role != user.RoleEmployee {
		response.Error(w, http.StatusForbidden, "only employees can cancel leave requests")
		return
	}

	// Get leave request ID from URL
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid leave request ID")
		return
	}

	// Cancel leave - Use UserID which represents the employee for this system
	err = h.service.CancelLeave(id, userCtx.UserID)
	if err != nil {
		if _, ok := err.(*errors.ForbiddenError); ok {
			response.Error(w, http.StatusForbidden, "you don't have permission to cancel this leave request")
			return
		}

		if validationErr, ok := err.(*errors.ValidationError); ok {
			response.ErrorWithFields(w, http.StatusBadRequest, "Validation failed", validationErr.Fields)
			return
		}

		if appErr, ok := err.(*errors.AppError); ok {
			if appErr.Code == 404 {
				response.Error(w, http.StatusNotFound, appErr.Message)
				return
			}
		}

		response.Error(w, http.StatusInternalServerError, "failed to cancel leave request")
		return
	}

	response.SuccessNoData(w, http.StatusOK, "Leave request cancelled successfully")
}

// GetAllLeaveRequests handles GET /leave/all (admin only)
func (h *LeaveHandler) GetAllLeaveRequests(w http.ResponseWriter, r *http.Request) {
	// Get user from context and verify admin role
	userCtx, err := middlewares.GetUserFromContext(r)
	if err != nil {
		response.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	if userCtx.Role != user.RoleAdmin {
		response.Error(w, http.StatusForbidden, "admin access required")
		return
	}

	// Get query parameters for filtering
	statusFilter := r.URL.Query().Get("status")

	// Get all leave requests
	leaveRequests, err := h.service.GetAllLeaveRequests(statusFilter)
	if err != nil {
		// Log the actual error for debugging
		errors.LogError("GetAllLeaveRequests failed", err)
		
		if validationErr, ok := err.(*errors.ValidationError); ok {
			response.ErrorWithFields(w, http.StatusBadRequest, "Validation failed", validationErr.Fields)
			return
		}

		if appErr, ok := err.(*errors.AppError); ok {
			if appErr.Code == 404 {
				response.Error(w, http.StatusNotFound, appErr.Message)
				return
			}
		}

		response.Error(w, http.StatusInternalServerError, "failed to retrieve leave requests: " + err.Error())
		return
	}

	response.Success(w, http.StatusOK, map[string]interface{}{
		"count":          len(leaveRequests),
		"leave_requests": leaveRequests,
	}, "Leave requests retrieved successfully")
}

// ApproveLeave handles POST /leave/approve/:id (admin only)
func (h *LeaveHandler) ApproveLeave(w http.ResponseWriter, r *http.Request) {
	// Get user from context and verify admin role
	userCtx, err := middlewares.GetUserFromContext(r)
	if err != nil {
		response.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	if userCtx.Role != user.RoleAdmin {
		response.Error(w, http.StatusForbidden, "admin access required")
		return
	}

	// Get leave request ID from URL
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid leave request ID")
		return
	}

	// Decode request body for optional notes
	var req leave.ApproveLeaveRequest
	var notes string
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		// If body is empty or invalid, continue without notes
		notes = ""
	} else {
		notes = req.Notes
	}

	// Approve leave - Use UserID which represents the admin user
	err = h.service.ApproveLeave(id, userCtx.UserID, notes)
	if err != nil {
		// Log the actual error for debugging
		errors.LogError("ApproveLeave failed", err)
		
		if validationErr, ok := err.(*errors.ValidationError); ok {
			response.ErrorWithFields(w, http.StatusBadRequest, "Validation failed", validationErr.Fields)
			return
		}

		if appErr, ok := err.(*errors.AppError); ok {
			if appErr.Code == 404 {
				response.Error(w, http.StatusNotFound, appErr.Message)
				return
			}
		}

		if _, ok := err.(*errors.NotFoundErrorType); ok {
			response.Error(w, http.StatusNotFound, err.Error())
			return
		}

		response.Error(w, http.StatusInternalServerError, "failed to approve leave request: " + err.Error())
		return
	}

	response.SuccessNoData(w, http.StatusOK, "Leave request approved successfully")
}

// RejectLeave handles POST /leave/reject/:id (admin only)
func (h *LeaveHandler) RejectLeave(w http.ResponseWriter, r *http.Request) {
	// Get user from context and verify admin role
	userCtx, err := middlewares.GetUserFromContext(r)
	if err != nil {
		response.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	if userCtx.Role != user.RoleAdmin {
		response.Error(w, http.StatusForbidden, "admin access required")
		return
	}

	// Get leave request ID from URL
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid leave request ID")
		return
	}

	// Decode request body for optional reason
	var req leave.RejectLeaveRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		// If body is empty or invalid, continue without reason
		req.Reason = ""
	}

	// Reject leave - Use UserID which represents the admin user
	err = h.service.RejectLeave(id, userCtx.UserID, req.Reason)
	if err != nil {
		// Log the actual error for debugging
		errors.LogError("RejectLeave failed", err)
		
		if validationErr, ok := err.(*errors.ValidationError); ok {
			response.ErrorWithFields(w, http.StatusBadRequest, "Validation failed", validationErr.Fields)
			return
		}

		if appErr, ok := err.(*errors.AppError); ok {
			if appErr.Code == 404 {
				response.Error(w, http.StatusNotFound, appErr.Message)
				return
			}
		}

		if _, ok := err.(*errors.NotFoundErrorType); ok {
			response.Error(w, http.StatusNotFound, err.Error())
			return
		}

		response.Error(w, http.StatusInternalServerError, "failed to reject leave request: " + err.Error())
		return
	}

	response.SuccessNoData(w, http.StatusOK, "Leave request rejected successfully")
}

// ... rest of the code remains the same ...
func (h *LeaveHandler) GetMyLeaveBalance(w http.ResponseWriter, r *http.Request) {
	// Get user from context
	userCtx, err := middlewares.GetUserFromContext(r)
	if err != nil {
		response.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	// Only employees can view their own leave balance
	if userCtx.Role != user.RoleEmployee {
		response.Error(w, http.StatusForbidden, "only employees can view leave balance")
		return
	}

	// Get leave balances
	balances, err := h.service.GetEmployeeLeaveBalances(userCtx.UserID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to retrieve leave balance")
		return
	}

	response.Success(w, http.StatusOK, map[string]interface{}{
		"count":           len(balances),
		"leave_balances": balances,
	}, "Leave balance retrieved successfully")
}

// GetMyLeaveBalanceByType handles GET /leave/balance/:type
func (h *LeaveHandler) GetMyLeaveBalanceByType(w http.ResponseWriter, r *http.Request) {
	// Get user from context
	userCtx, err := middlewares.GetUserFromContext(r)
	if err != nil {
		response.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	// Only employees can view their own leave balance
	if userCtx.Role != user.RoleEmployee {
		response.Error(w, http.StatusForbidden, "only employees can view leave balance")
		return
	}

	// Get leave type from URL
	leaveTypeStr := chi.URLParam(r, "type")
	leaveType := leave.LeaveType(leaveTypeStr)

	// Get leave balance
	balance, err := h.service.GetEmployeeLeaveBalance(userCtx.UserID, leaveType)
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			if appErr.Code == 404 {
				response.Error(w, http.StatusNotFound, appErr.Message)
				return
			}
		}

		response.Error(w, http.StatusInternalServerError, "failed to retrieve leave balance")
		return
	}

	response.Success(w, http.StatusOK, balance, "Leave balance retrieved successfully")
}

// ReviewLeaveRequests handles GET /leave/review (admin only)
// Returns pending leave requests for review
func (h *LeaveHandler) ReviewLeaveRequests(w http.ResponseWriter, r *http.Request) {
	// Get user from context and verify admin role
	userCtx, err := middlewares.GetUserFromContext(r)
	if err != nil {
		response.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	if userCtx.Role != user.RoleAdmin {
		response.Error(w, http.StatusForbidden, "admin access required")
		return
	}

	// Get pending leave requests only (for review)
	leaveRequests, err := h.service.GetAllLeaveRequests(string(leave.StatusPending))
	if err != nil {
		// Log the actual error for debugging
		errors.LogError("ReviewLeaveRequests failed", err)
		
		if validationErr, ok := err.(*errors.ValidationError); ok {
			response.ErrorWithFields(w, http.StatusBadRequest, "Validation failed", validationErr.Fields)
			return
		}

		if appErr, ok := err.(*errors.AppError); ok {
			if appErr.Code == 404 {
				response.Error(w, http.StatusNotFound, appErr.Message)
				return
			}
		}

		response.Error(w, http.StatusInternalServerError, "failed to retrieve pending leave requests: " + err.Error())
		return
	}

	response.Success(w, http.StatusOK, map[string]interface{}{
		"count":          len(leaveRequests),
		"leave_requests": leaveRequests,
	}, "Pending leave requests retrieved successfully")
}