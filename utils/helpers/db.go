package helpers

import (
	"database/sql"
	"fmt"
	"strings"

	_ "github.com/lib/pq"
	_ "modernc.org/sqlite"

	"employee-service/errors"
)


var DBType = "postgres"

// InitDB initializes a database connection
func InitDB(connectionString string) (*sql.DB, error) {
	db, err := sql.Open("postgres", connectionString)
	if err != nil {
		return nil, errors.WrapError("failed to open database", err)
	}

	// Test the connection
	err = db.Ping()
	if err != nil {
		// If database doesn't exist, try to create it
		if strings.Contains(err.Error(), "database") && strings.Contains(err.Error(), "does not exist") {
			errors.LogInfo("Database does not exist, attempting to create it...")
			db, err = createDatabase(connectionString)
			if err != nil {
				return nil, err
			}
		} else {
			// Authentication or other connection failure — fall back to SQLite local DB
			errors.LogInfo("Postgres connection failed (will fall back to local SQLite): " + err.Error())
			// Open SQLite file in project root
			sqlitePath := "employee_local.db"
			sqliteDSN := fmt.Sprintf("file:%s?_foreign_keys=1", sqlitePath)
			sqliteDB, err2 := sql.Open("sqlite", sqliteDSN)
			if err2 != nil {
				return nil, errors.WrapError("failed to open fallback sqlite database", err2)
			}
			// Ensure sqlite is reachable
			if err2 = sqliteDB.Ping(); err2 != nil {
				return nil, errors.WrapError("failed to ping fallback sqlite database", err2)
			}
			DBType = "sqlite"
			errors.LogInfo("Falling back to local SQLite database: " + sqlitePath)
			return sqliteDB, nil
		}
	}

	errors.LogInfo("✅ Successfully connected to PostgreSQL database")

	return db, nil
}

// createDatabase creates the database if it doesn't exist
func createDatabase(connectionString string) (*sql.DB, error) {
	// Extract database name from connection string
	dbName := extractDBName(connectionString)
	
	// Connect to postgres database to create the target database
	postgresConnStr := strings.ReplaceAll(connectionString, "dbname="+dbName, "dbname=postgres")
	
	db, err := sql.Open("postgres", postgresConnStr)
	if err != nil {
		return nil, errors.WrapError("failed to connect to postgres database", err)
	}
	defer db.Close()

	// Create the database
	query := fmt.Sprintf("CREATE DATABASE %s;", dbName)
	_, err = db.Exec(query)
	if err != nil {
		return nil, errors.WrapError("failed to create database", err)
	}

	errors.LogInfo("Database created successfully")

	// Now connect to the newly created database
	newDB, err := sql.Open("postgres", connectionString)
	if err != nil {
		return nil, errors.WrapError("failed to open new database", err)
	}

	err = newDB.Ping()
	if err != nil {
		return nil, errors.WrapError("failed to connect to new database", err)
	}

	return newDB, nil
}

// extractDBName extracts database name from connection string
func extractDBName(connStr string) string {
	parts := strings.Split(connStr, " ")
	for _, part := range parts {
		if strings.HasPrefix(part, "dbname=") {
			return strings.TrimPrefix(part, "dbname=")
		}
	}
	return "employee_db"
}

// CloseDB closes the database connection
func CloseDB(db *sql.DB) error {
	if db != nil {
		return db.Close()
	}
	return nil
}

// Placeholder for backward compatibility
func CloseDatabase(db *sql.DB) error {
	return CloseDB(db)
}

