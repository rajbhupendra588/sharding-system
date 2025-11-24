package discovery

import "context"

// DiscoveryService defines the interface for application discovery
type DiscoveryService interface {
	// DiscoverApplications discovers applications and databases
	DiscoverApplications(ctx context.Context) ([]DiscoveredApp, error)

	// UpdateRegisteredApps updates the list of registered application names
	UpdateRegisteredApps(names []string)
}
