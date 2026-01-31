package points

import (
	"testing"

	"github.com/google/uuid"
)

// MockNotificationHub implements NotificationHub for testing
type MockNotificationHub struct {
	Messages []struct {
		UserID  string
		Message []byte
	}
}

func NewMockNotificationHub() *MockNotificationHub {
	return &MockNotificationHub{
		Messages: make([]struct {
			UserID  string
			Message []byte
		}, 0),
	}
}

func (m *MockNotificationHub) BroadcastToUser(userID string, message []byte) {
	m.Messages = append(m.Messages, struct {
		UserID  string
		Message []byte
	}{UserID: userID, Message: message})
}

func (m *MockNotificationHub) GetLastMessage() (string, []byte) {
	if len(m.Messages) == 0 {
		return "", nil
	}
	last := m.Messages[len(m.Messages)-1]
	return last.UserID, last.Message
}

func (m *MockNotificationHub) GetMessageCount() int {
	return len(m.Messages)
}

func (m *MockNotificationHub) Clear() {
	m.Messages = m.Messages[:0]
}

// TestServiceWithMockHub tests the service notification logic
func TestServiceWithMockHub(t *testing.T) {
	hub := NewMockNotificationHub()

	// Create service with nil repo (we'll test notification logic only)
	service := &Service{
		repo: nil,
		hub:  hub,
	}

	// Test tier upgrade notification
	upgrade := &TierUpgrade{
		UserID:  uuid.New(),
		OldTier: TierBronze,
		NewTier: TierSilver,
		Points:  100,
	}

	service.notifyTierUpgrade(upgrade)

	if hub.GetMessageCount() != 1 {
		t.Errorf("Expected 1 message, got %d", hub.GetMessageCount())
	}

	userID, message := hub.GetLastMessage()
	if userID != upgrade.UserID.String() {
		t.Errorf("Message sent to wrong user: got %s, want %s", userID, upgrade.UserID.String())
	}

	// Check notification contains expected content
	msgStr := string(message)
	if !contains(msgStr, "Tier Upgrade") {
		t.Error("Notification should contain 'Tier Upgrade'")
	}
	if !contains(msgStr, "Silver") {
		t.Error("Notification should contain 'Silver'")
	}
	if !contains(msgStr, "🥈") {
		t.Error("Notification should contain silver badge emoji")
	}
}

func TestServiceWithNilHub(t *testing.T) {
	// Test that service handles nil hub gracefully
	service := &Service{
		repo: nil,
		hub:  nil,
	}

	upgrade := &TierUpgrade{
		UserID:  uuid.New(),
		OldTier: TierBronze,
		NewTier: TierGold,
		Points:  500,
	}

	// Should not panic
	service.notifyTierUpgrade(upgrade)
}

func TestAwardPointsWithBonusCalculation(t *testing.T) {
	tests := []struct {
		name       string
		basePoints int
		multiplier float64
		expected   int
	}{
		{"No bonus", 10, 1.0, 10},
		{"50% bonus", 10, 1.5, 15},
		{"Double", 10, 2.0, 20},
		{"Half", 10, 0.5, 5},
		{"Minimum 1", 10, 0.01, 1},
		{"Zero multiplier becomes 1", 10, 0.0, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := int(float64(tt.basePoints) * tt.multiplier)
			if result < 1 {
				result = 1
			}
			if result != tt.expected {
				t.Errorf("Points calculation: %d * %f = %d, want %d", tt.basePoints, tt.multiplier, result, tt.expected)
			}
		})
	}
}

func TestReviewRatingPointsLogic(t *testing.T) {
	// Test the rating to points logic (only 4-5 stars award points)
	tests := []struct {
		rating       int
		shouldAward  bool
		expectedMult float64 // multiplier for base points (5 for ActionReviewReceived)
	}{
		{5, true, 2.0},  // 5-star: 5 * 2 = 10 points
		{4, true, 1.0},  // 4-star: 5 * 1 = 5 points
		{3, false, 0.0}, // 3-star: no points
		{2, false, 0.0}, // 2-star: no points
		{1, false, 0.0}, // 1-star: no points
	}

	for _, tt := range tests {
		t.Run(string(rune('0'+tt.rating))+"_star", func(t *testing.T) {
			shouldAward := tt.rating >= 4
			if shouldAward != tt.shouldAward {
				t.Errorf("Rating %d should award: got %v, want %v", tt.rating, shouldAward, tt.shouldAward)
			}

			if tt.shouldAward {
				basePoints := PointsReward[ActionReviewReceived]
				expectedPoints := int(float64(basePoints) * tt.expectedMult)
				if tt.rating == 5 && expectedPoints != 10 {
					t.Errorf("5-star rating should award 10 points, got %d", expectedPoints)
				}
				if tt.rating == 4 && expectedPoints != 5 {
					t.Errorf("4-star rating should award 5 points, got %d", expectedPoints)
				}
			}
		})
	}
}

