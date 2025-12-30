package leave

import (
	"fmt"
	"time"

	"employee-service/errors"
	"employee-service/models/employee"
	"employee-service/models/leave"
	"employee-service/models/notification"
	"employee-service/models/user"
	"employee-service/repositories/postgres"
	"employee-service/services/email"
)

// Service handles business logic for leave requests
type Service struct {
	repository         *postgres.LeaveRepository
	employeeRepository *postgres.EmployeeRepository
	userRepository     *postgres.UserRepository
	notificationRepo   *postgres.NotificationRepository
	emailQueue         *email.EmailQueue
}

// NewService creates a new leave service
func NewService(
	repository *postgres.LeaveRepository,
	employeeRepository *postgres.EmployeeRepository,
	userRepository *postgres.UserRepository,
	notificationRepo *postgres.NotificationRepository,
	emailQueue *email.EmailQueue,
) *Service {
	return &Service{
		repository:         repository,
		employeeRepository: employeeRepository,
		userRepository:     userRepository,
		notificationRepo:   notificationRepo,
		emailQueue:         emailQueue,
	}
}

// ApplyLeave handles leave application
func (s *Service) ApplyLeave(userID int, req *leave.ApplyLeaveRequest) (*leave.LeaveRequest, error) {
	// Validate request
	if err := req.Validate(); err != nil {
		return nil, err
	}

	// Get employee by user_id
	emp, err := s.employeeRepository.GetEmployeeByUserID(userID)
	if err != nil {
		return nil, errors.NotFoundError("employee record")
	}

	// Validate maternity leave eligibility (only females and married)
	if req.LeaveType == leave.TypeMaternity {
		if emp.Gender != "Female" {
			validationErr := errors.NewValidationError()
			validationErr.AddField("leave_type", "maternity leave is only available for female employees")
			return nil, validationErr
		}
		if !emp.MaritalStatus {
			validationErr := errors.NewValidationError()
			validationErr.AddField("leave_type", "maternity leave is only available for married employees")
			return nil, validationErr
		}
	}

	// Validate paternity leave eligibility (only males and married)
	if req.LeaveType == leave.TypePaternity {
		if emp.Gender != "Male" {
			validationErr := errors.NewValidationError()
			validationErr.AddField("leave_type", "paternity leave is only available for male employees")
			return nil, validationErr
		}
		if !emp.MaritalStatus {
			validationErr := errors.NewValidationError()
			validationErr.AddField("leave_type", "paternity leave is only available for married employees")
			return nil, validationErr
		}
	}

	// Parse dates
	startDate, err := time.Parse("2006-01-02", req.StartDate)
	if err != nil {
		return nil, errors.NewValidationError().AddField("start_date", "invalid date format")
	}
	endDate, err := time.Parse("2006-01-02", req.EndDate)
	if err != nil {
		return nil, errors.NewValidationError().AddField("end_date", "invalid date format")
	}

	// Calculate days
	daysCount := leave.CalculateDays(startDate, endDate)
	if daysCount == 0 {
		return nil, errors.NewValidationError().AddField("dates", "leave period must include at least one working day")
	}

	// Check leave balance - Auto-initialize if not found for managed leave types
	balance, err := s.repository.GetLeaveBalance(emp.ID, req.LeaveType)
	if err != nil {
		// If balance doesn't exist, try to initialize it for managed leave types
		if req.LeaveType == leave.TypeAnnual || req.LeaveType == leave.TypeSick || req.LeaveType == leave.TypeCasual ||
			req.LeaveType == leave.TypeMaternity || req.LeaveType == leave.TypePaternity || req.LeaveType == leave.TypeUnpaid || req.LeaveType == leave.TypePersonal {
			// Auto-initialize leave balances for this employee
			errors.LogInfo(fmt.Sprintf("Auto-initializing leave balances for employee %d", emp.ID))
			if err := s.repository.InitializeLeaveBalances(emp.ID); err != nil {
				errors.LogError(fmt.Sprintf("Failed to initialize leave balances for employee %d", emp.ID), err)
				validationErr := errors.NewValidationError()
				validationErr.AddField("leave_balance", fmt.Sprintf("failed to initialize leave balance: %v", err))
				return nil, validationErr
			}
			// Retry getting the balance after initialization
			balance, err = s.repository.GetLeaveBalance(emp.ID, req.LeaveType)
			if err != nil {
				errors.LogError(fmt.Sprintf("Failed to get leave balance after initialization for employee %d, type %s", emp.ID, req.LeaveType), err)
				validationErr := errors.NewValidationError()
				validationErr.AddField("leave_balance", fmt.Sprintf("leave balance not found for this leave type: %v", err))
				return nil, validationErr
			}
		}
	}
	
	// Check if balance is sufficient
	if balance != nil && balance.Balance < daysCount {
		// Auto-reject if insufficient balance
		validationErr := errors.NewValidationError()
		validationErr.AddField("leave_balance", fmt.Sprintf("insufficient leave balance. Required: %d, Available: %d", daysCount, balance.Balance))
		return nil, validationErr
	}

	// Create leave request using the employee's ID (not user_id)
	var notes *string
	if req.Notes != "" {
		notes = &req.Notes
	}
	leaveRequest := &leave.LeaveRequest{
		EmployeeID: emp.ID,
		LeaveType:  req.LeaveType,
		StartDate:  startDate,
		EndDate:    endDate,
		Reason:     req.Reason,
		Notes:      notes,
		DaysCount:  daysCount,
	}

	// Calculate salary deduction for paid leaves (500 per day)
	if leave.IsPaidLeave(req.LeaveType) {
		leaveRequest.SalaryDeduction = float64(daysCount) * 500.0
	}

	result, err := s.repository.CreateLeaveRequest(leaveRequest)
	if err != nil {
		return nil, err
	}

	// Queue email notifications asynchronously (don't block on email errors)
	errors.LogInfo(fmt.Sprintf("🚀 LEAVE APPLIED: Leave request %d created | Employee: %d | Type: %s", result.ID, emp.ID, req.LeaveType))
	errors.LogInfo(fmt.Sprintf("📧 Queuing email notifications in background goroutine..."))
	go s.queueLeaveAppliedNotifications(result, emp)

	return result, nil
}

