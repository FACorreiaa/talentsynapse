package skills

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// SkillType represents whether a skill is offered or wanted
type SkillType string

const (
	SkillTypeOffered SkillType = "offered"
	SkillTypeWanted  SkillType = "wanted"
)

// Skill represents a skill in the catalog
type Skill struct {
	ID           string
	Name         string
	CategoryID   string
	CategoryName string
	Description  string
}

// UserSkill represents a user's skill with proficiency
type UserSkill struct {
	SkillID      string
	SkillName    string
	CategoryName string
	SkillType    SkillType
	Proficiency  int
}

// Repository handles database operations for skills
type Repository struct {
	pool *pgxpool.Pool
}

// NewRepository creates a new skills repository
func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

// GetUserSkills retrieves all skills for a user (both offered and wanted)
func (r *Repository) GetUserSkills(ctx context.Context, userID string) ([]UserSkill, error) {
	query := `
		SELECT 
			us.skill_id,
			s.name AS skill_name,
			sc.name AS category_name,
			us.skill_type,
			COALESCE(us.proficiency, 5) AS proficiency
		FROM user_skills us
		JOIN skills s ON us.skill_id = s.id
		JOIN skill_categories sc ON s.category_id = sc.id
		WHERE us.user_id = $1
		ORDER BY us.skill_type, s.name
	`

	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var skills []UserSkill
	for rows.Next() {
		var skill UserSkill
		err := rows.Scan(
			&skill.SkillID,
			&skill.SkillName,
			&skill.CategoryName,
			&skill.SkillType,
			&skill.Proficiency,
		)
		if err != nil {
			return nil, err
		}
		skills = append(skills, skill)
	}

	return skills, rows.Err()
}

// GetAllSkills retrieves all available skills from the catalog
func (r *Repository) GetAllSkills(ctx context.Context) ([]Skill, error) {
	query := `
		SELECT 
			s.id,
			s.name,
			s.category_id,
			sc.name AS category_name,
			COALESCE(s.description, '') AS description
		FROM skills s
		JOIN skill_categories sc ON s.category_id = sc.id
		ORDER BY sc.name, s.name
	`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var skills []Skill
	for rows.Next() {
		var skill Skill
		err := rows.Scan(
			&skill.ID,
			&skill.Name,
			&skill.CategoryID,
			&skill.CategoryName,
			&skill.Description,
		)
		if err != nil {
			return nil, err
		}
		skills = append(skills, skill)
	}

	return skills, rows.Err()
}

// AddUserSkill adds a skill to a user's profile
func (r *Repository) AddUserSkill(ctx context.Context, userID, skillID string, skillType SkillType, proficiency int) error {
	query := `
		INSERT INTO user_skills (user_id, skill_id, skill_type, proficiency)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (user_id, skill_id, skill_type) 
		DO UPDATE SET proficiency = $4, updated_at = NOW()
	`
	_, err := r.pool.Exec(ctx, query, userID, skillID, skillType, proficiency)
	return err
}

// RemoveUserSkill removes a skill from a user's profile
func (r *Repository) RemoveUserSkill(ctx context.Context, userID, skillID string, skillType SkillType) error {
	query := `DELETE FROM user_skills WHERE user_id = $1 AND skill_id = $2 AND skill_type = $3`
	_, err := r.pool.Exec(ctx, query, userID, skillID, skillType)
	return err
}

// GetSkillByID retrieves a skill by its ID
func (r *Repository) GetSkillByID(ctx context.Context, skillID string) (*Skill, error) {
	query := `
		SELECT 
			s.id,
			s.name,
			s.category_id,
			sc.name AS category_name,
			COALESCE(s.description, '') AS description
		FROM skills s
		JOIN skill_categories sc ON s.category_id = sc.id
		WHERE s.id = $1
	`

	var skill Skill
	err := r.pool.QueryRow(ctx, query, skillID).Scan(
		&skill.ID,
		&skill.Name,
		&skill.CategoryID,
		&skill.CategoryName,
		&skill.Description,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &skill, nil
}
