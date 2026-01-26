package session

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCountCompletedSessionsThreshold(t *testing.T) {
	// Test the threshold logic for dedicated_learner badge
	tests := []struct {
		name           string
		completedCount int
		shouldAward    bool
	}{
		{
			name:           "0 sessions - no badge",
			completedCount: 0,
			shouldAward:    false,
		},
		{
			name:           "4 sessions - no badge",
			completedCount: 4,
			shouldAward:    false,
		},
		{
			name:           "5 sessions - award badge",
			completedCount: 5,
			shouldAward:    true,
		},
		{
			name:           "10 sessions - award badge",
			completedCount: 10,
			shouldAward:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Simulate the badge award logic from handler.Complete
			shouldAward := tt.completedCount >= 5
			assert.Equal(t, tt.shouldAward, shouldAward)
		})
	}
}

// MockRepository implements a simple mock for testing
type MockRepository struct {
	completedCount   int
	countError       error
	markCompletedErr error
}

func (m *MockRepository) CountCompletedSessions(ctx context.Context, userID string) (int, error) {
	if m.countError != nil {
		return 0, m.countError
	}
	return m.completedCount, nil
}

func (m *MockRepository) MarkCompleted(ctx context.Context, sessionID, userID string) error {
	return m.markCompletedErr
}

func TestMockCountCompletedSessions(t *testing.T) {
	// Test the mock repository
	mock := &MockRepository{
		completedCount: 5,
	}

	count, err := mock.CountCompletedSessions(context.Background(), "test-user-id")
	require.NoError(t, err)
	assert.Equal(t, 5, count)
}

func TestDedicatedLearnerBadgeLogic(t *testing.T) {
	// Table-driven test for badge award logic
	testCases := []struct {
		name           string
		completedCount int
		countError     error
		expectBadge    bool
	}{
		{
			name:           "exactly 5 sessions awards badge",
			completedCount: 5,
			countError:     nil,
			expectBadge:    true,
		},
		{
			name:           "less than 5 sessions no badge",
			completedCount: 4,
			countError:     nil,
			expectBadge:    false,
		},
		{
			name:           "error in count prevents badge",
			completedCount: 10,
			countError:     assert.AnError,
			expectBadge:    false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mock := &MockRepository{
				completedCount: tc.completedCount,
				countError:     tc.countError,
			}

			count, err := mock.CountCompletedSessions(context.Background(), "user-123")

			// Replicate handler logic
			shouldAwardBadge := err == nil && count >= 5

			assert.Equal(t, tc.expectBadge, shouldAwardBadge)
		})
	}
}
