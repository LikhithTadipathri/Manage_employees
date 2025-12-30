package user

import "time"

// Role constants
const (
	RoleAdmin    = "admin"
	RoleEmployee = "employee"
	RoleUser     = "user"
)

// User represents a system user
type User struct {
	ID           int       `db:"id" json:"id"`
	Username     string    `db:"username" json:"username"`
	Email        string    `db:"email" json:"email"`
	PasswordHash string    `db:"password_hash" json:"-"` 
	Role         string    `db:"role" json:"role"`       
	CreatedAt    time.Time `db:"created_at" json:"created_at"`
	UpdatedAt    time.Time `db:"updated_at" json:"updated_at"`
}

// CreateUserRequest represents a user creation request
type CreateUserRequest struct {
	Username string `json:"username" validate:"required,min=3,max=50"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6"`
	Role     string `json:"role" validate:"required,oneof=admin employee user"`
}

// RegisterEmployeeRequest represents an employee registration request with both user and employee details
type RegisterEmployeeRequest struct {
	// User fields
	Username string `json:"username" validate:"required,min=3,max=50"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6"`
	Role     string `json:"role" validate:"required,oneof=employee user"`

	// Employee fields (required when role="employee")
	FirstName     string  `json:"first_name" validate:"required_if=Role employee,min=1,max=100"`
	LastName      string  `json:"last_name" validate:"required_if=Role employee,min=1,max=100"`
	Phone         string  `json:"phone" validate:"required_if=Role employee,min=5,max=20"`
	Position      string  `json:"position" validate:"required_if=Role employee,min=1,max=100"`
	Salary        float64 `json:"salary" validate:"required_if=Role employee,min=0"`
	Gender        string  `json:"gender" validate:"omitempty,oneof=Male Female"` // "Male" or "Female"
	MaritalStatus bool    `json:"marital_status"` // true = Married, false = Not Married
	HiredDate     *time.Time `json:"hired_date,omitempty"`
}

// LoginRequest represents login credentials
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// LoginResponse represents the login response
type LoginResponse struct {
	Token     string `json:"token"`
	User      User   `json:"user"`
	ExpiresAt int64  `json:"expires_at"`
}

// JWTClaims represents JWT token claims
type JWTClaims struct {
	UserID   int    `json:"user_id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Role     string `json:"role"`
}
