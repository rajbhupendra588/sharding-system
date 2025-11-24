package discovery

import (
	"context"

	"go.uber.org/zap"
)

// MockDiscovery provides mock discovery data for local development
type MockDiscovery struct {
	logger         *zap.Logger
	registeredApps map[string]bool
}

// NewMockDiscovery creates a new mock discovery service
func NewMockDiscovery(logger *zap.Logger) *MockDiscovery {
	return &MockDiscovery{
		logger:         logger,
		registeredApps: make(map[string]bool),
	}
}

// DiscoverApplications returns sample applications for local development
func (m *MockDiscovery) DiscoverApplications(ctx context.Context) ([]DiscoveredApp, error) {
	m.logger.Info("Returning mock discovery data")

	apps := []DiscoveredApp{
		{
			Namespace:    "default",
			Name:         "payment-service",
			Type:         "deployment",
			DatabaseName: "payment_db",
			DatabaseHost: "postgres-payment",
			DatabasePort: "5432",
			DatabaseUser: "postgres",
			Labels:       map[string]string{"app": "payment-service", "env": "dev"},
			Annotations:  map[string]string{"sharding.keyPrefix": "pay:"},
			IsRegistered: m.registeredApps["payment-service"],
		},
		{
			Namespace:    "default",
			Name:         "user-service",
			Type:         "deployment",
			DatabaseName: "users_db",
			DatabaseHost: "postgres-users",
			DatabasePort: "5432",
			DatabaseUser: "postgres",
			Labels:       map[string]string{"app": "user-service", "env": "dev"},
			Annotations:  map[string]string{"sharding.keyPrefix": "usr:"},
			IsRegistered: m.registeredApps["user-service"],
		},
		{
			Namespace:    "analytics",
			Name:         "analytics-worker",
			Type:         "statefulset",
			DatabaseName: "analytics_db",
			DatabaseHost: "postgres-analytics",
			DatabasePort: "5432",
			DatabaseUser: "postgres",
			Labels:       map[string]string{"app": "analytics", "component": "worker"},
			Annotations:  map[string]string{},
			IsRegistered: m.registeredApps["analytics-worker"],
		},
	}

	return apps, nil
}

// UpdateRegisteredApps updates the list of registered application names
func (m *MockDiscovery) UpdateRegisteredApps(names []string) {
	m.registeredApps = make(map[string]bool)
	for _, name := range names {
		m.registeredApps[name] = true
	}
}
