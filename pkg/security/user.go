package security

import (
	"errors"
	"sync"
)

// User represents a system user
type User struct {
	Username     string
	PasswordHash string
	Roles        []string
	Active       bool
}

// UserStore manages users
type UserStore struct {
	users map[string]*User
	mu    sync.RWMutex
}

// NewUserStore creates a new user store
func NewUserStore() *UserStore {
	store := &UserStore{
		users: make(map[string]*User),
	}
	
	// Initialize with default users (hashed passwords)
	// In production, these should be loaded from a database
	// Passwords: admin123, operator123, viewer123
	defaultUsers := []*User{
		{
			Username:     "admin",
			PasswordHash: "$2a$10$LtlhX7.r1Rf9Fl7XjR9VKeaZvwU7PJK6tlWF5rXdxe1fg55wurAnW", // admin123
			Roles:        []string{"admin"},
			Active:       true,
		},
		{
			Username:     "operator",
			PasswordHash: "$2a$10$oDZulSnupJh0OdVrJImYNO/HrxjmUx8QA.ICMSA/Pdskkdwd68.bu", // operator123
			Roles:        []string{"operator"},
			Active:       true,
		},
		{
			Username:     "viewer",
			PasswordHash: "$2a$10$QyJBIVEeUVYYYdRELwpeLe7E5y2vvDIWdIMlIoXOjQCYWj2ozssDG", // viewer123
			Roles:        []string{"viewer"},
			Active:       true,
		},
	}
	
	for _, user := range defaultUsers {
		store.users[user.Username] = user
	}
	
	return store
}

// GetUser retrieves a user by username
func (s *UserStore) GetUser(username string) (*User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	user, exists := s.users[username]
	if !exists {
		return nil, errors.New("user not found")
	}
	
	if !user.Active {
		return nil, errors.New("user is inactive")
	}
	
	return user, nil
}

// Authenticate verifies user credentials
func (s *UserStore) Authenticate(username, password string) (*User, error) {
	user, err := s.GetUser(username)
	if err != nil {
		return nil, err
	}
	
	if err := VerifyPassword(user.PasswordHash, password); err != nil {
		return nil, errors.New("invalid password")
	}
	
	return user, nil
}

// AddUser adds a new user (for future user management API)
func (s *UserStore) AddUser(user *User) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	if _, exists := s.users[user.Username]; exists {
		return errors.New("user already exists")
	}
	
	// Check admin limit (max 2 admins)
	isAdmin := false
	for _, role := range user.Roles {
		if role == "admin" {
			isAdmin = true
			break
		}
	}
	
	if isAdmin {
		adminCount := 0
		for _, u := range s.users {
			for _, role := range u.Roles {
				if role == "admin" {
					adminCount++
					break
				}
			}
		}
		if adminCount >= 2 {
			return errors.New("maximum of 2 admin users allowed")
		}
	}
	
	s.users[user.Username] = user
	return nil
}

// GetAdminCount returns the number of active admin users
func (s *UserStore) GetAdminCount() (int, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	count := 0
	for _, user := range s.users {
		if !user.Active {
			continue
		}
		for _, role := range user.Roles {
			if role == "admin" {
				count++
				break
			}
		}
	}
	return count, nil
}

// IsSetupRequired checks if the system needs initial setup (no users exist)
func (s *UserStore) IsSetupRequired() (bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.users) == 0, nil
}

