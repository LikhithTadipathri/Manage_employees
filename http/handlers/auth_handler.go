package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"employee-service/errors"
	"employee-service/http/middlewares"
	"employee-service/http/response"
	"employee-service/models/employee"
	usermodel "employee-service/models/user"
	"employee-service/repositories/postgres"
	leaveService "employee-service/services/leave"
	"employee-service/services/user"
	"employee-service/utils/jwt"
)

// AuthHandler handles authentication requests
type AuthHandler struct {
	userService      *user.UserService
	jwtManager       *jwt.JWTManager
	employeeRepo     *postgres.EmployeeRepository
	leaveService     *leaveService.Service
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(userService *user.UserService, jwtMgr *jwt.JWTManager) *AuthHandler {
	return &AuthHandler{
		userService:      userService,
		jwtManager:       jwtMgr,
		employeeRepo:     nil,
	}
}

// NewAuthHandlerWithEmployeeRepo creates a new auth handler with employee repository
func NewAuthHandlerWithEmployeeRepo(userService *user.UserService, jwtMgr *jwt.JWTManager, empRepo *postgres.EmployeeRepository) *AuthHandler {
	return &AuthHandler{
		userService:      userService,
		jwtManager:       jwtMgr,
		employeeRepo:     empRepo,
		leaveService:     nil,
	}
}

// NewAuthHandlerWithServices creates a new auth handler with all services
func NewAuthHandlerWithServices(userService *user.UserService, jwtMgr *jwt.JWTManager, empRepo *postgres.EmployeeRepository, leaveSvc *leaveService.Service) *AuthHandler {
	return &AuthHandler{
		userService:      userService,
		jwtManager:       jwtMgr,
		employeeRepo:     empRepo,
		leaveService:     leaveSvc,
	}
}

// Login handles POST /auth/login with real JWT and password verification
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req usermodel.LoginRequest

	// Decode request body
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate input
	if req.Username == "" || req.Password == "" {
		response.Error(w, http.StatusBadRequest, "Username and password are required")
		return
	}

	// Authenticate user (verifies password)
	userObj, err := h.userService.Authenticate(req.Username, req.Password)
	if err != nil {
		response.Error(w, http.StatusUnauthorized, "Invalid credentials")
		return
	}

	// Generate JWT token
	token, expiresAt, err := h.jwtManager.GenerateToken(userObj)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Failed to generate token")
		return
	}

	// Return response with token
	loginResp := usermodel.LoginResponse{
		Token:     token,
		User:      *userObj,
		ExpiresAt: expiresAt,
	}

	response.Success(w, http.StatusOK, loginResp, "Login successful")
}

// Logout handles POST /auth/logout
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	// Verify user is authenticated
	_, err := middlewares.GetUserFromContext(r)
	if err != nil {
		response.Error(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	
	response.SuccessNoData(w, http.StatusOK, "Logout successful")
}


// Register handles user/employee registration (Admin only)
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	// Verify user is authenticated AND is admin
	userCtx, err := middlewares.GetUserFromContext(r)
	if err != nil {
		response.Error(w, http.StatusUnauthorized, "Authentication required")
		return
	}

	// Check if user is admin
	if userCtx.Role != usermodel.RoleAdmin {
		response.Error(w, http.StatusForbidden, "Only admins can register new users")
		return
	}

	// Try to decode as RegisterEmployeeRequest first (extended format)
	var regEmpReq usermodel.RegisterEmployeeRequest
	if err := json.NewDecoder(r.Body).Decode(&regEmpReq); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate input
	if regEmpReq.Username == "" || regEmpReq.Password == "" || regEmpReq.Email == "" || regEmpReq.Role == "" {
		response.Error(w, http.StatusBadRequest, "Username, password, email, and role are required")
		return
	}

	// Validate role (only admin can create user or employee, not another admin)
	if regEmpReq.Role != usermodel.RoleUser && regEmpReq.Role != usermodel.RoleEmployee {
		response.Error(w, http.StatusBadRequest, "Only 'user' or 'employee' roles can be created by admin")
		return
	}

	// If role is employee, validate employee-specific fields
	if regEmpReq.Role == usermodel.RoleEmployee {
		if regEmpReq.FirstName == "" || regEmpReq.LastName == "" || regEmpReq.Phone == "" || regEmpReq.Position == "" {
			response.Error(w, http.StatusBadRequest, "For employee role, first_name, last_name, phone, and position are required")
			return
		}
		if regEmpReq.Salary < 0 {
			response.Error(w, http.StatusBadRequest, "Salary cannot be negative")
			return
		}
	}

	// Create basic user request
	userReq := &usermodel.CreateUserRequest{
		Username: regEmpReq.Username,
		Email:    regEmpReq.Email,
		Password: regEmpReq.Password,
		Role:     regEmpReq.Role,
	}

	// Register user
	userObj, err := h.userService.Register(userReq)
	if err != nil {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	// If employee role and employee repo is available, create employee record
	if regEmpReq.Role == usermodel.RoleEmployee && h.employeeRepo != nil {
		hiredDate := regEmpReq.HiredDate
		if hiredDate == nil {
			now := time.Now()
			hiredDate = &now
		}

		emp := &employee.Employee{
			UserID:        &userObj.ID,
			FirstName:     regEmpReq.FirstName,
			LastName:      regEmpReq.LastName,
			Email:         regEmpReq.Email,
			Phone:         regEmpReq.Phone,
			Position:      regEmpReq.Position,
			Salary:        regEmpReq.Salary,
			Gender:        regEmpReq.Gender,
			MaritalStatus: regEmpReq.MaritalStatus,
			Hired:         *hiredDate,
		}

		createdEmp, err := h.employeeRepo.CreateEmployee(emp)
		if err != nil {
			// Log error but don't fail - user is already created
			errors.LogError("Failed to create employee record", err)
			response.Error(w, http.StatusInternalServerError, "User created but failed to create employee record")
			return
		}

		// Initialize leave balances for the new employee
		if h.leaveService != nil {
			if err := h.leaveService.InitializeLeaveBalances(createdEmp.ID); err != nil {
				// Log the error but don't fail the registration
				errors.LogError("Failed to initialize leave balances for employee", err)
			}
		}
	}

	response.Success(w, http.StatusCreated, userObj, "User registered successfully")
}

// GetMe handles GET /auth/me - returns current authenticated user
func (h *AuthHandler) GetMe(w http.ResponseWriter, r *http.Request) {
	// Get user from context (set by JWT middleware)
	claims, err := middlewares.GetUserFromContext(r)
	if err != nil {
		response.Error(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	// Get full user data
	userObj, err := h.userService.GetUserByID(claims.UserID)
	if err != nil {
		response.Error(w, http.StatusNotFound, "User not found")
		return
	}

	response.Success(w, http.StatusOK, userObj, "User retrieved successfully")
}