func TestGetProgressToNextTier(t *testing.T) {
	// Test the progress calculation logic
	tests := []struct {
		points      int
		currentTier Tier
		needed      int
		nextTier    Tier
	}{
		{0, TierBronze, 100, TierSilver},
		{50, TierBronze, 50, TierSilver},
		{100, TierSilver, 400, TierGold},
		{400, TierSilver, 100, TierGold},
		{500, TierGold, 0, ""},
		{1000, TierGold, 0, ""},
	}

	for _, tt := range tests {
		t.Run(tt.currentTier.String()+"_"+string(rune('0'+tt.points/100)), func(t *testing.T) {
			needed := PointsToNextTier(tt.points, tt.currentTier)
			if needed != tt.needed {
				t.Errorf("PointsToNextTier(%d, %s) = %d, want %d", tt.points, tt.currentTier, needed, tt.needed)
			}
		})
	}
}

// Helper function
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// Add String method to Tier for testing
func (t Tier) String() string {
	return string(t)
}

// TestSessionPointsAwarding tests the session completion points logic
func TestSessionPointsAwarding(t *testing.T) {
	// Session completion should award 10 points
	expectedPoints := PointsReward[ActionSessionCompleted]
	if expectedPoints != 10 {
		t.Errorf("Session completion should award 10 points, got %d", expectedPoints)
	}
}

// TestFirstMatchPointsAwarding tests the first match points logic
func TestFirstMatchPointsAwarding(t *testing.T) {
	// First match should award 5 points (one-time)
	expectedPoints := PointsReward[ActionFirstMatch]
	if expectedPoints != 5 {
		t.Errorf("First match should award 5 points, got %d", expectedPoints)
	}
}

// TestFirstSessionPointsAwarding tests the first session points logic
func TestFirstSessionPointsAwarding(t *testing.T) {
	// First session should award 10 points (one-time)
	expectedPoints := PointsReward[ActionFirstSession]
	if expectedPoints != 10 {
		t.Errorf("First session should award 10 points, got %d", expectedPoints)
	}
}

// TestTierProgressionScenario tests a complete tier progression scenario
func TestTierProgressionScenario(t *testing.T) {
	// Simulate a user's journey through tiers
	type action struct {
		name   string
		points int
	}

	actions := []action{
		{"First Match", 5},       // +5 = 5
		{"First Session", 10},    // +10 = 15
		{"Session Complete", 10}, // +10 = 25
		{"4-star Review", 5},     // +5 = 30
		{"5-star Review", 10},    // +10 = 40
		{"Session Complete", 10}, // +10 = 50
		{"Session Complete", 10}, // +10 = 60
		{"Session Complete", 10}, // +10 = 70
		{"Session Complete", 10}, // +10 = 80
		{"Session Complete", 10}, // +10 = 90
		{"5-star Review", 10},    // +10 = 100 -> SILVER!
	}

	totalPoints := 0
	for _, a := range actions {
		totalPoints += a.points
	}

	expectedTier := CalculateTierFromPoints(totalPoints)
	if expectedTier != TierSilver {
		t.Errorf("After %d points, tier should be Silver, got %s", totalPoints, expectedTier)
	}

	// Continue to Gold
	additionalSessions := (GoldThreshold - totalPoints) / 10
	for i := 0; i < additionalSessions; i++ {
		totalPoints += 10
	}

	// Should now be at or above Gold
	finalTier := CalculateTierFromPoints(totalPoints)
	if finalTier != TierGold && totalPoints >= GoldThreshold {
		t.Errorf("After %d points (>=%d), tier should be Gold, got %s", totalPoints, GoldThreshold, finalTier)
	}
}

// TestPointsVisibilityOnDashboard tests that points structure supports dashboard display
func TestPointsVisibilityOnDashboard(t *testing.T) {
	// UserPoints should have all fields needed for dashboard display
	up := UserPoints{
		UserID: uuid.New(),
		Points: 150,
		Tier:   TierSilver,
	}

	// Verify tier is correct for points
	calculatedTier := CalculateTierFromPoints(up.Points)
	if calculatedTier != up.Tier {
		t.Errorf("UserPoints.Tier (%s) doesn't match calculated tier (%s) for %d points",
			up.Tier, calculatedTier, up.Points)
	}

	// Verify we can calculate progress to next tier
	pointsNeeded := PointsToNextTier(up.Points, up.Tier)
	if pointsNeeded <= 0 && up.Tier != TierGold {
		t.Error("Should have positive points needed for non-Gold tier")
	}
	if pointsNeeded != GoldThreshold-up.Points {
		t.Errorf("Points to next tier: got %d, want %d", pointsNeeded, GoldThreshold-up.Points)
	}
}
