package postgres

import (
	"database/sql"
	"fmt"
	"time"

	"employee-service/errors"
	"employee-service/models/leave"
	"employee-service/utils/helpers"
)

// LeaveRepository handles database operations for leave requests
type LeaveRepository struct {
	db *sql.DB
}

// NewLeaveRepository creates a new leave repository
func NewLeaveRepository(db *sql.DB) *LeaveRepository {
	return &LeaveRepository{db: db}
}

// CreateLeaveRequest creates a new leave request
func (r *LeaveRepository) CreateLeaveRequest(lr *leave.LeaveRequest) (*leave.LeaveRequest, error) {
	query := `
		INSERT INTO leave_requests (employee_id, leave_type, status, start_date, end_date, reason, days_count, notes, salary_deduction, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING id, created_at, updated_at
	`
	q := convertPlaceholders(query)

	now := time.Now()

	if helpers.DBType == "sqlite" {
		res, err := r.db.Exec(q,
			lr.EmployeeID,
			lr.LeaveType,
			leave.StatusPending,
			lr.StartDate,
			lr.EndDate,
			lr.Reason,
			lr.DaysCount,
			lr.Notes,
			lr.SalaryDeduction,
			now,
			now,
		)
		if err != nil {
			return nil, errors.WrapError("failed to create leave request", err)
		}

		lastID, err := res.LastInsertId()
		if err != nil {
			return nil, errors.WrapError("failed to get last insert id", err)
		}
		lr.ID = int(lastID)
		lr.Status = leave.StatusPending
		lr.CreatedAt = now
		lr.UpdatedAt = now

		return lr, nil
	}

	err := r.db.QueryRow(
		q,
		lr.EmployeeID,
		lr.LeaveType,
		leave.StatusPending,
		lr.StartDate,
		lr.EndDate,
		lr.Reason,
		lr.DaysCount,
		lr.Notes,
		lr.SalaryDeduction,
		now,
		now,
	).Scan(&lr.ID, &lr.CreatedAt, &lr.UpdatedAt)

	if err != nil {
		return nil, errors.WrapError("failed to create leave request", err)
	}

	lr.Status = leave.StatusPending
	return lr, nil
}

// GetLeaveRequest retrieves a leave request by ID
func (r *LeaveRepository) GetLeaveRequest(id int) (*leave.LeaveRequest, error) {
	query := `
		SELECT id, employee_id, leave_type, status, start_date, end_date, reason, days_count, 
		       notes, approved_by, approval_date, salary_deduction, created_at, updated_at
		FROM leave_requests
		WHERE id = $1
	`
	q := convertPlaceholders(query)

	var lr leave.LeaveRequest
	var approvedBy sql.NullInt64
	var approvalDate sql.NullTime
	var notes sql.NullString

	err := r.db.QueryRow(q, id).Scan(
		&lr.ID,
		&lr.EmployeeID,
		&lr.LeaveType,
		&lr.Status,
		&lr.StartDate,
		&lr.EndDate,
		&lr.Reason,
		&lr.DaysCount,
		&notes,
		&approvedBy,
		&approvalDate,
		&lr.SalaryDeduction,
		&lr.CreatedAt,
		&lr.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, errors.NewNotFoundError("leave request not found")
	}
	if err != nil {
		return nil, errors.WrapError("failed to get leave request", err)
	}

	if notes.Valid {
		lr.Notes = &notes.String
	}
	if approvedBy.Valid {
		approvedByInt := int(approvedBy.Int64)
		lr.ApprovedBy = &approvedByInt
	}
	if approvalDate.Valid {
		lr.ApprovalDate = &approvalDate.Time
	}

	return &lr, nil
}

