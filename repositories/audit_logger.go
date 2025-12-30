package repositories

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
)

// AuditLog represents an audit log entry for database changes
type AuditLog struct {
	ID        int64       `json:"id"`
	Table     string      `json:"table"`
	Operation string      `json:"operation"` // INSERT, UPDATE, DELETE
	RecordID  int64       `json:"record_id"`
	UserID    int64       `json:"user_id"`
	OldValues json.RawMessage `json:"old_values"` // JSON of previous values
	NewValues json.RawMessage `json:"new_values"` // JSON of new values
	IPAddress string      `json:"ip_address"`
	UserAgent string      `json:"user_agent,omitempty"`
	Reason    string      `json:"reason,omitempty"`
	CreatedAt time.Time   `json:"created_at"`
}

// AuditLogger handles audit logging for database operations
type AuditLogger struct {
	db     *sql.DB
	logger *logrus.Logger
}

// NewAuditLogger creates a new audit logger
func NewAuditLogger(db *sql.DB, logger *logrus.Logger) *AuditLogger {
	return &AuditLogger{
		db:     db,
		logger: logger,
	}
}

// LogInsert logs an INSERT operation
func (al *AuditLogger) LogInsert(ctx context.Context, table string, recordID int64, values map[string]interface{}, userID int64, ipAddress string) error {
	newValues, _ := json.Marshal(values)

	query := `
		INSERT INTO audit_logs (table_name, operation, record_id, user_id, old_values, new_values, ip_address, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	_, err := al.db.ExecContext(ctx, query,
		table,
		"INSERT",
		recordID,
		userID,
		nil,
		newValues,
		ipAddress,
		time.Now(),
	)

	if err != nil {
		al.logger.WithError(err).Errorf("Failed to log audit: INSERT on %s", table)
		return fmt.Errorf("failed to log audit insert: %w", err)
	}

	al.logger.WithFields(logrus.Fields{
		"table":       table,
		"operation":   "INSERT",
		"record_id":   recordID,
		"user_id":     userID,
		"correlation": ctx.Value("correlation_id"),
	}).Debug("Audit log: INSERT recorded")

	return nil
}

// LogUpdate logs an UPDATE operation
func (al *AuditLogger) LogUpdate(ctx context.Context, table string, recordID int64, oldValues, newValues map[string]interface{}, userID int64, ipAddress, reason string) error {
	oldValuesJSON, _ := json.Marshal(oldValues)
	newValuesJSON, _ := json.Marshal(newValues)

	query := `
		INSERT INTO audit_logs (table_name, operation, record_id, user_id, old_values, new_values, ip_address, reason, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`

	_, err := al.db.ExecContext(ctx, query,
		table,
		"UPDATE",
		recordID,
		userID,
		oldValuesJSON,
		newValuesJSON,
		ipAddress,
		reason,
		time.Now(),
	)

	if err != nil {
		al.logger.WithError(err).Errorf("Failed to log audit: UPDATE on %s", table)
		return fmt.Errorf("failed to log audit update: %w", err)
	}

	al.logger.WithFields(logrus.Fields{
		"table":       table,
		"operation":   "UPDATE",
		"record_id":   recordID,
		"user_id":     userID,
		"correlation": ctx.Value("correlation_id"),
	}).Debug("Audit log: UPDATE recorded")

	return nil
}

// LogDelete logs a DELETE operation
func (al *AuditLogger) LogDelete(ctx context.Context, table string, recordID int64, values map[string]interface{}, userID int64, ipAddress, reason string) error {
	oldValues, _ := json.Marshal(values)

	query := `
		INSERT INTO audit_logs (table_name, operation, record_id, user_id, old_values, new_values, ip_address, reason, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`

	_, err := al.db.ExecContext(ctx, query,
		table,
		"DELETE",
		recordID,
		userID,
		oldValues,
		nil,
		ipAddress,
		reason,
		time.Now(),
	)

	if err != nil {
		al.logger.WithError(err).Errorf("Failed to log audit: DELETE on %s", table)
		return fmt.Errorf("failed to log audit delete: %w", err)
	}

	al.logger.WithFields(logrus.Fields{
		"table":       table,
		"operation":   "DELETE",
		"record_id":   recordID,
		"user_id":     userID,
		"correlation": ctx.Value("correlation_id"),
	}).Debug("Audit log: DELETE recorded")

	return nil
}

// GetAuditLogs retrieves audit logs with filtering
func (al *AuditLogger) GetAuditLogs(ctx context.Context, table string, recordID int64, limit int, offset int) ([]AuditLog, error) {
	query := `
		SELECT id, table_name, operation, record_id, user_id, old_values, new_values, ip_address, user_agent, reason, created_at
		FROM audit_logs
		WHERE table_name = $1 AND record_id = $2
		ORDER BY created_at DESC
		LIMIT $3 OFFSET $4
	`

	rows, err := al.db.QueryContext(ctx, query, table, recordID, limit, offset)
	if err != nil {
		al.logger.WithError(err).Error("Failed to retrieve audit logs")
		return nil, fmt.Errorf("failed to query audit logs: %w", err)
	}
	defer rows.Close()

	var logs []AuditLog
	for rows.Next() {
		var log AuditLog
		err := rows.Scan(
			&log.ID,
			&log.Table,
			&log.Operation,
			&log.RecordID,
			&log.UserID,
			&log.OldValues,
			&log.NewValues,
			&log.IPAddress,
			&log.UserAgent,
			&log.Reason,
			&log.CreatedAt,
		)
		if err != nil {
			al.logger.WithError(err).Error("Failed to scan audit log row")
			return nil, fmt.Errorf("failed to scan audit log: %w", err)
		}
		logs = append(logs, log)
	}

	return logs, rows.Err()
}

// GetUserAuditActivity gets all audit logs for a specific user
func (al *AuditLogger) GetUserAuditActivity(ctx context.Context, userID int64, limit int, offset int) ([]AuditLog, error) {
	query := `
		SELECT id, table_name, operation, record_id, user_id, old_values, new_values, ip_address, user_agent, reason, created_at
		FROM audit_logs
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := al.db.QueryContext(ctx, query, userID, limit, offset)
	if err != nil {
		al.logger.WithError(err).Error("Failed to retrieve user audit activity")
		return nil, fmt.Errorf("failed to query user audit activity: %w", err)
	}
	defer rows.Close()

	var logs []AuditLog
	for rows.Next() {
		var log AuditLog
		err := rows.Scan(
			&log.ID,
			&log.Table,
			&log.Operation,
			&log.RecordID,
			&log.UserID,
			&log.OldValues,
			&log.NewValues,
			&log.IPAddress,
			&log.UserAgent,
			&log.Reason,
			&log.CreatedAt,
		)
		if err != nil {
			al.logger.WithError(err).Error("Failed to scan audit log row")
			return nil, fmt.Errorf("failed to scan audit log: %w", err)
		}
		logs = append(logs, log)
	}

	return logs, rows.Err()
}

// PurgeOldAuditLogs removes audit logs older than specified days
func (al *AuditLogger) PurgeOldAuditLogs(ctx context.Context, daysOld int) (int64, error) {
	query := `
		DELETE FROM audit_logs
		WHERE created_at < NOW() - INTERVAL '1 day' * $1
	`

	result, err := al.db.ExecContext(ctx, query, daysOld)
	if err != nil {
		al.logger.WithError(err).Errorf("Failed to purge audit logs older than %d days", daysOld)
		return 0, fmt.Errorf("failed to purge audit logs: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	al.logger.Infof("Purged %d audit log entries older than %d days", rowsAffected, daysOld)

	return rowsAffected, nil
}

// CompareChanges provides a readable summary of what changed
func (al *AuditLogger) CompareChanges(oldValues, newValues json.RawMessage) (map[string]interface{}, error) {
	var oldMap, newMap map[string]interface{}

	if oldValues != nil {
		if err := json.Unmarshal(oldValues, &oldMap); err != nil {
			return nil, err
		}
	}

	if newValues != nil {
		if err := json.Unmarshal(newValues, &newMap); err != nil {
			return nil, err
		}
	}

	changes := make(map[string]interface{})

	// Find changed fields
	for key := range newMap {
		oldVal := oldMap[key]
		newVal := newMap[key]

		if oldVal != newVal {
			changes[key] = map[string]interface{}{
				"old": oldVal,
				"new": newVal,
			}
		}
	}

	return changes, nil
}

// CreateAuditLogsTable creates the audit logs table if it doesn't exist
func CreateAuditLogsTable(db *sql.DB) error {
	query := `
		CREATE TABLE IF NOT EXISTS audit_logs (
			id SERIAL PRIMARY KEY,
			table_name VARCHAR(255) NOT NULL,
			operation VARCHAR(50) NOT NULL CHECK (operation IN ('INSERT', 'UPDATE', 'DELETE')),
			record_id BIGINT NOT NULL,
			user_id BIGINT NOT NULL,
			old_values JSONB,
			new_values JSONB,
			ip_address VARCHAR(45),
			user_agent TEXT,
			reason TEXT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE SET NULL,
			INDEX idx_table_record (table_name, record_id),
			INDEX idx_user_id (user_id),
			INDEX idx_created_at (created_at)
		);
	`

	_, err := db.Exec(query)
	return err
}
