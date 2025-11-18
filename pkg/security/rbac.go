package security

import (
	"strings"
)

// RBAC implements role-based access control
type RBAC struct {
	permissions map[string]map[string][]string // role -> resource -> actions
}

// NewRBAC creates a new RBAC instance with default permissions
func NewRBAC() *RBAC {
	rbac := &RBAC{
		permissions: make(map[string]map[string][]string),
	}

	// Define default roles and permissions
	rbac.AddPermission("admin", "*", []string{"*"}) // Admin can do everything
	rbac.AddPermission("operator", "shards", []string{"read", "create", "update"})
	rbac.AddPermission("operator", "reshard", []string{"read", "create"})
	rbac.AddPermission("viewer", "shards", []string{"read"})
	rbac.AddPermission("viewer", "reshard", []string{"read"})

	return rbac
}

// AddPermission adds a permission for a role
func (r *RBAC) AddPermission(role, resource string, actions []string) {
	if r.permissions[role] == nil {
		r.permissions[role] = make(map[string][]string)
	}
	r.permissions[role][resource] = actions
}

// IsAllowed checks if roles have permission for an action on a resource
func (r *RBAC) IsAllowed(roles []string, resource string, action string) bool {
	for _, role := range roles {
		if r.hasPermission(role, resource, action) {
			return true
		}
	}
	return false
}

// hasPermission checks if a role has permission
func (r *RBAC) hasPermission(role, resource, action string) bool {
	rolePerms, exists := r.permissions[role]
	if !exists {
		return false
	}

	// Check wildcard resource
	if actions, ok := rolePerms["*"]; ok {
		if contains(actions, "*") || contains(actions, action) {
			return true
		}
	}

	// Check specific resource
	if actions, ok := rolePerms[resource]; ok {
		return contains(actions, "*") || contains(actions, action)
	}

	return false
}

// contains checks if a slice contains a string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if strings.EqualFold(s, item) {
			return true
		}
	}
	return false
}

