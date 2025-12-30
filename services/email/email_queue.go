package email

import (
	"fmt"
	"sync"
	"time"

	"employee-service/errors"
	"employee-service/models/notification"
	"employee-service/repositories/postgres"
)

// EmailQueue manages the email sending queue
type EmailQueue struct {
	notificationRepo *postgres.NotificationRepository
	emailService     *EmailService
	queue            chan *notification.Notification
	mu               sync.Mutex
	running          bool
	wg               sync.WaitGroup
}

// NewEmailQueue creates a new email queue
func NewEmailQueue(notificationRepo *postgres.NotificationRepository, emailService *EmailService) *EmailQueue {
	return &EmailQueue{
		notificationRepo: notificationRepo,
		emailService:     emailService,
		queue:            make(chan *notification.Notification, 1000), // Buffer size for 1000 emails
		running:          false,
	}
}

// Start begins processing emails from the queue
func (eq *EmailQueue) Start(numWorkers int) error {
	eq.mu.Lock()
	if eq.running {
		eq.mu.Unlock()
		return fmt.Errorf("email queue already running")
	}
	eq.running = true
	eq.mu.Unlock()

	// Start worker goroutines
	for i := 0; i < numWorkers; i++ {
		eq.wg.Add(1)
		go eq.worker(i)
	}

	// Start retry scheduler
	eq.wg.Add(1)
	go eq.retryScheduler()

	errors.LogInfo(fmt.Sprintf("Email queue started with %d workers", numWorkers))
	return nil
}

// Stop gracefully stops the email queue
func (eq *EmailQueue) Stop() error {
	eq.mu.Lock()
	if !eq.running {
		eq.mu.Unlock()
		return fmt.Errorf("email queue not running")
	}
	eq.running = false
	eq.mu.Unlock()

	close(eq.queue)
	eq.wg.Wait()

	errors.LogInfo("Email queue stopped")
	return nil
}

// IsRunning returns whether the email queue is currently running
func (eq *EmailQueue) IsRunning() bool {
	eq.mu.Lock()
	defer eq.mu.Unlock()
	return eq.running
}

// Enqueue adds a notification to the queue
func (eq *EmailQueue) Enqueue(n *notification.Notification) error {
	eq.mu.Lock()
	if !eq.running {
		eq.mu.Unlock()
		err := fmt.Errorf("email queue not running")
		errors.LogError(fmt.Sprintf("Failed to enqueue notification %d", n.ID), err)
		return err
	}
	eq.mu.Unlock()

	errors.LogInfo(fmt.Sprintf("📧 ENQUEUE: Notification %d | Type: %s | To: %s | Subject: %s", n.ID, n.EventType, n.RecipientEmail, n.Subject))

	select {
	case eq.queue <- n:
		errors.LogInfo(fmt.Sprintf("✅ ENQUEUE SUCCESS: Notification %d added to queue (Depth: %d/%d)", n.ID, len(eq.queue), cap(eq.queue)))
		return nil
	default:
		err := fmt.Errorf("email queue is full")
		errors.LogError(fmt.Sprintf("ENQUEUE FAILED: Notification %d - queue is full", n.ID), err)
		return err
	}
}

// worker processes emails from the queue
func (eq *EmailQueue) worker(id int) {
	errors.LogInfo(fmt.Sprintf("🚀 WORKER %d STARTED", id))
	defer eq.wg.Done()

	for notification := range eq.queue {
		errors.LogInfo(fmt.Sprintf("🔄 WORKER %d: Processing notification %d", id, notification.ID))
		eq.processNotification(notification)
	}

	errors.LogInfo(fmt.Sprintf("⛔ WORKER %d STOPPED", id))
}

