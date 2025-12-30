package user

import (
	"golang.org/x/crypto/bcrypt"

	"employee-service/errors"
	usermodel "employee-service/models/user"
	"employee-service/repositories/postgres"
)

// UserService handles user business logic
type UserService struct {
	repo *postgres.UserRepository
}

// NewUserService creates a new user service
func NewUserService(repo *postgres.UserRepository) *UserService {
	return &UserService{repo: repo}
}

// Register creates a new user with hashed password
func (s *UserService) Register(req *usermodel.CreateUserRequest) (*usermodel.User, error) {
	// Hash password
	hashedPassword, err := s.HashPassword(req.Password)
	if err != nil {
		return nil, errors.WrapError("failed to hash password", err)
	}

	// Create user
	user, err := s.repo.CreateUser(req, hashedPassword)
	if err != nil {
		return nil, err
	}

	return user, nil
}

// Authenticate verifies username and password
func (s *UserService) Authenticate(username, password string) (*usermodel.User, error) {
	// Get user by username
	user, err := s.repo.GetUserByUsername(username)
	if err != nil {
		return nil, errors.UnauthorizedError("invalid credentials")
	}

	// Verify password
	if !s.VerifyPassword(user.PasswordHash, password) {
		return nil, errors.UnauthorizedError("invalid credentials")
	}

	return user, nil
}

// GetUserByID retrieves a user by ID
func (s *UserService) GetUserByID(id int) (*usermodel.User, error) {
	return s.repo.GetUserByID(id)
}

// GetUserByUsername retrieves a user by username
func (s *UserService) GetUserByUsername(username string) (*usermodel.User, error) {
	return s.repo.GetUserByUsername(username)
}

// GetAllUsers retrieves all users (admin only)
func (s *UserService) GetAllUsers() ([]usermodel.User, error) {
	return s.repo.GetAllUsers()
}

// UpdateUserRole updates a user's role (admin only)
func (s *UserService) UpdateUserRole(userID int, role string) error {
	// Validate role
	if role != "admin" && role != "user" {
		return errors.BadRequestError("role must be 'admin' or 'user'")
	}

	return s.repo.UpdateUserRole(userID, role)
}

// DeleteUser deletes a user (admin only)
func (s *UserService) DeleteUser(userID int) error {
	return s.repo.DeleteUser(userID)
}



// HashPassword hashes a password using bcrypt
func (s *UserService) HashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", errors.WrapError("failed to hash password", err)
	}
	return string(hashedPassword), nil
}

// VerifyPassword checks if a password matches a hash
func (s *UserService) VerifyPassword(hash, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
