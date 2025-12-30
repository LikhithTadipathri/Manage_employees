package repositories

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/sirupsen/logrus"
)

// Transaction represents a database transaction with error handling
type Transaction struct {
	tx     *sql.Tx
	logger *logrus.Logger
	ctx    context.Context
}

// TransactionManager handles transaction lifecycle
type TransactionManager struct {
	db     *sql.DB
	logger *logrus.Logger
}

// NewTransactionManager creates a new transaction manager
func NewTransactionManager(db *sql.DB, logger *logrus.Logger) *TransactionManager {
	return &TransactionManager{
		db:     db,
		logger: logger,
	}
}

// Begin starts a new transaction
func (tm *TransactionManager) Begin(ctx context.Context) (*Transaction, error) {
	tx, err := tm.db.BeginTx(ctx, &sql.TxOptions{
		Isolation: sql.LevelReadCommitted, // Default isolation level
	})
	if err != nil {
		tm.logger.WithError(err).Error("Failed to begin transaction")
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}

	tm.logger.WithField("correlation_id", ctx.Value("correlation_id")).Debug("Transaction started")

	return &Transaction{
		tx:     tx,
		logger: tm.logger,
		ctx:    ctx,
	}, nil
}

// BeginWithIsolation starts a transaction with specific isolation level
func (tm *TransactionManager) BeginWithIsolation(ctx context.Context, isolation sql.IsolationLevel) (*Transaction, error) {
	tx, err := tm.db.BeginTx(ctx, &sql.TxOptions{
		Isolation: isolation,
	})
	if err != nil {
		tm.logger.WithError(err).Error("Failed to begin transaction with isolation level")
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}

	tm.logger.WithFields(logrus.Fields{
		"correlation_id":   ctx.Value("correlation_id"),
		"isolation_level":  isolation.String(),
	}).Debug("Transaction started with isolation level")

	return &Transaction{
		tx:     tx,
		logger: tm.logger,
		ctx:    ctx,
	}, nil
}

// Commit commits the transaction
func (t *Transaction) Commit() error {
	if err := t.tx.Commit(); err != nil {
		t.logger.WithError(err).Error("Failed to commit transaction")
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	t.logger.WithField("correlation_id", t.ctx.Value("correlation_id")).Debug("Transaction committed successfully")
	return nil
}

// Rollback rolls back the transaction
func (t *Transaction) Rollback() error {
	if err := t.tx.Rollback(); err != nil && err != sql.ErrTxDone {
		t.logger.WithError(err).Error("Failed to rollback transaction")
		return fmt.Errorf("failed to rollback transaction: %w", err)
	}

	t.logger.WithField("correlation_id", t.ctx.Value("correlation_id")).Debug("Transaction rolled back")
	return nil
}

// RollbackOnError rolls back transaction on error, logs error, and returns original error
func (t *Transaction) RollbackOnError(err error) error {
	if err != nil {
		rollbackErr := t.Rollback()
		if rollbackErr != nil {
			t.logger.WithFields(logrus.Fields{
				"original_error": err,
				"rollback_error": rollbackErr,
			}).Error("Error during rollback after operation failure")
		}
		return err
	}
	return nil
}

// Exec executes a query without returning rows within a transaction
func (t *Transaction) Exec(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	result, err := t.tx.ExecContext(ctx, query, args...)
	if err != nil {
		t.logger.WithFields(logrus.Fields{
			"correlation_id": ctx.Value("correlation_id"),
			"query":          query,
			"error":          err,
		}).Error("Query execution failed in transaction")
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}
	return result, nil
}

// QueryRow queries a single row within a transaction
func (t *Transaction) QueryRow(ctx context.Context, query string, args ...interface{}) *sql.Row {
	return t.tx.QueryRowContext(ctx, query, args...)
}

// Query queries multiple rows within a transaction
func (t *Transaction) Query(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	rows, err := t.tx.QueryContext(ctx, query, args...)
	if err != nil {
		t.logger.WithFields(logrus.Fields{
			"correlation_id": ctx.Value("correlation_id"),
			"query":          query,
			"error":          err,
		}).Error("Query failed in transaction")
		return nil, fmt.Errorf("failed to query rows: %w", err)
	}
	return rows, nil
}

// GetTx returns the underlying *sql.Tx for custom operations
func (t *Transaction) GetTx() *sql.Tx {
	return t.tx
}

// GetContext returns the transaction context
func (t *Transaction) GetContext() context.Context {
	return t.ctx
}

// WithTx is a helper function to execute operations within a transaction with automatic rollback on error
func (tm *TransactionManager) WithTx(ctx context.Context, fn func(*Transaction) error) error {
	tx, err := tm.Begin(ctx)
	if err != nil {
		return err
	}

	defer func() {
		if r := recover(); r != nil {
			_ = tx.Rollback()
			panic(r)
		}
	}()

	if err := fn(tx); err != nil {
		_ = tx.Rollback()
		return err
	}

	return tx.Commit()
}

// WithTxIsolation is a helper function to execute operations within a transaction with specific isolation level
func (tm *TransactionManager) WithTxIsolation(ctx context.Context, isolation sql.IsolationLevel, fn func(*Transaction) error) error {
	tx, err := tm.BeginWithIsolation(ctx, isolation)
	if err != nil {
		return err
	}

	defer func() {
		if r := recover(); r != nil {
			_ = tx.Rollback()
			panic(r)
		}
	}()

	if err := fn(tx); err != nil {
		_ = tx.Rollback()
		return err
	}

	return tx.Commit()
}