// GetEmployeeLeaveRequests retrieves all leave requests for an employee
func (r *LeaveRepository) GetEmployeeLeaveRequests(employeeID int) ([]leave.LeaveRequest, error) {
	query := `
		SELECT id, employee_id, leave_type, status, start_date, end_date, reason, days_count, 
		       notes, approved_by, approval_date, salary_deduction, created_at, updated_at
		FROM leave_requests
		WHERE employee_id = $1
		ORDER BY created_at DESC
	`
	q := convertPlaceholders(query)

	rows, err := r.db.Query(q, employeeID)
	if err != nil {
		return nil, errors.WrapError("failed to query leave requests", err)
	}
	defer rows.Close()

	var leaveRequests []leave.LeaveRequest

	for rows.Next() {
		var lr leave.LeaveRequest
		var approvedBy sql.NullInt64
		var approvalDate sql.NullTime
		var notes sql.NullString

		err := rows.Scan(
			&lr.ID,
			&lr.EmployeeID,
			&lr.LeaveType,
			&lr.Status,
			&lr.StartDate,
			&lr.EndDate,
			&lr.Reason,
			&lr.DaysCount,
			&notes,
			&approvedBy,
			&approvalDate,
			&lr.SalaryDeduction,
			&lr.CreatedAt,
			&lr.UpdatedAt,
		)

		if err != nil {
			return nil, errors.WrapError("failed to scan leave request", err)
		}

		if notes.Valid {
			lr.Notes = &notes.String
		}
		if approvedBy.Valid {
			approvedByInt := int(approvedBy.Int64)
			lr.ApprovedBy = &approvedByInt
		}
		if approvalDate.Valid {
			lr.ApprovalDate = &approvalDate.Time
		}

		leaveRequests = append(leaveRequests, lr)
	}

	if err := rows.Err(); err != nil {
		return nil, errors.WrapError("error iterating leave requests", err)
	}

	return leaveRequests, nil
}

// GetAllLeaveRequests retrieves all leave requests with optional filtering
func (r *LeaveRepository) GetAllLeaveRequests(status string) ([]leave.LeaveRequestDetail, error) {
	var query string
	var args []interface{}
	
	// Use database-specific string concatenation
	if helpers.DBType == "sqlite" {
		query = `
			SELECT lr.id, lr.employee_id, lr.leave_type, lr.status, lr.start_date, lr.end_date, 
			       lr.reason, lr.days_count, lr.notes, lr.approved_by, lr.approval_date, lr.created_at, lr.updated_at,
			       (e.first_name || ' ' || e.last_name)
			FROM leave_requests lr
			JOIN employees e ON lr.employee_id = e.id
		`
	} else {
		query = `
			SELECT lr.id, lr.employee_id, lr.leave_type, lr.status, lr.start_date, lr.end_date, 
			       lr.reason, lr.days_count, lr.notes, lr.approved_by, lr.approval_date, lr.created_at, lr.updated_at,
			       CONCAT(e.first_name, ' ', e.last_name)
			FROM leave_requests lr
			JOIN employees e ON lr.employee_id = e.id
		`
	}

	paramIndex := 1
	if status != "" {
		if helpers.DBType == "sqlite" {
			query += " WHERE lr.status = ?"
		} else {
			query += fmt.Sprintf(" WHERE lr.status = $%d", paramIndex)
			paramIndex++
		}
		args = append(args, status)
	}

	query += " ORDER BY lr.created_at DESC"

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, errors.WrapError("failed to query leave requests", err)
	}
	defer rows.Close()

	var leaveRequests []leave.LeaveRequestDetail

	for rows.Next() {
		var lrd leave.LeaveRequestDetail
		var approvedBy sql.NullInt64
		var approvalDate sql.NullTime
		var notes sql.NullString

		err := rows.Scan(
			&lrd.ID,
			&lrd.EmployeeID,
			&lrd.LeaveType,
			&lrd.Status,
			&lrd.StartDate,
			&lrd.EndDate,
			&lrd.Reason,
			&lrd.DaysCount,
			&notes,
			&approvedBy,
			&approvalDate,
			&lrd.CreatedAt,
			&lrd.UpdatedAt,
			&lrd.EmployeeName,
		)

		if err != nil {
			return nil, errors.WrapError("failed to scan leave request", err)
		}

		if notes.Valid {
			lrd.Notes = &notes.String
		}
		if approvedBy.Valid {
			approvedByInt := int(approvedBy.Int64)
			lrd.ApprovedBy = &approvedByInt
		}
		if approvalDate.Valid {
			lrd.ApprovalDate = &approvalDate.Time
		}

		leaveRequests = append(leaveRequests, lrd)
	}

	if err := rows.Err(); err != nil {
		return nil, errors.WrapError("error iterating leave requests", err)
	}

	return leaveRequests, nil
}

