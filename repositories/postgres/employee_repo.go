package postgres

import (
	"database/sql"
	"regexp"
	"time"

	"employee-service/errors"
	"employee-service/models/employee"
	"employee-service/utils/helpers"
)

func convertPlaceholders(query string) string {
	if helpers.DBType == "sqlite" {
		re := regexp.MustCompile(`\$\d+`)
		return re.ReplaceAllString(query, "?")
	}
	return query
}

// EmployeeRepository handles database operations for employees
type EmployeeRepository struct {
	db *sql.DB
}

// NewEmployeeRepository creates a new employee repository
func NewEmployeeRepository(db *sql.DB) *EmployeeRepository {
	return &EmployeeRepository{db: db}
}

// CreateEmployee creates a new employee in the database
func (r *EmployeeRepository) CreateEmployee(emp *employee.Employee) (*employee.Employee, error) {
	query := `
		INSERT INTO employees (user_id, first_name, last_name, email, phone, position, salary, gender, marital_status, hired_date, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		RETURNING id, created_at, updated_at
	`
	q := convertPlaceholders(query)

	now := time.Now()

	if helpers.DBType == "sqlite" {
		// SQLite: Exec then use LastInsertId and fetch timestamps
		res, err := r.db.Exec(q,
			emp.UserID,
			emp.FirstName,
			emp.LastName,
			emp.Email,
			emp.Phone,
			emp.Position,
			emp.Salary,
			emp.Gender,
			emp.MaritalStatus,
			emp.Hired,
			now,
			now,
		)
		if err != nil {
			return nil, errors.WrapError("failed to create employee", err)
		}

		lastID, err := res.LastInsertId()
		if err != nil {
			return nil, errors.WrapError("failed to get last insert id", err)
		}
		emp.ID = int(lastID)

		// fetch created_at and updated_at for sqlite
		row := r.db.QueryRow("SELECT created_at, updated_at FROM employees WHERE id = ?", emp.ID)
		if err := row.Scan(&emp.CreatedAt, &emp.UpdatedAt); err != nil {
			return nil, errors.WrapError("failed to retrieve timestamps", err)
		}

		return emp, nil
	}

	// Postgres path (RETURNING)
	err := r.db.QueryRow(
		q,
		emp.UserID,
		emp.FirstName,
		emp.LastName,
		emp.Email,
		emp.Phone,
		emp.Position,
		emp.Salary,
		emp.Gender,
		emp.MaritalStatus,
		emp.Hired,
		now,
		now,
	).Scan(&emp.ID, &emp.CreatedAt, &emp.UpdatedAt)

	if err != nil {
		return nil, errors.WrapError("failed to create employee", err)
	}

	return emp, nil
}

// GetEmployeeByID retrieves an employee by ID
func (r *EmployeeRepository) GetEmployeeByID(id int) (*employee.Employee, error) {
	query := `
		SELECT id, user_id, first_name, last_name, email, phone, position, salary, gender, marital_status, hired_date, created_at, updated_at
		FROM employees
		WHERE id = $1
	`
	q := convertPlaceholders(query)

	emp := &employee.Employee{}
	err := r.db.QueryRow(q, id).Scan(
		&emp.ID,
		&emp.UserID,
		&emp.FirstName,
		&emp.LastName,
		&emp.Email,
		&emp.Phone,
		&emp.Position,
		&emp.Salary,
		&emp.Gender,
		&emp.MaritalStatus,
		&emp.Hired,
		&emp.CreatedAt,
		&emp.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, errors.NotFoundError("Employee")
	}

	if err != nil {
		return nil, errors.WrapError("failed to fetch employee", err)
	}

	return emp, nil
}

// GetEmployeeByUserID retrieves an employee by user_id
func (r *EmployeeRepository) GetEmployeeByUserID(userID int) (*employee.Employee, error) {
	query := `
		SELECT id, user_id, first_name, last_name, email, phone, position, salary, gender, marital_status, hired_date, created_at, updated_at
		FROM employees
		WHERE user_id = $1
	`
	q := convertPlaceholders(query)

	emp := &employee.Employee{}
	err := r.db.QueryRow(q, userID).Scan(
		&emp.ID,
		&emp.UserID,
		&emp.FirstName,
		&emp.LastName,
		&emp.Email,
		&emp.Phone,
		&emp.Position,
		&emp.Salary,
		&emp.Gender,
		&emp.MaritalStatus,
		&emp.Hired,
		&emp.CreatedAt,
		&emp.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, errors.NotFoundError("Employee")
	}

	if err != nil {
		return nil, errors.WrapError("failed to fetch employee by user_id", err)
	}

	return emp, nil
}

