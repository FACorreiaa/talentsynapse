-- +goose Up
-- +goose NO TRANSACTION
-- Migration: Add performance indexes for frequently queried columns
-- Optimizes common query patterns for user lookups, skills, matches, sessions, reviews, and messages
-- Created: 2026-01-31

-- ============================================================================
-- USERS TABLE - Optimizing user lookups
-- ============================================================================

-- Composite index for active user searches (common in auth and user listings)
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_users_active_email
    ON users (email)
    WHERE is_active = true AND deleted_at IS NULL;

-- Composite index for verified users (common filter in matching)
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_users_verified_active
    ON users (is_verified, is_active)
    WHERE deleted_at IS NULL;

-- Index for username search with case-insensitive pattern matching
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_users_username_lower
    ON users (LOWER(username));

-- Index for email search with case-insensitive pattern matching
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_users_email_lower
    ON users (LOWER(email));

-- Composite index for user role filtering with activity
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_users_role_active
    ON users (role, is_active)
    WHERE deleted_at IS NULL;

-- ============================================================================
-- USER_SKILLS TABLE - Optimizing skill-based queries
-- ============================================================================

-- Composite index for finding users who offer specific skills (matching algorithm)
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_user_skills_skill_offered
    ON user_skills (skill_id, proficiency DESC)
    WHERE skill_type = 'offered';

-- Composite index for finding users who want specific skills (matching algorithm)
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_user_skills_skill_wanted
    ON user_skills (skill_id, proficiency DESC)
    WHERE skill_type = 'wanted';

-- Composite index for user profile skill display (ordered by proficiency)
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_user_skills_user_proficiency
    ON user_skills (user_id, skill_type, proficiency DESC);

-- ============================================================================
-- SKILLS TABLE - Optimizing skill catalog queries
-- ============================================================================

-- Composite index for category browsing with popularity
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_skills_category_popular
    ON skills (category_id, users_offering_count DESC, users_wanting_count DESC);

-- Index for skill name text search (partial match)
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_skills_name_lower
    ON skills (LOWER(name));

-- ============================================================================
-- SESSIONS TABLE - Optimizing session queries
-- ============================================================================

-- Composite index for user's upcoming sessions
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_sessions_user_upcoming
    ON sessions (initiator_id, scheduled_start)
    WHERE status IN ('pending', 'confirmed');

-- Composite index for partner's upcoming sessions
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_sessions_partner_upcoming
    ON sessions (partner_id, scheduled_start)
    WHERE status IN ('pending', 'confirmed');

-- Index for finding sessions by date range (calendar view)
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_sessions_date_range
    ON sessions (scheduled_start, scheduled_end)
    WHERE status NOT IN ('cancelled', 'no_show');

-- Composite index for completed sessions count (statistics)
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_sessions_completed_user
    ON sessions (initiator_id)
    WHERE status = 'completed';

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_sessions_completed_partner
    ON sessions (partner_id)
    WHERE status = 'completed';

-- Index for session time overlap detection (scheduling conflicts)
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_sessions_time_overlap
    ON sessions USING gist (tstzrange(scheduled_start, scheduled_end, '[)'))
    WHERE status NOT IN ('cancelled', 'no_show');

-- ============================================================================
-- MATCH_HISTORY TABLE - Optimizing match queries
-- ============================================================================

-- Composite index for user match history with score (sorted results)
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_match_history_user_a_score
    ON match_history (user_id_a, match_score DESC, created_at DESC);

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_match_history_user_b_score
    ON match_history (user_id_b, match_score DESC, created_at DESC);

-- Index for finding mutual matches (where both users interacted)
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_match_history_interaction
    ON match_history (user_id_a, user_id_b)
    WHERE interaction_initiated = true;

-- Index for high-quality matches (score threshold)
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_match_history_high_score
    ON match_history (match_score DESC)
    WHERE match_score >= 0.7;

-- ============================================================================
-- REVIEWS TABLE - Optimizing review queries
-- ============================================================================

