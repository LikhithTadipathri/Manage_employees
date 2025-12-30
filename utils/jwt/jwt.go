package jwt

import (
	"fmt"
	"time"

	jwtlib "github.com/golang-jwt/jwt/v5"

	"employee-service/config"
	"employee-service/errors"
	usermodel "employee-service/models/user"
)

// JWTManager handles JWT token generation and verification
type JWTManager struct {
	secret           string
	expirationHours  int
}

// NewJWTManager creates a new JWT manager
func NewJWTManager(cfg *config.JWTConfig) *JWTManager {
	return &JWTManager{
		secret:          cfg.Secret,
		expirationHours: cfg.ExpirationHours,
	}
}

// GenerateToken generates a JWT token for a user
func (m *JWTManager) GenerateToken(user *usermodel.User) (string, int64, error) {
	expirationTime := time.Now().Add(time.Duration(m.expirationHours) * time.Hour)

	claims := jwtlib.MapClaims{
		"user_id":  user.ID,
		"username": user.Username,
		"email":    user.Email,
		"role":     user.Role,
		"exp":      expirationTime.Unix(),
		"iat":      time.Now().Unix(),
	}

	token := jwtlib.NewWithClaims(jwtlib.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(m.secret))
	if err != nil {
		return "", 0, errors.WrapError("failed to sign token", err)
	}

	return tokenString, expirationTime.Unix(), nil
}

// VerifyToken verifies a JWT token and returns claims
func (m *JWTManager) VerifyToken(tokenString string) (jwtlib.MapClaims, error) {
	token, err := jwtlib.ParseWithClaims(tokenString, jwtlib.MapClaims{}, func(token *jwtlib.Token) (interface{}, error) {
		// Verify signing method
		if _, ok := token.Method.(*jwtlib.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(m.secret), nil
	})

	if err != nil {
		return nil, errors.WrapError("failed to parse token", err)
	}

	claims, ok := token.Claims.(jwtlib.MapClaims)
	if !ok || !token.Valid {
		return nil, errors.UnauthorizedError("token is invalid or expired")
	}

	return claims, nil
}

// ExtractClaims extracts claims from a token string
func (m *JWTManager) ExtractClaims(tokenString string) (*usermodel.JWTClaims, error) {
	mapClaims, err := m.VerifyToken(tokenString)
	if err != nil {
		return nil, err
	}

	// Safe type assertions with error handling
	userID, ok := mapClaims["user_id"].(float64)
	if !ok {
		return nil, fmt.Errorf("invalid user_id claim in token")
	}

	username, ok := mapClaims["username"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid username claim in token")
	}

	email, ok := mapClaims["email"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid email claim in token")
	}

	role, ok := mapClaims["role"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid role claim in token")
	}

	claims := &usermodel.JWTClaims{
		UserID:   int(userID),
		Username: username,
		Email:    email,
		Role:     role,
	}

	return claims, nil
}