// GetAllEmployees retrieves all employees with pagination
func (r *EmployeeRepository) GetAllEmployees(limit, offset int) ([]*employee.Employee, error) {
	query := `
		SELECT id, user_id, first_name, last_name, email, phone, position, salary, gender, marital_status, hired_date, created_at, updated_at
		FROM employees
		ORDER BY id DESC
		LIMIT $1 OFFSET $2
	`

	q := convertPlaceholders(query)

	rows, err := r.db.Query(q, limit, offset)
	if err != nil {
		return nil, errors.WrapError("failed to fetch employees", err)
	}
	defer rows.Close()

	var employees []*employee.Employee

	for rows.Next() {
		emp := &employee.Employee{}
		err := rows.Scan(
			&emp.ID,
			&emp.UserID,
			&emp.FirstName,
			&emp.LastName,
			&emp.Email,
			&emp.Phone,
			&emp.Position,
			&emp.Salary,
			&emp.Gender,
			&emp.MaritalStatus,
			&emp.Hired,
			&emp.CreatedAt,
			&emp.UpdatedAt,
		)

		if err != nil {
			return nil, errors.WrapError("failed to scan employee", err)
		}

		employees = append(employees, emp)
	}

	if err = rows.Err(); err != nil {
		return nil, errors.WrapError("error iterating employees", err)
	}

	return employees, nil
}

// UpdateEmployee updates an employee record
func (r *EmployeeRepository) UpdateEmployee(id int, updates *employee.UpdateEmployeeRequest) (*employee.Employee, error) {
	// First check if employee exists
	emp, err := r.GetEmployeeByID(id)
	if err != nil {
		return nil, err
	}

	// Apply updates
	if updates.FirstName != nil {
		emp.FirstName = *updates.FirstName//do it service, only queirs
		
	}
	if updates.LastName != nil {
		emp.LastName = *updates.LastName
	}
	if updates.Email != nil {
		emp.Email = *updates.Email
	}
	if updates.Phone != nil {
		emp.Phone = *updates.Phone
	}
	if updates.Position != nil {
		emp.Position = *updates.Position
	}
	if updates.Salary != nil {
		emp.Salary = *updates.Salary
	}
	if updates.Gender != nil {
		emp.Gender = *updates.Gender
	}
	if updates.MaritalStatus != nil {
		emp.MaritalStatus = *updates.MaritalStatus
	}

	emp.UpdatedAt = time.Now()

	query := `
		UPDATE employees
		SET first_name = $1, last_name = $2, email = $3, phone = $4, position = $5, salary = $6, gender = $7, marital_status = $8, updated_at = $9
		WHERE id = $10
		RETURNING updated_at
	`
	q := convertPlaceholders(query)

	if helpers.DBType == "sqlite" {
		// sqlite: Exec and return the updated object (UpdatedAt already set)
		_, err := r.db.Exec(q,
			emp.FirstName,
			emp.LastName,
			emp.Email,
			emp.Phone,
			emp.Position,
			emp.Salary,
			emp.Gender,
			emp.MaritalStatus,
			emp.UpdatedAt,
			id,
		)
		if err != nil {
			return nil, errors.WrapError("failed to update employee", err)
		}
		return emp, nil
	}

	// Postgres: use RETURNING
	err = r.db.QueryRow(
		q,
		emp.FirstName,
		emp.LastName,
		emp.Email,
		emp.Phone,
		emp.Position,
		emp.Salary,
		emp.Gender,
		emp.MaritalStatus,
		emp.UpdatedAt,
		id,
	).Scan(&emp.UpdatedAt)

	if err != nil {
		return nil, errors.WrapError("failed to update employee", err)
	}

	return emp, nil
}

