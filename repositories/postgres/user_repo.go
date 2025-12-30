package postgres

import (
	"database/sql"
	"time"

	"employee-service/errors"
	usermodel "employee-service/models/user"
)

// UserRepository handles user data access
type UserRepository struct {
	db *sql.DB
}

// NewUserRepository creates a new user repository
func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

// GetUserByUsername retrieves a user by username
func (r *UserRepository) GetUserByUsername(username string) (*usermodel.User, error) {
	query := `
		SELECT id, username, email, password_hash, role, created_at, updated_at
		FROM users
		WHERE username = $1
	`

	user := &usermodel.User{}
	err := r.db.QueryRow(query, username).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.PasswordHash,
		&user.Role,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, errors.NotFoundError("user")
	}
	if err != nil {
		return nil, errors.WrapError("failed to get user by username", err)
	}

	return user, nil
}

// GetUserByID retrieves a user by ID
func (r *UserRepository) GetUserByID(id int) (*usermodel.User, error) {
	query := `
		SELECT id, username, email, password_hash, role, created_at, updated_at
		FROM users
		WHERE id = $1
	`

	user := &usermodel.User{}
	err := r.db.QueryRow(query, id).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.PasswordHash,
		&user.Role,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, errors.NotFoundError("user")
	}
	if err != nil {
		return nil, errors.WrapError("failed to get user by id", err)
	}

	return user, nil
}

// CreateUser creates a new user
func (r *UserRepository) CreateUser(req *usermodel.CreateUserRequest, passwordHash string) (*usermodel.User, error) {
	query := `
		INSERT INTO users (username, email, password_hash, role, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, username, email, password_hash, role, created_at, updated_at
	`

	now := time.Now()
	user := &usermodel.User{}
	err := r.db.QueryRow(
		query,
		req.Username,
		req.Email,
		passwordHash,
		req.Role,
		now,
		now,
	).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.PasswordHash,
		&user.Role,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		return nil, errors.WrapError("failed to create user", err)
	}

	return user, nil
}

// UpdateUserRole updates a user's role
func (r *UserRepository) UpdateUserRole(userID int, role string) error {
	query := `
		UPDATE users
		SET role = $1, updated_at = $2
		WHERE id = $3
	`

	result, err := r.db.Exec(query, role, time.Now(), userID)
	if err != nil {
		return errors.WrapError("failed to update user role", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return errors.WrapError("failed to check rows affected", err)
	}

	if rowsAffected == 0 {
		return errors.NotFoundError("user")
	}

	return nil
}

// GetAllUsers retrieves all users
func (r *UserRepository) GetAllUsers() ([]usermodel.User, error) {
	query := `
		SELECT id, username, email, password_hash, role, created_at, updated_at
		FROM users
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, errors.WrapError("failed to get all users", err)
	}
	defer rows.Close()

	var users []usermodel.User
	for rows.Next() {
		var user usermodel.User
		err := rows.Scan(
			&user.ID,
			&user.Username,
			&user.Email,
			&user.PasswordHash,
			&user.Role,
			&user.CreatedAt,
			&user.UpdatedAt,
		)
		if err != nil {
			return nil, errors.WrapError("failed to scan user row", err)
		}
		users = append(users, user)
	}

	if err = rows.Err(); err != nil {
		return nil, errors.WrapError("error iterating user rows", err)
	}

	return users, nil
}

// DeleteUser deletes a user
func (r *UserRepository) DeleteUser(userID int) error {
	query := `DELETE FROM users WHERE id = $1`

	result, err := r.db.Exec(query, userID)
	if err != nil {
		return errors.WrapError("failed to delete user", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return errors.WrapError("failed to check rows affected", err)
	}

	if rowsAffected == 0 {
		return errors.NotFoundError("user")
	}

	return nil
}

// GetUsersByRole retrieves all users with a specific role
func (r *UserRepository) GetUsersByRole(role string) ([]usermodel.User, error) {
	query := `
		SELECT id, username, email, password_hash, role, created_at, updated_at
		FROM users
		WHERE role = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(query, role)
	if err != nil {
		return nil, errors.WrapError("failed to get users by role", err)
	}
	defer rows.Close()

	var users []usermodel.User
	for rows.Next() {
		var user usermodel.User
		err := rows.Scan(
			&user.ID,
			&user.Username,
			&user.Email,
			&user.PasswordHash,
			&user.Role,
			&user.CreatedAt,
			&user.UpdatedAt,
		)
		if err != nil {
			return nil, errors.WrapError("failed to scan user row", err)
		}
		users = append(users, user)
	}

	if err = rows.Err(); err != nil {
		return nil, errors.WrapError("error iterating user rows", err)
	}

	return users, nil
}