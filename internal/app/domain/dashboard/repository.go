package dashboard

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// DashboardStats contains all dashboard statistics for a user
type DashboardStats struct {
	Points   int
	Tier     string
	Matches  int
	Sessions int
}

// TopMatch represents a top matching user for the dashboard
type TopMatch struct {
	UserID      string
	DisplayName string
	Username    string
	AvatarURL   string
	Skills      string
	MatchScore  int // Percentage (0-100)
}

// ActivityItem represents a recent activity entry
type ActivityItem struct {
	Type      string // "match", "session", "message", "info"
	UserName  string
	Action    string
	TimeAgo   string
	Timestamp time.Time
}

// ProfileStrength contains profile completion data
type ProfileStrength struct {
	Percentage       int
	HasAvatar        bool
	HasBio           bool
	HasOfferedSkills bool // Has at least 3 offered skills
	HasWantedSkills  bool // Has at least 3 wanted skills
	HasSession       bool // Has completed at least one session
	HasSocialLinks   bool
}

// Repository handles database operations for dashboard data
type Repository struct {
	pool *pgxpool.Pool
}

// NewRepository creates a new dashboard repository
func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

// GetDashboardStats retrieves points, tier, match count, and session count
func (r *Repository) GetDashboardStats(ctx context.Context, userID string) (*DashboardStats, error) {
	stats := &DashboardStats{
		Tier: "Bronze",
	}

	// Get points and tier
	pointsQuery := `SELECT COALESCE(points, 0), COALESCE(tier, 'Bronze') FROM user_stats WHERE user_id = $1`
	err := r.pool.QueryRow(ctx, pointsQuery, userID).Scan(&stats.Points, &stats.Tier)
	if err != nil && err != pgx.ErrNoRows {
		return nil, err
	}

	// Get mutual matches count
	matchesQuery := `
		SELECT COUNT(DISTINCT 
			CASE 
				WHEN mh1.user_id_a = $1 THEN mh1.user_id_b 
				ELSE mh1.user_id_a 
			END
		)
		FROM match_history mh1
		JOIN match_history mh2 ON mh1.user_id_a = mh2.user_id_b AND mh1.user_id_b = mh2.user_id_a
		WHERE 
			(mh1.user_id_a = $1 OR mh1.user_id_b = $1)
			AND mh1.interaction_initiated = true
			AND mh2.interaction_initiated = true
	`
	err = r.pool.QueryRow(ctx, matchesQuery, userID).Scan(&stats.Matches)
	if err != nil && err != pgx.ErrNoRows {
		return nil, err
	}

	// Get sessions count (all non-cancelled sessions)
	sessionsQuery := `
		SELECT COUNT(*) FROM sessions 
		WHERE (initiator_id = $1 OR partner_id = $1)
		AND status != 'cancelled'
	`
	err = r.pool.QueryRow(ctx, sessionsQuery, userID).Scan(&stats.Sessions)
	if err != nil && err != pgx.ErrNoRows {
		return nil, err
	}

	return stats, nil
}