// queueLeaveAppliedNotifications queues email notifications for leave application
func (s *Service) queueLeaveAppliedNotifications(leaveReq *leave.LeaveRequest, emp *employee.Employee) {
	errors.LogInfo(fmt.Sprintf("\n📧 ========== QUEUEING LEAVE APPLIED NOTIFICATIONS =========="))
	errors.LogInfo(fmt.Sprintf("Leave Request ID: %d | Employee: %s (%d)\n", leaveReq.ID, emp.FirstName+" "+emp.LastName, emp.ID))
	
	// Prepare template data
	templateData := notification.TemplateData{
		EmployeeName:    fmt.Sprintf("%s %s", emp.FirstName, emp.LastName),
		EmployeeEmail:   emp.Email,
		EmployeeID:      emp.ID,
		LeaveType:       string(leaveReq.LeaveType),
		StartDate:       leaveReq.StartDate.Format("2006-01-02"),
		EndDate:         leaveReq.EndDate.Format("2006-01-02"),
		TotalDays:       leaveReq.DaysCount,
		Reason:          leaveReq.Reason,
		IsPaidLeave:     leave.IsPaidLeave(leaveReq.LeaveType),
	}

	// 1. Send email to employee
	errors.LogInfo(fmt.Sprintf("1️⃣  Sending notification to EMPLOYEE"))
	empTemplate := notification.GetTemplate(notification.EventLeaveApplied, leave.IsPaidLeave(leaveReq.LeaveType))
	empSubject, empBody, err := email.RenderTemplate(empTemplate, templateData)
	if err != nil {
		errors.LogError("Failed to render employee notification template", err)
	} else {
		empNotif := &notification.Notification{
			LeaveRequestID:  &leaveReq.ID,
			RecipientEmail:  emp.Email,
			RecipientName:   fmt.Sprintf("%s %s", emp.FirstName, emp.LastName),
			EventType:       notification.EventLeaveApplied,
			TemplateName:    empTemplate.Name,
			DeliveryChannel: notification.ChannelSMTP,
			Status:          notification.StatusPending,
			Subject:         empSubject,
			Body:            empBody,
			MaxRetries:      3,
		}

		createdEmpNotif, err := s.notificationRepo.CreateNotification(empNotif)
		if err != nil {
			errors.LogError("Failed to create employee notification record in DB", err)
		} else {
			errors.LogInfo(fmt.Sprintf("   ✅ Notification record created (ID will be assigned by DB)"))
			// Enqueue for sending - USE THE RETURNED NOTIFICATION WITH ID!
			err = s.emailQueue.Enqueue(createdEmpNotif)
			if err != nil {
				errors.LogError("Failed to enqueue employee notification", err)
			} else {
				errors.LogInfo(fmt.Sprintf("   ✅ Enqueued for email delivery to: %s\n", emp.Email))
			}
		}
	}

	// 2. Send email to admin
	errors.LogInfo(fmt.Sprintf("2️⃣  Sending notifications to ALL ADMINS"))
	adminTemplate := notification.GetAdminTemplate(notification.EventLeaveApplied)
	// Get first admin user for now (in production, you might want to send to all admins)
	admins, err := s.userRepository.GetUsersByRole(user.RoleAdmin)
	if err != nil {
		errors.LogError("Failed to get admin users", err)
		return
	}

	if len(admins) == 0 {
		errors.LogInfo("   ⚠️  No admin users found to notify")
		return
	}

	errors.LogInfo(fmt.Sprintf("   Found %d admin(s) to notify", len(admins)))
	for _, admin := range admins {
		templateData.AdminName = admin.Username
		templateData.AdminEmail = admin.Email

		adminSubject, adminBody, err := email.RenderTemplate(adminTemplate, templateData)
		if err != nil {
			errors.LogError("Failed to render admin notification template", err)
			continue
		}

		adminNotif := &notification.Notification{
			LeaveRequestID:  &leaveReq.ID,
			RecipientEmail:  admin.Email,
			RecipientName:   admin.Username,
			EventType:       notification.EventLeaveApplied,
			TemplateName:    adminTemplate.Name,
			DeliveryChannel: notification.ChannelSMTP,
			Status:          notification.StatusPending,
			Subject:         adminSubject,
			Body:            adminBody,
			MaxRetries:      3,
		}

		createdAdminNotif, err := s.notificationRepo.CreateNotification(adminNotif)
		if err != nil {
			errors.LogError("Failed to create admin notification record in DB", err)
			continue
		} else {
			errors.LogInfo(fmt.Sprintf("   ✅ Notification record created for admin: %s", admin.Username))
		}

		// Enqueue for sending - USE THE RETURNED NOTIFICATION WITH ID!
		err = s.emailQueue.Enqueue(createdAdminNotif)
		if err != nil {
			errors.LogError("Failed to enqueue admin notification", err)
		} else {
			errors.LogInfo(fmt.Sprintf("   ✅ Enqueued for email delivery to: %s\n", admin.Email))
		}
	}
	
	errors.LogInfo(fmt.Sprintf("========== LEAVE APPLIED NOTIFICATIONS QUEUED ==========\n"))
}

