-- +goose Up
CREATE TABLE badges (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4 (),
    code VARCHAR(50) UNIQUE NOT NULL, -- e.g., 'first_match', 'top_teacher'
    name VARCHAR(100) NOT NULL,
    description TEXT,
    icon_url VARCHAR(255),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE user_badges (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4 (),
    user_id UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    badge_id UUID NOT NULL REFERENCES badges (id) ON DELETE CASCADE,
    awarded_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (user_id, badge_id)
);

-- Seed initial badges
INSERT INTO
    badges (
        code,
        name,
        description,
        icon_url
    )
VALUES (
        'early_adopter',
        'Early Adopter',
        'Joined SkillSphere in the early days.',
        '🚀'
    ),
    (
        'first_match',
        'First Connection',
        'Successfully connected with another user.',
        '🤝'
    ),
    (
        'top_teacher',
        'Top Teacher',
        'Received high ratings for teaching skills.',
        '🎓'
    ),
    (
        'dedicated_learner',
        'Dedicated Learner',
        'Completed 5 learning sessions.',
        '📚'
    ),
    (
        'verified_expert',
        'Verified Expert',
        'Verified as an expert in a skill.',
        '✅'
    );