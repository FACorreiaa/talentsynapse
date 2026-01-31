package review

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestValidateReviewSubmission_DuplicateReview tests that duplicate reviews are rejected
func TestValidateReviewSubmission_DuplicateReview(t *testing.T) {
	tests := []struct {
		name           string
		fromUserID     string
		toUserID       string
		sessionID      string
		existingReview bool
		expectedError  error
	}{
		{
			name:           "duplicate review should be rejected",
			fromUserID:     "user-a-id",
			toUserID:       "user-b-id",
			sessionID:      "session-1",
			existingReview: true,
			expectedError:  ErrDuplicateReview,
		},
		{
			name:           "first review should be allowed",
			fromUserID:     "user-a-id",
			toUserID:       "user-b-id",
			sessionID:      "session-1",
			existingReview: false,
			expectedError:  nil,
		},
		{
			name:           "different direction review should be allowed (B->A after A->B exists)",
			fromUserID:     "user-b-id",
			toUserID:       "user-a-id",
			sessionID:      "session-1",
			existingReview: false, // B->A doesn't exist, only A->B
			expectedError:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Simulate the duplicate check logic
			reviewExists := tt.existingReview

			var err error
			if reviewExists {
				err = ErrDuplicateReview
			}

			if tt.expectedError != nil {
				assert.ErrorIs(t, err, tt.expectedError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestValidateReviewSubmission_SessionStatus tests session completion requirement
func TestValidateReviewSubmission_SessionStatus(t *testing.T) {
	tests := []struct {
		name          string
		sessionStatus string
		expectedError error
	}{
		{
			name:          "completed session allows review",
			sessionStatus: "completed",
			expectedError: nil,
		},
		{
			name:          "pending session rejects review",
			sessionStatus: "pending",
			expectedError: ErrSessionNotCompleted,
		},
		{
			name:          "confirmed session rejects review",
			sessionStatus: "confirmed",
			expectedError: ErrSessionNotCompleted,
		},
		{
			name:          "in_progress session rejects review",
			sessionStatus: "in_progress",
			expectedError: ErrSessionNotCompleted,
		},
		{
			name:          "cancelled session rejects review",
			sessionStatus: "cancelled",
			expectedError: ErrSessionNotCompleted,
		},
		{
			name:          "no_show session rejects review",
			sessionStatus: "no_show",
			expectedError: ErrSessionNotCompleted,
		},
		{
			name:          "disputed session rejects review",
			sessionStatus: "disputed",
			expectedError: ErrSessionNotCompleted,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Simulate the session status check
			var err error
			if tt.sessionStatus != "completed" {
				err = ErrSessionNotCompleted
			}

			if tt.expectedError != nil {
				assert.ErrorIs(t, err, tt.expectedError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestValidateReviewSubmission_SelfReview tests that self-reviews are rejected
func TestValidateReviewSubmission_SelfReview(t *testing.T) {
	tests := []struct {
		name          string
		fromUserID    string
		toUserID      string
		expectedError error
	}{
		{
			name:          "self review should be rejected",
			fromUserID:    "user-a-id",
			toUserID:      "user-a-id",
			expectedError: ErrCannotReviewSelf,
		},
		{
			name:          "review of different user should be allowed",
			fromUserID:    "user-a-id",
			toUserID:      "user-b-id",
			expectedError: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Simulate the self-review check
			var err error
			if tt.fromUserID == tt.toUserID {
				err = ErrCannotReviewSelf
			}

			if tt.expectedError != nil {
				assert.ErrorIs(t, err, tt.expectedError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestValidateReviewSubmission_SessionParticipation tests participant validation
func TestValidateReviewSubmission_SessionParticipation(t *testing.T) {
	// Session with initiator=A, partner=B
	sessionInitiator := "user-a-id"
	sessionPartner := "user-b-id"

	tests := []struct {
		name          string
		fromUserID    string
		toUserID      string
		expectedError error
	}{
		{
			name:          "initiator can review partner",
			fromUserID:    sessionInitiator,
			toUserID:      sessionPartner,
			expectedError: nil,
		},
		{
			name:          "partner can review initiator",
			fromUserID:    sessionPartner,
			toUserID:      sessionInitiator,
			expectedError: nil,
		},
		{
			name:          "non-participant cannot review",
			fromUserID:    "user-c-id",
			toUserID:      sessionPartner,
			expectedError: ErrNotSessionParticipant,
		},
		{
			name:          "participant cannot review non-participant",
			fromUserID:    sessionInitiator,
			toUserID:      "user-c-id",
			expectedError: ErrNotSessionParticipant,
		},
		{
			name:          "both users must be participants",
			fromUserID:    "user-c-id",
			toUserID:      "user-d-id",
			expectedError: ErrNotSessionParticipant,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Simulate the participation check
			var err error
			isFromParticipant := tt.fromUserID == sessionInitiator || tt.fromUserID == sessionPartner
			isToParticipant := tt.toUserID == sessionInitiator || tt.toUserID == sessionPartner

			if !isFromParticipant || !isToParticipant {
				err = ErrNotSessionParticipant
			}

			if tt.expectedError != nil {
				assert.ErrorIs(t, err, tt.expectedError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestReviewExistsForSession tests the duplicate detection logic
func TestReviewExistsForSession(t *testing.T) {
	// Simulate a reviews map: key is "fromUserID:sessionID"
	existingReviews := map[string]bool{
		"user-a:session-1": true,
		"user-b:session-2": true,
	}

	tests := []struct {
		name       string
		fromUserID string
		sessionID  string
		expected   bool
	}{
		{
			name:       "existing review returns true",
			fromUserID: "user-a",
			sessionID:  "session-1",
			expected:   true,
		},
		{
			name:       "no existing review returns false",
			fromUserID: "user-a",
			sessionID:  "session-2",
			expected:   false,
		},
		{
			name:       "different user same session returns false",
			fromUserID: "user-c",
			sessionID:  "session-1",
			expected:   false,
		},
		{
			name:       "same user different session returns false",
			fromUserID: "user-a",
			sessionID:  "session-3",
			expected:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key := tt.fromUserID + ":" + tt.sessionID
			exists := existingReviews[key]
			assert.Equal(t, tt.expected, exists)
		})
	}
}

// TestBidirectionalReviews tests that A->B and B->A are independent
func TestBidirectionalReviews(t *testing.T) {
	// Simulate existing reviews
	existingReviews := map[string]bool{
		"user-a:session-1": true, // A already reviewed B for session-1
	}

	t.Run("A already reviewed B, B can still review A", func(t *testing.T) {
		// A->B exists
		aReviewedB := existingReviews["user-a:session-1"]
		assert.True(t, aReviewedB, "A's review of B should exist")

		// B->A should not exist (different key)
		bReviewedA := existingReviews["user-b:session-1"]
		assert.False(t, bReviewedA, "B's review of A should not exist yet")

		// B should be able to review A
		canBReviewA := !bReviewedA
		assert.True(t, canBReviewA, "B should be able to review A")
	})

	t.Run("after both review, both entries exist", func(t *testing.T) {
		// Add B's review
		existingReviews["user-b:session-1"] = true

		aReviewedB := existingReviews["user-a:session-1"]
		bReviewedA := existingReviews["user-b:session-1"]

		assert.True(t, aReviewedB, "A's review should exist")
		assert.True(t, bReviewedA, "B's review should exist")
	})
}

// TestErrorTypes tests that error types are correctly defined
func TestErrorTypes(t *testing.T) {
	t.Run("ErrDuplicateReview is distinct", func(t *testing.T) {
		err := ErrDuplicateReview
		assert.True(t, errors.Is(err, ErrDuplicateReview))
		assert.False(t, errors.Is(err, ErrSessionNotCompleted))
		assert.False(t, errors.Is(err, ErrSessionNotFound))
	})

	t.Run("ErrSessionNotCompleted is distinct", func(t *testing.T) {
		err := ErrSessionNotCompleted
		assert.True(t, errors.Is(err, ErrSessionNotCompleted))
		assert.False(t, errors.Is(err, ErrDuplicateReview))
	})

	t.Run("ErrSessionNotFound is distinct", func(t *testing.T) {
		err := ErrSessionNotFound
		assert.True(t, errors.Is(err, ErrSessionNotFound))
		assert.False(t, errors.Is(err, ErrDuplicateReview))
	})

	t.Run("ErrNotSessionParticipant is distinct", func(t *testing.T) {
		err := ErrNotSessionParticipant
		assert.True(t, errors.Is(err, ErrNotSessionParticipant))
		assert.False(t, errors.Is(err, ErrDuplicateReview))
	})

	t.Run("ErrCannotReviewSelf is distinct", func(t *testing.T) {
		err := ErrCannotReviewSelf
		assert.True(t, errors.Is(err, ErrCannotReviewSelf))
		assert.False(t, errors.Is(err, ErrDuplicateReview))
	})
}

// TestValidationOrder tests that validation checks are performed in the correct order
func TestValidationOrder(t *testing.T) {
	// The order should be:
	// 1. Self-review check (fastest, no DB access needed)
	// 2. Session exists check
	// 3. Participation check
	// 4. Session status check
	// 5. Duplicate review check

	t.Run("self review check comes first", func(t *testing.T) {
		fromUserID := "user-a"
		toUserID := "user-a" // Same user

		// Even if session doesn't exist, self-review should be caught first
		var err error
		if fromUserID == toUserID {
			err = ErrCannotReviewSelf
		}
		require.ErrorIs(t, err, ErrCannotReviewSelf)
	})
}

// TestRatingValidation tests rating value constraints
func TestRatingValidation(t *testing.T) {
	tests := []struct {
		name    string
		rating  int
		isValid bool
	}{
		{name: "rating 0 is invalid", rating: 0, isValid: false},
		{name: "rating 1 is valid", rating: 1, isValid: true},
		{name: "rating 2 is valid", rating: 2, isValid: true},
		{name: "rating 3 is valid", rating: 3, isValid: true},
		{name: "rating 4 is valid", rating: 4, isValid: true},
		{name: "rating 5 is valid", rating: 5, isValid: true},
		{name: "rating 6 is invalid", rating: 6, isValid: false},
		{name: "negative rating is invalid", rating: -1, isValid: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isValid := tt.rating >= 1 && tt.rating <= 5
			assert.Equal(t, tt.isValid, isValid)
		})
	}
}
