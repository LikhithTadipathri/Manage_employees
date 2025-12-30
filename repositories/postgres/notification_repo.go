package postgres

import (
	"database/sql"
	"fmt"
	"time"

	"employee-service/errors"
	"employee-service/models/notification"
)

// NotificationRepository handles notification database operations
type NotificationRepository struct {
	db *sql.DB
}

// NewNotificationRepository creates a new notification repository
func NewNotificationRepository(db *sql.DB) *NotificationRepository {
	return &NotificationRepository{db: db}
}

// CreateNotification creates a new notification record
func (r *NotificationRepository) CreateNotification(n *notification.Notification) (*notification.Notification, error) {
	now := time.Now()
	n.CreatedAt = now
	n.UpdatedAt = now
	n.Status = notification.StatusPending
	n.RetryCount = 0

	errors.LogInfo(fmt.Sprintf("📝 CREATING NOTIFICATION in DB"))
	errors.LogInfo(fmt.Sprintf("   To: %s | Event: %s | Subject: %s", n.RecipientEmail, n.EventType, n.Subject))

	query := `
		INSERT INTO notifications (
			leave_request_id, recipient_email, recipient_name, event_type, 
			template_name, delivery_channel, status, subject, body, 
			retry_count, max_retries, error_message, sent_at, next_retry_at, 
			created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)
		RETURNING id, created_at, updated_at
	`

	err := r.db.QueryRow(
		query,
		n.LeaveRequestID,
		n.RecipientEmail,
		n.RecipientName,
		n.EventType,
		n.TemplateName,
		n.DeliveryChannel,
		n.Status,
		n.Subject,
		n.Body,
		n.RetryCount,
		n.MaxRetries,
		n.ErrorMessage,
		n.SentAt,
		n.NextRetryAt,
		n.CreatedAt,
		n.UpdatedAt,
	).Scan(&n.ID, &n.CreatedAt, &n.UpdatedAt)

	if err != nil {
		errors.LogError("❌ FAILED to create notification in DB", err)
		return nil, errors.WrapError("failed to create notification", err)
	}

	errors.LogInfo(fmt.Sprintf("✅ NOTIFICATION CREATED in DB with ID: %d", n.ID))
	return n, nil
}

// GetNotification retrieves a notification by ID
func (r *NotificationRepository) GetNotification(id int) (*notification.Notification, error) {
	query := `
		SELECT id, leave_request_id, recipient_email, recipient_name, event_type, 
		       template_name, delivery_channel, status, subject, body, 
		       retry_count, max_retries, error_message, sent_at, next_retry_at, 
		       created_at, updated_at
		FROM notifications WHERE id = $1
	`

	n := &notification.Notification{}
	err := r.db.QueryRow(query, id).Scan(
		&n.ID, &n.LeaveRequestID, &n.RecipientEmail, &n.RecipientName, &n.EventType,
		&n.TemplateName, &n.DeliveryChannel, &n.Status, &n.Subject, &n.Body,
		&n.RetryCount, &n.MaxRetries, &n.ErrorMessage, &n.SentAt, &n.NextRetryAt,
		&n.CreatedAt, &n.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.NotFoundError("notification")
		}
		return nil, errors.WrapError("failed to get notification", err)
	}

	return n, nil
}

// GetPendingNotifications retrieves pending notifications for retry
func (r *NotificationRepository) GetPendingNotifications() ([]notification.Notification, error) {
	query := `
		SELECT id, leave_request_id, recipient_email, recipient_name, event_type, 
		       template_name, delivery_channel, status, subject, body, 
		       retry_count, max_retries, error_message, sent_at, next_retry_at, 
		       created_at, updated_at
		FROM notifications 
		WHERE status IN ($1, $2)
		AND (next_retry_at IS NULL OR next_retry_at <= NOW())
		AND retry_count < max_retries
		ORDER BY created_at ASC
		LIMIT 100
	`

	rows, err := r.db.Query(query, notification.StatusPending, notification.StatusRetry)
	if err != nil {
		return nil, errors.WrapError("failed to get pending notifications", err)
	}
	defer rows.Close()

	notifications := []notification.Notification{}
	for rows.Next() {
		n := notification.Notification{}
		err := rows.Scan(
			&n.ID, &n.LeaveRequestID, &n.RecipientEmail, &n.RecipientName, &n.EventType,
			&n.TemplateName, &n.DeliveryChannel, &n.Status, &n.Subject, &n.Body,
			&n.RetryCount, &n.MaxRetries, &n.ErrorMessage, &n.SentAt, &n.NextRetryAt,
			&n.CreatedAt, &n.UpdatedAt,
		)
		if err != nil {
			return nil, errors.WrapError("failed to scan notification", err)
		}
		notifications = append(notifications, n)
	}

	return notifications, nil
}