// InitializeSchema initializes the database schema
func InitializeSchema(db *sql.DB) error {
	if DBType == "sqlite" {
		// SQLite users schema
		usersSchema := `
		CREATE TABLE IF NOT EXISTS users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			username TEXT NOT NULL UNIQUE,
			email TEXT NOT NULL UNIQUE,
			password_hash TEXT NOT NULL,
			role TEXT NOT NULL DEFAULT 'user',
			created_at DATETIME NOT NULL,
			updated_at DATETIME NOT NULL
		);`

		_, err := db.Exec(usersSchema)
		if err != nil {
			return errors.WrapError("failed to create users table (sqlite)", err)
		}

		// SQLite schema
		tableSchema := `
		CREATE TABLE IF NOT EXISTS employees (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER UNIQUE,
			first_name TEXT NOT NULL,
			last_name TEXT NOT NULL,
			email TEXT NOT NULL UNIQUE,
			phone TEXT NOT NULL,
			position TEXT NOT NULL,
			salary REAL NOT NULL,
			gender VARCHAR(10) NOT NULL DEFAULT 'Male',
			marital_status BOOLEAN NOT NULL DEFAULT FALSE,
			hired_date DATETIME NOT NULL,
			created_at DATETIME NOT NULL,
			updated_at DATETIME NOT NULL,
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
			CHECK (gender IN ('Male', 'Female'))
		);`

		_, err = db.Exec(tableSchema)
		if err != nil {
			return errors.WrapError("failed to create employees table (sqlite)", err)
		}

		// SQLite indexes
		indexQueries := []string{
			"CREATE INDEX IF NOT EXISTS idx_employees_user_id ON employees(user_id);",
			"CREATE INDEX IF NOT EXISTS idx_employees_email ON employees(email);",
			"CREATE INDEX IF NOT EXISTS idx_employees_created_at ON employees(created_at);",
		}

		for _, query := range indexQueries {
			_, err := db.Exec(query)
			if err != nil {
				return errors.WrapError("failed to create index (sqlite)", err)
			}
		}

		// SQLite leave_requests table
		leaveRequestsSchema := `
		CREATE TABLE IF NOT EXISTS leave_requests (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			employee_id INTEGER NOT NULL,
			leave_type TEXT NOT NULL,
			status TEXT NOT NULL DEFAULT 'pending',
			start_date DATETIME NOT NULL,
			end_date DATETIME NOT NULL,
			reason TEXT,
			days_count INTEGER NOT NULL,
			notes TEXT,
			approved_by INTEGER,
			approval_date DATETIME,
			salary_deduction REAL DEFAULT 0,
			created_at DATETIME NOT NULL,
			updated_at DATETIME NOT NULL,
			FOREIGN KEY (employee_id) REFERENCES employees(id) ON DELETE CASCADE,
			FOREIGN KEY (approved_by) REFERENCES users(id) ON DELETE SET NULL
		);`

		_, err = db.Exec(leaveRequestsSchema)
		if err != nil {
			return errors.WrapError("failed to create leave_requests table (sqlite)", err)
		}

		// SQLite leave_requests indexes
		leaveIndexQueries := []string{
			"CREATE INDEX IF NOT EXISTS idx_leave_requests_employee_id ON leave_requests(employee_id);",
			"CREATE INDEX IF NOT EXISTS idx_leave_requests_status ON leave_requests(status);",
			"CREATE INDEX IF NOT EXISTS idx_leave_requests_start_date ON leave_requests(start_date);",
			"CREATE INDEX IF NOT EXISTS idx_leave_requests_created_at ON leave_requests(created_at);",
		}

		for _, query := range leaveIndexQueries {
			_, err := db.Exec(query)
			if err != nil {
				return errors.WrapError("failed to create leave_requests index (sqlite)", err)
			}
		}

		// SQLite leave_balances table
		leaveBalancesSchema := `
		CREATE TABLE IF NOT EXISTS leave_balances (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			employee_id INTEGER NOT NULL,
			leave_type TEXT NOT NULL,
			balance INTEGER NOT NULL DEFAULT 0,
			created_at DATETIME NOT NULL,
			updated_at DATETIME NOT NULL,
			FOREIGN KEY (employee_id) REFERENCES employees(id) ON DELETE CASCADE,
			UNIQUE(employee_id, leave_type)
		);`

		_, err = db.Exec(leaveBalancesSchema)
		if err != nil {
			return errors.WrapError("failed to create leave_balances table (sqlite)", err)
		}

		// SQLite leave_balances indexes
		leaveBalancesIndexQueries := []string{
			"CREATE INDEX IF NOT EXISTS idx_leave_balances_employee_id ON leave_balances(employee_id);",
			"CREATE INDEX IF NOT EXISTS idx_leave_balances_employee_leave_type ON leave_balances(employee_id, leave_type);",
		}

		for _, query := range leaveBalancesIndexQueries {
			_, err := db.Exec(query)
			if err != nil {
				return errors.WrapError("failed to create leave_balances index (sqlite)", err)
			}
		}

		// SQLite notifications table
		notificationsSchema := `
		CREATE TABLE IF NOT EXISTS notifications (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			leave_request_id INTEGER,
			recipient_email TEXT NOT NULL,
			recipient_name TEXT NOT NULL,
			event_type TEXT NOT NULL,
			template_name TEXT NOT NULL,
			delivery_channel TEXT NOT NULL DEFAULT 'SMTP',
			status TEXT NOT NULL DEFAULT 'PENDING',
			subject TEXT NOT NULL,
			body TEXT NOT NULL,
			retry_count INTEGER DEFAULT 0,
			max_retries INTEGER DEFAULT 3,
			error_message TEXT,
			sent_at DATETIME,
			next_retry_at DATETIME,
			created_at DATETIME NOT NULL,
			updated_at DATETIME NOT NULL,
			FOREIGN KEY (leave_request_id) REFERENCES leave_requests(id) ON DELETE SET NULL
		);`

		_, err = db.Exec(notificationsSchema)
		if err != nil {
			return errors.WrapError("failed to create notifications table (sqlite)", err)
		}

		// SQLite notifications indexes
		notificationsIndexQueries := []string{
			"CREATE INDEX IF NOT EXISTS idx_notifications_leave_request_id ON notifications(leave_request_id);",
			"CREATE INDEX IF NOT EXISTS idx_notifications_status ON notifications(status);",
			"CREATE INDEX IF NOT EXISTS idx_notifications_created_at ON notifications(created_at);",
		}

		for _, query := range notificationsIndexQueries {
			_, err := db.Exec(query)
			if err != nil {
				return errors.WrapError("failed to create notifications index (sqlite)", err)
			}
		}

		errors.LogInfo("SQLite database schema initialized successfully")
		return nil
	}

	// Default: Postgres schema
	usersTableSchema := `
	CREATE TABLE IF NOT EXISTS users (
		id SERIAL PRIMARY KEY,
		username VARCHAR(50) NOT NULL UNIQUE,
		email VARCHAR(100) NOT NULL UNIQUE,
		password_hash VARCHAR(255) NOT NULL,
		role VARCHAR(20) NOT NULL DEFAULT 'user',
		created_at TIMESTAMP NOT NULL,
		updated_at TIMESTAMP NOT NULL
	);`

	_, err := db.Exec(usersTableSchema)
	if err != nil {
		return errors.WrapError("failed to create users table", err)
	}

	tableSchema := `
	CREATE TABLE IF NOT EXISTS employees (
		id SERIAL PRIMARY KEY,
		user_id INTEGER UNIQUE,
		first_name VARCHAR(100) NOT NULL,
		last_name VARCHAR(100) NOT NULL,
		email VARCHAR(100) NOT NULL UNIQUE,
		phone VARCHAR(20) NOT NULL,
		position VARCHAR(100) NOT NULL,
		salary DECIMAL(10, 2) NOT NULL,
		gender VARCHAR(10),
		marital_status BOOLEAN DEFAULT FALSE,
		hired_date TIMESTAMP NOT NULL,
		created_at TIMESTAMP NOT NULL,
		updated_at TIMESTAMP NOT NULL,
		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
	);`

	_, err = db.Exec(tableSchema)
	if err != nil {
		return errors.WrapError("failed to create employees table", err)
	}

	// Create indexes
	indexQueries := []string{
		"CREATE INDEX IF NOT EXISTS idx_employees_user_id ON employees(user_id);",
		"CREATE INDEX IF NOT EXISTS idx_employees_email ON employees(email);",
		"CREATE INDEX IF NOT EXISTS idx_employees_created_at ON employees(created_at);",
	}

	for _, query := range indexQueries {
		_, err := db.Exec(query)
		if err != nil {
			// Skip user_id index if column doesn't exist yet (migration not applied)
			if strings.Contains(err.Error(), "user_id") && strings.Contains(err.Error(), "does not exist") {
				errors.LogInfo("Skipping user_id index - migration 005 not yet applied. Run 'migrate up' to apply pending migrations.")
				continue
			}
			return errors.WrapError("failed to create index", err)
		}
	}

	// Create leave_requests table
	leaveRequestsTableSchema := `
	CREATE TABLE IF NOT EXISTS leave_requests (
		id SERIAL PRIMARY KEY,
		employee_id INTEGER NOT NULL,
		leave_type VARCHAR(50) NOT NULL,
		status VARCHAR(20) NOT NULL DEFAULT 'pending',
		start_date TIMESTAMP NOT NULL,
		end_date TIMESTAMP NOT NULL,
		reason TEXT,
		days_count INTEGER NOT NULL,
		notes TEXT,
		approved_by INTEGER,
		approval_date TIMESTAMP,
		salary_deduction DECIMAL(10, 2) DEFAULT 0,
		created_at TIMESTAMP NOT NULL,
		updated_at TIMESTAMP NOT NULL,
		FOREIGN KEY (employee_id) REFERENCES employees(id) ON DELETE CASCADE,
		FOREIGN KEY (approved_by) REFERENCES users(id) ON DELETE SET NULL
	);`

	_, err = db.Exec(leaveRequestsTableSchema)
	if err != nil {
		return errors.WrapError("failed to create leave_requests table", err)
	}

	// Create leave_requests indexes
	leaveIndexQueries := []string{
		"CREATE INDEX IF NOT EXISTS idx_leave_requests_employee_id ON leave_requests(employee_id);",
		"CREATE INDEX IF NOT EXISTS idx_leave_requests_status ON leave_requests(status);",
		"CREATE INDEX IF NOT EXISTS idx_leave_requests_start_date ON leave_requests(start_date);",
		"CREATE INDEX IF NOT EXISTS idx_leave_requests_created_at ON leave_requests(created_at);",
	}

	for _, query := range leaveIndexQueries {
		_, err := db.Exec(query)
		if err != nil {
			return errors.WrapError("failed to create leave_requests index", err)
		}
	}

	// Create leave_balances table
	leaveBalancesTableSchema := `
	CREATE TABLE IF NOT EXISTS leave_balances (
		id SERIAL PRIMARY KEY,
		employee_id INTEGER NOT NULL,
		leave_type VARCHAR(50) NOT NULL,
		balance INTEGER NOT NULL DEFAULT 0,
		created_at TIMESTAMP NOT NULL,
		updated_at TIMESTAMP NOT NULL,
		FOREIGN KEY (employee_id) REFERENCES employees(id) ON DELETE CASCADE,
		UNIQUE(employee_id, leave_type)
	);`

	_, err = db.Exec(leaveBalancesTableSchema)
	if err != nil {
		return errors.WrapError("failed to create leave_balances table", err)
	}
	errors.LogInfo("✅ leave_balances table created successfully")

	// Create leave_balances indexes
	leaveBalancesIndexQueries := []string{
		"CREATE INDEX IF NOT EXISTS idx_leave_balances_employee_id ON leave_balances(employee_id);",
		"CREATE INDEX IF NOT EXISTS idx_leave_balances_employee_leave_type ON leave_balances(employee_id, leave_type);",
	}

	for _, query := range leaveBalancesIndexQueries {
		_, err := db.Exec(query)
		if err != nil {
			return errors.WrapError("failed to create leave_balances index", err)
		}
	}
	errors.LogInfo("✅ leave_balances indexes created successfully")

	// Create notifications table
	notificationsTableSchema := `
	CREATE TABLE IF NOT EXISTS notifications (
		id SERIAL PRIMARY KEY,
		leave_request_id INTEGER,
		recipient_email VARCHAR(100) NOT NULL,
		recipient_name VARCHAR(100) NOT NULL,
		event_type VARCHAR(50) NOT NULL,
		template_name VARCHAR(100) NOT NULL,
		delivery_channel VARCHAR(20) NOT NULL DEFAULT 'SMTP',
		status VARCHAR(20) NOT NULL DEFAULT 'PENDING',
		subject TEXT NOT NULL,
		body TEXT NOT NULL,
		retry_count INTEGER DEFAULT 0,
		max_retries INTEGER DEFAULT 3,
		error_message TEXT,
		sent_at TIMESTAMP,
		next_retry_at TIMESTAMP,
		created_at TIMESTAMP NOT NULL,
		updated_at TIMESTAMP NOT NULL,
		FOREIGN KEY (leave_request_id) REFERENCES leave_requests(id) ON DELETE SET NULL
	);`

	_, err = db.Exec(notificationsTableSchema)
	if err != nil {
		return errors.WrapError("failed to create notifications table", err)
	}
	errors.LogInfo("✅ notifications table created successfully")

	// Create notifications indexes
	notificationsIndexQueries := []string{
		"CREATE INDEX IF NOT EXISTS idx_notifications_leave_request_id ON notifications(leave_request_id);",
		"CREATE INDEX IF NOT EXISTS idx_notifications_status ON notifications(status);",
		"CREATE INDEX IF NOT EXISTS idx_notifications_created_at ON notifications(created_at);",
	}

	for _, query := range notificationsIndexQueries {
		_, err := db.Exec(query)
		if err != nil {
			return errors.WrapError("failed to create notifications index", err)
		}
	}
	errors.LogInfo("✅ notifications indexes created successfully")

	errors.LogInfo("Database schema initialized successfully")
	return nil
}
