package points

import (
	"testing"
)

func TestCalculateTierFromPoints(t *testing.T) {
	tests := []struct {
		name     string
		points   int
		expected Tier
	}{
		{"Zero points", 0, TierBronze},
		{"Low bronze", 50, TierBronze},
		{"High bronze", 99, TierBronze},
		{"Silver threshold", 100, TierSilver},
		{"Mid silver", 300, TierSilver},
		{"High silver", 499, TierSilver},
		{"Gold threshold", 500, TierGold},
		{"High gold", 1000, TierGold},
		{"Very high gold", 9999, TierGold},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CalculateTierFromPoints(tt.points)
			if result != tt.expected {
				t.Errorf("CalculateTierFromPoints(%d) = %s, want %s", tt.points, result, tt.expected)
			}
		})
	}
}

func TestNextTierThreshold(t *testing.T) {
	tests := []struct {
		name     string
		tier     Tier
		expected int
	}{
		{"Bronze to Silver", TierBronze, SilverThreshold},
		{"Silver to Gold", TierSilver, GoldThreshold},
		{"Gold max", TierGold, -1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NextTierThreshold(tt.tier)
			if result != tt.expected {
				t.Errorf("NextTierThreshold(%s) = %d, want %d", tt.tier, result, tt.expected)
			}
		})
	}
}

func TestPointsToNextTier(t *testing.T) {
	tests := []struct {
		name          string
		currentPoints int
		currentTier   Tier
		expected      int
	}{
		{"Bronze needs 100 to silver", 0, TierBronze, 100},
		{"Bronze needs 50 to silver", 50, TierBronze, 50},
		{"Bronze needs 1 to silver", 99, TierBronze, 1},
		{"Silver needs 400 to gold", 100, TierSilver, 400},
		{"Silver needs 100 to gold", 400, TierSilver, 100},
		{"Silver needs 1 to gold", 499, TierSilver, 1},
		{"Gold at max", 500, TierGold, 0},
		{"Gold way past max", 1000, TierGold, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := PointsToNextTier(tt.currentPoints, tt.currentTier)
			if result != tt.expected {
				t.Errorf("PointsToNextTier(%d, %s) = %d, want %d", tt.currentPoints, tt.currentTier, result, tt.expected)
			}
		})
	}
}

func TestGetTierBadge(t *testing.T) {
	tests := []struct {
		tier     Tier
		expected string
	}{
		{TierBronze, "🥉"},
		{TierSilver, "🥈"},
		{TierGold, "🥇"},
		{Tier("Unknown"), "🥉"}, // Default
	}

	for _, tt := range tests {
		t.Run(string(tt.tier), func(t *testing.T) {
			result := GetTierBadge(tt.tier)
			if result != tt.expected {
				t.Errorf("GetTierBadge(%s) = %s, want %s", tt.tier, result, tt.expected)
			}
		})
	}
}

func TestGetTierDescription(t *testing.T) {
	tests := []struct {
		tier     Tier
		expected string
	}{
		{TierBronze, "Basic verified badge"},
		{TierSilver, "Enhanced visibility in search"},
		{TierGold, "Featured in recommendations"},
	}

	for _, tt := range tests {
		t.Run(string(tt.tier), func(t *testing.T) {
			result := GetTierDescription(tt.tier)
			if result != tt.expected {
				t.Errorf("GetTierDescription(%s) = %s, want %s", tt.tier, result, tt.expected)
			}
		})
	}
}

func TestPointsReward(t *testing.T) {
	// Verify all actions have rewards defined with correct values
	expectedRewards := map[PointsAction]int{
		ActionSessionCompleted: 10, // +10 per completed session
		ActionReviewReceived:   5,  // +5 per positive review
		ActionProfileComplete:  25, // +25 for profile completion
		ActionFirstMatch:       5,  // +5 for first match
		ActionFirstSession:     10, // +10 for first session
	}

	for action, expectedPoints := range expectedRewards {
		t.Run(string(action), func(t *testing.T) {
			reward, ok := PointsReward[action]
			if !ok {
				t.Errorf("PointsReward missing action: %s", action)
			}
			if reward != expectedPoints {
				t.Errorf("PointsReward[%s] = %d, want %d", action, reward, expectedPoints)
			}
		})
	}
}

func TestTierInfoMap(t *testing.T) {
	// Verify all tiers have info defined
	tiers := []Tier{TierBronze, TierSilver, TierGold}

	for _, tier := range tiers {
		t.Run(string(tier), func(t *testing.T) {
			info, ok := TierInfoMap[tier]
			if !ok {
				t.Errorf("TierInfoMap missing tier: %s", tier)
			}
			if info.Badge == "" {
				t.Errorf("TierInfoMap[%s].Badge is empty", tier)
			}
			if info.Description == "" {
				t.Errorf("TierInfoMap[%s].Description is empty", tier)
			}
		})
	}

	// Verify tier thresholds are correct
	if TierInfoMap[TierBronze].MinPoints != 0 {
		t.Errorf("Bronze MinPoints = %d, want 0", TierInfoMap[TierBronze].MinPoints)
	}
	if TierInfoMap[TierSilver].MinPoints != 100 {
		t.Errorf("Silver MinPoints = %d, want 100", TierInfoMap[TierSilver].MinPoints)
	}
	if TierInfoMap[TierGold].MinPoints != 500 {
		t.Errorf("Gold MinPoints = %d, want 500", TierInfoMap[TierGold].MinPoints)
	}
}

func TestTierUpgradeEdgeCases(t *testing.T) {
	// Test edge cases for tier transitions
	tests := []struct {
		name          string
		oldPoints     int
		newPoints     int
		shouldUpgrade bool
		oldTier       Tier
		newTier       Tier
	}{
		{"Stay in bronze", 50, 80, false, TierBronze, TierBronze},
		{"Bronze to silver exact", 99, 100, true, TierBronze, TierSilver},
		{"Bronze to silver over", 90, 110, true, TierBronze, TierSilver},
		{"Stay in silver", 200, 300, false, TierSilver, TierSilver},
		{"Silver to gold exact", 499, 500, true, TierSilver, TierGold},
		{"Silver to gold over", 450, 550, true, TierSilver, TierGold},
		{"Bronze to gold skip", 50, 600, true, TierBronze, TierGold},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			oldTier := CalculateTierFromPoints(tt.oldPoints)
			newTier := CalculateTierFromPoints(tt.newPoints)

			if oldTier != tt.oldTier {
				t.Errorf("oldTier = %s, want %s", oldTier, tt.oldTier)
			}
			if newTier != tt.newTier {
				t.Errorf("newTier = %s, want %s", newTier, tt.newTier)
			}

			upgraded := oldTier != newTier
			if upgraded != tt.shouldUpgrade {
				t.Errorf("upgrade = %v, want %v", upgraded, tt.shouldUpgrade)
			}
		})
	}
}
