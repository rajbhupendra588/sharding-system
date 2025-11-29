package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sharding-system/pkg/security"
	"go.uber.org/zap/zaptest"
)

// MockUserStore is a mock implementation of UserStore
type MockUserStore struct {
	users          map[string]*security.User
	adminCount     int
	setupRequired  bool
	authenticateFn func(username, password string) (*security.User, error)
}

func NewMockUserStore() *MockUserStore {
	return &MockUserStore{
		users: make(map[string]*security.User),
	}
}

func (m *MockUserStore) GetUser(username string) (*security.User, error) {
	if user, ok := m.users[username]; ok {
		return user, nil
	}
	return nil, errors.New("user not found")
}

func (m *MockUserStore) Authenticate(username, password string) (*security.User, error) {
	if m.authenticateFn != nil {
		return m.authenticateFn(username, password)
	}
	user, ok := m.users[username]
	if !ok {
		return nil, errors.New("invalid credentials")
	}
	// In a real mock we might check password, but for now just return the user if found
	return user, nil
}

func (m *MockUserStore) AddUser(user *security.User) error {
	if _, ok := m.users[user.Username]; ok {
		return errors.New("user already exists")
	}
	m.users[user.Username] = user
	return nil
}

func (m *MockUserStore) GetAdminCount() (int, error) {
	return m.adminCount, nil
}

func (m *MockUserStore) IsSetupRequired() (bool, error) {
	return m.setupRequired, nil
}

func TestAuthHandler_Login(t *testing.T) {
	logger := zaptest.NewLogger(t)
	authManager := security.NewAuthManager("test-secret")

	tests := []struct {
		name           string
		setupMock      func(*MockUserStore)
		requestBody    map[string]interface{}
		expectedStatus int
		expectedError  string
	}{
		{
			name: "Successful Login",
			setupMock: func(m *MockUserStore) {
				m.authenticateFn = func(username, password string) (*security.User, error) {
					return &security.User{
						Username: "admin",
						Roles:    []string{"admin"},
					}, nil
				}
			},
			requestBody: map[string]interface{}{
				"username": "admin",
				"password": "password123",
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "Invalid Credentials",
			setupMock: func(m *MockUserStore) {
				m.authenticateFn = func(username, password string) (*security.User, error) {
					return nil, errors.New("invalid credentials")
				}
			},
			requestBody: map[string]interface{}{
				"username": "admin",
				"password": "wrongpassword",
			},
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "Invalid credentials",
		},
		{
			name:      "Missing Username",
			setupMock: func(m *MockUserStore) {},
			requestBody: map[string]interface{}{
				"password": "password123",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Username is required",
		},
		{
			name:      "Missing Password",
			setupMock: func(m *MockUserStore) {},
			requestBody: map[string]interface{}{
				"username": "admin",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Password is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStore := NewMockUserStore()
			if tt.setupMock != nil {
				tt.setupMock(mockStore)
			}

			handler, err := NewAuthHandler(authManager, "", logger)
			if err != nil {
				t.Fatalf("Failed to create handler: %v", err)
			}
			// Inject mock store
			handler.userStore = mockStore

			body, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest("POST", "/api/v1/auth/login", bytes.NewBuffer(body))
			w := httptest.NewRecorder()

			handler.Login(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.expectedError != "" {
				var resp map[string]interface{}
				json.NewDecoder(w.Body).Decode(&resp)
				if errMap, ok := resp["error"].(map[string]interface{}); ok {
					if msg, ok := errMap["message"].(string); ok {
						if msg != tt.expectedError {
							t.Errorf("Expected error message %q, got %q", tt.expectedError, msg)
						}
					}
				} else {
					t.Errorf("Expected error response, got %v", resp)
				}
			}
		})
	}
}

func TestAuthHandler_Setup(t *testing.T) {
	logger := zaptest.NewLogger(t)
	authManager := security.NewAuthManager("test-secret")

	tests := []struct {
		name           string
		setupMock      func(*MockUserStore)
		requestBody    map[string]interface{}
		expectedStatus int
		expectedError  string
	}{
		{
			name: "Successful Setup",
			setupMock: func(m *MockUserStore) {
				m.setupRequired = true
			},
			requestBody: map[string]interface{}{
				"username": "admin",
				"password": "strongpassword123",
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name: "Setup Not Required",
			setupMock: func(m *MockUserStore) {
				m.setupRequired = false
			},
			requestBody: map[string]interface{}{
				"username": "admin",
				"password": "strongpassword123",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "System already initialized. Setup can only be performed when no users exist.",
		},
		{
			name: "Weak Password",
			setupMock: func(m *MockUserStore) {
				m.setupRequired = true
			},
			requestBody: map[string]interface{}{
				"username": "admin",
				"password": "123",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Password must be at least 8 characters",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStore := NewMockUserStore()
			if tt.setupMock != nil {
				tt.setupMock(mockStore)
			}

			handler, err := NewAuthHandler(authManager, "", logger)
			if err != nil {
				t.Fatalf("Failed to create handler: %v", err)
			}
			handler.userStore = mockStore

			body, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest("POST", "/api/v1/auth/setup", bytes.NewBuffer(body))
			w := httptest.NewRecorder()

			handler.Setup(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.expectedError != "" {
				var resp map[string]interface{}
				json.NewDecoder(w.Body).Decode(&resp)
				if errMap, ok := resp["error"].(map[string]interface{}); ok {
					if msg, ok := errMap["message"].(string); ok {
						if msg != tt.expectedError {
							t.Errorf("Expected error message %q, got %q", tt.expectedError, msg)
						}
					}
				}
			}
		})
	}
}
