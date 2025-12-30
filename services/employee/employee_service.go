package employee

import (
	"strings"
	"time"

	"employee-service/errors"
	"employee-service/models/employee"
	usermodel "employee-service/models/user"
	"employee-service/repositories/postgres"
	userService "employee-service/services/user"
)

// Service handles business logic for employees
type Service struct {
	repo        *postgres.EmployeeRepository
	userRepo    *postgres.UserRepository
	userService *userService.UserService
}

// NewService creates a new employee service
func NewService(repo *postgres.EmployeeRepository) *Service {
	return &Service{repo: repo}
}

// NewServiceWithUser creates a new employee service with user repository and service
func NewServiceWithUser(repo *postgres.EmployeeRepository, userRepo *postgres.UserRepository, userSvc *userService.UserService) *Service {
	return &Service{
		repo:        repo,
		userRepo:    userRepo,
		userService: userSvc,
	}
}

// CreateEmployee creates a new employee after validation and creates a user account for login
func (s *Service) CreateEmployee(req *employee.CreateEmployeeRequest) (*employee.Employee, error) {
	// Validate request
	if err := req.Validate(); err != nil {
		return nil, err
	}

	var userID *int

	// Create user account if username and password are provided
	if s.userService != nil && req.Username != "" && req.Password != "" {
		userReq := &usermodel.CreateUserRequest{
			Username: req.Username,
			Email:    req.Email,
			Password: req.Password,
			Role:     "employee", // Auto-assign employee role
		}

		user, err := s.userService.Register(userReq)
		if err != nil {
			return nil, err
		}
		userID = &user.ID
		req.UserID = userID
	}

	hiredDate := time.Now()
	if req.HiredDate != nil {
		hiredDate = *req.HiredDate
	}

	emp := &employee.Employee{
		UserID:        req.UserID,
		FirstName:     req.FirstName,
		LastName:      req.LastName,
		Email:         req.Email,
		Phone:         req.Phone,
		Position:      req.Position,
		Salary:        req.Salary,
		Gender:        req.Gender,
		MaritalStatus: req.MaritalStatus,
		Hired:         hiredDate,
	}

	return s.repo.CreateEmployee(emp)
}

// GetEmployee retrieves an employee by ID
func (s *Service) GetEmployee(id int) (*employee.Employee, error) {
	if id <= 0 {
		return nil, errors.BadRequestError("Invalid employee ID")
	}

	return s.repo.GetEmployeeByID(id)
}

// ListEmployees retrieves all employees with pagination
func (s *Service) ListEmployees(limit, offset int) ([]*employee.Employee, int, error) {
	// Validate pagination parameters
	if limit <= 0 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}

	employees, err := s.repo.GetAllEmployees(limit, offset)
	if err != nil {
		return nil, 0, err
	}

	total, err := s.repo.GetEmployeeCount()
	if err != nil {
		return nil, 0, err
	}

	return employees, total, nil
}

// UpdateEmployee updates an employee record
func (s *Service) UpdateEmployee(id int, req *employee.UpdateEmployeeRequest) (*employee.Employee, error) {
	if id <= 0 {
		return nil, errors.BadRequestError("Invalid employee ID")
	}

	// Validate update request
	if err := req.Validate(); err != nil {
		return nil, err
	}

	return s.repo.UpdateEmployee(id, req)
}

// DeleteEmployee deletes an employee record
func (s *Service) DeleteEmployee(id int) error {
	if id <= 0 {
		return errors.BadRequestError("Invalid employee ID")
	}

	return s.repo.DeleteEmployee(id)
}

// GetEmployeeCount returns the total count of employees
func (s *Service) GetEmployeeCount() (int, error) {
	return s.repo.GetEmployeeCount()
}


func (s *Service) SearchEmployees(query string, limit, offset int) ([]*employee.Employee, int, error) {
	// Trim and validate query
	query = strings.TrimSpace(query)
	if query == "" {
		// If query is empty, return all employees (fallback to ListEmployees)
		return s.ListEmployees(limit, offset)
	}

	// Validate pagination parameters
	if limit <= 0 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}

	employees, err := s.repo.SearchEmployees(query, limit, offset)
	if err != nil {
		return nil, 0, err
	}

	total, err := s.repo.GetSearchEmployeeCount(query)
	if err != nil {
		return nil, 0, err
	}

	return employees, total, nil
}

