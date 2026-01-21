-- +goose Up
CREATE TYPE skill_type AS ENUM ('offered', 'wanted');

CREATE TABLE skill_categories (
                                id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
                                name VARCHAR(100) UNIQUE NOT NULL, -- e.g., "Technology", "Creative Arts", "Languages"
                                description TEXT,
                                created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE skills (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    category_id UUID NOT NULL REFERENCES skill_categories(id) ON DELETE RESTRICT,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    -- Vector embedding support requires pgvector extension
    -- embedding VECTOR(768),
    users_offering_count INT NOT NULL DEFAULT 0,
    users_wanting_count INT NOT NULL DEFAULT 0,
    total_sessions_count INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (category_id, name)
);

CREATE INDEX idx_skills_name_trgm ON skills USING GIN (name gin_trgm_ops);
-- CREATE INDEX idx_skills_embedding_cosine ON skills USING IVFFLAT (embedding vector_cosine_ops);
CREATE INDEX idx_skills_category_id ON skills (category_id);

CREATE TABLE user_skills (
                           user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
                           skill_id UUID NOT NULL REFERENCES skills(id) ON DELETE CASCADE,
                           skill_type skill_type NOT NULL,

  -- Proficiency on a scale of 1-10.
                           proficiency SMALLINT CHECK (proficiency >= 1 AND proficiency <= 10),

                           created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
                           updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

                           PRIMARY KEY (user_id, skill_id, skill_type)
);

-- Indexes for quickly finding skills for a user, or users for a skill.
CREATE INDEX idx_user_skills_user_id ON user_skills (user_id);
CREATE INDEX idx_user_skills_skill_id ON user_skills (skill_id);

CREATE TABLE skill_tags (
                          id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
                          name VARCHAR(50) UNIQUE NOT NULL -- e.g., "backend", "frontend", "data-analysis"
);

CREATE TABLE skill_to_tags (
                             skill_id UUID NOT NULL REFERENCES skills(id) ON DELETE CASCADE,
                             tag_id UUID NOT NULL REFERENCES skill_tags(id) ON DELETE CASCADE,
                             PRIMARY KEY (skill_id, tag_id)
);
