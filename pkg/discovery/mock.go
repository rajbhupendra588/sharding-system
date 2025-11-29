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

// DiscoverApplications returns empty list - no mock data
// In production, this should be replaced with actual Kubernetes discovery
func (m *MockDiscovery) DiscoverApplications(ctx context.Context) ([]DiscoveredApp, error) {
	m.logger.Info("MockDiscovery: returning empty list (no mock data)")
	return []DiscoveredApp{}, nil
}

// UpdateRegisteredApps updates the list of registered application names
func (m *MockDiscovery) UpdateRegisteredApps(names []string) {
	m.registeredApps = make(map[string]bool)
	for _, name := range names {
		m.registeredApps[name] = true
	}
}
