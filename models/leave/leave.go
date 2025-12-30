package leave

import (
	"employee-service/errors"
	"time"
)

// LeaveStatus represents the status of a leave request
type LeaveStatus string

const (
	StatusPending   LeaveStatus = "PENDING"
	StatusApproved  LeaveStatus = "APPROVED"
	StatusRejected  LeaveStatus = "REJECTED"
	StatusCancelled LeaveStatus = "CANCELLED"
)

// LeaveType represents the type of leave
type LeaveType string

const (
	TypeAnnual    LeaveType = "ANNUAL"
	TypeSick      LeaveType = "SICK"
	TypePersonal  LeaveType = "PERSONAL"
	TypeMaternity LeaveType = "MATERNITY"
	TypePaternity LeaveType = "PATERNITY"
	TypeUnpaid    LeaveType = "UNPAID"
	TypeCasual    LeaveType = "CASUAL"
)

// LeaveRequest represents a leave request
type LeaveRequest struct {
	ID           int       `json:"id"`
	EmployeeID   int       `json:"employee_id"`
	LeaveType    LeaveType `json:"leave_type"`
	StartDate    time.Time `json:"start_date"`
	EndDate      time.Time `json:"end_date"`
	Reason       string    `json:"reason"`
	DaysCount    int       `json:"days_count"`
	Status       LeaveStatus `json:"status"`
	Notes        *string   `json:"notes"`
	ApprovedBy   *int      `json:"approved_by"`
	ApprovalDate *time.Time `json:"approval_date"`
	SalaryDeduction float64 `json:"salary_deduction"` // Amount deducted from salary for paid leave
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// LeaveRequestDetail provides detailed view of leave request (for admin)
type LeaveRequestDetail struct {
	ID           int       `json:"id"`
	EmployeeID   int       `json:"employee_id"`
	EmployeeName string    `json:"employee_name"`
	LeaveType    LeaveType `json:"leave_type"`
	StartDate    time.Time `json:"start_date"`
	EndDate      time.Time `json:"end_date"`
	Reason       string    `json:"reason"`
	DaysCount    int       `json:"days_count"`
	Status       LeaveStatus `json:"status"`
	Notes        *string   `json:"notes"`
	ApprovedBy   *int      `json:"approved_by"`
	ApprovalDate *time.Time `json:"approval_date"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// ApplyLeaveRequest represents the request to apply for leave
type ApplyLeaveRequest struct {
	LeaveType LeaveType `json:"leave_type"`
	StartDate string    `json:"start_date"` // format: YYYY-MM-DD
	EndDate   string    `json:"end_date"`   // format: YYYY-MM-DD
	Reason    string    `json:"reason"`
	Notes     string    `json:"notes"`
}

// ApproveLeaveRequest represents the request to approve a leave
type ApproveLeaveRequest struct {
	Notes string `json:"notes"`
}

// RejectLeaveRequest represents the request to reject a leave
type RejectLeaveRequest struct {
	Reason string `json:"reason"`
}

// LeaveBalance represents the leave balance for an employee
type LeaveBalance struct {
	ID         int       `json:"id"`
	EmployeeID int       `json:"employee_id"`
	LeaveType  LeaveType `json:"leave_type"`
	Balance    int       `json:"balance"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// DefaultLeaveBalances returns the default leave balances
func DefaultLeaveBalances() map[LeaveType]int {
	return map[LeaveType]int{
		TypeAnnual:    10,
		TypeSick:      15,
		TypeCasual:    10,
		TypeMaternity: 90,
		TypePaternity: 7,
		TypeUnpaid:    10,
		TypePersonal:  10,
	}
}

// Validate validates the ApplyLeaveRequest
func (r *ApplyLeaveRequest) Validate() error {
	validationErr := errors.NewValidationError()

	if r.LeaveType == "" {
		validationErr.AddField("leave_type", "leave_type is required")
	}

	if r.StartDate == "" {
		validationErr.AddField("start_date", "start_date is required")
	}

	if r.EndDate == "" {
		validationErr.AddField("end_date", "end_date is required")
	}

	if r.Reason == "" {
		validationErr.AddField("reason", "reason is required")
	}

	// Parse dates if provided
	if r.StartDate != "" {
		startDate, err := time.Parse("2006-01-02", r.StartDate)
		if err != nil {
			validationErr.AddField("start_date", "invalid start_date format (use YYYY-MM-DD)")
		} else if startDate.Before(time.Now().Truncate(24 * time.Hour)) {
			validationErr.AddField("start_date", "start_date cannot be in the past")
		}
	}

	if r.EndDate != "" {
		_, err := time.Parse("2006-01-02", r.EndDate)
		if err != nil {
			validationErr.AddField("end_date", "invalid end_date format (use YYYY-MM-DD)")
		}
	}

	// Validate date range if both dates are valid
	if r.StartDate != "" && r.EndDate != "" {
		startDate, errStart := time.Parse("2006-01-02", r.StartDate)
		endDate, errEnd := time.Parse("2006-01-02", r.EndDate)
		if errStart == nil && errEnd == nil && startDate.After(endDate) {
			validationErr.AddField("end_date", "end_date must be after start_date")
		}
	}

	return validationErr.Validate()
}

// CalculateDays calculates working days between two dates (excludes weekends)
func CalculateDays(startDate, endDate time.Time) int {
	days := 0
	current := startDate

	for !current.After(endDate) {
		// 0 = Sunday, 6 = Saturday
		if current.Weekday() != time.Saturday && current.Weekday() != time.Sunday {
			days++
		}
		current = current.AddDate(0, 0, 1)
	}

	return days
}

// IsPaidLeave checks if the leave type is paid leave (deducts from salary)
// Only ANNUAL and SICK leaves have salary deduction
// MATERNITY, PATERNITY, UNPAID, PERSONAL, and CASUAL leaves do NOT deduct salary
func IsPaidLeave(leaveType LeaveType) bool {
	switch leaveType {
	case TypeAnnual, TypeSick:
		return true
	default:
		return false
	}
}

// GetLeaveLimit returns the maximum days allowed for each leave type
func GetLeaveLimit(leaveType LeaveType) int {
	switch leaveType {
	case TypeAnnual:
		return 20 // Default annual leave limit
	case TypeSick:
		return 10 // Sick leave limit
	case TypePersonal:
		return 10 // Personal leave limit
	case TypeMaternity:
		return 90 // Maternity leave limit
	case TypePaternity:
		return 7 // Paternity leave limit
	case TypeUnpaid:
		return 10 // Unpaid leave limit
	case TypeCasual:
		return 5 // Casual leave limit
	default:
		return 0
	}
}

// IsManagedLeave checks if the leave type has balance limits
func IsManagedLeave(leaveType LeaveType) bool {
	switch leaveType {
	case TypeAnnual, TypeSick, TypeCasual, TypeMaternity, TypePaternity, TypeUnpaid, TypePersonal:
		return true
	default:
		return false
	}
}

// CanApplyMaternityLeave checks if employee is eligible for maternity leave
// Returns (canApply, errorMessage)
func CanApplyMaternityLeave(gender string, isMarried bool) (bool, string) {
	if gender != "Female" && gender != "female" {
		return false, "maternity leave can only be applied by female employees"
	}
	if !isMarried {
		return false, "maternity leave can only be applied by married female employees"
	}
	return true, ""
}