// UpdateNotificationStatus updates notification status and tracking info
func (r *NotificationRepository) UpdateNotificationStatus(
	id int,
	status notification.NotificationStatus,
	sentAt *time.Time,
	errorMsg *string,
) error {
	now := time.Now()
	query := `
		UPDATE notifications 
		SET status = $1, sent_at = $2, error_message = $3, updated_at = $4
		WHERE id = $5
	`

	_, err := r.db.Exec(query, status, sentAt, errorMsg, now, id)
	if err != nil {
		return errors.WrapError("failed to update notification status", err)
	}

	return nil
}

// UpdateNotificationRetry updates notification retry information
func (r *NotificationRepository) UpdateNotificationRetry(
	id int,
	retryCount int,
	nextRetryAt *time.Time,
	errorMsg *string,
) error {
	now := time.Now()
	query := `
		UPDATE notifications 
		SET retry_count = $1, next_retry_at = $2, error_message = $3, 
		    status = $4, updated_at = $5
		WHERE id = $6
	`

	status := notification.StatusRetry
	_, err := r.db.Exec(query, retryCount, nextRetryAt, errorMsg, status, now, id)
	if err != nil {
		return errors.WrapError("failed to update notification retry info", err)
	}

	return nil
}

// DeleteNotification deletes a notification (soft or hard)
func (r *NotificationRepository) DeleteNotification(id int) error {
	query := `DELETE FROM notifications WHERE id = $1`
	_, err := r.db.Exec(query, id)
	if err != nil {
		return errors.WrapError("failed to delete notification", err)
	}
	return nil
}

// GetNotificationsByEvent retrieves notifications by event type
func (r *NotificationRepository) GetNotificationsByEvent(eventType notification.EventType) ([]notification.Notification, error) {
	query := `
		SELECT id, leave_request_id, recipient_email, recipient_name, event_type, 
		       template_name, delivery_channel, status, subject, body, 
		       retry_count, max_retries, error_message, sent_at, next_retry_at, 
		       created_at, updated_at
		FROM notifications 
		WHERE event_type = $1
		ORDER BY created_at DESC
		LIMIT 100
	`

	rows, err := r.db.Query(query, eventType)
	if err != nil {
		return nil, errors.WrapError("failed to get notifications by event", err)
	}
	defer rows.Close()

	notifications := []notification.Notification{}
	for rows.Next() {
		n := notification.Notification{}
		err := rows.Scan(
			&n.ID, &n.LeaveRequestID, &n.RecipientEmail, &n.RecipientName, &n.EventType,
			&n.TemplateName, &n.DeliveryChannel, &n.Status, &n.Subject, &n.Body,
			&n.RetryCount, &n.MaxRetries, &n.ErrorMessage, &n.SentAt, &n.NextRetryAt,
			&n.CreatedAt, &n.UpdatedAt,
		)
		if err != nil {
			return nil, errors.WrapError("failed to scan notification", err)
		}
		notifications = append(notifications, n)
	}

	return notifications, nil
}

// GetFailedNotifications retrieves notifications that have failed after max retries
func (r *NotificationRepository) GetFailedNotifications() ([]notification.Notification, error) {
	query := `
		SELECT id, leave_request_id, recipient_email, recipient_name, event_type, 
		       template_name, delivery_channel, status, subject, body, 
		       retry_count, max_retries, error_message, sent_at, next_retry_at, 
		       created_at, updated_at
		FROM notifications 
		WHERE status = $1 OR (retry_count >= max_retries AND status != $2)
		ORDER BY created_at DESC
		LIMIT 100
	`

	rows, err := r.db.Query(query, notification.StatusFailed, notification.StatusSent)
	if err != nil {
		return nil, errors.WrapError("failed to get failed notifications", err)
	}
	defer rows.Close()

	notifications := []notification.Notification{}
	for rows.Next() {
		n := notification.Notification{}
		err := rows.Scan(
			&n.ID, &n.LeaveRequestID, &n.RecipientEmail, &n.RecipientName, &n.EventType,
			&n.TemplateName, &n.DeliveryChannel, &n.Status, &n.Subject, &n.Body,
			&n.RetryCount, &n.MaxRetries, &n.ErrorMessage, &n.SentAt, &n.NextRetryAt,
			&n.CreatedAt, &n.UpdatedAt,
		)
		if err != nil {
			return nil, errors.WrapError("failed to scan notification", err)
		}
		notifications = append(notifications, n)
	}

	return notifications, nil
}
