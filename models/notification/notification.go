package notification

import (
	"time"
)

// EventType represents the type of notification event
type EventType string

const (
	EventLeaveApplied   EventType = "LEAVE_APPLIED"
	EventLeaveApproved  EventType = "LEAVE_APPROVED"
	EventLeaveRejected  EventType = "LEAVE_REJECTED"
	EventLeaveCancelled EventType = "LEAVE_CANCELLED"
	EventLowBalance     EventType = "LOW_BALANCE"
	EventApprovalReminder EventType = "APPROVAL_REMINDER"
)

// DeliveryChannel represents how the notification is sent
type DeliveryChannel string

const (
	ChannelSMTP DeliveryChannel = "SMTP"
)

// NotificationStatus represents the status of a notification
type NotificationStatus string

const (
	StatusPending NotificationStatus = "PENDING"
	StatusSent    NotificationStatus = "SENT"
	StatusFailed  NotificationStatus = "FAILED"
	StatusRetry   NotificationStatus = "RETRY"
)

// Notification represents a notification record
type Notification struct {
	ID              int                 `json:"id"`
	LeaveRequestID  *int                `json:"leave_request_id"` // nullable for non-leave events
	RecipientEmail  string              `json:"recipient_email"`
	RecipientName   string              `json:"recipient_name"`
	EventType       EventType           `json:"event_type"`
	TemplateName    string              `json:"template_name"`
	DeliveryChannel DeliveryChannel     `json:"delivery_channel"`
	Status          NotificationStatus  `json:"status"`
	Subject         string              `json:"subject"`
	Body            string              `json:"body"`
	RetryCount      int                 `json:"retry_count"`
	MaxRetries      int                 `json:"max_retries"`
	ErrorMessage    *string             `json:"error_message"`
	SentAt          *time.Time          `json:"sent_at"`
	NextRetryAt     *time.Time          `json:"next_retry_at"`
	CreatedAt       time.Time           `json:"created_at"`
	UpdatedAt       time.Time           `json:"updated_at"`
}

// TemplateData contains all variables for template rendering
type TemplateData struct {
	EmployeeName     string
	EmployeeEmail    string
	EmployeeID       int
	AdminName        string
	AdminEmail       string
	LeaveType        string
	StartDate        string
	EndDate          string
	TotalDays        int
	Reason           string
	RejectionReason  string
	TotalDeduction   float64
	IsPaidLeave      bool
	CurrentBalance   int
	LowBalanceThreshold int
}

// EmailTemplate represents an email template
type EmailTemplate struct {
	Name    string
	Subject string
	Body    string
}

