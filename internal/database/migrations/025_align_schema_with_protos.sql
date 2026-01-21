-- +goose Up
-- Align enums and add tables so the schema matches the proto contracts and README features.

-- Normalize session_status values to match skillsphere.common.v1.SessionStatus
ALTER TYPE session_status RENAME TO session_status_old;
CREATE TYPE session_status AS ENUM ('scheduled', 'in_progress', 'completed', 'cancelled', 'no_show');
ALTER TABLE sessions
  ALTER COLUMN status DROP DEFAULT,
  ALTER COLUMN status TYPE session_status USING (
    CASE
      WHEN status::text IN ('pending', 'confirmed') THEN 'scheduled'::session_status
      WHEN status::text = 'disputed' THEN 'cancelled'::session_status
      ELSE status::text::session_status
    END
  ),
  ALTER COLUMN status SET DEFAULT 'scheduled';
DROP TYPE session_status_old;

-- Make proficiency_level consistent with skillsphere.common.v1.ProficiencyLevel
ALTER TYPE proficiency_level RENAME TO proficiency_level_old;
CREATE TYPE proficiency_level AS ENUM ('beginner', 'intermediate', 'expert');
ALTER TABLE gigs
  ALTER COLUMN required_proficiency TYPE proficiency_level USING (
    CASE
      WHEN required_proficiency::text = 'advanced' THEN 'expert'::proficiency_level
      ELSE required_proficiency::text::proficiency_level
    END
  );
DROP TYPE proficiency_level_old;

-- Rename subscription tier value so it lines up with proto-defined names
-- +goose StatementBegin
DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM pg_type t
        JOIN pg_enum e ON t.oid = e.enumtypid
        WHERE t.typname = 'subscription_tier'
          AND e.enumlabel = 'pro'
    ) THEN
        ALTER TYPE subscription_tier RENAME VALUE 'pro' TO 'professional';
    END IF;
END $$;
-- +goose StatementEnd

-- Align payment_status enum with skillsphere.common.v1.PaymentStatus
ALTER TYPE payment_status RENAME TO payment_status_old;
CREATE TYPE payment_status AS ENUM ('pending', 'completed', 'failed', 'refunded', 'held_in_escrow');
ALTER TABLE payments
  ALTER COLUMN status TYPE payment_status USING (
    CASE
      WHEN status::text IN ('pending', 'requires_action') THEN 'pending'::payment_status
      WHEN status::text = 'succeeded' THEN 'completed'::payment_status
      WHEN status::text = 'failed' THEN 'failed'::payment_status
      WHEN status::text = 'refunded' THEN 'refunded'::payment_status
      ELSE 'pending'::payment_status
    END
  ),
  ALTER COLUMN status SET DEFAULT 'pending';
DROP TYPE payment_status_old;

-- Add the missing subscription payment purpose used by PaymentPurpose
ALTER TYPE payment_purpose ADD VALUE IF NOT EXISTS 'subscription' BEFORE 'workshop';

-- Extend subscriptions and payments to include proto-visible fields
ALTER TABLE subscriptions
  ADD COLUMN IF NOT EXISTS monthly_amount NUMERIC(10, 2),
  ADD COLUMN IF NOT EXISTS is_active BOOLEAN NOT NULL DEFAULT true;

ALTER TABLE escrow_payments
  ADD COLUMN IF NOT EXISTS release_condition VARCHAR(100);

-- Payment method metadata backing skillsphere.payment.v1.PaymentMethodInfo
CREATE TYPE payment_method_type AS ENUM ('card', 'paypal', 'bank_transfer', 'crypto');

CREATE TABLE payment_methods (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  type payment_method_type NOT NULL,
  provider_payment_method_id VARCHAR(255),
  last_four VARCHAR(4),
  brand VARCHAR(50),
  exp_month SMALLINT,
  exp_year SMALLINT,
  is_default BOOLEAN NOT NULL DEFAULT false,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE (user_id, provider_payment_method_id)
);

CREATE TRIGGER update_payment_methods_updated_at
  BEFORE UPDATE ON payment_methods
  FOR EACH ROW
  EXECUTE FUNCTION set_updated_at();

ALTER TABLE payments
  ADD COLUMN IF NOT EXISTS payment_method payment_method_type,
  ADD COLUMN IF NOT EXISTS payment_method_id UUID REFERENCES payment_methods(id);

ALTER TABLE customers
  ADD COLUMN IF NOT EXISTS payment_methods_synced_at TIMESTAMPTZ,
  ADD COLUMN IF NOT EXISTS default_payment_method_uuid UUID REFERENCES payment_methods(id);

