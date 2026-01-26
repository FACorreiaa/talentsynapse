package algorithm

import (
	"math"
	"testing"
)

func TestCosineSimilarity(t *testing.T) {
	tests := []struct {
		name     string
		vectorA  []float64
		vectorB  []float64
		expected float64
	}{
		{
			name:     "identical vectors",
			vectorA:  []float64{1, 2, 3},
			vectorB:  []float64{1, 2, 3},
			expected: 1.0,
		},
		{
			name:     "opposite vectors",
			vectorA:  []float64{1, 2, 3},
			vectorB:  []float64{-1, -2, -3},
			expected: -1.0,
		},
		{
			name:     "orthogonal vectors",
			vectorA:  []float64{1, 0, 0},
			vectorB:  []float64{0, 1, 0},
			expected: 0.0,
		},
		{
			name:     "zero vector",
			vectorA:  []float64{1, 2, 3},
			vectorB:  []float64{0, 0, 0},
			expected: 0.0,
		},
		{
			name:     "partial match",
			vectorA:  []float64{8, 5, 0},
			vectorB:  []float64{7, 6, 3},
			expected: 0.9402, // approximate
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CosineSimilarity(tt.vectorA, tt.vectorB)

			// Use approximate comparison for floating point
			if math.Abs(result-tt.expected) > 0.01 {
				t.Errorf("CosineSimilarity() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestEuclideanDistance(t *testing.T) {
	tests := []struct {
		name     string
		vectorA  []float64
		vectorB  []float64
		expected float64
	}{
		{
			name:     "identical vectors",
			vectorA:  []float64{1, 2, 3},
			vectorB:  []float64{1, 2, 3},
			expected: 0.0,
		},
		{
			name:     "unit distance in 1D",
			vectorA:  []float64{0},
			vectorB:  []float64{1},
			expected: 1.0,
		},
		{
			name:     "3-4-5 triangle",
			vectorA:  []float64{0, 0},
			vectorB:  []float64{3, 4},
			expected: 5.0,
		},
		{
			name:     "skill proficiency difference",
			vectorA:  []float64{8, 5, 0},
			vectorB:  []float64{7, 6, 3},
			expected: 3.3166, // approximate sqrt(1 + 1 + 9)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := EuclideanDistance(tt.vectorA, tt.vectorB)

			if math.Abs(result-tt.expected) > 0.01 {
				t.Errorf("EuclideanDistance() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestMatchScore(t *testing.T) {
	tests := []struct {
		name     string
		userA    UserProfile
		userB    UserProfile
		minScore float64
		maxScore float64
	}{
		{
			name: "perfect complementary match",
			userA: UserProfile{
				UserID:        "user-a",
				OfferedVector: []float64{8, 0, 0}, // Offers Python (skill 1)
				WantedVector:  []float64{0, 7, 0}, // Wants JavaScript (skill 2)
			},
			userB: UserProfile{
				UserID:        "user-b",
				OfferedVector: []float64{0, 9, 0}, // Offers JavaScript (skill 2)
				WantedVector:  []float64{6, 0, 0}, // Wants Python (skill 1)
			},
			minScore: 0.7,
			maxScore: 1.0,
		},
		{
			name: "no overlap",
			userA: UserProfile{
				UserID:        "user-a",
				OfferedVector: []float64{8, 0, 0},
				WantedVector:  []float64{0, 7, 0},
			},
			userB: UserProfile{
				UserID:        "user-b",
				OfferedVector: []float64{0, 0, 5},
				WantedVector:  []float64{0, 0, 4},
			},
			minScore: 0.0,
			maxScore: 0.1,
		},
		{
			name: "partial match",
			userA: UserProfile{
				UserID:        "user-a",
				OfferedVector: []float64{8, 5, 0},
				WantedVector:  []float64{0, 0, 7},
			},
			userB: UserProfile{
				UserID:        "user-b",
				OfferedVector: []float64{0, 0, 6},
				WantedVector:  []float64{7, 6, 0},
			},
			minScore: 0.5,
			maxScore: 1.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score := MatchScore(tt.userA, tt.userB)

			if score < tt.minScore || score > tt.maxScore {
				t.Errorf("MatchScore() = %v, want between %v and %v",
					score, tt.minScore, tt.maxScore)
			}
		})
	}
}

func TestFindBestMatches(t *testing.T) {
	queryUser := UserProfile{
		UserID:        "query-user",
		OfferedVector: []float64{8, 0, 0}, // Offers Python
		WantedVector:  []float64{0, 7, 0}, // Wants JavaScript
	}

	candidates := []UserProfile{
		{
			UserID:        "perfect-match",
			DisplayName:   "Perfect Match",
			Username:      "perfect",
			OfferedVector: []float64{0, 9, 0}, // Offers JavaScript
			WantedVector:  []float64{8, 0, 0}, // Wants Python
		},
		{
			UserID:        "no-match",
			DisplayName:   "No Match",
			Username:      "nomatch",
			OfferedVector: []float64{0, 0, 5},
			WantedVector:  []float64{0, 0, 4},
		},
		{
			UserID:        "partial-match",
			DisplayName:   "Partial Match",
			Username:      "partial",
			OfferedVector: []float64{0, 6, 0}, // Offers JavaScript
			WantedVector:  []float64{5, 0, 3}, // Wants Python (and something else)
		},
		{
			UserID:        "query-user", // Should be filtered out
			DisplayName:   "Self",
			Username:      "self",
			OfferedVector: []float64{8, 0, 0},
			WantedVector:  []float64{0, 7, 0},
		},
	}

	t.Run("find all above threshold", func(t *testing.T) {
		results := FindBestMatches(queryUser, candidates, 0.5, 0)

		// Should find perfect-match and partial-match, but not no-match or self
		if len(results) < 1 {
			t.Errorf("Expected at least 1 match, got %d", len(results))
		}

		// Verify results are sorted by score (descending)
		for i := 0; i < len(results)-1; i++ {
			if results[i].Score < results[i+1].Score {
				t.Errorf("Results not sorted: results[%d].Score (%v) < results[%d].Score (%v)",
					i, results[i].Score, i+1, results[i+1].Score)
			}
		}

		// Verify self is not included
		for _, result := range results {
			if result.UserID == "query-user" {
				t.Error("Self-match should be filtered out")
			}
		}
	})

	t.Run("limit results", func(t *testing.T) {
		results := FindBestMatches(queryUser, candidates, 0.0, 1)

		if len(results) != 1 {
			t.Errorf("Expected 1 result with limit, got %d", len(results))
		}
	})

	t.Run("high threshold filters matches", func(t *testing.T) {
		results := FindBestMatches(queryUser, candidates, 0.99, 0)

		// With very high threshold, might get 0 or 1 results
		if len(results) > 2 {
			t.Errorf("Expected few results with high threshold, got %d", len(results))
		}
	})

	t.Run("algorithm name is set", func(t *testing.T) {
		results := FindBestMatches(queryUser, candidates, 0.0, 1)

		if len(results) > 0 && results[0].Algorithm != "cosine_similarity" {
			t.Errorf("Expected algorithm 'cosine_similarity', got '%s'", results[0].Algorithm)
		}
	})
}

func TestPadVector(t *testing.T) {
	tests := []struct {
		name      string
		vector    []float64
		targetLen int
		wantLen   int
	}{
		{
			name:      "pad short vector",
			vector:    []float64{1, 2},
			targetLen: 5,
			wantLen:   5,
		},
		{
			name:      "no padding needed",
			vector:    []float64{1, 2, 3},
			targetLen: 3,
			wantLen:   3,
		},
		{
			name:      "vector longer than target",
			vector:    []float64{1, 2, 3, 4, 5},
			targetLen: 3,
			wantLen:   5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := padVector(tt.vector, tt.targetLen)

			if len(result) != tt.wantLen {
				t.Errorf("padVector() length = %v, want %v", len(result), tt.wantLen)
			}

			// Verify original values are preserved
			for i := 0; i < len(tt.vector) && i < len(result); i++ {
				if result[i] != tt.vector[i] {
					t.Errorf("padVector() changed original value at index %d", i)
				}
			}
		})
	}
}

// Benchmark tests
func BenchmarkCosineSimilarity(b *testing.B) {
	v1 := []float64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	v2 := []float64{2, 3, 4, 5, 6, 7, 8, 9, 10, 11}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		CosineSimilarity(v1, v2)
	}
}

func BenchmarkMatchScore(b *testing.B) {
	userA := UserProfile{
		UserID:        "user-a",
		OfferedVector: []float64{8, 5, 0, 7, 3},
		WantedVector:  []float64{0, 0, 7, 0, 9},
	}
	userB := UserProfile{
		UserID:        "user-b",
		OfferedVector: []float64{0, 0, 6, 0, 8},
		WantedVector:  []float64{7, 6, 0, 8, 0},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		MatchScore(userA, userB)
	}
}

func BenchmarkFindBestMatches(b *testing.B) {
	queryUser := UserProfile{
		UserID:        "query",
		OfferedVector: []float64{8, 0, 0, 5, 7},
		WantedVector:  []float64{0, 7, 0, 0, 0},
	}

	// Create 100 candidate users
	candidates := make([]UserProfile, 100)
	for i := 0; i < 100; i++ {
		candidates[i] = UserProfile{
			UserID:        string(rune(i)),
			OfferedVector: []float64{float64(i % 10), float64((i + 1) % 10), float64((i + 2) % 10), 0, 0},
			WantedVector:  []float64{0, 0, float64((i + 3) % 10), float64((i + 4) % 10), 0},
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		FindBestMatches(queryUser, candidates, 0.3, 10)
	}
}

// TestComplementarySkillMatch tests the specific scenario where:
// User A: knows Go, wants Piano
// User B: knows Piano, wants Go
// These users should match with a high score
func TestComplementarySkillMatch(t *testing.T) {
	// Simulate a skill catalog with 5 skills:
	// Index 0: Go
	// Index 1: JavaScript
	// Index 2: Piano
	// Index 3: Python
	// Index 4: Spanish

	// User A: knows Go (index 0), wants Piano (index 2)
	userA := UserProfile{
		UserID:        "user-a",
		DisplayName:   "Alice",
		Username:      "alice",
		OfferedVector: []float64{8, 0, 0, 0, 0}, // Offers Go with proficiency 8
		WantedVector:  []float64{0, 0, 7, 0, 0}, // Wants Piano with proficiency 7
	}

	// User B: knows Piano (index 2), wants Go (index 0)
	userB := UserProfile{
		UserID:        "user-b",
		DisplayName:   "Bob",
		Username:      "bob",
		OfferedVector: []float64{0, 0, 9, 0, 0}, // Offers Piano with proficiency 9
		WantedVector:  []float64{6, 0, 0, 0, 0}, // Wants Go with proficiency 6
	}

	t.Run("direct match score calculation", func(t *testing.T) {
		score := MatchScore(userA, userB)

		// With perfect complementary skills, we expect a high score
		// B offers Piano (index 2) at position 2, A wants Piano at position 2 -> similarity = 1.0
		// A offers Go (index 0) at position 0, B wants Go at position 0 -> similarity = 1.0
		// Average = 1.0
		if score < 0.9 {
			t.Errorf("Expected high match score (>0.9) for complementary skills, got %v", score)
		}
	})

	t.Run("bidirectional similarity", func(t *testing.T) {
		// The score should be symmetric
		scoreAB := MatchScore(userA, userB)
		scoreBA := MatchScore(userB, userA)

		if math.Abs(scoreAB-scoreBA) > 0.001 {
			t.Errorf("Match score should be symmetric: A->B = %v, B->A = %v", scoreAB, scoreBA)
		}
	})

	t.Run("find matches includes complementary user", func(t *testing.T) {
		candidates := []UserProfile{userB}
		results := FindBestMatches(userA, candidates, 0.3, 10)

		if len(results) == 0 {
			t.Fatal("Expected to find at least one match for complementary skills")
		}

		if results[0].UserID != "user-b" {
			t.Errorf("Expected user-b to be the top match, got %s", results[0].UserID)
		}

		if results[0].Score < 0.9 {
			t.Errorf("Expected high score for complementary match, got %v", results[0].Score)
		}
	})
}

// TestNoOverlapNoMatch tests that users with no skill overlap don't match
func TestNoOverlapNoMatch(t *testing.T) {
	// User A: knows Go, wants Piano
	userA := UserProfile{
		UserID:        "user-a",
		OfferedVector: []float64{8, 0, 0, 0, 0}, // Go
		WantedVector:  []float64{0, 0, 7, 0, 0}, // Piano
	}

	// User B: knows Spanish, wants JavaScript (no overlap with A)
	userB := UserProfile{
		UserID:        "user-b",
		OfferedVector: []float64{0, 0, 0, 0, 9}, // Spanish
		WantedVector:  []float64{0, 8, 0, 0, 0}, // JavaScript
	}

	score := MatchScore(userA, userB)

	// With zero overlap, score should be very low
	if score > 0.1 {
		t.Errorf("Expected low score for no overlap, got %v", score)
	}
}

// TestPartialMatch tests users with partial skill overlap
func TestPartialMatch(t *testing.T) {
	// User A: knows Go and Python, wants Piano
	userA := UserProfile{
		UserID:        "user-a",
		OfferedVector: []float64{8, 0, 0, 7, 0}, // Go, Python
		WantedVector:  []float64{0, 0, 7, 0, 0}, // Piano
	}

	// User B: knows Piano, wants Python (partial overlap - only one direction)
	userB := UserProfile{
		UserID:        "user-b",
		OfferedVector: []float64{0, 0, 9, 0, 0}, // Piano
		WantedVector:  []float64{0, 0, 0, 6, 0}, // Python
	}

	score := MatchScore(userA, userB)

	// Partial match should have a moderate score
	if score < 0.3 || score > 0.99 {
		t.Errorf("Expected moderate score for partial match (0.3-0.99), got %v", score)
	}

	t.Logf("Partial match score: %v", score)
}

// TestEmptyVectors tests handling of users with no skills
func TestEmptyVectors(t *testing.T) {
	userA := UserProfile{
		UserID:        "user-a",
		OfferedVector: []float64{8, 0, 0, 0, 0},
		WantedVector:  []float64{0, 0, 7, 0, 0},
	}

	userEmpty := UserProfile{
		UserID:        "user-empty",
		OfferedVector: []float64{0, 0, 0, 0, 0}, // No skills offered
		WantedVector:  []float64{0, 0, 0, 0, 0}, // No skills wanted
	}

	score := MatchScore(userA, userEmpty)

	// Empty vectors should result in zero score (handled by CosineSimilarity)
	if score != 0 {
		t.Errorf("Expected zero score for empty vectors, got %v", score)
	}
}

// TestMultipleSkillsMatch tests matching with multiple offered/wanted skills
func TestMultipleSkillsMatch(t *testing.T) {
	// User A: knows Go, Python, JavaScript; wants Piano, Spanish
	userA := UserProfile{
		UserID:        "user-a",
		OfferedVector: []float64{8, 7, 0, 9, 0}, // Go, JS, Python
		WantedVector:  []float64{0, 0, 6, 0, 5}, // Piano, Spanish
	}

	// User B: knows Piano, Spanish; wants Go, Python
	userB := UserProfile{
		UserID:        "user-b",
		OfferedVector: []float64{0, 0, 8, 0, 7}, // Piano, Spanish
		WantedVector:  []float64{6, 0, 0, 5, 0}, // Go, Python
	}

	score := MatchScore(userA, userB)

	// Should have a high score due to multiple complementary skills
	if score < 0.7 {
		t.Errorf("Expected high score for multiple complementary skills, got %v", score)
	}

	t.Logf("Multiple skills match score: %v", score)
}
