package database

// DatabaseTemplate defines a pre-configured template for database creation
type DatabaseTemplate struct {
	Name            string   `json:"name"`
	DisplayName     string   `json:"display_name"`
	Description     string   `json:"description"`
	ShardCount      int      `json:"shard_count"`
	ReplicasPerShard int     `json:"replicas_per_shard"`
	VNodeCount      int      `json:"vnode_count"`
	AutoScale       bool     `json:"auto_scale"`
	EstimatedCost   string   `json:"estimated_cost"` // Monthly estimate
}

// PredefinedTemplates contains templates for different use cases
var PredefinedTemplates = map[string]DatabaseTemplate{
	"starter": {
		Name:            "starter",
		DisplayName:     "Starter",
		Description:     "Perfect for small applications and development. 2 shards with 1 replica each.",
		ShardCount:      2,
		ReplicasPerShard: 1,
		VNodeCount:      256,
		AutoScale:       true,
		EstimatedCost:   "$99/month",
	},
	"production": {
		Name:            "production",
		DisplayName:     "Production",
		Description:     "For production workloads. 4 shards with 2 replicas each for high availability.",
		ShardCount:      4,
		ReplicasPerShard: 2,
		VNodeCount:      512,
		AutoScale:       true,
		EstimatedCost:   "$299/month",
	},
	"enterprise": {
		Name:            "enterprise",
		DisplayName:     "Enterprise",
		Description:     "For large-scale enterprise applications. 8 shards with 2 replicas each, multi-region ready.",
		ShardCount:      8,
		ReplicasPerShard: 2,
		VNodeCount:      1024,
		AutoScale:       true,
		EstimatedCost:   "Custom pricing",
	},
}

// GetTemplate returns a template by name, or starter if not found
func GetTemplate(name string) DatabaseTemplate {
	if template, ok := PredefinedTemplates[name]; ok {
		return template
	}
	return PredefinedTemplates["starter"]
}

// ListTemplates returns all available templates
func ListTemplates() []DatabaseTemplate {
	templates := make([]DatabaseTemplate, 0, len(PredefinedTemplates))
	for _, template := range PredefinedTemplates {
		templates = append(templates, template)
	}
	return templates
}