-- Search infrastructure backing SearchService
CREATE TABLE user_search_documents (
  user_id UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
  document TSVECTOR NOT NULL,
  skill_ids UUID[] NOT NULL DEFAULT '{}',
  skill_names TEXT[] NOT NULL DEFAULT '{}',
  -- Geometry and vector support requires PostGIS and pgvector extensions
  -- location_geom GEOMETRY(Point, 4326),
  -- embedding VECTOR(768),
  latitude DOUBLE PRECISION,
  longitude DOUBLE PRECISION,
  subscription_tier subscription_tier NOT NULL DEFAULT 'free',
  average_rating NUMERIC(3, 2) NOT NULL DEFAULT 0,
  total_sessions INT NOT NULL DEFAULT 0,
  total_reviews INT NOT NULL DEFAULT 0,
  indexed_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_user_search_documents_tsv ON user_search_documents USING GIN (document);
-- CREATE INDEX idx_user_search_documents_embedding ON user_search_documents USING IVFFLAT (embedding vector_cosine_ops);
-- CREATE INDEX idx_user_search_documents_location ON user_search_documents USING GIST (location_geom);
CREATE INDEX idx_user_search_documents_location ON user_search_documents (latitude, longitude) WHERE latitude IS NOT NULL AND longitude IS NOT NULL;

CREATE TABLE search_queries (
  id BIGSERIAL PRIMARY KEY,
  user_id UUID REFERENCES users(id) ON DELETE SET NULL,
  query TEXT NOT NULL,
  filters JSONB,
  result_count INT,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_search_queries_user_id ON search_queries (user_id);
CREATE INDEX idx_search_queries_created_at ON search_queries (created_at);

-- Workshop tables (skillsphere.workshop.v1)
CREATE TYPE workshop_status AS ENUM ('draft', 'published', 'scheduled', 'live', 'completed', 'cancelled');
CREATE TYPE workshop_format AS ENUM ('lecture', 'interactive', 'hands_on', 'q_and_a');
CREATE TYPE workshop_difficulty AS ENUM ('beginner', 'intermediate', 'advanced');

CREATE TABLE workshops (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  host_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  title TEXT NOT NULL,
  description TEXT,
  skill_ids UUID[] NOT NULL DEFAULT '{}',
  scheduled_start TIMESTAMPTZ,
  scheduled_end TIMESTAMPTZ,
  duration_minutes INT,
  timezone VARCHAR(100),
  max_participants INT NOT NULL DEFAULT 1,
  current_participants INT NOT NULL DEFAULT 0,
  ticket_price NUMERIC(10, 2),
  is_free BOOLEAN NOT NULL DEFAULT false,
  cover_image_url TEXT,
  format workshop_format NOT NULL DEFAULT 'lecture',
  difficulty workshop_difficulty NOT NULL DEFAULT 'beginner',
  prerequisites TEXT[] DEFAULT '{}',
  learning_outcomes TEXT[] DEFAULT '{}',
  meeting_url TEXT,
  recording_url TEXT,
  status workshop_status NOT NULL DEFAULT 'draft',
  average_rating NUMERIC(3, 2) NOT NULL DEFAULT 0,
  total_reviews INT NOT NULL DEFAULT 0,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_workshops_host_id ON workshops (host_id);
CREATE INDEX idx_workshops_status ON workshops (status);
CREATE INDEX idx_workshops_scheduled_start ON workshops (scheduled_start);

CREATE TRIGGER update_workshops_updated_at
  BEFORE UPDATE ON workshops
  FOR EACH ROW
  EXECUTE FUNCTION set_updated_at();

CREATE TABLE workshop_materials (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  workshop_id UUID NOT NULL REFERENCES workshops(id) ON DELETE CASCADE,
  title TEXT NOT NULL,
  url TEXT NOT NULL,
  mime_type VARCHAR(100),
  size_bytes BIGINT,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE workshop_registrations (
  workshop_id UUID NOT NULL REFERENCES workshops(id) ON DELETE CASCADE,
  user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  status VARCHAR(20) NOT NULL DEFAULT 'registered',
  payment_id UUID REFERENCES payments(id) ON DELETE SET NULL,
  joined_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  PRIMARY KEY (workshop_id, user_id)
);

-- Certifications (skillsphere.certification.v1)
CREATE TYPE certification_type AS ENUM ('skill_completion', 'expert_verified', 'milestone', 'achievement');
CREATE TYPE certification_status AS ENUM ('pending', 'issued', 'revoked');
CREATE TYPE blockchain_network AS ENUM ('polygon', 'ethereum', 'solana');

CREATE TABLE certifications (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  skill_id UUID REFERENCES skills(id) ON DELETE SET NULL,
  skill_name VARCHAR(255),
  type certification_type NOT NULL,
  status certification_status NOT NULL DEFAULT 'pending',
  blockchain blockchain_network NOT NULL DEFAULT 'polygon',
  contract_address VARCHAR(255),
  token_id VARCHAR(255),
  transaction_hash VARCHAR(255),
  issuer_id UUID REFERENCES users(id) ON DELETE SET NULL,
  issuer_name VARCHAR(255),
  badge_image_url TEXT,
  certificate_url TEXT,
  metadata JSONB,
  issued_at TIMESTAMPTZ,
  expires_at TIMESTAMPTZ,
  revoked_at TIMESTAMPTZ,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_certifications_user_id ON certifications (user_id);
CREATE INDEX idx_certifications_status ON certifications (status);
CREATE INDEX idx_certifications_skill_id ON certifications (skill_id);

CREATE TRIGGER update_certifications_updated_at
  BEFORE UPDATE ON certifications
  FOR EACH ROW
  EXECUTE FUNCTION set_updated_at();

-- Challenges (skillsphere.challenge.v1)
CREATE TYPE challenge_status AS ENUM ('upcoming', 'active', 'voting', 'completed', 'cancelled');
CREATE TYPE challenge_type AS ENUM ('weekly', 'monthly', 'special_event', 'sponsored');
CREATE TYPE challenge_entry_status AS ENUM ('draft', 'submitted', 'disqualified', 'winner');
CREATE TYPE challenge_difficulty AS ENUM ('easy', 'medium', 'hard');

CREATE TABLE challenges (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  title TEXT NOT NULL,
  description TEXT,
  skill_ids UUID[] NOT NULL DEFAULT '{}',
  difficulty challenge_difficulty NOT NULL DEFAULT 'easy',
  start_date TIMESTAMPTZ,
  end_date TIMESTAMPTZ,
  voting_end_date TIMESTAMPTZ,
  rules TEXT[] DEFAULT '{}',
  judging_criteria TEXT[] DEFAULT '{}',
  max_participants INT,
  requires_verification BOOLEAN NOT NULL DEFAULT false,
  has_sponsor BOOLEAN NOT NULL DEFAULT false,
  sponsor_name TEXT,
  sponsor_logo_url TEXT,
  prizes JSONB,
  status challenge_status NOT NULL DEFAULT 'upcoming',
  type challenge_type NOT NULL DEFAULT 'weekly',
  participant_count INT NOT NULL DEFAULT 0,
  entry_count INT NOT NULL DEFAULT 0,
  banner_image_url TEXT,
  tags TEXT[] DEFAULT '{}',
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_challenges_status ON challenges (status);
CREATE INDEX idx_challenges_type ON challenges (type);
CREATE INDEX idx_challenges_start_date ON challenges (start_date);

CREATE TRIGGER update_challenges_updated_at
  BEFORE UPDATE ON challenges
  FOR EACH ROW
  EXECUTE FUNCTION set_updated_at();

CREATE TABLE challenge_entries (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  challenge_id UUID NOT NULL REFERENCES challenges(id) ON DELETE CASCADE,
  user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  status challenge_entry_status NOT NULL DEFAULT 'draft',
  submission_url TEXT,
  summary TEXT,
  score NUMERIC(5, 2),
  votes_count INT NOT NULL DEFAULT 0,
  submitted_at TIMESTAMPTZ,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE (challenge_id, user_id)
);

CREATE INDEX idx_challenge_entries_challenge ON challenge_entries (challenge_id);
CREATE INDEX idx_challenge_entries_user ON challenge_entries (user_id);

CREATE TABLE challenge_votes (
  challenge_id UUID NOT NULL REFERENCES challenges(id) ON DELETE CASCADE,
  entry_id UUID NOT NULL REFERENCES challenge_entries(id) ON DELETE CASCADE,
  voter_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  voted_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  PRIMARY KEY (challenge_id, voter_id)
);

-- Analytics storage for skillsphere.analytics.v1
CREATE TYPE metric_granularity AS ENUM ('hourly', 'daily', 'weekly', 'monthly');

CREATE TABLE platform_metrics (
  id BIGSERIAL PRIMARY KEY,
  metric_date TIMESTAMPTZ NOT NULL,
  granularity metric_granularity NOT NULL,
  active_users INT NOT NULL DEFAULT 0,
  matches_run INT NOT NULL DEFAULT 0,
  sessions_completed INT NOT NULL DEFAULT 0,
  premium_subscribers INT NOT NULL DEFAULT 0,
  revenue NUMERIC(12, 2) NOT NULL DEFAULT 0,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE (granularity, metric_date)
);

CREATE TABLE skill_trend_metrics (
  id BIGSERIAL PRIMARY KEY,
  skill_id UUID REFERENCES skills(id) ON DELETE CASCADE,
  captured_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  searches INT NOT NULL DEFAULT 0,
  matches INT NOT NULL DEFAULT 0,
  workshops_hosted INT NOT NULL DEFAULT 0
);

CREATE INDEX idx_skill_trend_metrics_skill_id ON skill_trend_metrics (skill_id);