-- Composite index for reviewee lookup (profile reviews display)
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_reviews_reviewer_completed
    ON reviews (reviewer_id, completed_at DESC)
    WHERE status = 'completed';

-- Index for session-based reviews lookup
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_reviews_session
    ON reviews (session_id)
    WHERE session_id IS NOT NULL;

-- Unique index to prevent duplicate reviews (one review per user per session)
-- This enforces the business rule at the database level
CREATE UNIQUE INDEX CONCURRENTLY IF NOT EXISTS idx_reviews_unique_session_reviewer
    ON reviews (requester_id, session_id)
    WHERE session_id IS NOT NULL;

-- Composite index for finding pending reviews for a user
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_reviews_requester_pending
    ON reviews (requester_id, requested_at DESC)
    WHERE status = 'pending';

-- Index for reviewer workload (pending assignments)
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_reviews_reviewer_pending
    ON reviews (reviewer_id, deadline)
    WHERE status IN ('pending', 'accepted', 'in_progress');

-- ============================================================================
-- CONVERSATIONS TABLE - Optimizing chat queries
-- ============================================================================

-- Composite index for user's recent conversations (inbox view)
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_conversations_user_a_recent
    ON conversations (user_a_id, last_message_at DESC NULLS LAST);

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_conversations_user_b_recent
    ON conversations (user_b_id, last_message_at DESC NULLS LAST);

-- ============================================================================
-- MESSAGES TABLE - Optimizing message queries
-- ============================================================================

-- Composite index for unread messages count
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_messages_conversation_unread
    ON messages (conversation_id, sent_at DESC)
    WHERE is_deleted = false;

-- Index for user's sent messages (message history)
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_messages_sender_recent
    ON messages (sender_id, sent_at DESC)
    WHERE is_deleted = false;

-- ============================================================================
-- CONVERSATION_PARTICIPANTS TABLE - Optimizing unread counts
-- ============================================================================

-- Index for finding unread conversations for a user
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_conversation_participants_unread
    ON conversation_participants (user_id, last_read_at)
    WHERE is_archived = false;

-- ============================================================================
-- USER_STATS TABLE - Optimizing leaderboard/ranking queries
-- ============================================================================

-- Index for leaderboard queries (top users by points)
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_user_stats_points_ranking
    ON user_stats (points DESC NULLS LAST);

-- Index for tier-based filtering
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_user_stats_tier_points
    ON user_stats (tier, points DESC NULLS LAST);

-- Index for rating-based queries (top rated users)
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_user_stats_rating_ranking
    ON user_stats (average_rating DESC NULLS LAST)
    WHERE total_reviews >= 3;

-- ============================================================================
-- USER_MFA TABLE - Optimizing MFA lookups
-- ============================================================================

-- Index for enabled MFA users
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_user_mfa_enabled
    ON user_mfa (user_id)
    WHERE enabled = true;

-- ============================================================================
-- CALENDAR_EVENTS TABLE - Optimizing calendar queries
-- ============================================================================

-- Composite index for user's events in a date range
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_calendar_events_user_range
    ON calendar_events (user_id, start_time, end_time);

-- ============================================================================
-- PORTFOLIO_ITEMS TABLE - Optimizing portfolio queries
-- ============================================================================

-- Composite index for user portfolio display (ordered)
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_portfolio_items_user_ordered
    ON portfolio_items (user_id, created_at DESC);

-- ============================================================================
-- USER_TOKENS TABLE - Optimizing token validation
-- ============================================================================

-- Index for token lookup (expiry check done in application layer)
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_user_tokens_hash
    ON user_tokens (token_hash, expires_at);

-- ============================================================================
-- USER_SESSIONS TABLE (auth sessions) - Optimizing session validation
-- ============================================================================

-- Index for session lookup (expiry check done in application layer)
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_user_sessions_refresh_token
    ON user_sessions (hashed_refresh_token, expires_at);

-- ============================================================================
-- ANALYZE TABLES to update statistics after adding indexes
-- ============================================================================

