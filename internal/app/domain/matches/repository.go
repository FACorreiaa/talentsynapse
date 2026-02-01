package matches

import (
	"context"

	"github.com/FACorreiaa/talentsynapse/internal/app/domain/matches/algorithm"
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
		overlapSkills, err := r.GetOverlapSkills(ctx, userID, result.UserID)
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
// This queries user_skills directly to ensure fresh data (not stale materialized view)
func (r *Repository) getUserProfile(ctx context.Context, userID string) (algorithm.UserProfile, error) {
	// First get user info
	userQuery := `
		SELECT
			u.id,
			u.display_name,
			u.username,
			COALESCE(u.avatar_url, '') as avatar_url
		FROM users u
		WHERE u.id = $1 AND u.is_active = true
	`

	var profile algorithm.UserProfile
	err := r.pool.QueryRow(ctx, userQuery, userID).Scan(
		&profile.UserID,
		&profile.DisplayName,
		&profile.Username,
		&profile.AvatarURL,
	)
	if err != nil {
		return algorithm.UserProfile{}, err
	}

	// Get total skill count to determine vector dimension
	var skillCount int
	err = r.pool.QueryRow(ctx, "SELECT COUNT(*) FROM skills").Scan(&skillCount)
	if err != nil {
		return algorithm.UserProfile{}, err
	}

	// Initialize zero vectors with proper dimension
	profile.OfferedVector = make([]float64, skillCount)
	profile.WantedVector = make([]float64, skillCount)

	// Query user's skills with their position in the global skill list
	skillsQuery := `
		WITH ranked_skills AS (
			SELECT id, ROW_NUMBER() OVER (ORDER BY name) - 1 as skill_index
			FROM skills
		)
		SELECT 
			rs.skill_index,
			us.skill_type,
			COALESCE(us.proficiency, 5) as proficiency
		FROM user_skills us
		JOIN ranked_skills rs ON us.skill_id = rs.id
		WHERE us.user_id = $1
	`

	rows, err := r.pool.Query(ctx, skillsQuery, userID)
	if err != nil {
		return algorithm.UserProfile{}, err
	}
	defer rows.Close()

	for rows.Next() {
		var skillIndex int
		var skillType string
		var proficiency int

		if err := rows.Scan(&skillIndex, &skillType, &proficiency); err != nil {
			return algorithm.UserProfile{}, err
		}

		// Place proficiency at the correct index in the vector
		if skillIndex < skillCount {
			switch skillType {
			case "offered":
				profile.OfferedVector[skillIndex] = float64(proficiency)
			case "wanted":
				profile.WantedVector[skillIndex] = float64(proficiency)
			}
		}
	}

	if err := rows.Err(); err != nil {
		return algorithm.UserProfile{}, err
	}

	profile.SkillCount = skillCount
	return profile, nil
}

// getAllCandidateProfiles retrieves all potential candidate users with their skill vectors
// This queries user_skills directly to ensure fresh data
// Excludes users that the current user has already acted on (accepted or rejected)
func (r *Repository) getAllCandidateProfiles(ctx context.Context, excludeUserID string) ([]algorithm.UserProfile, error) {
	// Get total skill count to determine vector dimension
	var skillCount int
	err := r.pool.QueryRow(ctx, "SELECT COUNT(*) FROM skills").Scan(&skillCount)
	if err != nil {
		return nil, err
	}

	// Get all active users except:
	// 1. The current user
	// 2. Users the current user has already acted on (accepted or rejected)
	usersQuery := `
		SELECT DISTINCT
			u.id,
			u.display_name,
			u.username,
			COALESCE(u.avatar_url, '') as avatar_url
		FROM users u
		INNER JOIN user_skills us ON u.id = us.user_id
		WHERE u.id != $1 
		AND u.is_active = true
		AND NOT EXISTS (
			SELECT 1 FROM match_history mh 
			WHERE mh.user_id_a = $1 AND mh.user_id_b = u.id
		)
	`

	rows, err := r.pool.Query(ctx, usersQuery, excludeUserID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var candidates []algorithm.UserProfile
	for rows.Next() {
		var profile algorithm.UserProfile
		err := rows.Scan(
			&profile.UserID,
			&profile.DisplayName,
			&profile.Username,
			&profile.AvatarURL,
		)
		if err != nil {
			return nil, err
		}
		profile.SkillCount = skillCount
		candidates = append(candidates, profile)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	// Now populate skill vectors for each candidate
	skillsQuery := `
		WITH ranked_skills AS (
			SELECT id, ROW_NUMBER() OVER (ORDER BY name) - 1 as skill_index
			FROM skills
		)
		SELECT 
			us.user_id,
			rs.skill_index,
			us.skill_type,
			COALESCE(us.proficiency, 5) as proficiency
		FROM user_skills us
		JOIN ranked_skills rs ON us.skill_id = rs.id
		WHERE us.user_id = ANY($1)
	`

	// Collect user IDs
	userIDs := make([]string, len(candidates))
	userIndexMap := make(map[string]int)
	for i, c := range candidates {
		userIDs[i] = c.UserID
		userIndexMap[c.UserID] = i
		// Initialize vectors
		candidates[i].OfferedVector = make([]float64, skillCount)
		candidates[i].WantedVector = make([]float64, skillCount)
	}

	if len(userIDs) == 0 {
		return candidates, nil
	}

	skillRows, err := r.pool.Query(ctx, skillsQuery, userIDs)
	if err != nil {
		return nil, err
	}
	defer skillRows.Close()

	for skillRows.Next() {
		var userID string
		var skillIndex int
		var skillType string
		var proficiency int

		if err := skillRows.Scan(&userID, &skillIndex, &skillType, &proficiency); err != nil {
			return nil, err
		}

		if idx, ok := userIndexMap[userID]; ok && skillIndex < skillCount {
			switch skillType {
			case "offered":
				candidates[idx].OfferedVector[skillIndex] = float64(proficiency)
			case "wanted":
				candidates[idx].WantedVector[skillIndex] = float64(proficiency)
			}
		}
	}

	return candidates, skillRows.Err()
}

// GetOverlapSkills retrieves the skills that overlap between two users
func (r *Repository) GetOverlapSkills(ctx context.Context, userID, matchedUserID string) ([]OverlapSkill, error) {
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

// UserInfo represents basic user information
type UserInfo struct {
	UserID      string
	DisplayName string
	Username    string
	AvatarURL   string
}

// GetUserInfo retrieves basic user information
func (r *Repository) GetUserInfo(ctx context.Context, userID string) (*UserInfo, error) {
	query := `
		SELECT id, display_name, username, avatar_url
		FROM users
		WHERE id = $1
	`
	var user UserInfo
	err := r.pool.QueryRow(ctx, query, userID).Scan(
		&user.UserID,
		&user.DisplayName,
		&user.Username,
		&user.AvatarURL,
	)
	if err != nil {
		return nil, err
	}
	return &user, nil
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
	// Find users where both have accepted each other
	// We look for any record where userA accepted userB AND userB accepted userA
	query := `
		SELECT DISTINCT
			u.id,
			u.display_name,
			u.username,
			COALESCE(u.avatar_url, '') as avatar_url,
			COALESCE(u.bio, '') as bio
		FROM users u
		WHERE u.id != $1
		AND u.is_active = true
		-- Current user accepted them
		AND EXISTS (
			SELECT 1 FROM match_history mh1 
			WHERE mh1.user_id_a = $1 
			AND mh1.user_id_b = u.id 
			AND mh1.interaction_initiated = true
		)
		-- They accepted current user
		AND EXISTS (
			SELECT 1 FROM match_history mh2 
			WHERE mh2.user_id_a = u.id 
			AND mh2.user_id_b = $1 
			AND mh2.interaction_initiated = true
		)
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
			SELECT 1 WHERE
			-- userA accepted userB
			EXISTS (
				SELECT 1 FROM match_history 
				WHERE user_id_a = $1 AND user_id_b = $2 AND interaction_initiated = true
			)
			AND
			-- userB accepted userA
			EXISTS (
				SELECT 1 FROM match_history 
				WHERE user_id_a = $2 AND user_id_b = $1 AND interaction_initiated = true
			)
		)
	`
	var exists bool
	err := r.pool.QueryRow(ctx, query, userA, userB).Scan(&exists)
	return exists, err
}

// CountMutualMatches counts the number of mutual matches for a user
func (r *Repository) CountMutualMatches(ctx context.Context, userID string) (int, error) {
	query := `
		SELECT COUNT(DISTINCT 
			CASE 
				WHEN mh1.user_id_a = $1 THEN mh1.user_id_b 
				ELSE mh1.user_id_a 
			END
		)
		FROM match_history mh1
		JOIN match_history mh2 ON mh1.user_id_a = mh2.user_id_b AND mh1.user_id_b = mh2.user_id_a
		WHERE (mh1.user_id_a = $1 OR mh1.user_id_b = $1)
		AND mh1.interaction_initiated = true
		AND mh2.interaction_initiated = true
	`
	var count int
	err := r.pool.QueryRow(ctx, query, userID).Scan(&count)
	return count, err
}

// PendingMatch represents a user who has expressed interest but hasn't been responded to
type PendingMatch struct {
	UserID      string
	DisplayName string
	Username    string
	AvatarURL   string
	MatchScore  float64
}

// GetPendingReceivedMatches retrieves users who have accepted the current user but the current user hasn't responded
func (r *Repository) GetPendingReceivedMatches(ctx context.Context, userID string, limit int) ([]PendingMatch, error) {
	query := `
		SELECT DISTINCT
			u.id,
			u.display_name,
			u.username,
			COALESCE(u.avatar_url, '') as avatar_url,
			COALESCE(mh.match_score, 0) as match_score
		FROM match_history mh
		JOIN users u ON mh.user_id_a = u.id
		WHERE mh.user_id_b = $1 
		AND mh.interaction_initiated = true
		AND u.is_active = true
		-- Current user hasn't responded yet
		AND NOT EXISTS (
			SELECT 1 FROM match_history mh2 
			WHERE mh2.user_id_a = $1 AND mh2.user_id_b = u.id
		)
		ORDER BY mh.match_score DESC
		LIMIT $2
	`

	rows, err := r.pool.Query(ctx, query, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var matches []PendingMatch
	for rows.Next() {
		var m PendingMatch
		err := rows.Scan(&m.UserID, &m.DisplayName, &m.Username, &m.AvatarURL, &m.MatchScore)
		if err != nil {
			return nil, err
		}
		matches = append(matches, m)
	}

	return matches, rows.Err()
}

// GetConnectionCount returns the number of mutual connections for a user
func (r *Repository) GetConnectionCount(ctx context.Context, userID string) (int, error) {
	query := `
		SELECT COUNT(DISTINCT u.id)
		FROM users u
		WHERE u.id != $1
		AND u.is_active = true
		-- Current user accepted them
		AND EXISTS (
			SELECT 1 FROM match_history mh1 
			WHERE mh1.user_id_a = $1 
			AND mh1.user_id_b = u.id 
			AND mh1.interaction_initiated = true
		)
		-- They accepted current user
		AND EXISTS (
			SELECT 1 FROM match_history mh2 
			WHERE mh2.user_id_a = u.id 
			AND mh2.user_id_b = $1 
			AND mh2.interaction_initiated = true
		)
	`
	var count int
	err := r.pool.QueryRow(ctx, query, userID).Scan(&count)
	return count, err
}

// HasConnections checks if a user has any mutual connections
func (r *Repository) HasConnections(ctx context.Context, userID string) (bool, error) {
	count, err := r.GetConnectionCount(ctx, userID)
	return count > 0, err
}

// GetPendingSentMatches retrieves users that the current user has liked but who haven't responded yet
func (r *Repository) GetPendingSentMatches(ctx context.Context, userID string, limit int) ([]PendingMatch, error) {
	query := `
		SELECT DISTINCT
			u.id,
			u.display_name,
			u.username,
			COALESCE(u.avatar_url, '') as avatar_url,
			COALESCE(mh.match_score, 0) as match_score
		FROM match_history mh
		JOIN users u ON mh.user_id_b = u.id
		WHERE mh.user_id_a = $1 
		AND mh.interaction_initiated = true
		AND u.is_active = true
		-- They haven't responded yet (no record from them to current user with acceptance)
		AND NOT EXISTS (
			SELECT 1 FROM match_history mh2 
			WHERE mh2.user_id_a = u.id AND mh2.user_id_b = $1 AND mh2.interaction_initiated = true
		)
		ORDER BY mh.match_score DESC
		LIMIT $2
	`

	rows, err := r.pool.Query(ctx, query, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var matches []PendingMatch
	for rows.Next() {
		var m PendingMatch
		err := rows.Scan(&m.UserID, &m.DisplayName, &m.Username, &m.AvatarURL, &m.MatchScore)
		if err != nil {
			return nil, err
		}
		matches = append(matches, m)
	}

	return matches, rows.Err()
}
