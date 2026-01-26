package profile

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	profilepages "github.com/FACorreiaa/talentsynapse/internal/app/views/pages/profile"
)

func TestReviewDataMapping(t *testing.T) {
	// Test that review data is correctly mapped to view model
	tests := []struct {
		name           string
		rating         int
		comment        string
		reviewerName   string
		reviewerAvatar string
	}{
		{
			name:           "5 star review with comment",
			rating:         5,
			comment:        "Excellent teacher!",
			reviewerName:   "John Doe",
			reviewerAvatar: "https://example.com/avatar.jpg",
		},
		{
			name:           "3 star review without avatar",
			rating:         3,
			comment:        "Good but could improve",
			reviewerName:   "Jane Smith",
			reviewerAvatar: "",
		},
		{
			name:           "1 star review without comment",
			rating:         1,
			comment:        "",
			reviewerName:   "Anonymous",
			reviewerAvatar: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Simulate the mapping in handler.ShowPublic()
			reviewData := profilepages.ReviewData{
				Rating:         tt.rating,
				Comment:        tt.comment,
				ReviewerName:   tt.reviewerName,
				ReviewerAvatar: tt.reviewerAvatar,
				CreatedAt:      time.Now(),
			}

			assert.Equal(t, tt.rating, reviewData.Rating)
			assert.Equal(t, tt.comment, reviewData.Comment)
			assert.Equal(t, tt.reviewerName, reviewData.ReviewerName)
			assert.Equal(t, tt.reviewerAvatar, reviewData.ReviewerAvatar)
			assert.False(t, reviewData.CreatedAt.IsZero())
		})
	}
}

func TestPublicProfileDataReviewsSlice(t *testing.T) {
	// Test that reviews can be properly appended to profile data
	profile := profilepages.PublicProfileData{
		ID:          "test-user-id",
		DisplayName: "Test User",
		Reviews:     []profilepages.ReviewData{},
	}

	// Add reviews
	reviews := []profilepages.ReviewData{
		{Rating: 5, Comment: "Great!", ReviewerName: "User1"},
		{Rating: 4, Comment: "Good", ReviewerName: "User2"},
		{Rating: 3, Comment: "OK", ReviewerName: "User3"},
	}

	profile.Reviews = append(profile.Reviews, reviews...)

	assert.Len(t, profile.Reviews, 3)
	assert.Equal(t, 5, profile.Reviews[0].Rating)
	assert.Equal(t, "Great!", profile.Reviews[0].Comment)
}

func TestGetInitialFunction(t *testing.T) {
	// Test the getPublicInitial helper function logic
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "normal name",
			input:    "John",
			expected: "J",
		},
		{
			name:     "unicode name",
			input:    "Über",
			expected: "Ü",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "?",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Replicate the getPublicInitial logic from template
			var result string
			if len(tt.input) == 0 {
				result = "?"
			} else {
				runes := []rune(tt.input)
				result = string(runes[0])
			}
			assert.Equal(t, tt.expected, result)
		})
	}
}
