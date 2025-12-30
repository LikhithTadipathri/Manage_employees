package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"employee-service/errors"
	"employee-service/http/middlewares"
	"employee-service/http/response"
	"employee-service/models/employee"
	employeeService "employee-service/services/employee"
	leaveService "employee-service/services/leave"

	"github.com/go-chi/chi/v5"
)

// EmployeeHandler handles HTTP requests
type EmployeeHandler struct {
	service      *employeeService.Service
	leaveService *leaveService.Service
}

// NewEmployeeHandler creates a new employee handler
func NewEmployeeHandler(service *employeeService.Service) *EmployeeHandler {
	return &EmployeeHandler{service: service}
}

// NewEmployeeHandlerWithLeave creates a new employee handler with leave service
func NewEmployeeHandlerWithLeave(service *employeeService.Service, leaveService *leaveService.Service) *EmployeeHandler {
	return &EmployeeHandler{service: service, leaveService: leaveService}
}

// CreateEmployee handles POST /employees
func (h *EmployeeHandler) CreateEmployee(w http.ResponseWriter, r *http.Request) {
	// Authorization: any authenticated user can create
	userCtx, err := middlewares.GetUserFromContext(r)
	if err != nil {
		response.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	
	var req employee.CreateEmployeeRequest

	// Decode request body
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// ✔ Auto-link employee to current user
	userIDCopy := userCtx.UserID
	req.UserID = &userIDCopy

	// Create employee
	emp, err := h.service.CreateEmployee(&req)
	if err != nil {
		// Check if it's a validation error
		if validationErr, ok := err.(*errors.ValidationError); ok {
			response.ErrorWithFields(w, http.StatusBadRequest, "Validation failed", validationErr.Fields)
			return
		}

		// Check if it's a duplicate email error
		errStr := err.Error()
		if strings.Contains(errStr, "duplicate") || strings.Contains(errStr, "unique constraint") {
			response.Error(w, http.StatusConflict, "Email address already exists or you already have an employee record")
			return
		}

		errors.LogError("Failed to create employee", err)
		response.Error(w, http.StatusInternalServerError, "Failed to create employee")
		return
	}

	// Initialize leave balances for the new employee
	if h.leaveService != nil {
		if err := h.leaveService.InitializeLeaveBalances(emp.ID); err != nil {
			// Log the error but don't fail the employee creation
			errors.LogError("Failed to initialize leave balances for employee", err)
		}
	}

	response.Success(w, http.StatusCreated, emp, "Employee created successfully")
}

// GetEmployee handles GET /employees/{id}
func (h *EmployeeHandler) GetEmployee(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)

	if err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid employee ID")
		return
	}

	emp, err := h.service.GetEmployee(id)
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			response.Error(w, appErr.Code, appErr.Message)
			return
		}

		errors.LogError("Failed to fetch employee", err)
		response.Error(w, http.StatusInternalServerError, "Failed to fetch employee")
		return
	}

	response.Success(w, http.StatusOK, emp, "Employee retrieved successfully")
}

// 2Employees handles GET /employees
func (h *EmployeeHandler) ListEmployees(w http.ResponseWriter, r *http.Request) {
	// Get query parameters
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")
	
	// Check for search parameter (for convenience)
	searchQuery := r.URL.Query().Get("search")
	if searchQuery == "" {
		searchQuery = r.URL.Query().Get("q")
	}

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

	// If search query is provided, use search instead
	if searchQuery != "" {
		employees, total, err := h.service.SearchEmployees(searchQuery, limit, offset)
		if err != nil {
			errors.LogError("Failed to search employees", err)
			response.Error(w, http.StatusInternalServerError, "Failed to search employees")
			return
		}

		// Prepare response with pagination info
		data := map[string]interface{}{
			"employees": employees,
			"pagination": map[string]interface{}{
				"limit":  limit,
				"offset": offset,
				"total":  total,
			},
			"search_query": searchQuery,
		}

		response.Success(w, http.StatusOK, data, "Employees retrieved successfully (filtered by search)")
		return
	}

	employees, total, err := h.service.ListEmployees(limit, offset)
	if err != nil {
		errors.LogError("Failed to list employees", err)
		response.Error(w, http.StatusInternalServerError, "Failed to fetch employees")
		return
	}

	// Prepare response with pagination info
	data := map[string]interface{}{
		"employees": employees,
		"pagination": map[string]interface{}{
			"limit":  limit,
			"offset": offset,
			"total":  total,
		},
	}

	response.Success(w, http.StatusOK, data, "Employees retrieved successfully")
}

// UpdateEmployee handles PUT /employees/{id}
func (h *EmployeeHandler) UpdateEmployee(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)

	if err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid employee ID")
		return
	}

	var req employee.UpdateEmployeeRequest

	// Decode request body
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Authorization: only admin can update
	claims, err := middlewares.GetUserFromContext(r)
	if err != nil || claims.Role != "admin" {
		response.Error(w, http.StatusForbidden, "forbidden: admin only")
		return
	}

	// Update employee
	emp, err := h.service.UpdateEmployee(id, &req)
	if err != nil {
		// Check if it's a validation error
		if validationErr, ok := err.(*errors.ValidationError); ok {
			response.ErrorWithFields(w, http.StatusBadRequest, "Validation failed", validationErr.Fields)
			return
		}

		// Check if it's an app error
		if appErr, ok := err.(*errors.AppError); ok {
			response.Error(w, appErr.Code, appErr.Message)
			return
		}

		errors.LogError("Failed to update employee", err)
		response.Error(w, http.StatusInternalServerError, "Failed to update employee")
		return
	}

	response.Success(w, http.StatusOK, emp, "Employee updated successfully")
}

// DeleteEmployee handles DELETE /employees/{id}
func (h *EmployeeHandler) DeleteEmployee(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)

	if err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid employee ID")
		return
	}

	// Authorization: only admin can delete
	claims, err := middlewares.GetUserFromContext(r)
	if err != nil || claims.Role != "admin" {
		response.Error(w, http.StatusForbidden, "forbidden: admin only")
		return
	}

	// Delete employee
	err = h.service.DeleteEmployee(id)
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			response.Error(w, appErr.Code, appErr.Message)
			return
		}

		errors.LogError("Failed to delete employee", err)
		response.Error(w, http.StatusInternalServerError, "Failed to delete employee")
		return
	}

	response.SuccessNoData(w, http.StatusOK, "Employee deleted successfully")
}

// SearchEmployees handles GET /employees/search?q=search_term&limit=10&offset=0
func (h *EmployeeHandler) SearchEmployees(w http.ResponseWriter, r *http.Request) {
	// Get query parameters
	query := r.URL.Query().Get("q")
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

	// If no query provided, return error
	if query == "" {
		response.Error(w, http.StatusBadRequest, "Search query (q parameter) is required")
		return
	}

	employees, total, err := h.service.SearchEmployees(query, limit, offset)
	if err != nil {
		errors.LogError("Failed to search employees", err)
		response.Error(w, http.StatusInternalServerError, "Failed to search employees")
		return
	}

	// Prepare response with pagination info
	data := map[string]interface{}{
		"employees": employees,
		"pagination": map[string]interface{}{
			"limit":  limit,
			"offset": offset,
			"total":  total,
		},
		"query": query,
	}

	response.Success(w, http.StatusOK, data, "Search results retrieved successfully")
}