// GetEmployeeLeaveRequests retrieves all leave requests for an employee
func (s *Service) GetEmployeeLeaveRequests(userID int) ([]leave.LeaveRequest, error) {
	// Get employee by user_id
	emp, err := s.employeeRepository.GetEmployeeByUserID(userID)
	if err != nil {
		return nil, errors.NotFoundError("employee record")
	}

	requests, err := s.repository.GetEmployeeLeaveRequests(emp.ID)
	if err != nil {
		return nil, err
	}

	if len(requests) == 0 {
		return []leave.LeaveRequest{}, nil
	}

	return requests, nil
}

// GetLeaveRequest retrieves a single leave request
func (s *Service) GetLeaveRequest(id int) (*leave.LeaveRequest, error) {
	return s.repository.GetLeaveRequest(id)
}

// CancelLeave cancels a pending leave request
func (s *Service) CancelLeave(id int, userID int) error {
	// Get employee by user_id
	emp, err := s.employeeRepository.GetEmployeeByUserID(userID)
	if err != nil {
		return errors.NotFoundError("employee record")
	}

	// Verify the leave request belongs to the employee
	leaveRequest, err := s.repository.GetLeaveRequest(id)
	if err != nil {
		return err
	}

	if leaveRequest.EmployeeID != emp.ID {
		return errors.NewForbiddenError("you don't have permission to cancel this leave request")
	}

	if leaveRequest.Status != leave.StatusPending {
		return errors.NewValidationError().AddField("status", "only pending leave requests can be cancelled")
	}

	return s.repository.CancelLeaveRequest(id)
}

