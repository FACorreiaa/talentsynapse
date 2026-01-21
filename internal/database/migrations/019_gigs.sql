-- +goose Up
CREATE TYPE gig_status AS ENUM ('open', 'in_progress', 'submitted', 'revision_requested', 'completed', 'disputed', 'cancelled');
CREATE TYPE gig_type AS ENUM ('tutoring', 'project', 'consulting', 'review');
CREATE TYPE proficiency_level AS ENUM ('beginner', 'intermediate', 'advanced', 'expert'); -- Assuming from common.proto

CREATE TABLE gigs (
                    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
                    creator_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,

                    title TEXT NOT NULL,
                    description TEXT NOT NULL,

                    type gig_type NOT NULL,
                    budget NUMERIC(10, 2) NOT NULL,
                    is_hourly BOOLEAN NOT NULL DEFAULT false,
                    estimated_hours INT,

                    required_proficiency proficiency_level,
  -- Storing simple requirements as a text array is efficient.
                    requirements TEXT[],

                    deadline TIMESTAMPTZ,
                    status gig_status NOT NULL DEFAULT 'open',

  -- This is populated once an application is accepted.
                    assigned_to_id UUID REFERENCES users(id) ON DELETE SET NULL,

  -- Denormalized count for performance, updated by triggers or application logic.
                    application_count INT NOT NULL DEFAULT 0,

                    completed_at TIMESTAMPTZ,

                    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
                    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Indexes for common query patterns in ListGigs/SearchGigs
CREATE INDEX idx_gigs_status ON gigs (status);
CREATE INDEX idx_gigs_creator_id ON gigs (creator_id);
CREATE INDEX idx_gigs_assigned_to_id ON gigs (assigned_to_id);
-- For full-text search, similar to the skills table
CREATE INDEX idx_gigs_title_description_trgm ON gigs USING GIN ((title || ' ' || description) gin_trgm_ops);

CREATE TABLE gig_skills (
                          gig_id UUID NOT NULL REFERENCES gigs(id) ON DELETE CASCADE,
                          skill_id UUID NOT NULL REFERENCES skills(id) ON DELETE CASCADE,
                          PRIMARY KEY (gig_id, skill_id)
);

CREATE TYPE application_status AS ENUM ('pending', 'accepted', 'rejected', 'withdrawn');

CREATE TABLE gig_applications (
                                id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
                                gig_id UUID NOT NULL REFERENCES gigs(id) ON DELETE CASCADE,
                                freelancer_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,

                                cover_letter TEXT NOT NULL,
                                proposed_amount NUMERIC(10, 2),
                                estimated_hours INT,

                                status application_status NOT NULL DEFAULT 'pending',

                                applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
                                responded_at TIMESTAMPTZ -- When the creator accepted/rejected
);

-- A freelancer can only apply to a gig once.
CREATE UNIQUE INDEX idx_gig_applications_unique ON gig_applications (gig_id, freelancer_id);
CREATE INDEX idx_gig_applications_freelancer_id ON gig_applications (freelancer_id);

CREATE TABLE gig_work_submissions (
                                    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
                                    gig_id UUID NOT NULL REFERENCES gigs(id) ON DELETE CASCADE,
                                    freelancer_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,

                                    description TEXT,

  -- Will be incremented for each new revision.
                                    revision_number INT NOT NULL DEFAULT 1,

                                    submitted_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE gig_deliverables (
                                id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
                                submission_id UUID NOT NULL REFERENCES gig_work_submissions(id) ON DELETE CASCADE,

  -- Corresponds to the common.v1.Attachment message
                                file_name VARCHAR(255) NOT NULL,
                                url TEXT NOT NULL, -- URL to the file in cloud storage (S3, GCS, etc.)
                                mime_type VARCHAR(100),
                                size_bytes BIGINT,

                                uploaded_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE gig_reviews (
                           gig_id UUID PRIMARY KEY REFERENCES gigs(id) ON DELETE CASCADE,
                           creator_id UUID NOT NULL REFERENCES users(id),
                           freelancer_id UUID NOT NULL REFERENCES users(id),

  -- Rating given by the creator to the freelancer
                           rating SMALLINT NOT NULL CHECK (rating >= 1 AND rating <= 5),
                           review TEXT,

                           created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