// GetTopMatches retrieves the top 3 pending matches with highest scores
func (r *Repository) GetTopMatches(ctx context.Context, userID string, limit int) ([]TopMatch, error) {
	// Get users who haven't been actioned yet but have overlapping skills
	query := `
		WITH user_offered AS (
			SELECT skill_id FROM user_skills WHERE user_id = $1 AND skill_type = 'offered'
		),
		user_wanted AS (
			SELECT skill_id FROM user_skills WHERE user_id = $1 AND skill_type = 'wanted'
		),
		potential_matches AS (
			SELECT DISTINCT 
				u.id,
				u.display_name,
				u.username,
				COALESCE(u.avatar_url, '') as avatar_url,
				-- Calculate overlap score
				(
					SELECT COUNT(*) FROM user_skills us 
					WHERE us.user_id = u.id 
					AND us.skill_type = 'offered' 
					AND us.skill_id IN (SELECT skill_id FROM user_wanted)
				) + (
					SELECT COUNT(*) FROM user_skills us 
					WHERE us.user_id = u.id 
					AND us.skill_type = 'wanted' 
					AND us.skill_id IN (SELECT skill_id FROM user_offered)
				) as overlap_count
			FROM users u
			INNER JOIN user_skills us ON u.id = us.user_id
			WHERE u.id != $1 
			AND u.is_active = true
			-- Exclude users already actioned
			AND NOT EXISTS (
				SELECT 1 FROM match_history mh 
				WHERE mh.user_id_a = $1 AND mh.user_id_b = u.id
			)
		),
		matched_skills AS (
			SELECT 
				pm.id as user_id,
				string_agg(DISTINCT s.name, ', ' ORDER BY s.name) as skills
			FROM potential_matches pm
			JOIN user_skills us ON pm.id = us.user_id AND us.skill_type = 'offered'
			JOIN skills s ON us.skill_id = s.id
			GROUP BY pm.id
		)
		SELECT 
			pm.id,
			pm.display_name,
			pm.username,
			pm.avatar_url,
			COALESCE(ms.skills, ''),
			-- Normalize overlap to percentage (max 100)
			LEAST(pm.overlap_count * 20 + 50, 99) as match_score
		FROM potential_matches pm
		LEFT JOIN matched_skills ms ON pm.id = ms.user_id
		WHERE pm.overlap_count > 0
		ORDER BY pm.overlap_count DESC, pm.display_name
		LIMIT $2
	`

	rows, err := r.pool.Query(ctx, query, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var matches []TopMatch
	for rows.Next() {
		var m TopMatch
		err := rows.Scan(&m.UserID, &m.DisplayName, &m.Username, &m.AvatarURL, &m.Skills, &m.MatchScore)
		if err != nil {
			return nil, err
		}
		matches = append(matches, m)
	}

	return matches, rows.Err()
}

// GetRecentActivity retrieves recent activity for the user
func (r *Repository) GetRecentActivity(ctx context.Context, userID string, limit int) ([]ActivityItem, error) {
	var activities []ActivityItem

	// Get recent pending matches (people who want to learn from user)
	matchQuery := `
		SELECT 
			u.display_name,
			'wants to connect with you',
			mh.created_at
		FROM match_history mh
		JOIN users u ON mh.user_id_a = u.id
		WHERE mh.user_id_b = $1 
		AND mh.interaction_initiated = true
		AND NOT EXISTS (
			SELECT 1 FROM match_history mh2 
			WHERE mh2.user_id_a = $1 AND mh2.user_id_b = mh.user_id_a
		)
		ORDER BY mh.created_at DESC
		LIMIT 2
	`
	rows, err := r.pool.Query(ctx, matchQuery, userID)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var a ActivityItem
		a.Type = "match"
		err := rows.Scan(&a.UserName, &a.Action, &a.Timestamp)
		if err != nil {
			rows.Close()
			return nil, err
		}
		a.TimeAgo = formatTimeAgo(a.Timestamp)
		activities = append(activities, a)
	}
	rows.Close()

	// Get recent completed sessions
	sessionQuery := `
		SELECT 
			CASE WHEN s.initiator_id = $1 THEN u2.display_name ELSE u1.display_name END,
			'completed a session with you',
			s.actual_end
		FROM sessions s
		JOIN users u1 ON s.initiator_id = u1.id
		JOIN users u2 ON s.partner_id = u2.id
		WHERE (s.initiator_id = $1 OR s.partner_id = $1)
		AND s.status = 'completed'
		AND s.actual_end IS NOT NULL
		ORDER BY s.actual_end DESC
		LIMIT 2
	`
	rows2, err := r.pool.Query(ctx, sessionQuery, userID)
	if err != nil {
		return nil, err
	}
	for rows2.Next() {
		var a ActivityItem
		var ts *time.Time
		a.Type = "session"
		err := rows2.Scan(&a.UserName, &a.Action, &ts)
		if err != nil {
			rows2.Close()
			return nil, err
		}
		if ts != nil {
			a.Timestamp = *ts
			a.TimeAgo = formatTimeAgo(*ts)
		} else {
			a.TimeAgo = "Recently"
		}
		activities = append(activities, a)
	}
	rows2.Close()

	// Get recent messages received
	messageQuery := `
		SELECT 
			u.display_name,
			'sent you a message',
			m.created_at
		FROM messages m
		JOIN users u ON m.sender_id = u.id
		WHERE m.receiver_id = $1
		ORDER BY m.created_at DESC
		LIMIT 2
	`
	rows3, err := r.pool.Query(ctx, messageQuery, userID)
	if err != nil {
		return nil, err
	}
	for rows3.Next() {
		var a ActivityItem
		a.Type = "message"
		err := rows3.Scan(&a.UserName, &a.Action, &a.Timestamp)
		if err != nil {
			rows3.Close()
			return nil, err
		}
		a.TimeAgo = formatTimeAgo(a.Timestamp)
		activities = append(activities, a)
	}
	rows3.Close()

	// If no activity, add a default info item
	if len(activities) == 0 {
		activities = append(activities, ActivityItem{
			Type:     "info",
			UserName: "Welcome",
			Action:   "Start by adding skills you can teach and want to learn!",
			TimeAgo:  "Just now",
		})
	}

	// Sort by timestamp and limit
	// For simplicity, we'll just take first 'limit' items
	if len(activities) > limit {
		activities = activities[:limit]
	}

	return activities, nil
}

