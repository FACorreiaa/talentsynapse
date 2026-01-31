package points

import (
	"time"

	"github.com/google/uuid"
)

// Tier represents user verification tier levels
type Tier string

const (
	// TierBronze is the basic tier (0-99 points)
	TierBronze Tier = "Bronze"
	// TierSilver is the enhanced tier (100-499 points)
	TierSilver Tier = "Silver"
	// TierGold is the premium tier (500+ points)
	TierGold Tier = "Gold"
)

// Tier thresholds
const (
	SilverThreshold = 100
	GoldThreshold   = 500
)

// TierInfo contains tier metadata
type TierInfo struct {
	Name        Tier
	MinPoints   int
	MaxPoints   int // -1 for no max
	Badge       string
	Description string
}

// TierInfoMap contains all tier information
var TierInfoMap = map[Tier]TierInfo{
	TierBronze: {
		Name:        TierBronze,
		MinPoints:   0,
		MaxPoints:   99,
		Badge:       "🥉",
		Description: "Basic verified badge",
	},
	TierSilver: {
		Name:        TierSilver,
		MinPoints:   100,
		MaxPoints:   499,
		Badge:       "🥈",
		Description: "Enhanced visibility in search",
	},
	TierGold: {
		Name:        TierGold,
		MinPoints:   500,
		MaxPoints:   -1,
		Badge:       "🥇",
		Description: "Featured in recommendations",
	},
}

// UserPoints represents a user's points and tier information
type UserPoints struct {
	UserID    uuid.UUID
	Points    int
	Tier      Tier
	UpdatedAt time.Time
}

// TierUpgrade represents a tier upgrade event
type TierUpgrade struct {
	UserID     uuid.UUID
	OldTier    Tier
	NewTier    Tier
	Points     int
	UpgradedAt time.Time
}

// PointsAction represents actions that award points
type PointsAction string

const (
	ActionSessionCompleted PointsAction = "session_completed"
	ActionReviewReceived   PointsAction = "review_received"
	ActionProfileComplete  PointsAction = "profile_complete"
	ActionFirstMatch       PointsAction = "first_match"
	ActionFirstSession     PointsAction = "first_session"
)

// PointsReward contains the reward for each action
var PointsReward = map[PointsAction]int{
	ActionSessionCompleted: 10, // +10 points per completed session
	ActionReviewReceived:   5,  // +5 points per positive review (4-5 stars)
	ActionProfileComplete:  25, // +25 points for completing profile
	ActionFirstMatch:       5,  // +5 points for first mutual match (one-time)
	ActionFirstSession:     10, // +10 points for first session
}

// CalculateTierFromPoints determines the tier based on point total
func CalculateTierFromPoints(points int) Tier {
	switch {
	case points >= GoldThreshold:
		return TierGold
	case points >= SilverThreshold:
		return TierSilver
	default:
		return TierBronze
	}
}

// NextTierThreshold returns the points needed for the next tier
func NextTierThreshold(currentTier Tier) int {
	switch currentTier {
	case TierBronze:
		return SilverThreshold
	case TierSilver:
		return GoldThreshold
	default:
		return -1 // Already at max tier
	}
}

// PointsToNextTier calculates how many points needed to reach next tier
func PointsToNextTier(currentPoints int, currentTier Tier) int {
	threshold := NextTierThreshold(currentTier)
	if threshold == -1 {
		return 0 // Already at max tier
	}
	return threshold - currentPoints
}