// DeleteEmployee deletes an employee record
func (r *EmployeeRepository) DeleteEmployee(id int) error {
	// Check if employee exists first
	_, err := r.GetEmployeeByID(id)
	if err != nil {
		return err
	}

	query := "DELETE FROM employees WHERE id = $1"
	q := convertPlaceholders(query)

	result, err := r.db.Exec(q, id)
	if err != nil {
		return errors.WrapError("failed to delete employee", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return errors.WrapError("failed to get rows affected", err)
	}

	if rowsAffected == 0 {
		return errors.NotFoundError("Employee")
	}

	return nil
}

// GetEmployeeCount returns the total number of employees
func (r *EmployeeRepository) GetEmployeeCount() (int, error) {
	query := "SELECT COUNT(*) FROM employees"

	var count int
	err := r.db.QueryRow(query).Scan(&count)
	if err != nil {
		return 0, errors.WrapError("failed to count employees", err)
	}

	return count, nil
}

// SearchEmployees searches employees by query string and returns paginated results
func (r *EmployeeRepository) SearchEmployees(query string, limit, offset int) ([]*employee.Employee, error) {
	// Add wildcard for partial matching
	searchTerm := "%" + query + "%"

	var searchQuery string
	var rows *sql.Rows
	var err error

	if helpers.DBType == "sqlite" {
		// SQLite uses ? placeholders
		searchQuery = `
			SELECT id, first_name, last_name, email, phone, position, salary, hired_date, created_at, updated_at
			FROM employees
			WHERE first_name LIKE ?
			   OR last_name LIKE ?
			   OR email LIKE ?
			   OR phone LIKE ?
			   OR position LIKE ?
			ORDER BY id DESC
			LIMIT ? OFFSET ?
		`
		rows, err = r.db.Query(searchQuery, searchTerm, searchTerm, searchTerm, searchTerm, searchTerm, limit, offset)
	} else {
		// PostgreSQL uses $1, $2, etc.
		searchQuery = `
			SELECT id, first_name, last_name, email, phone, position, salary, hired_date, created_at, updated_at
			FROM employees
			WHERE first_name ILIKE $1
			   OR last_name ILIKE $1
			   OR email ILIKE $1
			   OR phone ILIKE $1
			   OR position ILIKE $1
			ORDER BY id DESC
			LIMIT $2 OFFSET $3
		`
		rows, err = r.db.Query(searchQuery, searchTerm, limit, offset)
	}

	if err != nil {
		return nil, errors.WrapError("failed to search employees", err)
	}
	defer rows.Close()

	var employees []*employee.Employee

	for rows.Next() {
		emp := &employee.Employee{}
		err := rows.Scan(
			&emp.ID,
			&emp.FirstName,
			&emp.LastName,
			&emp.Email,
			&emp.Phone,
			&emp.Position,
			&emp.Salary,
			&emp.Hired,
			&emp.CreatedAt,
			&emp.UpdatedAt,
		)

		if err != nil {
			return nil, errors.WrapError("failed to scan employee", err)
		}

		employees = append(employees, emp)
	}

	if err = rows.Err(); err != nil {
		return nil, errors.WrapError("error iterating search results", err)
	}

	return employees, nil
}

// GetSearchEmployeeCount returns the total count for search results
func (r *EmployeeRepository) GetSearchEmployeeCount(query string) (int, error) {
	// Add wildcard for partial matching
	searchTerm := "%" + query + "%"

	var countQuery string
	var count int
	var err error

	if helpers.DBType == "sqlite" {
		// SQLite uses ? placeholders
		countQuery = `
			SELECT COUNT(*) FROM employees
			WHERE first_name LIKE ?
			   OR last_name LIKE ?
			   OR email LIKE ?
			   OR phone LIKE ?
			   OR position LIKE ?
		`
		err = r.db.QueryRow(countQuery, searchTerm, searchTerm, searchTerm, searchTerm, searchTerm).Scan(&count)
	} else {
		// PostgreSQL uses $1, $2, etc.
		countQuery = `
			SELECT COUNT(*) FROM employees
			WHERE first_name ILIKE $1
			   OR last_name ILIKE $1
			   OR email ILIKE $1
			   OR phone ILIKE $1
			   OR position ILIKE $1
		`
		err = r.db.QueryRow(countQuery, searchTerm).Scan(&count)
	}

	if err != nil {
		return 0, errors.WrapError("failed to count search results", err)
	}

	return count, nil
}

// UpdateEmployeeSalary updates only the salary field of an employee
func (r *EmployeeRepository) UpdateEmployeeSalary(id int, newSalary float64) error {
	query := `
		UPDATE employees
		SET salary = $1, updated_at = $2
		WHERE id = $3
	`
	q := convertPlaceholders(query)

	now := time.Now()

	result, err := r.db.Exec(q, newSalary, now, id)
	if err != nil {
		return errors.WrapError("failed to update employee salary", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return errors.WrapError("failed to get rows affected", err)
	}

	if rowsAffected == 0 {
		return errors.NewNotFoundError("employee not found")
	}

	return nil
}


