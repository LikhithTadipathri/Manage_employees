// Package docs contains swagger documentation for the Employee Management API
// @title Employee Management API
// @version 1.0
// @description Production-ready Employee Management System with Leave Management
// @termsOfService http://swagger.io/terms/
// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io
// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html
// @host localhost:8080
// @BasePath /api/v1
// @schemes http https
// @securityDefinitions.apikey Bearer
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.
package docs

import (
	"time"
)

// ============================================================================
// Authentication Models
// ============================================================================

// LoginRequest represents user login request
// @Description User login credentials
type LoginRequest struct {
	Email    string `json:"email" example:"user@example.com" binding:"required,email"`
	Password string `json:"password" example:"MyPassword!2025" binding:"required,min=12"`
}

// LoginResponse represents login response with JWT token
// @Description Successful login response with authentication token
type LoginResponse struct {
	Token     string    `json:"token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	UserID    int64     `json:"user_id" example:"1"`
	Email     string    `json:"email" example:"user@example.com"`
	ExpiresAt time.Time `json:"expires_at" example:"2025-12-31T23:59:59Z"`
}

// ============================================================================
// Employee Models
// ============================================================================

// CreateEmployeeRequest represents employee creation request
// @Description Request body for creating a new employee
type CreateEmployeeRequest struct {
	FirstName string `json:"first_name" example:"John" binding:"required,min=2,max=50"`
	LastName  string `json:"last_name" example:"Doe" binding:"required,min=2,max=50"`
	Email     string `json:"email" example:"john.doe@example.com" binding:"required,email"`
	Phone     string `json:"phone" example:"+919876543210" binding:"required"`
	PAN       string `json:"pan" example:"ABCDE1234F" binding:"required,len=10"`
	Aadhaar   string `json:"aadhaar" example:"123456789012" binding:"required,len=12"`
	Gender    string `json:"gender" example:"Male" binding:"required,oneof=Male Female Other"`
	DOB       string `json:"dob" example:"1990-05-15" binding:"required,datetime=2006-01-02"`
	Position  string `json:"position" example:"Software Engineer" binding:"required,min=3,max=100"`
	Department string `json:"department" example:"Engineering" binding:"required,min=3,max=50"`
	Salary    float64 `json:"salary" example:"50000.00" binding:"required,gt=0"`
}

// EmployeeResponse represents employee response
// @Description Employee data response
type EmployeeResponse struct {
	ID         int64     `json:"id" example:"1"`
	FirstName  string    `json:"first_name" example:"John"`
	LastName   string    `json:"last_name" example:"Doe"`
	Email      string    `json:"email" example:"john.doe@example.com"`
	Phone      string    `json:"phone" example:"+919876543210"`
	PAN        string    `json:"pan" example:"ABCDE1234F"`
	Aadhaar    string    `json:"aadhaar" example:"123456789012"`
	Gender     string    `json:"gender" example:"Male"`
	DOB        time.Time `json:"dob" example:"1990-05-15T00:00:00Z"`
	Position   string    `json:"position" example:"Software Engineer"`
	Department string    `json:"department" example:"Engineering"`
	Salary     float64   `json:"salary" example:"50000.00"`
	CreatedAt  time.Time `json:"created_at" example:"2025-12-30T15:05:33Z"`
	UpdatedAt  time.Time `json:"updated_at" example:"2025-12-30T15:05:33Z"`
}

// ============================================================================
// Leave Management Models
// ============================================================================

// CreateLeaveRequestRequest represents leave request creation
// @Description Request body for creating a leave request
type CreateLeaveRequestRequest struct {
	LeaveTypeID int64  `json:"leave_type_id" example:"1" binding:"required,gt=0"`
	StartDate   string `json:"start_date" example:"2025-01-15" binding:"required,datetime=2006-01-02"`
	EndDate     string `json:"end_date" example:"2025-01-20" binding:"required,datetime=2006-01-02"`
	Reason      string `json:"reason" example:"Maternity Leave" binding:"required,min=10,max=500"`
}

// LeaveRequestResponse represents leave request response
// @Description Leave request data
type LeaveRequestResponse struct {
	ID          int64     `json:"id" example:"1"`
	EmployeeID  int64     `json:"employee_id" example:"1"`
	LeaveTypeID int64     `json:"leave_type_id" example:"1"`
	StartDate   time.Time `json:"start_date" example:"2025-01-15T00:00:00Z"`
	EndDate     time.Time `json:"end_date" example:"2025-01-20T00:00:00Z"`
	Reason      string    `json:"reason" example:"Maternity Leave"`
	Status      string    `json:"status" example:"pending" enums:"pending,approved,rejected"`
	CreatedAt   time.Time `json:"created_at" example:"2025-12-30T15:05:33Z"`
	UpdatedAt   time.Time `json:"updated_at" example:"2025-12-30T15:05:33Z"`
}

// ApproveLeaveRequest represents leave approval request
// @Description Request to approve a leave request
type ApproveLeaveRequest struct {
	Reason string `json:"reason" example:"Approved by HR" binding:"min=5,max=200"`
}

// LeaveBalanceResponse represents employee leave balance
// @Description Current leave balance for an employee
type LeaveBalanceResponse struct {
	EmployeeID    int64                  `json:"employee_id" example:"1"`
	EmployeeName  string                 `json:"employee_name" example:"John Doe"`
	LeaveBalances map[string]interface{} `json:"leave_balances"`
	LastUpdated   time.Time              `json:"last_updated" example:"2025-12-30T15:05:33Z"`
}

// ============================================================================
// Health Check Models
// ============================================================================

// HealthCheckResponse represents server health status
// @Description Server health check response
type HealthCheckResponse struct {
	Status    string                 `json:"status" example:"healthy"`
	Timestamp time.Time              `json:"timestamp" example:"2025-12-30T15:05:33Z"`
	Checks    map[string]interface{} `json:"checks"`
}

// ============================================================================
// Error Models
// ============================================================================

// ErrorResponse represents error response
// @Description Standard error response format
type ErrorResponse struct {
	Status  string                 `json:"status" example:"error"`
	Message string                 `json:"message" example:"Invalid request"`
	Code    string                 `json:"code" example:"VALIDATION_ERROR"`
	Details map[string]interface{} `json:"details,omitempty"`
}

// ValidationError represents field validation error
// @Description Validation error details
type ValidationError struct {
	Field   string `json:"field" example:"email"`
	Message string `json:"message" example:"Invalid email format"`
}

// ============================================================================
// Success Response Models
// ============================================================================

// SuccessResponse represents successful response
// @Description Standard success response format
type SuccessResponse struct {
	Status string      `json:"status" example:"success"`
	Data   interface{} `json:"data"`
}

// ListResponse represents paginated list response
// @Description Paginated list response with metadata
type ListResponse struct {
	Status string        `json:"status" example:"success"`
	Data   interface{}   `json:"data"`
	Meta   PaginationMeta `json:"meta"`
}

// PaginationMeta represents pagination metadata
// @Description Pagination information
type PaginationMeta struct {
	Page      int   `json:"page" example:"1"`
	PageSize  int   `json:"page_size" example:"20"`
	Total     int64 `json:"total" example:"100"`
	TotalPages int  `json:"total_pages" example:"5"`
	HasNext   bool  `json:"has_next" example:"true"`
	HasPrev   bool  `json:"has_prev" example:"false"`
}

// ============================================================================
// API Endpoints Documentation (handlers implement these)
// ============================================================================

/*
AUTH ENDPOINTS:

POST /api/v1/auth/login
  @Description User login
  @Accept json
  @Produce json
  @Param request body LoginRequest true "Login credentials"
  @Success 200 {object} LoginResponse
  @Failure 400 {object} ErrorResponse "Invalid credentials"
  @Failure 500 {object} ErrorResponse "Internal server error"

EMPLOYEE ENDPOINTS:

POST /api/v1/employees
  @Description Create new employee
  @Security Bearer
  @Accept json
  @Produce json
  @Param request body CreateEmployeeRequest true "Employee data"
  @Success 201 {object} EmployeeResponse
  @Failure 400 {object} ErrorResponse "Validation error"
  @Failure 401 {object} ErrorResponse "Unauthorized"
  @Failure 409 {object} ErrorResponse "Employee already exists"

GET /api/v1/employees
  @Description List all employees (paginated)
  @Security Bearer
  @Produce json
  @Param page query int false "Page number" default(1)
  @Param page_size query int false "Records per page" default(20)
  @Success 200 {object} ListResponse
  @Failure 401 {object} ErrorResponse "Unauthorized"

GET /api/v1/employees/:id
  @Description Get employee by ID
  @Security Bearer
  @Produce json
  @Param id path int64 true "Employee ID"
  @Success 200 {object} EmployeeResponse
  @Failure 404 {object} ErrorResponse "Employee not found"
  @Failure 401 {object} ErrorResponse "Unauthorized"

PUT /api/v1/employees/:id
  @Description Update employee
  @Security Bearer
  @Accept json
  @Produce json
  @Param id path int64 true "Employee ID"
  @Param request body CreateEmployeeRequest true "Updated employee data"
  @Success 200 {object} EmployeeResponse
  @Failure 404 {object} ErrorResponse "Employee not found"
  @Failure 401 {object} ErrorResponse "Unauthorized"

DELETE /api/v1/employees/:id
  @Description Delete employee
  @Security Bearer
  @Param id path int64 true "Employee ID"
  @Success 204 "Employee deleted successfully"
  @Failure 404 {object} ErrorResponse "Employee not found"
  @Failure 401 {object} ErrorResponse "Unauthorized"

LEAVE REQUEST ENDPOINTS:

POST /api/v1/employees/:id/leave-requests
  @Description Create leave request
  @Security Bearer
  @Accept json
  @Produce json
  @Param id path int64 true "Employee ID"
  @Param request body CreateLeaveRequestRequest true "Leave request data"
  @Success 201 {object} LeaveRequestResponse
  @Failure 400 {object} ErrorResponse "Validation error"
  @Failure 401 {object} ErrorResponse "Unauthorized"

GET /api/v1/employees/:id/leave-requests
  @Description List leave requests for employee
  @Security Bearer
  @Produce json
  @Param id path int64 true "Employee ID"
  @Param status query string false "Filter by status" enums(pending,approved,rejected)
  @Success 200 {object} ListResponse
  @Failure 401 {object} ErrorResponse "Unauthorized"

PUT /api/v1/leave-requests/:id/approve
  @Description Approve leave request
  @Security Bearer
  @Accept json
  @Produce json
  @Param id path int64 true "Leave request ID"
  @Param request body ApproveLeaveRequest true "Approval reason"
  @Success 200 {object} LeaveRequestResponse
  @Failure 404 {object} ErrorResponse "Leave request not found"
  @Failure 401 {object} ErrorResponse "Unauthorized"

PUT /api/v1/leave-requests/:id/reject
  @Description Reject leave request
  @Security Bearer
  @Accept json
  @Produce json
  @Param id path int64 true "Leave request ID"
  @Param request body ApproveLeaveRequest true "Rejection reason"
  @Success 200 {object} LeaveRequestResponse
  @Failure 404 {object} ErrorResponse "Leave request not found"
  @Failure 401 {object} ErrorResponse "Unauthorized"

GET /api/v1/employees/:id/leave-balance
  @Description Get employee leave balance
  @Security Bearer
  @Produce json
  @Param id path int64 true "Employee ID"
  @Success 200 {object} LeaveBalanceResponse
  @Failure 404 {object} ErrorResponse "Employee not found"
  @Failure 401 {object} ErrorResponse "Unauthorized"

HEALTH CHECK ENDPOINTS:

GET /health
  @Description Health check endpoint
  @Produce json
  @Success 200 {object} HealthCheckResponse

GET /readiness
  @Description Readiness check endpoint
  @Produce json
  @Success 200 {object} HealthCheckResponse
  @Failure 503 {object} ErrorResponse "Service not ready"
*/