// GetAllLeaveRequests retrieves all leave requests (admin only)
func (s *Service) GetAllLeaveRequests(status string) ([]leave.LeaveRequestDetail, error) {
	return s.repository.GetAllLeaveRequests(status)
}

// ApproveLeave approves a leave request (admin only)
func (s *Service) ApproveLeave(id int, approvedByUserID int, notes string) error {
	leaveRequest, err := s.repository.GetLeaveRequest(id)
	if err != nil {
		return err
	}

	if leaveRequest.Status != leave.StatusPending {
		return errors.NewValidationError().AddField("status", "only pending leave requests can be approved")
	}

	// Deduct from leave balance for all managed leave types
	if leave.IsManagedLeave(leaveRequest.LeaveType) {
		err := s.repository.DeductLeaveBalance(leaveRequest.EmployeeID, leaveRequest.LeaveType, leaveRequest.DaysCount)
		if err != nil {
			return err
		}
	}

	// Deduct salary for paid leaves
	if leave.IsPaidLeave(leaveRequest.LeaveType) {
		emp, err := s.employeeRepository.GetEmployeeByID(leaveRequest.EmployeeID)
		if err != nil {
			return errors.WrapError("failed to get employee for salary deduction", err)
		}

		salaryDeduction := float64(leaveRequest.DaysCount) * 500
		newSalary := emp.Salary - salaryDeduction

		// Update employee salary
		err = s.employeeRepository.UpdateEmployeeSalary(emp.ID, newSalary)
		if err != nil {
			return errors.WrapError("failed to deduct salary for paid leave", err)
		}

		// Add deduction note to leave request
		deductionNote := fmt.Sprintf("Your paid leave for %d days is approved. An amount of 500 per day (total: %.2f) from %s to %s of your leave has been deducted from your salary.",
			leaveRequest.DaysCount, salaryDeduction, leaveRequest.StartDate.Format("2006-01-02"), leaveRequest.EndDate.Format("2006-01-02"))
		
		// Append admin notes if provided
		if notes != "" {
			deductionNote = deductionNote + " Admin notes: " + notes
		}
		
		err = s.repository.UpdateLeaveRequestNotes(leaveRequest.ID, deductionNote)
		if err != nil {
			// Log but don't fail the approval
			errors.LogError("Failed to update leave notes with deduction info", err)
		}
	} else if notes != "" {
		// For non-paid leaves, just add the admin notes
		err := s.repository.UpdateLeaveRequestNotes(leaveRequest.ID, notes)
		if err != nil {
			errors.LogError("Failed to update leave notes", err)
		}
	}

	err = s.repository.UpdateLeaveRequestStatus(id, string(leave.StatusApproved), &approvedByUserID)
	if err != nil {
		return err
	}

	// Queue approval notification asynchronously
	go s.queueLeaveApprovedNotification(leaveRequest, approvedByUserID)

	return nil
}

