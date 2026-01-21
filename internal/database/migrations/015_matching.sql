-- +goose Up
-- This materialized view transforms the user_skills data into a dense vector format
-- that is ready for mathematical comparison.
CREATE MATERIALIZED VIEW user_skill_vectors AS
WITH all_skills AS (
  -- Get a distinct, ordered list of all skills in the catalog
  SELECT id, name, ROW_NUMBER() OVER (ORDER BY name) as rn
  FROM skills
),
     user_skills_ranked AS (
       -- For each user, get their skills and join them with the ranked skill list
       SELECT
         us.user_id,
         ask.rn AS skill_index,
         us.proficiency,
         us.skill_type
       FROM user_skills us
              JOIN all_skills ask ON us.skill_id = ask.id
     )
-- The final vector generation using array aggregation
SELECT
  uv.user_id,
  -- Create an array for skills offered. The index is the skill's rank, and the value is the proficiency.
  -- This creates a sparse array, e.g., {..., 15:8, ...} where skill #15 has proficiency 8.
  ARRAY_AGG(uv.proficiency) FILTER (WHERE uv.skill_type = 'offered') AS offered_vector,

    -- Create another array for skills wanted.
  ARRAY_AGG(uv.proficiency) FILTER (WHERE uv.skill_type = 'wanted') AS wanted_vector

    -- Vector embedding support requires pgvector extension
    -- NULL::VECTOR(768) AS embedding_vector

FROM user_skills_ranked uv
GROUP BY uv.user_id;

-- Create indexes for fast lookups
CREATE UNIQUE INDEX idx_user_skill_vectors_user_id ON user_skill_vectors (user_id);

  CREATE TABLE match_history (
                               id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
                               user_id_a UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
                               user_id_b UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,

                               algorithm_used VARCHAR(50) NOT NULL,
                               match_score NUMERIC(5, 4) NOT NULL,

    -- Did the match result in an interaction?
                               interaction_initiated BOOLEAN NOT NULL DEFAULT false,
                               session_id UUID REFERENCES user_sessions(id), -- If a session was scheduled

                               created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

                               UNIQUE (user_id_a, user_id_b, created_at) -- Prevent duplicate logging
  );

CREATE INDEX idx_match_history_user_a ON match_history (user_id_a);
CREATE INDEX idx_match_history_user_b ON match_history (user_id_b);

CREATE TABLE recommendation_cache (
                                    user_id UUID NOT NULL,
                                    item_id TEXT NOT NULL,
                                    item_type VARCHAR(50) NOT NULL, -- 'user', 'skill'

                                    relevance_score NUMERIC(5, 4) NOT NULL,
                                    reason TEXT,

                                    generated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

                                    PRIMARY KEY (user_id, item_type, item_id)
);
