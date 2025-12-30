package email

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"net/mail"
	"net/smtp"
	"strings"
	"text/template"
	"time"

	"employee-service/errors"
	"employee-service/models/notification"
)

// SMTPConfig holds SMTP configuration
type SMTPConfig struct {
	Host     string
	Port     int
	Username string
	Password string
	FromName string
	FromAddr string
}

// EmailService handles email sending via SMTP
type EmailService struct {
	config SMTPConfig
}

// NewEmailService creates a new email service
func NewEmailService(host string, port int, username, password, fromAddr, fromName string) *EmailService {
	return &EmailService{
		config: SMTPConfig{
			Host:     host,
			Port:     port,
			Username: username,
			Password: password,
			FromAddr: fromAddr,
			FromName: fromName,
		},
	}
}

// SendEmail sends an email with retry support
func (s *EmailService) SendEmail(to, subject, body string) error {
	// Validate email address
	if !isValidEmail(to) {
		return fmt.Errorf("invalid recipient email address: %s", to)
	}

	// Sanitize body to prevent injection
	body = sanitizeEmailBody(body)

	// Create email message
	from := mail.Address{
		Name:    s.config.FromName,
		Address: s.config.FromAddr,
	}

	headers := make(map[string]string)
	headers["From"] = from.String()
	headers["To"] = to
	headers["Subject"] = subject
	headers["MIME-Version"] = "1.0"
	headers["Content-Type"] = "text/plain; charset=\"utf-8\""
	headers["Content-Transfer-Encoding"] = "base64"

	message := ""
	for k, v := range headers {
		message += fmt.Sprintf("%s: %s\r\n", k, v)
	}
	message += "\r\n" + body

	// Connect to SMTP server
	addr := fmt.Sprintf("%s:%d", s.config.Host, s.config.Port)
	
	errors.LogInfo(fmt.Sprintf("📧 SMTP: Connecting to %s (Port %d)", s.config.Host, s.config.Port))
	
	var conn *smtp.Client
	var err error
	
	// Handle both port 465 (implicit TLS) and port 587 (STARTTLS)
	if s.config.Port == 465 {
		// Port 465: Implicit TLS (SSL from the start)
		errors.LogInfo("🔒 Using implicit TLS (port 465)")
		tlsConfig := &tls.Config{
			ServerName: s.config.Host,
			InsecureSkipVerify: false,
		}
		tlsConn, err := tls.Dial("tcp", addr, tlsConfig)
		if err != nil {
			return errors.WrapError("failed to connect to SMTP server with TLS", err)
		}
		conn, err = smtp.NewClient(tlsConn, s.config.Host)
		if err != nil {
			tlsConn.Close()
			return errors.WrapError("failed to create SMTP client", err)
		}
	} else {
		// Port 587: STARTTLS
		errors.LogInfo("🔐 Using STARTTLS (port 587)")
		conn, err = smtp.Dial(addr)
		if err != nil {
			return errors.WrapError("failed to connect to SMTP server", err)
		}
		
		// Start TLS
		tlsConfig := &tls.Config{
			ServerName: s.config.Host,
			InsecureSkipVerify: false,
		}
		err = conn.StartTLS(tlsConfig)
		if err != nil {
			conn.Close()
			return errors.WrapError("failed to start TLS", err)
		}
	}
	defer conn.Close()
	
	// Authenticate
	auth := smtp.PlainAuth("", s.config.Username, s.config.Password, s.config.Host)
	errors.LogInfo(fmt.Sprintf("🔑 SMTP: Authenticating as %s", s.config.Username))
	err = conn.Auth(auth)
	if err != nil {
		return errors.WrapError("SMTP authentication failed", err)
	}
	errors.LogInfo("✅ SMTP: Authentication successful")

	// Send email
	err = conn.Mail(s.config.FromAddr)
	if err != nil {
		return errors.WrapError("failed to set SMTP sender", err)
	}

	err = conn.Rcpt(to)
	if err != nil {
		return errors.WrapError("failed to set SMTP recipient", err)
	}

	wc, err := conn.Data()
	if err != nil {
		return errors.WrapError("failed to open SMTP data writer", err)
	}

	_, err = wc.Write([]byte(message))
	if err != nil {
		return errors.WrapError("failed to write email data", err)
	}

	err = wc.Close()
	if err != nil {
		return errors.WrapError("failed to close SMTP data writer", err)
	}

	err = conn.Quit()
	if err != nil {
		return errors.WrapError("failed to quit SMTP connection", err)
	}

	errors.LogInfo(fmt.Sprintf("Email sent successfully to %s with subject: %s", to, subject))
	return nil
}

// RenderTemplate renders an email template with data
func RenderTemplate(tmpl notification.EmailTemplate, data notification.TemplateData) (string, string, error) {
	// Create a map from struct for easier template access
	dataMap := map[string]interface{}{
		"employee_name":        data.EmployeeName,
		"employee_email":       data.EmployeeEmail,
		"employee_id":          data.EmployeeID,
		"admin_name":           data.AdminName,
		"admin_email":          data.AdminEmail,
		"leave_type":           data.LeaveType,
		"start_date":           data.StartDate,
		"end_date":             data.EndDate,
		"total_days":           data.TotalDays,
		"reason":               data.Reason,
		"rejection_reason":     data.RejectionReason,
		"total_deduction":      data.TotalDeduction,
		"is_paid_leave":        data.IsPaidLeave,
		"current_balance":      data.CurrentBalance,
		"low_balance_threshold": data.LowBalanceThreshold,
	}

	// Render subject
	subjectTmpl, err := template.New("subject").Parse(tmpl.Subject)
	if err != nil {
		return "", "", errors.WrapError("failed to parse email subject template", err)
	}

	var subjectBuf bytes.Buffer
	err = subjectTmpl.Execute(&subjectBuf, dataMap)
	if err != nil {
		return "", "", errors.WrapError("failed to render email subject", err)
	}

	// Render body
	bodyTmpl, err := template.New("body").Parse(tmpl.Body)
	if err != nil {
		return "", "", errors.WrapError("failed to parse email body template", err)
	}

	var bodyBuf bytes.Buffer
	err = bodyTmpl.Execute(&bodyBuf, dataMap)
	if err != nil {
		return "", "", errors.WrapError("failed to render email body", err)
	}

	return subjectBuf.String(), bodyBuf.String(), nil
}

// isValidEmail validates an email address
func isValidEmail(email string) bool {
	_, err := mail.ParseAddress(email)
	return err == nil
}

// sanitizeEmailBody prevents email injection
func sanitizeEmailBody(body string) string {
	// Remove line breaks at the beginning of body
	body = strings.TrimSpace(body)
	
	// Replace CRLF injection attempts
	body = strings.ReplaceAll(body, "\r\n\r\n", "\n")
	
	return body
}

// GetRetryDuration returns the duration to wait before next retry
func GetRetryDuration(retryCount int) time.Duration {
	switch retryCount {
	case 0:
		return 5 * time.Minute
	case 1:
		return 15 * time.Minute
	case 2:
		return 1 * time.Hour
	default:
		return 24 * time.Hour
	}
}
