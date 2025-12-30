package employee

import (
	"time"

	customErr "employee-service/errors"
)

// Employee represents an employee record
type Employee struct {
	ID             int       `json:"id"`
	UserID         *int      `json:"user_id"`
	FirstName      string    `json:"first_name"`
	LastName       string    `json:"last_name"`
	Email          string    `json:"email"`
	Phone          string    `json:"phone"`
	Position       string    `json:"position"`
	Salary         float64   `json:"salary"`
	Gender         string    `json:"gender"` // "Male" or "Female"
	MaritalStatus  bool      `json:"marital_status"` // true = Married, false = Not Married
	Hired          time.Time `json:"hired_date"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}




// CreateEmployeeRequest represents the request for creating an employee
type CreateEmployeeRequest struct {
	UserID         *int       `json:"user_id,omitempty"`
	Username       string     `json:"username"` // Required for login
	Password       string     `json:"password"` // Required for login
	FirstName      string     `json:"first_name"`
	LastName       string     `json:"last_name"`
	Email          string     `json:"email"`
	Phone          string     `json:"phone"`
	Position       string     `json:"position"`
	Salary         float64    `json:"salary"`
	Gender         string     `json:"gender"` // "Male" or "Female"
	MaritalStatus  bool       `json:"marital_status"` // true = Married, false = Not Married
	HiredDate      *time.Time `json:"hired_date,omitempty"`
}

// UpdateEmployeeRequest represents the request for updating an employee
type UpdateEmployeeRequest struct {
	FirstName      *string  `json:"first_name,omitempty"`
	LastName       *string  `json:"last_name,omitempty"`
	Email          *string  `json:"email,omitempty"`
	Phone          *string  `json:"phone,omitempty"`
	Position       *string  `json:"position,omitempty"`
	Salary         *float64 `json:"salary,omitempty"`
	Gender         *string  `json:"gender,omitempty"` // "Male" or "Female"
	MaritalStatus  *bool    `json:"marital_status,omitempty"` // true = Married, false = Not Married
}

// Validate validates the create employee request
func (c *CreateEmployeeRequest) Validate() error {
	validationErr := customErr.NewValidationError()

	if c.Username == "" {
		validationErr.AddFieldError("username", "Username is required")
	} else if len(c.Username) < 3 {
		validationErr.AddFieldError("username", "Username must be at least 3 characters long")
	}

	if c.Password == "" {
		validationErr.AddFieldError("password", "Password is required")
	} else if len(c.Password) < 6 {
		validationErr.AddFieldError("password", "Password must be at least 6 characters long")
	}

	if c.FirstName == "" {
		validationErr.AddFieldError("first_name", "First name is required")
	}

	if c.LastName == "" {
		validationErr.AddFieldError("last_name", "Last name is required")
	}

	if c.Email == "" {
		validationErr.AddFieldError("email", "Email is required")
	}

	if c.Phone == "" {
		validationErr.AddFieldError("phone", "Phone is required")
	}

	if c.Position == "" {
		validationErr.AddFieldError("position", "Position is required")
	}

	if c.Salary < 0 {
		validationErr.AddFieldError("salary", "Salary cannot be negative")
	}

	if c.Gender == "" {
		validationErr.AddFieldError("gender", "Gender is required (Male or Female)")
	} else if c.Gender != "Male" && c.Gender != "Female" {
		validationErr.AddFieldError("gender", "Gender must be 'Male' or 'Female'")
	}

	if c.MaritalStatus && (c.Gender == "Male" || c.Gender == "Female") {
		// MaritalStatus is just a boolean, no special validation needed
		// But we can add a note that it's required for maternity/paternity leave
	}

	return validationErr.Validate()
}

// ValidateUpdate validates the update employee request
func (u *UpdateEmployeeRequest) Validate() error {
	validationErr := customErr.NewValidationError()

	if u.FirstName != nil && *u.FirstName == "" {
		validationErr.AddFieldError("first_name", "First name cannot be empty")
	}

	if u.LastName != nil && *u.LastName == "" {
		validationErr.AddFieldError("last_name", "Last name cannot be empty")
	}

	if u.Email != nil && *u.Email == "" {
		validationErr.AddFieldError("email", "Email cannot be empty")
	}

	if u.Phone != nil && *u.Phone == "" {
		validationErr.AddFieldError("phone", "Phone cannot be empty")
	}

	if u.Position != nil && *u.Position == "" {
		validationErr.AddFieldError("position", "Position cannot be empty")
	}

	if u.Salary != nil && *u.Salary < 0 {
		validationErr.AddFieldError("salary", "Salary cannot be negative")
	}

	if u.Gender != nil && *u.Gender == "" {
		validationErr.AddFieldError("gender", "Gender cannot be empty")
	} else if u.Gender != nil && *u.Gender != "Male" && *u.Gender != "Female" {
		validationErr.AddFieldError("gender", "Gender must be 'Male' or 'Female'")
	}

	return validationErr.Validate()
}

// SearchEmployeeRequest represents the request for searching employees
type SearchEmployeeRequest struct {
	Query  string `json:"query"`
	Limit  int    `json:"limit"`
	Offset int    `json:"offset"`
}

