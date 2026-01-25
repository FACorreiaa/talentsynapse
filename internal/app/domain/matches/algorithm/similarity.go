package algorithm

import (
	"math"

	"gonum.org/v1/gonum/floats"
)

// UserProfile represents a skill vector for matching
type UserProfile struct {
	UserID        string
	OfferedVector []float64
	WantedVector  []float64
	SkillCount    int
	DisplayName   string
	Username      string
	AvatarURL     string
}

// MatchResult contains the result of a matching operation
type MatchResult struct {
	UserID      string
	DisplayName string
	Username    string
	AvatarURL   string
	Score       float64
	Algorithm   string
}

// EuclideanDistance calculates the Euclidean distance between two vectors
// Smaller distances indicate better matches
func EuclideanDistance(a, b []float64) float64 {
	if len(a) != len(b) {
		panic("Vectors must be same length")
	}

	diff := make([]float64, len(a))
	floats.SubTo(diff, a, b) // diff = a - b
	floats.Mul(diff, diff)   // square each element
	return math.Sqrt(floats.Sum(diff))
}

// CosineSimilarity calculates the cosine similarity between two vectors
// The similarity S between two users A and B is calculated as:
// S(A, B) = (A · B) / (||A|| ||B||)
// Where A · B is the dot product and ||A|| is the magnitude
// Returns a value between -1 and 1, where 1 means identical direction
func CosineSimilarity(a, b []float64) float64 {
	if len(a) != len(b) {
		panic("Vectors must be same length")
	}

	dot := floats.Dot(a, b)
	normA := floats.Norm(a, 2) // L2 norm (Euclidean norm)
	normB := floats.Norm(b, 2)

	// Handle zero vectors (no skills)
	if normA == 0 || normB == 0 {
		return 0
	}

	return dot / (normA * normB)
}

// MatchScore calculates a comprehensive match score between two users
// This considers both what they offer vs what you want, and vice versa
func MatchScore(userA, userB UserProfile) float64 {
	// Ensure vectors are the same length by padding with zeros if needed
	maxLen := max(len(userA.OfferedVector), len(userB.OfferedVector),
		len(userA.WantedVector), len(userB.WantedVector))

	offeredA := padVector(userA.OfferedVector, maxLen)
	wantedA := padVector(userA.WantedVector, maxLen)
	offeredB := padVector(userB.OfferedVector, maxLen)
	wantedB := padVector(userB.WantedVector, maxLen)

	// Calculate bidirectional similarity:
	// 1. How well does B's offered skills match A's wanted skills?
	// 2. How well does A's offered skills match B's wanted skills?
	similarityBtoA := CosineSimilarity(offeredB, wantedA)
	similarityAtoB := CosineSimilarity(offeredA, wantedB)

	// Take the average for a balanced score
	// This ensures mutual benefit in the skill exchange
	return (similarityBtoA + similarityAtoB) / 2.0
}

// scoredMatch represents a user profile with their match score
type scoredMatch struct {
	profile UserProfile
	score   float64
}

// FindBestMatches finds the best matching users based on cosine similarity
// Returns sorted results by score (highest first)
func FindBestMatches(queryUser UserProfile, candidates []UserProfile, threshold float64, limit int) []MatchResult {
	var matches []scoredMatch
	for _, candidate := range candidates {
		// Skip self-matching
		if candidate.UserID == queryUser.UserID {
			continue
		}

		score := MatchScore(queryUser, candidate)

		// Only include matches above threshold
		if score >= threshold {
			matches = append(matches, scoredMatch{
				profile: candidate,
				score:   score,
			})
		}
	}

	// Sort by score descending (highest similarity first)
	sortByScore(matches)

	// Limit results
	if limit > 0 && len(matches) > limit {
		matches = matches[:limit]
	}

	// Convert to MatchResult
	results := make([]MatchResult, 0, len(matches))
	for _, m := range matches {
		results = append(results, MatchResult{
			UserID:      m.profile.UserID,
			DisplayName: m.profile.DisplayName,
			Username:    m.profile.Username,
			AvatarURL:   m.profile.AvatarURL,
			Score:       m.score,
			Algorithm:   "cosine_similarity",
		})
	}

	return results
}

// padVector pads a vector with zeros to reach the target length
func padVector(v []float64, targetLen int) []float64 {
	if len(v) >= targetLen {
		return v
	}

	padded := make([]float64, targetLen)
	copy(padded, v)
	return padded
}

// sortByScore sorts matches by score in descending order (bubble sort for simplicity)
func sortByScore(matches []scoredMatch) {
	n := len(matches)
	for i := 0; i < n-1; i++ {
		for j := 0; j < n-i-1; j++ {
			if matches[j].score < matches[j+1].score {
				matches[j], matches[j+1] = matches[j+1], matches[j]
			}
		}
	}
}

// max returns the maximum of the given integers
func max(nums ...int) int {
	if len(nums) == 0 {
		return 0
	}
	maxVal := nums[0]
	for _, n := range nums[1:] {
		if n > maxVal {
			maxVal = n
		}
	}
	return maxVal
}
