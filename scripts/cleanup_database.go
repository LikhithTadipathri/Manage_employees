package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"

	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
)

func main() {
	// Try PostgreSQL first
	var db *sql.DB
	var err error

	dbURL := os.Getenv("DATABASE_URL")
	var dbType string

	if dbURL == "" {
		// Fallback to SQLite
		dbType = "sqlite"
		db, err = sql.Open("sqlite3", "employee_local.db")
		if err != nil {
			log.Fatalf("Failed to connect to SQLite: %v", err)
		}
		fmt.Println("✓ Connected to SQLite database")
	} else {
		dbType = "postgres"
		db, err = sql.Open("postgres", dbURL)
		if err != nil {
			log.Fatalf("Failed to connect to PostgreSQL: %v", err)
		}
		fmt.Println("✓ Connected to PostgreSQL database")
	}
	defer db.Close()

	// Test connection
	err = db.Ping()
	if err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	fmt.Println("\n⚠️  CLEANING UP DATABASE - DELETING ALL RECORDS\n")

	// Delete in correct order (respecting foreign keys)
	queries := []struct {
		name  string
		query string
	}{
		{
			"leave_requests",
			"DELETE FROM leave_requests",
		},
		{
			"employees",
			"DELETE FROM employees",
		},
		{
			"users",
			"DELETE FROM users",
		},
	}

	// Execute delete queries
	for _, q := range queries {
		fmt.Printf("Deleting %s... ", q.name)
		result, err := db.Exec(q.query)
		if err != nil {
			fmt.Printf("ERROR: %v\n", err)
			continue
		}

		rowsAffected, err := result.RowsAffected()
		if err != nil {
			fmt.Printf("ERROR getting rows affected: %v\n", err)
			continue
		}

		fmt.Printf("✓ Deleted %d records\n", rowsAffected)
	}

	// Reset sequences if PostgreSQL
	if dbType == "postgres" {
		fmt.Println("\nResetting PostgreSQL sequences...")
		sequenceQueries := []string{
			"ALTER SEQUENCE IF EXISTS leave_requests_id_seq RESTART WITH 1",
			"ALTER SEQUENCE IF EXISTS employees_id_seq RESTART WITH 1",
			"ALTER SEQUENCE IF EXISTS users_id_seq RESTART WITH 1",
		}

		for _, query := range sequenceQueries {
			if err := db.QueryRow(query).Err(); err != nil {
				// Some sequences might not exist, that's ok
				fmt.Printf("Sequence reset skipped: %v\n", err)
				continue
			}
			fmt.Printf("✓ %s\n", strings.Split(query, " ")[2])
		}
	}

	// Verify deletion
	fmt.Println("\n📊 VERIFICATION:")

	counts := []struct {
		table string
		query string
	}{
		{"leave_requests", "SELECT COUNT(*) FROM leave_requests"},
		{"employees", "SELECT COUNT(*) FROM employees"},
		{"users", "SELECT COUNT(*) FROM users"},
	}

	for _, c := range counts {
		var count int
		err := db.QueryRow(c.query).Scan(&count)
		if err != nil {
			fmt.Printf("✗ %s: ERROR %v\n", c.table, err)
			continue
		}
		fmt.Printf("✓ %s: %d records\n", c.table, count)
	}

	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("✅ DATABASE CLEANUP COMPLETE - READY FOR FRESH START")
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println("\nNext steps:")
	fmt.Println("1. Restart the server: go run main.go")
	fmt.Println("2. Server will auto-seed superadmin account")
	fmt.Println("3. Start fresh in Postman!")
}
