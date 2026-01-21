-- +goose Up
CREATE TYPE review_status AS ENUM ('pending', 'accepted', 'in_progress', 'completed', 'revision_requested', 'declined', 'cancelled');
CREATE TYPE review_type AS ENUM ('code_review', 'design_review', 'writing_review', 'video_review', 'portfolio_review', 'resume_review');
CREATE TYPE review_depth AS ENUM ('quick', 'standard', 'comprehensive');

CREATE TABLE reviews (
                       id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
                       requester_id UUID NOT NULL REFERENCES users(id) ON DELETE SET NULL,
                       reviewer_id UUID REFERENCES users(id) ON DELETE SET NULL,
                       skill_id UUID NOT NULL REFERENCES skills(id),

  -- Request Details
                       type review_type NOT NULL,
                       depth review_depth NOT NULL,
                       title TEXT NOT NULL,
                       description TEXT,

  -- Storing specific questions as a text array is simple and effective.
                       specific_questions TEXT[],

  -- Pricing and Payment
                       price NUMERIC(10, 2) NOT NULL,
                       payment_id VARCHAR(255), -- Stripe PaymentIntent ID or similar

  -- Timeline
                       requested_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
                       accepted_at TIMESTAMPTZ,
                       deadline TIMESTAMPTZ,
                       completed_at TIMESTAMPTZ,

  -- Status & Workflow
                       status review_status NOT NULL DEFAULT 'pending',
                       decline_reason TEXT, -- Populated if a reviewer declines

  -- Overall review feedback (summary)
                       feedback_summary TEXT,
                       video_feedback_url TEXT,

  -- Ratings
                       requester_rating SMALLINT CHECK (requester_rating >= 1 AND requester_rating <= 5),
                       requester_comment TEXT,
                       reviewer_rating SMALLINT CHECK (reviewer_rating >= 1 AND reviewer_rating <= 5),
                       reviewer_comment TEXT
);

-- Indexes for common query patterns
CREATE INDEX idx_reviews_requester_id ON reviews (requester_id);
CREATE INDEX idx_reviews_reviewer_id ON reviews (reviewer_id);
CREATE INDEX idx_reviews_status ON reviews (status);

CREATE TABLE review_content (
                              id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
                              review_id UUID NOT NULL REFERENCES reviews(id) ON DELETE CASCADE,

  -- Corresponds to the common.v1.Attachment message
                              file_name VARCHAR(255) NOT NULL,
                              url TEXT NOT NULL, -- URL to the file in cloud storage (e.g., S3, Google Cloud Storage)
                              mime_type VARCHAR(100),
                              size_bytes BIGINT,

                              uploaded_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE review_feedback_sections (
                                        id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
                                        review_id UUID NOT NULL REFERENCES reviews(id) ON DELETE CASCADE,

                                        title VARCHAR(255) NOT NULL, -- e.g., "Code Quality", "Architecture"
                                        content TEXT NOT NULL,
                                        rating SMALLINT CHECK (rating >= 1 AND rating <= 5),

  -- Storing lists as text arrays is efficient for this use case.
                                        strengths TEXT[],
                                        improvements TEXT[],
                                        resources TEXT[] -- e.g., list of URLs
);

CREATE TABLE review_revisions (
                                id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
                                review_id UUID NOT NULL REFERENCES reviews(id) ON DELETE CASCADE,
                                requester_id UUID NOT NULL REFERENCES users(id),

                                revision_notes TEXT NOT NULL,
                                specific_areas TEXT[], -- e.g., {"Code style", "Scalability concerns"}

                                new_deadline TIMESTAMPTZ,
                                requested_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