// GetTemplate returns the email template for a given event
func GetTemplate(eventType EventType, isPaidLeave bool) EmailTemplate {
	switch eventType {
	case EventLeaveApplied:
		return EmailTemplate{
			Name:    "leave_applied_employee",
			Subject: "Leave Request Submitted – Pending Approval",
			Body: `Hello {{.employee_name}},

This is an automated notification sent via our SMTP mail service.

Your leave request has been submitted successfully and is currently under review.

Leave Details:
- Leave Type: {{.leave_type}}
- Start Date: {{.start_date}}
- End Date: {{.end_date}}
- Total Days: {{.total_days}}
- Reason: {{.reason}}

Current Status: PENDING

Your request has been forwarded to the Admin/Manager for approval.
You will receive another SMTP email once a decision is made.

Regards,
HR Management System
(no-reply – This mailbox is not monitored)`,
		}

	case EventLeaveApproved:
		if isPaidLeave {
			return EmailTemplate{
				Name:    "leave_approved_employee",
				Subject: "Leave Approved",
				Body: `Hello {{.employee_name}},

This is an automated SMTP notification.

Your leave request has been APPROVED.

Leave Details:
- Leave Type: {{.leave_type}}
- Duration: {{.start_date}} to {{.end_date}}
- Total Days: {{.total_days}}

Salary Deduction Policy:
- Paid Leave Deduction: ₹500 per day
- Total Deduction: ₹{{.total_deduction}}

This amount will be deducted from your salary.

Approved By: {{.admin_name}}

Regards,
HR Management System
(no-reply)`,
			}
		} else {
			// For unpaid leaves (MATERNITY, PATERNITY, UNPAID, etc)
			return EmailTemplate{
				Name:    "leave_approved_employee",
				Subject: "Leave Approved",
				Body: `Hello {{.employee_name}},

This is an automated SMTP notification.

Your leave request has been APPROVED.

Leave Details:
- Leave Type: {{.leave_type}}
- Duration: {{.start_date}} to {{.end_date}}
- Total Days: {{.total_days}}

Salary Deduction Policy:
- No salary deduction applies to this leave type
- You will receive your full salary during this period

Approved By: {{.admin_name}}

Regards,
HR Management System
(no-reply)`,
			}
		}

	case EventLeaveRejected:
		return EmailTemplate{
			Name:    "leave_rejected_employee",
			Subject: "Leave Request Rejected",
			Body: `Hello {{.employee_name}},

This is an automated SMTP notification.

Your leave request has been REJECTED.

Leave Type: {{.leave_type}}
Requested Days: {{.total_days}}

Reason for Rejection:
{{.rejection_reason}}

For further clarification, please contact HR/Admin.

Regards,
HR Management System
(no-reply)`,
		}

	case EventLeaveCancelled:
		return EmailTemplate{
			Name:    "leave_cancelled_employee",
			Subject: "Leave Request Cancelled",
			Body: `Hello {{.employee_name}},

This is an automated SMTP notification.

Your leave request has been CANCELLED.

Leave Details:
- Leave Type: {{.leave_type}}
- Duration: {{.start_date}} to {{.end_date}}
- Total Days: {{.total_days}}

If this was unexpected, please contact HR/Admin for assistance.

Regards,
HR Management System
(no-reply)`,
		}

	case EventLowBalance:
		return EmailTemplate{
			Name:    "low_balance_warning",
			Subject: "Low Leave Balance Warning",
			Body: `Hello {{.employee_name}},

This is an automated SMTP notification.

Your {{.leave_type}} leave balance is running low.

Current Balance: {{.current_balance}} days

Please plan your leaves accordingly. Contact HR for more information.

Regards,
HR Management System
(no-reply)`,
		}

	case EventApprovalReminder:
		return EmailTemplate{
			Name:    "approval_reminder_admin",
			Subject: "Action Required: Pending Leave Approvals",
			Body: `Hello {{.admin_name}},

This is an automated SMTP notification.

You have pending leave requests that require your attention:

Employee: {{.employee_name}}
Leave Type: {{.leave_type}}
Duration: {{.start_date}} to {{.end_date}} ({{.total_days}} days)

Please login to the admin portal to approve or reject these requests.

HR Management System
(no-reply)`,
		}

	default:
		return EmailTemplate{
			Name:    "default",
			Subject: "Notification",
			Body:    "{{message}}",
		}
	}
}

// GetAdminTemplate returns the admin notification template
func GetAdminTemplate(eventType EventType) EmailTemplate {
	if eventType == EventLeaveApplied {
		return EmailTemplate{
			Name:    "leave_applied_admin",
			Subject: "Action Required: New Leave Request Submitted",
			Body: `Hello Admin,

This is an automated SMTP notification.

A new leave request has been submitted and requires your action.

Employee Name: {{.employee_name}}
Employee ID: {{.employee_id}}
Leave Type: {{.leave_type}}
Duration: {{.start_date}} to {{.end_date}} ({{.total_days}} days)
Reason: {{.reason}}

Current Status: PENDING

Please login to the admin portal to approve or reject this request.

HR Management System`,
		}
	}
	return EmailTemplate{}
}