// queueLeaveApprovedNotification queues email notification for leave approval
func (s *Service) queueLeaveApprovedNotification(leaveReq *leave.LeaveRequest, approvedByUserID int) {
	// Get employee and approver
	emp, err := s.employeeRepository.GetEmployeeByID(leaveReq.EmployeeID)
	if err != nil {
		errors.LogError("Failed to get employee for notification", err)
		return
	}

	approver, err := s.userRepository.GetUserByID(approvedByUserID)
	if err != nil {
		errors.LogError("Failed to get approver for notification", err)
		return
	}

	// Prepare template data
	templateData := notification.TemplateData{
		EmployeeName:   fmt.Sprintf("%s %s", emp.FirstName, emp.LastName),
		EmployeeEmail:  emp.Email,
		EmployeeID:     emp.ID,
		AdminName:      approver.Username,
		AdminEmail:     approver.Email,
		LeaveType:      string(leaveReq.LeaveType),
		StartDate:      leaveReq.StartDate.Format("2006-01-02"),
		EndDate:        leaveReq.EndDate.Format("2006-01-02"),
		TotalDays:      leaveReq.DaysCount,
		TotalDeduction: leaveReq.SalaryDeduction,
		IsPaidLeave:    leave.IsPaidLeave(leaveReq.LeaveType),
	}

	// Send email to employee
	tmpl := notification.GetTemplate(notification.EventLeaveApproved, leave.IsPaidLeave(leaveReq.LeaveType))
	subject, body, err := email.RenderTemplate(tmpl, templateData)
	if err != nil {
		errors.LogError("Failed to render approval notification template", err)
		return
	}

	notif := &notification.Notification{
		LeaveRequestID:  &leaveReq.ID,
		RecipientEmail:  emp.Email,
		RecipientName:   fmt.Sprintf("%s %s", emp.FirstName, emp.LastName),
		EventType:       notification.EventLeaveApproved,
		TemplateName:    tmpl.Name,
		DeliveryChannel: notification.ChannelSMTP,
		Status:          notification.StatusPending,
		Subject:         subject,
		Body:            body,
		MaxRetries:      3,
	}

	_, err = s.notificationRepo.CreateNotification(notif)
	if err != nil {
		errors.LogError("Failed to create approval notification record", err)
		return
	}

	// Enqueue for sending
	err = s.emailQueue.Enqueue(notif)
	if err != nil {
		errors.LogError("Failed to enqueue approval notification", err)
	}
}

// RejectLeave rejects a leave request (admin only)
func (s *Service) RejectLeave(id int, approvedByUserID int, reason string) error {
	leaveRequest, err := s.repository.GetLeaveRequest(id)
	if err != nil {
		return err
	}

	if leaveRequest.Status != leave.StatusPending {
		return errors.NewValidationError().AddField("status", "only pending leave requests can be rejected")
	}

	// Add rejection reason to notes if provided
	if reason != "" {
		err := s.repository.UpdateLeaveRequestNotes(leaveRequest.ID, "Rejection reason: "+reason)
		if err != nil {
			errors.LogError("Failed to update leave notes with rejection reason", err)
		}
	}

	err = s.repository.UpdateLeaveRequestStatus(id, string(leave.StatusRejected), &approvedByUserID)
	if err != nil {
		return err
	}

	// Queue rejection notification asynchronously
	go s.queueLeaveRejectedNotification(leaveRequest, reason)

	return nil
}