// UpdateLeaveRequestStatus updates the status of a leave request
func (r *LeaveRepository) UpdateLeaveRequestStatus(id int, status string, approvedBy *int) error {
	query := `
		UPDATE leave_requests
		SET status = $1, approved_by = $2, approval_date = $3, updated_at = $4
		WHERE id = $5
	`
	q := convertPlaceholders(query)

	now := time.Now()
	
	// Only set approval_date if status is APPROVED
	var approvalDate interface{}
	if status == "APPROVED" {
		approvalDate = now
	} else {
		approvalDate = nil
	}

	result, err := r.db.Exec(
		q,
		status,
		approvedBy,
		approvalDate,
		now,
		id,
	)

	if err != nil {
		return errors.WrapError("failed to update leave request status", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return errors.WrapError("failed to get rows affected", err)
	}

	if rowsAffected == 0 {
		return errors.NewNotFoundError("leave request not found")
	}

	return nil
}

// CancelLeaveRequest cancels a leave request
func (r *LeaveRepository) CancelLeaveRequest(id int) error {
	query := `
		UPDATE leave_requests
		SET status = $1, updated_at = $2
		WHERE id = $3 AND status = $4
	`
	q := convertPlaceholders(query)

	now := time.Now()

	result, err := r.db.Exec(
		q,
		leave.StatusCancelled,
		now,
		id,
		leave.StatusPending,
	)

	if err != nil {
		return errors.WrapError("failed to cancel leave request", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return errors.WrapError("failed to get rows affected", err)
	}

	if rowsAffected == 0 {
		validationErr := errors.NewValidationError()
		validationErr.AddField("status", "only pending leave requests can be cancelled")
		return validationErr
	}

	return nil
}

// GetLeaveBalance retrieves the leave balance for an employee by leave type
func (r *LeaveRepository) GetLeaveBalance(employeeID int, leaveType leave.LeaveType) (*leave.LeaveBalance, error) {
	query := `
		SELECT id, employee_id, leave_type, balance, created_at, updated_at
		FROM leave_balances
		WHERE employee_id = $1 AND leave_type = $2
	`
	q := convertPlaceholders(query)

	var lb leave.LeaveBalance

	err := r.db.QueryRow(q, employeeID, leaveType).Scan(
		&lb.ID,
		&lb.EmployeeID,
		&lb.LeaveType,
		&lb.Balance,
		&lb.CreatedAt,
		&lb.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, errors.NewNotFoundError("leave balance not found")
	}
	if err != nil {
		return nil, errors.WrapError("failed to get leave balance", err)
	}

	return &lb, nil
}

// GetEmployeeLeaveBalances retrieves all leave balances for an employee
func (r *LeaveRepository) GetEmployeeLeaveBalances(employeeID int) ([]leave.LeaveBalance, error) {
	query := `
		SELECT id, employee_id, leave_type, balance, created_at, updated_at
		FROM leave_balances
		WHERE employee_id = $1
		ORDER BY leave_type
	`
	q := convertPlaceholders(query)

	rows, err := r.db.Query(q, employeeID)
	if err != nil {
		return nil, errors.WrapError("failed to query leave balances", err)
	}
	defer rows.Close()

	var balances []leave.LeaveBalance

	for rows.Next() {
		var lb leave.LeaveBalance

		err := rows.Scan(
			&lb.ID,
			&lb.EmployeeID,
			&lb.LeaveType,
			&lb.Balance,
			&lb.CreatedAt,
			&lb.UpdatedAt,
		)

		if err != nil {
			return nil, errors.WrapError("failed to scan leave balance", err)
		}

		balances = append(balances, lb)
	}

	if err := rows.Err(); err != nil {
		return nil, errors.WrapError("error iterating leave balances", err)
	}

	return balances, nil
}

// InitializeLeaveBalances initializes default leave balances for a new employee
func (r *LeaveRepository) InitializeLeaveBalances(employeeID int) error {
	// First verify employee exists
	var empExists int
	err := r.db.QueryRow("SELECT 1 FROM employees WHERE id = $1", employeeID).Scan(&empExists)
	if err != nil {
		if err == sql.ErrNoRows {
			return errors.WrapError("employee not found", err)
		}
		return errors.WrapError("failed to verify employee", err)
	}

	defaultBalances := leave.DefaultLeaveBalances()
	now := time.Now()

	for leaveType, balance := range defaultBalances {
		// Try to insert - if it already exists (ON CONFLICT), skip it silently
		query := `
			INSERT INTO leave_balances (employee_id, leave_type, balance, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5)
			ON CONFLICT (employee_id, leave_type) DO UPDATE SET updated_at = $5
		`
		q := convertPlaceholders(query)

		_, err := r.db.Exec(q, employeeID, leaveType, balance, now, now)
		if err != nil {
			// Log error but continue with other types
			errors.LogError(fmt.Sprintf("Failed to initialize %s balance for employee %d", leaveType, employeeID), err)
			// Don't return - continue with other leave types
		}
	}

	return nil
}

// UpdateLeaveBalance updates the leave balance for an employee
func (r *LeaveRepository) UpdateLeaveBalance(employeeID int, leaveType leave.LeaveType, newBalance int) error {
	query := `
		UPDATE leave_balances
		SET balance = $1, updated_at = $2
		WHERE employee_id = $3 AND leave_type = $4
	`
	q := convertPlaceholders(query)

	now := time.Now()

	result, err := r.db.Exec(q, newBalance, now, employeeID, leaveType)
	if err != nil {
		return errors.WrapError("failed to update leave balance", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return errors.WrapError("failed to get rows affected", err)
	}

	if rowsAffected == 0 {
		return errors.NewNotFoundError("leave balance not found")
	}

	return nil
}

// DeductLeaveBalance deducts the specified days from the employee's leave balance
func (r *LeaveRepository) DeductLeaveBalance(employeeID int, leaveType leave.LeaveType, days int) error {
	query := `
		UPDATE leave_balances
		SET balance = balance - $1, updated_at = $2
		WHERE employee_id = $3 AND leave_type = $4 AND balance >= $1
	`
	q := convertPlaceholders(query)

	now := time.Now()

	result, err := r.db.Exec(q, days, now, employeeID, leaveType)
	if err != nil {
		return errors.WrapError("failed to deduct leave balance", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return errors.WrapError("failed to get rows affected", err)
	}

	if rowsAffected == 0 {
		return errors.NewValidationError().AddField("balance", "insufficient leave balance")
	}

	return nil
}
// UpdateLeaveRequestNotes updates the notes field of a leave request
func (r *LeaveRepository) UpdateLeaveRequestNotes(id int, notes string) error {
	query := `
		UPDATE leave_requests
		SET notes = $1, updated_at = $2
		WHERE id = $3
	`
	q := convertPlaceholders(query)

	now := time.Now()

	result, err := r.db.Exec(q, notes, now, id)
	if err != nil {
		return errors.WrapError("failed to update leave request notes", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return errors.WrapError("failed to get rows affected", err)
	}

	if rowsAffected == 0 {
		return errors.NewNotFoundError("leave request not found")
	}

	return nil
}