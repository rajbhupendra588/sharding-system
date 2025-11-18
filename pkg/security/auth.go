package security

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Claims represents JWT claims
type Claims struct {
	Username string   `json:"username"`
	Roles    []string `json:"roles"`
	jwt.RegisteredClaims
}

// AuthManager handles authentication and authorization
type AuthManager struct {
	jwtSecret []byte
	rbac      *RBAC
}

// NewAuthManager creates a new auth manager
func NewAuthManager(jwtSecret string) *AuthManager {
	return &AuthManager{
		jwtSecret: []byte(jwtSecret),
		rbac:      NewRBAC(),
	}
}

// GenerateToken generates a JWT token for a user
func (a *AuthManager) GenerateToken(username string, roles []string) (string, error) {
	claims := &Claims{
		Username: username,
		Roles:    roles,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(a.jwtSecret)
}

// ValidateToken validates a JWT token
func (a *AuthManager) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return a.jwtSecret, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}

// Authorize checks if a user has permission for an action
func (a *AuthManager) Authorize(claims *Claims, resource string, action string) bool {
	return a.rbac.IsAllowed(claims.Roles, resource, action)
}