// queueLeaveRejectedNotification queues email notification for leave rejection
func (s *Service) queueLeaveRejectedNotification(leaveReq *leave.LeaveRequest, reason string) {
	// Get employee
	emp, err := s.employeeRepository.GetEmployeeByID(leaveReq.EmployeeID)
	if err != nil {
		errors.LogError("Failed to get employee for rejection notification", err)
		return
	}

	// Prepare template data
	templateData := notification.TemplateData{
		EmployeeName:    fmt.Sprintf("%s %s", emp.FirstName, emp.LastName),
		EmployeeEmail:   emp.Email,
		EmployeeID:      emp.ID,
		LeaveType:       string(leaveReq.LeaveType),
		StartDate:       leaveReq.StartDate.Format("2006-01-02"),
		EndDate:         leaveReq.EndDate.Format("2006-01-02"),
		TotalDays:       leaveReq.DaysCount,
		RejectionReason: reason,
	}

	// Send email to employee
	tmpl := notification.GetTemplate(notification.EventLeaveRejected, false)
	subject, body, err := email.RenderTemplate(tmpl, templateData)
	if err != nil {
		errors.LogError("Failed to render rejection notification template", err)
		return
	}

	notif := &notification.Notification{
		LeaveRequestID:  &leaveReq.ID,
		RecipientEmail:  emp.Email,
		RecipientName:   fmt.Sprintf("%s %s", emp.FirstName, emp.LastName),
		EventType:       notification.EventLeaveRejected,
		TemplateName:    tmpl.Name,
		DeliveryChannel: notification.ChannelSMTP,
		Status:          notification.StatusPending,
		Subject:         subject,
		Body:            body,
		MaxRetries:      3,
	}

	_, err = s.notificationRepo.CreateNotification(notif)
	if err != nil {
		errors.LogError("Failed to create rejection notification record", err)
		return
	}

	// Enqueue for sending
	err = s.emailQueue.Enqueue(notif)
	if err != nil {
		errors.LogError("Failed to enqueue rejection notification", err)
	}
}

// GetEmployeeLeaveBalance retrieves a specific leave balance for an employee
func (s *Service) GetEmployeeLeaveBalance(userID int, leaveType leave.LeaveType) (*leave.LeaveBalance, error) {
	emp, err := s.employeeRepository.GetEmployeeByUserID(userID)
	if err != nil {
		return nil, errors.NotFoundError("employee record")
	}

	balance, err := s.repository.GetLeaveBalance(emp.ID, leaveType)
	if err != nil {
		// Auto-initialize leave balances for managed leave types
		if leaveType == leave.TypeAnnual || leaveType == leave.TypeSick || leaveType == leave.TypeCasual ||
			leaveType == leave.TypeMaternity || leaveType == leave.TypePaternity || leaveType == leave.TypeUnpaid || leaveType == leave.TypePersonal {
			// Auto-initialize leave balances for this employee
			if err := s.repository.InitializeLeaveBalances(emp.ID); err != nil {
				errors.LogError("Failed to initialize leave balances", err)
				return nil, errors.WrapError("failed to initialize leave balance", err)
			}
			// Retry getting the balance after initialization
			balance, err := s.repository.GetLeaveBalance(emp.ID, leaveType)
			if err != nil {
				return nil, err
			}
			return balance, nil
		}
		return nil, err
	}
	return balance, nil
}

// GetEmployeeLeaveBalances retrieves all leave balances for an employee
func (s *Service) GetEmployeeLeaveBalances(userID int) ([]leave.LeaveBalance, error) {
	emp, err := s.employeeRepository.GetEmployeeByUserID(userID)
	if err != nil {
		return nil, errors.NotFoundError("employee record")
	}

	return s.repository.GetEmployeeLeaveBalances(emp.ID)
}

// InitializeLeaveBalances initializes default leave balances for a new employee
func (s *Service) InitializeLeaveBalances(employeeID int) error {
	return s.repository.InitializeLeaveBalances(employeeID)
}