// processNotification sends an email and handles retries
func (eq *EmailQueue) processNotification(n *notification.Notification) {
	errors.LogInfo(fmt.Sprintf("\n═══════════════════════════════════════════════════════"))
	errors.LogInfo(fmt.Sprintf("📧 PROCESS NOTIFICATION %d", n.ID))
	errors.LogInfo(fmt.Sprintf("  Event Type: %s", n.EventType))
	errors.LogInfo(fmt.Sprintf("  Recipient: %s (%s)", n.RecipientName, n.RecipientEmail))
	errors.LogInfo(fmt.Sprintf("  Subject: %s", n.Subject))
	errors.LogInfo(fmt.Sprintf("  Status: %s | Retries: %d/%d", n.Status, n.RetryCount, n.MaxRetries))
	errors.LogInfo(fmt.Sprintf("═══════════════════════════════════════════════════════\n"))

	// Send email
	errors.LogInfo(fmt.Sprintf("📤 ATTEMPTING TO SEND EMAIL to %s", n.RecipientEmail))
	err := eq.emailService.SendEmail(n.RecipientEmail, n.Subject, n.Body)

	if err == nil {
		// Email sent successfully
		now := time.Now()
		errors.LogInfo(fmt.Sprintf("✅ EMAIL SENT SUCCESSFULLY for notification %d", n.ID))
		errors.LogInfo(fmt.Sprintf("   Updating database status to SENT at %s", now.Format(time.RFC3339)))
		err = eq.notificationRepo.UpdateNotificationStatus(
			n.ID,
			notification.StatusSent,
			&now,
			nil,
		)
		if err != nil {
			errors.LogError(fmt.Sprintf("❌ FAILED to update notification %d status in DB", n.ID), err)
		} else {
			errors.LogInfo(fmt.Sprintf("✅ DATABASE UPDATED: Notification %d marked as SENT", n.ID))
		}
	} else {
		// Email send failed - handle retry
		errors.LogError(fmt.Sprintf("❌ SEND EMAIL FAILED for notification %d (Attempt %d/%d)", n.ID, n.RetryCount+1, n.MaxRetries), err)

		newRetryCount := n.RetryCount + 1
		if newRetryCount >= n.MaxRetries {
			// Max retries reached
			errMsg := fmt.Sprintf("Failed after %d retries: %s", newRetryCount, err.Error())
			errors.LogInfo(fmt.Sprintf("⛔ MAX RETRIES REACHED for notification %d", n.ID))
			errors.LogInfo(fmt.Sprintf("   Marking as FAILED in database..."))
			err = eq.notificationRepo.UpdateNotificationStatus(
				n.ID,
				notification.StatusFailed,
				nil,
				&errMsg,
			)
			if err != nil {
				errors.LogError(fmt.Sprintf("❌ FAILED to update notification %d as FAILED in DB", n.ID), err)
			} else {
				errors.LogInfo(fmt.Sprintf("✅ Notification %d marked as FAILED in database", n.ID))
			}
			errors.LogError(fmt.Sprintf("❌ FINAL FAILURE: Notification %d failed after %d attempts", n.ID, newRetryCount), err)
		} else {
			// Schedule retry
			nextRetryDuration := GetRetryDuration(newRetryCount)
			nextRetryAt := time.Now().Add(nextRetryDuration)
			errMsg := fmt.Sprintf("Retry %d of %d: %s", newRetryCount, n.MaxRetries, err.Error())

			errors.LogInfo(fmt.Sprintf("⏳ SCHEDULING RETRY %d/%d for notification %d", newRetryCount, n.MaxRetries, n.ID))
			errors.LogInfo(fmt.Sprintf("   Next retry at: %s (in %s)", nextRetryAt.Format(time.RFC3339), nextRetryDuration.String()))
			err = eq.notificationRepo.UpdateNotificationRetry(
				n.ID,
				newRetryCount,
				&nextRetryAt,
				&errMsg,
			)
			if err != nil {
				errors.LogError(fmt.Sprintf("❌ FAILED to update retry info for notification %d in DB", n.ID), err)
			} else {
				errors.LogInfo(fmt.Sprintf("✅ Notification %d retry info saved to DB", n.ID))
			}
		}
	}
	errors.LogInfo(fmt.Sprintf("═══════════════════════════════════════════════════════\n"))
}

// retryScheduler periodically checks for notifications that need retry
func (eq *EmailQueue) retryScheduler() {
	defer eq.wg.Done()

	errors.LogInfo("🔄 RETRY SCHEDULER STARTED (checks every 2 minutes)")

	// Check every 2 minutes for notifications to retry
	ticker := time.NewTicker(2 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		eq.mu.Lock()
		if !eq.running {
			eq.mu.Unlock()
			break
		}
		eq.mu.Unlock()

		errors.LogInfo("🔍 RETRY SCHEDULER: Checking for pending notifications...")

		// Get pending notifications for retry
		notifications, err := eq.notificationRepo.GetPendingNotifications()
		if err != nil {
			errors.LogError("❌ RETRY SCHEDULER: Failed to get pending notifications", err)
			continue
		}

		if len(notifications) > 0 {
			errors.LogInfo(fmt.Sprintf("📋 RETRY SCHEDULER: Found %d pending notifications to retry", len(notifications)))
		} else {
			errors.LogInfo("✅ RETRY SCHEDULER: No pending notifications (all caught up!)")
		}

		// Enqueue them for processing
		for _, n := range notifications {
			notifCopy := n // Create a copy for the closure
			errors.LogInfo(fmt.Sprintf("   → Re-enqueuing notification %d (attempt %d/%d)", n.ID, n.RetryCount+1, n.MaxRetries))
			err := eq.Enqueue(&notifCopy)
			if err != nil {
				errors.LogError(fmt.Sprintf("❌ Failed to enqueue notification %d for retry", n.ID), err)
			} else {
				errors.LogInfo(fmt.Sprintf("✅ Successfully re-enqueued notification %d", n.ID))
			}
		}
	}

	errors.LogInfo("⛔ RETRY SCHEDULER STOPPED")
}

// GetQueueStats returns statistics about the email queue
func (eq *EmailQueue) GetQueueStats() map[string]interface{} {
	return map[string]interface{}{
		"running":  eq.running,
		"queued":   len(eq.queue),
		"capacity": cap(eq.queue),
	}
}
