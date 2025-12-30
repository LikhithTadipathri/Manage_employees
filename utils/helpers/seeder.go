package helpers

import (
	"database/sql"
	"time"

	"employee-service/errors"

	"golang.org/x/crypto/bcrypt"
)

// User struct for seeding
type TestUser struct {
	Username string
	Email    string
	Password string
	Role     string
}

// SeedUsers creates test users in the database
func SeedUsers(db *sql.DB) error {
	testUsers := []TestUser{
		{
			Username: "superadmin",
			Email:    "superadmin@system.local",
			Password: "passkey",
			Role:     "admin",
		},
		{
			Username: "admin",
			Email:    "admin@example.com",
			Password: "admin123",
			Role:     "admin",
		},
		{
			Username: "john_doe",
			Email:    "john@example.com",
			Password: "john123",
			Role:     "employee",
		},
		{
			Username: "jane_smith",
			Email:    "jane@example.com",
			Password: "jane123",
			Role:     "employee",
		},
	}

	for _, user := range testUsers {
		// Check if user already exists
		var count int
		var userID int
		err := db.QueryRow("SELECT COUNT(*) FROM users WHERE username = $1", user.Username).Scan(&count)
		if err != nil {
			return errors.WrapError("failed to check existing user", err)
		}

		if count > 0 {
			errors.LogInfo("User " + user.Username + " already exists, skipping...")
			// Get the user ID for employee creation
			err = db.QueryRow("SELECT id FROM users WHERE username = $1", user.Username).Scan(&userID)
			if err != nil {
				return errors.WrapError("failed to get user id", err)
			}
		} else {
			// Hash password
			hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
			if err != nil {
				return errors.WrapError("failed to hash password", err)
			}

			// Insert user
			query := `
				INSERT INTO users (username, email, password_hash, role, created_at, updated_at)
				VALUES ($1, $2, $3, $4, $5, $6)
				RETURNING id
			`

			now := time.Now()
			err = db.QueryRow(query, user.Username, user.Email, string(hashedPassword), user.Role, now, now).Scan(&userID)
			if err != nil {
				return errors.WrapError("failed to insert test user: "+user.Username, err)
			}

			errors.LogInfo("Test user created: " + user.Username + " (role: " + user.Role + ")")
		}

		// Create employee record for employees if not already exists
		if user.Role == "employee" {
			var empCount int
			err := db.QueryRow("SELECT COUNT(*) FROM employees WHERE user_id = $1", userID).Scan(&empCount)
			if err != nil {
				errors.LogError("failed to check existing employee for user "+user.Username, err)
				continue
			}

			if empCount == 0 {
				// Create employee record
				empQuery := `
					INSERT INTO employees (user_id, first_name, last_name, email, phone, position, salary, gender, marital_status, hired_date, created_at, updated_at)
					VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
				`

				now := time.Now()
				firstName := "Test"
				lastName := "User"
				gender := "Male"
				maritalStatus := false
				
				if user.Username == "john_doe" {
					firstName = "John"
					lastName = "Doe"
					gender = "Male"
					maritalStatus = true
				} else if user.Username == "jane_smith" {
					firstName = "Jane"
					lastName = "Smith"
					gender = "Female"
					maritalStatus = true
				}

				_, err := db.Exec(empQuery, 
					userID, 
					firstName, 
					lastName, 
					user.Email, 
					"555-0000", 
					"Software Engineer", 
					50000.00, 
					gender,
					maritalStatus,
					now,
					now, 
					now,
				)
				if err != nil {
					errors.LogError("failed to create employee record for user "+user.Username, err)
					continue
				}

				errors.LogInfo("Employee record created for user: " + user.Username)
			}
		}
	}

	errors.LogInfo("Database seeding completed successfully")
	return nil
}
