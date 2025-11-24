package pricing

import "strings"

// Tier represents a pricing tier
type Tier string

const (
	TierFree       Tier = "free"
	TierPro        Tier = "pro"
	TierEnterprise Tier = "enterprise"
)

// Limits defines the capabilities for a specific tier
type Limits struct {
	MaxShards              int
	MaxRPS                 int
	AllowStrongConsistency bool
	Name                   string
}

// GetLimits returns the limits for a given tier
func GetLimits(tierStr string) Limits {
	tier := Tier(strings.ToLower(tierStr))
	switch tier {
	case TierPro:
		return Limits{
			MaxShards:              10,
			MaxRPS:                 100,
			AllowStrongConsistency: true,
			Name:                   "Pro",
		}
	case TierEnterprise:
		return Limits{
			MaxShards:              -1, // Unlimited
			MaxRPS:                 -1, // Unlimited
			AllowStrongConsistency: true,
			Name:                   "Enterprise",
		}
	default: // Default to Free
		return Limits{
			MaxShards:              2,
			MaxRPS:                 10,
			AllowStrongConsistency: false,
			Name:                   "Free",
		}
	}
}