ANALYZE users;
ANALYZE user_skills;
ANALYZE skills;
ANALYZE sessions;
ANALYZE match_history;
ANALYZE reviews;
ANALYZE conversations;
ANALYZE messages;
ANALYZE conversation_participants;
ANALYZE user_stats;
ANALYZE calendar_events;
ANALYZE portfolio_items;
ANALYZE user_tokens;
ANALYZE user_sessions;

-- +goose Down
-- +goose NO TRANSACTION
-- Drop all indexes created in this migration

DROP INDEX CONCURRENTLY IF EXISTS idx_users_active_email;
DROP INDEX CONCURRENTLY IF EXISTS idx_users_verified_active;
DROP INDEX CONCURRENTLY IF EXISTS idx_users_username_lower;
DROP INDEX CONCURRENTLY IF EXISTS idx_users_email_lower;
DROP INDEX CONCURRENTLY IF EXISTS idx_users_role_active;

DROP INDEX CONCURRENTLY IF EXISTS idx_user_skills_skill_offered;
DROP INDEX CONCURRENTLY IF EXISTS idx_user_skills_skill_wanted;
DROP INDEX CONCURRENTLY IF EXISTS idx_user_skills_user_proficiency;

DROP INDEX CONCURRENTLY IF EXISTS idx_skills_category_popular;
DROP INDEX CONCURRENTLY IF EXISTS idx_skills_name_lower;

DROP INDEX CONCURRENTLY IF EXISTS idx_sessions_user_upcoming;
DROP INDEX CONCURRENTLY IF EXISTS idx_sessions_partner_upcoming;
DROP INDEX CONCURRENTLY IF EXISTS idx_sessions_date_range;
DROP INDEX CONCURRENTLY IF EXISTS idx_sessions_completed_user;
DROP INDEX CONCURRENTLY IF EXISTS idx_sessions_completed_partner;
DROP INDEX CONCURRENTLY IF EXISTS idx_sessions_time_overlap;

DROP INDEX CONCURRENTLY IF EXISTS idx_match_history_user_a_score;
DROP INDEX CONCURRENTLY IF EXISTS idx_match_history_user_b_score;
DROP INDEX CONCURRENTLY IF EXISTS idx_match_history_interaction;
DROP INDEX CONCURRENTLY IF EXISTS idx_match_history_high_score;

DROP INDEX CONCURRENTLY IF EXISTS idx_reviews_reviewer_completed;
DROP INDEX CONCURRENTLY IF EXISTS idx_reviews_session;
DROP INDEX CONCURRENTLY IF EXISTS idx_reviews_unique_session_reviewer;
DROP INDEX CONCURRENTLY IF EXISTS idx_reviews_requester_pending;
DROP INDEX CONCURRENTLY IF EXISTS idx_reviews_reviewer_pending;

DROP INDEX CONCURRENTLY IF EXISTS idx_conversations_user_a_recent;
DROP INDEX CONCURRENTLY IF EXISTS idx_conversations_user_b_recent;

DROP INDEX CONCURRENTLY IF EXISTS idx_messages_conversation_unread;
DROP INDEX CONCURRENTLY IF EXISTS idx_messages_sender_recent;

DROP INDEX CONCURRENTLY IF EXISTS idx_conversation_participants_unread;

DROP INDEX CONCURRENTLY IF EXISTS idx_user_stats_points_ranking;
DROP INDEX CONCURRENTLY IF EXISTS idx_user_stats_tier_points;
DROP INDEX CONCURRENTLY IF EXISTS idx_user_stats_rating_ranking;

DROP INDEX CONCURRENTLY IF EXISTS idx_user_mfa_enabled;

DROP INDEX CONCURRENTLY IF EXISTS idx_calendar_events_user_range;

DROP INDEX CONCURRENTLY IF EXISTS idx_portfolio_items_user_ordered;

DROP INDEX CONCURRENTLY IF EXISTS idx_user_tokens_hash;

DROP INDEX CONCURRENTLY IF EXISTS idx_user_sessions_refresh_token;
