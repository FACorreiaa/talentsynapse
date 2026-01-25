package matches

import (
	"context"

	"github.com/FACorreiaa/skillsphere/internal/app/domain/matches/algorithm"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Match represents a potential skill exchange match
type Match struct {
	UserID        string
	DisplayName   string
	Username      string
	AvatarURL     string
	MatchScore    float64
	OverlapSkills []OverlapSkill
}

// OverlapSkill represents a skill that overlaps between users
type OverlapSkill struct {
	Name      string
	TheyTeach bool // true if they teach it (you want), false if they want it (you teach)
}

// Repository handles database operations for matches
type Repository struct {
	pool *pgxpool.Pool
}

// NewRepository creates a new matches repository
func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

// GetPendingMatches retrieves potential matches for a user using cosine similarity algorithm
// This implementation uses the user_skill_vectors materialized view for efficient vector-based matching
func (r *Repository) GetPendingMatches(ctx context.Context, userID string, limit int) ([]Match, error) {
	// First, get the current user's skill vectors
	queryUser, err := r.getUserProfile(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Get all potential candidate users with their skill vectors
	candidates, err := r.getAllCandidateProfiles(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Use cosine similarity algorithm to find best matches
	// Threshold of 0.3 means we need at least 30% similarity
	const similarityThreshold = 0.3
	matchResults := algorithm.FindBestMatches(queryUser, candidates, similarityThreshold, limit)

	// Convert algorithm results to Match objects with overlap skills
	var matches []Match
	for _, result := range matchResults {
		overlapSkills, err := r.getOverlapSkills(ctx, userID, result.UserID)
		if err != nil {
			return nil, err
		}

		matches = append(matches, Match{
			UserID:        result.UserID,
			DisplayName:   result.DisplayName,
			Username:      result.Username,
			AvatarURL:     result.AvatarURL,
			MatchScore:    result.Score,
			OverlapSkills: overlapSkills,
		})
	}

	return matches, nil
}

// getUserProfile retrieves a user's profile as a vector for matching
func (r *Repository) getUserProfile(ctx context.Context, userID string) (algorithm.UserProfile, error) {
	query := `
		SELECT
			u.id,
			u.display_name,
			u.username,
			COALESCE(u.avatar_url, '') as avatar_url,
			COALESCE(usv.offered_vector, ARRAY[]::integer[]) as offered_vector,
			COALESCE(usv.wanted_vector, ARRAY[]::integer[]) as wanted_vector
		FROM users u
		LEFT JOIN user_skill_vectors usv ON u.id = usv.user_id
		WHERE u.id = $1 AND u.is_active = true
	`

	var profile algorithm.UserProfile
	var offeredInts, wantedInts []int

	err := r.pool.QueryRow(ctx, query, userID).Scan(
		&profile.UserID,
		&profile.DisplayName,
		&profile.Username,
		&profile.AvatarURL,
		&offeredInts,
		&wantedInts,
	)
	if err != nil {
		return algorithm.UserProfile{}, err
	}

	// Convert int arrays to float64 for algorithm
	profile.OfferedVector = intsToFloats(offeredInts)
	profile.WantedVector = intsToFloats(wantedInts)
	profile.SkillCount = len(offeredInts) + len(wantedInts)

	return profile, nil
}

// getAllCandidateProfiles retrieves all potential candidate users with their skill vectors
func (r *Repository) getAllCandidateProfiles(ctx context.Context, excludeUserID string) ([]algorithm.UserProfile, error) {
	query := `
		SELECT
			u.id,
			u.display_name,
			u.username,
			COALESCE(u.avatar_url, '') as avatar_url,
			COALESCE(usv.offered_vector, ARRAY[]::integer[]) as offered_vector,
			COALESCE(usv.wanted_vector, ARRAY[]::integer[]) as wanted_vector
		FROM users u
		LEFT JOIN user_skill_vectors usv ON u.id = usv.user_id
		WHERE u.id != $1
		AND u.is_active = true
		AND (usv.offered_vector IS NOT NULL OR usv.wanted_vector IS NOT NULL)
	`

	rows, err := r.pool.Query(ctx, query, excludeUserID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var candidates []algorithm.UserProfile
	for rows.Next() {
		var profile algorithm.UserProfile
		var offeredInts, wantedInts []int

		err := rows.Scan(
			&profile.UserID,
			&profile.DisplayName,
			&profile.Username,
			&profile.AvatarURL,
			&offeredInts,
			&wantedInts,
		)
		if err != nil {
			return nil, err
		}

		// Convert int arrays to float64 for algorithm
		profile.OfferedVector = intsToFloats(offeredInts)
		profile.WantedVector = intsToFloats(wantedInts)
		profile.SkillCount = len(offeredInts) + len(wantedInts)

		candidates = append(candidates, profile)
	}

	return candidates, rows.Err()
}

// intsToFloats converts an int slice to float64 slice
func intsToFloats(ints []int) []float64 {
	floats := make([]float64, len(ints))
	for i, v := range ints {
		floats[i] = float64(v)
	}
	return floats
}

// getOverlapSkills retrieves the skills that overlap between two users
func (r *Repository) getOverlapSkills(ctx context.Context, userID, matchedUserID string) ([]OverlapSkill, error) {
	query := `
		SELECT 
			s.name,
			CASE WHEN us_them.skill_type = 'offered' THEN true ELSE false END as they_teach
		FROM user_skills us_them
		JOIN skills s ON us_them.skill_id = s.id
		WHERE us_them.user_id = $2
		AND (
			-- They offer what user wants
			(us_them.skill_type = 'offered' AND us_them.skill_id IN (
				SELECT skill_id FROM user_skills WHERE user_id = $1 AND skill_type = 'wanted'
			))
			OR
			-- They want what user offers
			(us_them.skill_type = 'wanted' AND us_them.skill_id IN (
				SELECT skill_id FROM user_skills WHERE user_id = $1 AND skill_type = 'offered'
			))
		)
		LIMIT 5
	`

	rows, err := r.pool.Query(ctx, query, userID, matchedUserID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var skills []OverlapSkill
	for rows.Next() {
		var skill OverlapSkill
		err := rows.Scan(&skill.Name, &skill.TheyTeach)
		if err != nil {
			return nil, err
		}
		skills = append(skills, skill)
	}

	return skills, rows.Err()
}

// RecordMatchAction records a user's decision (accept/reject) on a match
func (r *Repository) RecordMatchAction(ctx context.Context, userID, matchedUserID string, accepted bool) error {
	// Calculate the actual match score using cosine similarity
	userProfile, err := r.getUserProfile(ctx, userID)
	if err != nil {
		return err
	}

	matchedProfile, err := r.getUserProfile(ctx, matchedUserID)
	if err != nil {
		return err
	}

	matchScore := algorithm.MatchScore(userProfile, matchedProfile)

	query := `
		INSERT INTO match_history (user_id_a, user_id_b, algorithm_used, match_score, interaction_initiated)
		VALUES ($1, $2, 'cosine_similarity', $3, $4)
		ON CONFLICT (user_id_a, user_id_b, created_at) DO NOTHING
	`
	_, err = r.pool.Exec(ctx, query, userID, matchedUserID, matchScore, accepted)
	return err
}

// MutualMatch represents a user with mutual connection
type MutualMatch struct {
	UserID      string
	DisplayName string
	Username    string
	AvatarURL   string
	Bio         string
}

// GetMutualMatches retrieves users who have mutually accepted each other
func (r *Repository) GetMutualMatches(ctx context.Context, userID string, limit int) ([]MutualMatch, error) {
	// Find pairs where both users have accepted each other
	query := `
		SELECT DISTINCT
			u.id,
			u.display_name,
			u.username,
			COALESCE(u.avatar_url, '') as avatar_url,
			COALESCE(u.bio, '') as bio
		FROM match_history mh1
		JOIN match_history mh2 ON mh1.user_id_a = mh2.user_id_b AND mh1.user_id_b = mh2.user_id_a
		JOIN users u ON (
			CASE WHEN mh1.user_id_a = $1 THEN mh1.user_id_b ELSE mh1.user_id_a END = u.id
		)
		WHERE 
			(mh1.user_id_a = $1 OR mh1.user_id_b = $1)
			AND mh1.interaction_initiated = true
			AND mh2.interaction_initiated = true
			AND u.is_active = true
		ORDER BY u.display_name
		LIMIT $2
	`

	rows, err := r.pool.Query(ctx, query, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var matches []MutualMatch
	for rows.Next() {
		var m MutualMatch
		err := rows.Scan(&m.UserID, &m.DisplayName, &m.Username, &m.AvatarURL, &m.Bio)
		if err != nil {
			return nil, err
		}
		matches = append(matches, m)
	}

	return matches, rows.Err()
}

// AreConnected checks if two users have mutually accepted each other
func (r *Repository) AreConnected(ctx context.Context, userA, userB string) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1
			FROM match_history mh1
			JOIN match_history mh2 ON mh1.user_id_a = mh2.user_id_b AND mh1.user_id_b = mh2.user_id_a
			WHERE 
				mh1.user_id_a = $1 AND mh1.user_id_b = $2
				AND mh1.interaction_initiated = true
				AND mh2.interaction_initiated = true
		)
	`
	var exists bool
	err := r.pool.QueryRow(ctx, query, userA, userB).Scan(&exists)
	return exists, err
}
