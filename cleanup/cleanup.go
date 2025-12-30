package main

import (
	"database/sql"
	"fmt"
	"log"

	"employee-service/config"

	_ "github.com/lib/pq"
)

func main() {
	// Load application config (loads .env if present)
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	dbURL := cfg.GetDatabaseURL()
	fmt.Printf("Using DB connection: %s\n", dbURL)

	// Connect to database
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer db.Close()

	// Delete old records
	queries := []string{
		"DELETE FROM leave_requests",
		"DELETE FROM leave_balances",
		"DELETE FROM employees WHERE id > 2",
		"DELETE FROM users WHERE id > 4",
	}

	for _, query := range queries {
		result, err := db.Exec(query)
		if err != nil {
			log.Printf("Error executing %s: %v", query, err)
		} else {
			rows, _ := result.RowsAffected()
			fmt.Printf("✓ Executed: %s (affected %d rows)\n", query, rows)
		}
	}

	// Check remaining
	var count int
	db.QueryRow("SELECT COUNT(*) FROM employees").Scan(&count)
	fmt.Printf("\n✓ Remaining employees: %d\n", count)

	var userCount int
	db.QueryRow("SELECT COUNT(*) FROM users").Scan(&userCount)
	fmt.Printf("✓ Remaining users: %d\n", userCount)

	fmt.Println("\n✓ Database cleanup complete!")
}
