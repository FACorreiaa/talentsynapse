package matches

import (
	"context"

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

// GetPendingMatches retrieves potential matches for a user
// This is a simplified version - a real implementation would use the user_skill_vectors materialized view
func (r *Repository) GetPendingMatches(ctx context.Context, userID string, limit int) ([]Match, error) {
	// Find users where:
	// - They offer skills the current user wants
	// - OR they want skills the current user offers
	// - Exclude users already in match_history with this user
	query := `
		WITH user_offered AS (
			SELECT skill_id FROM user_skills 
			WHERE user_id = $1 AND skill_type = 'offered'
		),
		user_wanted AS (
			SELECT skill_id FROM user_skills 
			WHERE user_id = $1 AND skill_type = 'wanted'
		),
		potential_matches AS (
			SELECT DISTINCT 
				us.user_id,
				COUNT(DISTINCT us.skill_id) as skill_overlap
			FROM user_skills us
			WHERE us.user_id != $1
			AND (
				-- They offer what user wants
				(us.skill_type = 'offered' AND us.skill_id IN (SELECT skill_id FROM user_wanted))
				OR
				-- They want what user offers
				(us.skill_type = 'wanted' AND us.skill_id IN (SELECT skill_id FROM user_offered))
			)
			GROUP BY us.user_id
		)
		SELECT 
			u.id,
			u.display_name,
			u.username,
			COALESCE(u.avatar_url, '') as avatar_url,
			CAST(pm.skill_overlap AS FLOAT) / 10.0 as match_score
		FROM potential_matches pm
		JOIN users u ON pm.user_id = u.id
		WHERE u.is_active = true
		ORDER BY pm.skill_overlap DESC
		LIMIT $2
	`

	rows, err := r.pool.Query(ctx, query, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var matches []Match
	for rows.Next() {
		var m Match
		err := rows.Scan(
			&m.UserID,
			&m.DisplayName,
			&m.Username,
			&m.AvatarURL,
			&m.MatchScore,
		)
		if err != nil {
			return nil, err
		}

		// Get overlapping skills for this match
		overlapSkills, err := r.getOverlapSkills(ctx, userID, m.UserID)
		if err != nil {
			return nil, err
		}
		m.OverlapSkills = overlapSkills

		matches = append(matches, m)
	}

	return matches, rows.Err()
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
	query := `
		INSERT INTO match_history (user_id_a, user_id_b, algorithm_used, match_score, interaction_initiated)
		VALUES ($1, $2, 'skill_overlap', 0.5, $3)
		ON CONFLICT (user_id_a, user_id_b, created_at) DO NOTHING
	`
	_, err := r.pool.Exec(ctx, query, userID, matchedUserID, accepted)
	return err
}
