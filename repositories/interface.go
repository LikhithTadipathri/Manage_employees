package repositories

import (
	"employee-service/models/employee"
	"employee-service/models/user"
)

// EmployeeRepository defines the interface for employee data operations
type EmployeeRepository interface {
	CreateEmployee(emp *employee.Employee) (*employee.Employee, error)
	GetEmployeeByID(id int) (*employee.Employee, error)
	GetAllEmployees(limit, offset int) ([]*employee.Employee, error)
	UpdateEmployee(id int, req *employee.UpdateEmployeeRequest) (*employee.Employee, error)
	DeleteEmployee(id int) error
	GetEmployeeCount() (int, error)
	SearchEmployees(query string, limit, offset int) ([]*employee.Employee, error)
	GetSearchEmployeeCount(query string) (int, error)
}

// UserRepository defines the interface for user data operations
type UserRepository interface {
	GetUserByUsername(username string) (*user.User, error)
	CreateUser(req *user.CreateUserRequest, passwordHash string) (*user.User, error)
	GetAllUsers() ([]*user.User, error)
	GetUserByID(id int) (*user.User, error)
	UpdateUserRole(userID int, role string) error
	DeleteUser(userID int) error
}
