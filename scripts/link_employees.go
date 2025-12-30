package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	// Try PostgreSQL first
	var db *sql.DB
	var err error

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		// Fallback to SQLite
		db, err = sql.Open("sqlite3", "employee_local.db")
		if err != nil {
			log.Fatalf("Failed to connect to SQLite: %v", err)
		}
	} else {
		db, err = sql.Open("postgres", dbURL)
		if err != nil {
			log.Fatalf("Failed to connect to PostgreSQL: %v", err)
		}
	}
	defer db.Close()

	// Test connection
	err = db.Ping()
	if err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	fmt.Println("\n✓ Database connection successful\n")

	// Get all employees without user_id
	rows, err := db.Query(`
		SELECT id, first_name, last_name, email, phone
		FROM employees
		WHERE user_id IS NULL
		ORDER BY id
	`)
	if err != nil {
		log.Fatalf("Failed to query employees: %v", err)
	}
	defer rows.Close()

	type Employee struct {
		ID        int
		FirstName string
		LastName  string
		Email     string
		Phone     string
	}

	var employees []Employee
	for rows.Next() {
		var emp Employee
		if err := rows.Scan(&emp.ID, &emp.FirstName, &emp.LastName, &emp.Email, &emp.Phone); err != nil {
			log.Fatalf("Failed to scan employee: %v", err)
		}
		employees = append(employees, emp)
	}

	if len(employees) == 0 {
		fmt.Println("✓ All employees already have user accounts!")
		return
	}

	fmt.Printf("\nFound %d employees without user accounts\n\n", len(employees))
	fmt.Println("Linking existing employees to user accounts...")
	fmt.Println(strings.Repeat("=", 70))

	tx, err := db.Begin()
	if err != nil {
		log.Fatalf("Failed to begin transaction: %v", err)
	}

	linkedCount := 0
	createdUsers := make(map[string]interface{})

	for i, emp := range employees {
		username := generateUsername(emp.FirstName, emp.LastName)
		password := "DefaultPass@123"

		// Hash password
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			fmt.Printf("❌ Failed to hash password for %s: %v\n", username, err)
			tx.Rollback()
			return
		}

		// Insert user
		var userID int
		err = tx.QueryRow(`
			INSERT INTO users (username, email, password_hash, role, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6)
			RETURNING id
		`, username, emp.Email, string(hashedPassword), "employee", time.Now(), time.Now()).Scan(&userID)

		if err != nil {
			fmt.Printf("❌ Failed to create user for %s %s: %v\n", emp.FirstName, emp.LastName, err)
			tx.Rollback()
			return
		}

		// Link employee to user
		_, err = tx.Exec(`
			UPDATE employees
			SET user_id = $1, updated_at = $2
			WHERE id = $3
		`, userID, time.Now(), emp.ID)

		if err != nil {
			fmt.Printf("❌ Failed to link employee %s %s: %v\n", emp.FirstName, emp.LastName, err)
			tx.Rollback()
			return
		}

		linkedCount++
		fmt.Printf("[%d/%d] ✓ %s %s\n", i+1, len(employees), emp.FirstName, emp.LastName)
		fmt.Printf("       Username: %s\n", username)
		fmt.Printf("       Password: %s\n", password)
		fmt.Printf("       User ID: %d\n", userID)
		fmt.Printf("       Email: %s\n\n", emp.Email)

		createdUsers[username] = map[string]interface{}{
			"username": username,
			"password": password,
			"email":    emp.Email,
			"user_id":  userID,
		}
	}

	err = tx.Commit()
	if err != nil {
		log.Fatalf("Failed to commit transaction: %v", err)
	}

	fmt.Println(strings.Repeat("=", 70))
	fmt.Printf("✓ Successfully linked %d employees to user accounts\n", linkedCount)
	fmt.Println("\nAll employees can now log in and apply for leave!")
	fmt.Println("\nLogin with any of the generated credentials above.")
}

// generateUsername creates a username from first and last name
func generateUsername(firstName, lastName string) string {
	// Remove spaces and combine
	first := strings.ToLower(strings.TrimSpace(firstName))
	last := strings.ToLower(strings.TrimSpace(lastName))
	username := first + "_" + last
	return strings.ReplaceAll(username, " ", "_")
}