// GetProfileStrength calculates the profile completion percentage
func (r *Repository) GetProfileStrength(ctx context.Context, userID string) (*ProfileStrength, error) {
	ps := &ProfileStrength{}

	// Check user profile fields
	profileQuery := `
		SELECT 
			COALESCE(avatar_url, '') != '' as has_avatar,
			COALESCE(bio, '') != '' as has_bio,
			(social_links->>'github' IS NOT NULL AND social_links->>'github' != '') OR
			(social_links->>'linkedin' IS NOT NULL AND social_links->>'linkedin' != '') OR
			(social_links->>'twitter' IS NOT NULL AND social_links->>'twitter' != '') OR
			(social_links->>'website' IS NOT NULL AND social_links->>'website' != '') as has_social
		FROM users
		WHERE id = $1
	`
	err := r.pool.QueryRow(ctx, profileQuery, userID).Scan(&ps.HasAvatar, &ps.HasBio, &ps.HasSocialLinks)
	if err != nil && err != pgx.ErrNoRows {
		return nil, err
	}

	// Check skill counts
	skillsQuery := `
		SELECT 
			COUNT(*) FILTER (WHERE skill_type = 'offered') >= 3,
			COUNT(*) FILTER (WHERE skill_type = 'wanted') >= 3
		FROM user_skills
		WHERE user_id = $1
	`
	err = r.pool.QueryRow(ctx, skillsQuery, userID).Scan(&ps.HasOfferedSkills, &ps.HasWantedSkills)
	if err != nil && err != pgx.ErrNoRows {
		return nil, err
	}

	// Check if user has completed at least one session
	sessionQuery := `
		SELECT EXISTS(
			SELECT 1 FROM sessions 
			WHERE (initiator_id = $1 OR partner_id = $1) 
			AND status = 'completed'
		)
	`
	err = r.pool.QueryRow(ctx, sessionQuery, userID).Scan(&ps.HasSession)
	if err != nil && err != pgx.ErrNoRows {
		return nil, err
	}

	// Calculate percentage (each item is ~17%)
	count := 0
	if ps.HasAvatar {
		count++
	}
	if ps.HasBio {
		count++
	}
	if ps.HasOfferedSkills {
		count++
	}
	if ps.HasWantedSkills {
		count++
	}
	if ps.HasSession {
		count++
	}
	if ps.HasSocialLinks {
		count++
	}

	// 6 items total, each worth ~17%
	ps.Percentage = (count * 100) / 6
	if ps.Percentage > 100 {
		ps.Percentage = 100
	}

	return ps, nil
}

// formatTimeAgo converts a timestamp to a human-readable "time ago" string
func formatTimeAgo(t time.Time) string {
	now := time.Now()
	diff := now.Sub(t)

	switch {
	case diff < time.Minute:
		return "Just now"
	case diff < time.Hour:
		mins := int(diff.Minutes())
		if mins == 1 {
			return "1 minute ago"
		}
		return fmt.Sprintf("%d minutes ago", mins)
	case diff < 24*time.Hour:
		hours := int(diff.Hours())
		if hours == 1 {
			return "1 hour ago"
		}
		return fmt.Sprintf("%d hours ago", hours)
	case diff < 48*time.Hour:
		return "Yesterday"
	case diff < 7*24*time.Hour:
		days := int(diff.Hours() / 24)
		return fmt.Sprintf("%d days ago", days)
	case diff < 30*24*time.Hour:
		weeks := int(diff.Hours() / (24 * 7))
		if weeks == 1 {
			return "1 week ago"
		}
		return fmt.Sprintf("%d weeks ago", weeks)
	default:
		return t.Format("Jan 2, 2006")
	}
}
