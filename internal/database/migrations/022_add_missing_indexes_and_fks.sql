-- +goose Up
-- Migration to add missing indexes and foreign keys for performance and data integrity
-- Based on comprehensive analysis of query patterns and access patterns

-- ============================================================================
-- 002_users.sql - Users Table Indexes
-- ============================================================================

CREATE INDEX idx_users_role ON users (role);
CREATE INDEX idx_users_is_active ON users (is_active);
CREATE INDEX idx_users_email_verified_at ON users (email_verified_at);
CREATE INDEX idx_users_created_at ON users (created_at);
CREATE INDEX idx_users_last_login_at ON users (last_login_at);
CREATE INDEX idx_users_is_verified ON users (is_verified); -- From 008_users_columns
CREATE INDEX idx_users_deleted_at ON users (deleted_at) WHERE deleted_at IS NOT NULL; -- Partial index for soft deletes

-- User OAuth Identities: Composite index for OAuth login lookups
-- Note: idx_user_oauth_updated_at moved to 024_fix_user_oauth_identities.sql after adding missing columns

-- ============================================================================
-- 003_users_sessions.sql - User Sessions
-- ============================================================================

CREATE INDEX idx_user_sessions_composite ON user_sessions (user_id, expires_at);

-- ============================================================================
-- 004_users_tokens.sql - User Tokens
-- ============================================================================

CREATE INDEX idx_user_tokens_type ON user_tokens (type);
CREATE INDEX idx_user_tokens_user_type ON user_tokens (user_id, type);
CREATE INDEX idx_user_tokens_user_expires ON user_tokens (user_id, expires_at);

-- ============================================================================
-- 005_user_moderation_actions.sql - Moderation
-- ============================================================================

-- User Moderation Actions
CREATE INDEX idx_user_moderation_admin_id ON user_moderation_actions (admin_id);
CREATE INDEX idx_user_moderation_action_type ON user_moderation_actions (action_type);
CREATE INDEX idx_user_moderation_expires_at ON user_moderation_actions (expires_at);
CREATE INDEX idx_user_moderation_reversed_by ON user_moderation_actions (reversed_by_admin_id);
CREATE INDEX idx_user_moderation_created_at ON user_moderation_actions (created_at);
CREATE INDEX idx_user_moderation_user_active ON user_moderation_actions (user_id, is_active);
CREATE INDEX idx_user_moderation_user_type_active ON user_moderation_actions (user_id, action_type, is_active);

-- Content Reports
CREATE INDEX idx_content_reports_reporter_id ON content_reports (reporter_id);
CREATE INDEX idx_content_reports_content_type ON content_reports (content_type);
CREATE INDEX idx_content_reports_report_type ON content_reports (report_type);
CREATE INDEX idx_content_reports_assigned_admin ON content_reports (assigned_admin_id);
CREATE INDEX idx_content_reports_moderation_action ON content_reports (moderation_action_id);
CREATE INDEX idx_content_reports_created_at ON content_reports (created_at);
CREATE INDEX idx_content_reports_reviewed_at ON content_reports (reviewed_at);
CREATE INDEX idx_content_reports_resolved_at ON content_reports (resolved_at);
CREATE INDEX idx_content_reports_status_created ON content_reports (status, created_at);
CREATE INDEX idx_content_reports_admin_status ON content_reports (assigned_admin_id, status);
CREATE INDEX idx_content_reports_content_lookup ON content_reports (content_type, content_id);

-- ============================================================================
-- 006_user_disputes.sql - Disputes
-- ============================================================================

CREATE INDEX idx_disputes_session_id ON disputes (session_id);
CREATE INDEX idx_disputes_disputing_user ON disputes (disputing_user_id);
CREATE INDEX idx_disputes_disputed_user ON disputes (disputed_user_id);
CREATE INDEX idx_disputes_assigned_admin ON disputes (assigned_admin_id);
CREATE INDEX idx_disputes_winner_id ON disputes (winner_id);
CREATE INDEX idx_disputes_created_at ON disputes (created_at);
CREATE INDEX idx_disputes_resolved_at ON disputes (resolved_at);
CREATE INDEX idx_disputes_status_created ON disputes (status, created_at);
CREATE INDEX idx_disputes_disputing_status ON disputes (disputing_user_id, status);
CREATE INDEX idx_disputes_disputed_status ON disputes (disputed_user_id, status);

-- ============================================================================
-- 007_platform_settings.sql - Platform Management
-- ============================================================================

-- Feature Flags
CREATE INDEX idx_feature_flags_is_enabled ON feature_flags (is_enabled);
CREATE INDEX idx_feature_flags_updated_at ON feature_flags (updated_at);

-- Feature Flag Users (reverse lookup)
CREATE INDEX idx_feature_flag_users_user_id ON feature_flag_users (user_id);

-- Audit Logs
CREATE INDEX idx_audit_logs_action ON audit_logs (action);
CREATE INDEX idx_audit_logs_timestamp_desc ON audit_logs (timestamp DESC);
CREATE INDEX idx_audit_logs_admin_timestamp ON audit_logs (admin_id, timestamp DESC);
CREATE INDEX idx_audit_logs_details_gin ON audit_logs USING GIN (details); -- For JSONB queries

-- Announcements
CREATE INDEX idx_announcements_admin_id ON announcements (admin_id);
CREATE INDEX idx_announcements_priority ON announcements (priority);
CREATE INDEX idx_announcements_created_at ON announcements (created_at);
CREATE INDEX idx_announcements_expires_at ON announcements (expires_at);
CREATE INDEX idx_announcements_priority_expires ON announcements (priority, expires_at) WHERE expires_at IS NOT NULL;

-- Announcement Recipients (reverse lookup)
CREATE INDEX idx_announcement_recipients_user_id ON announcement_recipients (user_id);

-- ============================================================================
-- 009_user_skills.sql - Skills System
-- ============================================================================

-- Skill Categories
CREATE INDEX idx_skill_categories_created_at ON skill_categories (created_at);

-- Skills
CREATE INDEX idx_skills_created_at ON skills (created_at);
CREATE INDEX idx_skills_updated_at ON skills (updated_at);
CREATE INDEX idx_skills_popularity ON skills (users_offering_count, users_wanting_count);
CREATE INDEX idx_skills_category_popularity ON skills (category_id, users_offering_count);

-- User Skills
CREATE INDEX idx_user_skills_type ON user_skills (skill_type);
CREATE INDEX idx_user_skills_proficiency ON user_skills (proficiency);
CREATE INDEX idx_user_skills_user_type ON user_skills (user_id, skill_type);
CREATE INDEX idx_user_skills_skill_type ON user_skills (skill_id, skill_type);
CREATE INDEX idx_user_skills_created_at ON user_skills (created_at);
CREATE INDEX idx_user_skills_updated_at ON user_skills (updated_at);

-- Skill to Tags (reverse lookup)
CREATE INDEX idx_skill_to_tags_tag_id ON skill_to_tags (tag_id);

-- ============================================================================
-- 010_user_availability.sql - Availability
-- ============================================================================

CREATE INDEX idx_user_availability_timezone ON user_availability (timezone);
CREATE INDEX idx_user_availability_updated_at ON user_availability (updated_at);

-- ============================================================================
-- 011_user_notification_preferences.sql - Notifications
-- ============================================================================

CREATE INDEX idx_user_notif_prefs_channel ON user_notification_preferences (channel_preference);
CREATE INDEX idx_user_notif_prefs_updated_at ON user_notification_preferences (updated_at);

-- ============================================================================
-- 012_user_stats.sql - User Statistics
-- ============================================================================

CREATE INDEX idx_user_stats_rating ON user_stats (average_rating);
CREATE INDEX idx_user_stats_completed_sessions ON user_stats (completed_sessions);
CREATE INDEX idx_user_stats_last_updated ON user_stats (last_updated_at);

-- ============================================================================
-- 013_user_verifications.sql - Verifications
-- ============================================================================

CREATE INDEX idx_user_verifications_method ON user_verifications (method);
CREATE INDEX idx_user_verifications_status ON user_verifications (status);
CREATE INDEX idx_user_verifications_reviewed_by ON user_verifications (reviewed_by_admin_id);
CREATE INDEX idx_user_verifications_submitted_at ON user_verifications (submitted_at);
CREATE INDEX idx_user_verifications_reviewed_at ON user_verifications (reviewed_at);
CREATE INDEX idx_user_verifications_status_submitted ON user_verifications (status, submitted_at);
CREATE INDEX idx_user_verifications_user_status ON user_verifications (user_id, status);

-- ============================================================================
-- 015_matching.sql - Matching System
-- ============================================================================

-- Match History
CREATE INDEX idx_match_history_algorithm ON match_history (algorithm_used);
CREATE INDEX idx_match_history_score ON match_history (match_score);
CREATE INDEX idx_match_history_interaction ON match_history (interaction_initiated);
CREATE INDEX idx_match_history_session_id ON match_history (session_id);
CREATE INDEX idx_match_history_created_at ON match_history (created_at);
CREATE INDEX idx_match_history_user_a_created ON match_history (user_id_a, created_at DESC);
CREATE INDEX idx_match_history_user_b_created ON match_history (user_id_b, created_at DESC);

-- Recommendation Cache
CREATE INDEX idx_recommendation_cache_item_type ON recommendation_cache (item_type);
CREATE INDEX idx_recommendation_cache_score ON recommendation_cache (relevance_score);
CREATE INDEX idx_recommendation_cache_generated_at ON recommendation_cache (generated_at);
CREATE INDEX idx_recommendation_cache_user_score ON recommendation_cache (user_id, relevance_score DESC);

-- Add missing FK for recommendation_cache
ALTER TABLE recommendation_cache
  ADD CONSTRAINT fk_recommendation_cache_user
  FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;

-- ============================================================================
-- 016_reviews.sql - Review System
-- ============================================================================

-- Reviews
CREATE INDEX idx_reviews_skill_id ON reviews (skill_id);
CREATE INDEX idx_reviews_type ON reviews (type);
CREATE INDEX idx_reviews_depth ON reviews (depth);
CREATE INDEX idx_reviews_requested_at ON reviews (requested_at);
CREATE INDEX idx_reviews_accepted_at ON reviews (accepted_at);
CREATE INDEX idx_reviews_deadline ON reviews (deadline);
CREATE INDEX idx_reviews_completed_at ON reviews (completed_at);
CREATE INDEX idx_reviews_payment_id ON reviews (payment_id);
CREATE INDEX idx_reviews_requester_status ON reviews (requester_id, status);
CREATE INDEX idx_reviews_reviewer_status ON reviews (reviewer_id, status);
CREATE INDEX idx_reviews_status_deadline ON reviews (status, deadline);
CREATE INDEX idx_reviews_skill_status ON reviews (skill_id, status);

-- Review Content
CREATE INDEX idx_review_content_review_id ON review_content (review_id);
CREATE INDEX idx_review_content_uploaded_at ON review_content (uploaded_at);

-- Review Feedback Sections
CREATE INDEX idx_review_feedback_review_id ON review_feedback_sections (review_id);
CREATE INDEX idx_review_feedback_rating ON review_feedback_sections (rating);

-- Review Revisions
CREATE INDEX idx_review_revisions_review_id ON review_revisions (review_id);
CREATE INDEX idx_review_revisions_requester_id ON review_revisions (requester_id);
CREATE INDEX idx_review_revisions_requested_at ON review_revisions (requested_at);
CREATE INDEX idx_review_revisions_new_deadline ON review_revisions (new_deadline);
CREATE INDEX idx_review_revisions_review_requested ON review_revisions (review_id, requested_at DESC);

-- ============================================================================
-- 018_chat.sql - Chat System
-- ============================================================================

-- Conversations
CREATE INDEX idx_conversations_last_message_at ON conversations (last_message_at DESC);
CREATE INDEX idx_conversations_created_at ON conversations (created_at);
CREATE INDEX idx_conversations_user_a_last_message ON conversations (user_a_id, last_message_at DESC);
CREATE INDEX idx_conversations_user_b_last_message ON conversations (user_b_id, last_message_at DESC);

-- Messages
CREATE INDEX idx_messages_type ON messages (type);
CREATE INDEX idx_messages_is_deleted ON messages (is_deleted);
CREATE INDEX idx_messages_conversation_sent_deleted ON messages (conversation_id, sent_at DESC, is_deleted);
CREATE INDEX idx_messages_sender_sent ON messages (sender_id, sent_at DESC);

-- Message Read Status (reverse lookups)
CREATE INDEX idx_message_read_status_user_id ON message_read_status (user_id);
CREATE INDEX idx_message_read_status_read_at ON message_read_status (read_at);

-- Conversation Participants (reverse lookups)
CREATE INDEX idx_conversation_participants_user_id ON conversation_participants (user_id);
CREATE INDEX idx_conversation_participants_is_archived ON conversation_participants (is_archived);
CREATE INDEX idx_conversation_participants_last_read ON conversation_participants (last_read_at);
CREATE INDEX idx_conversation_participants_user_archived ON conversation_participants (user_id, is_archived);

-- Message Reactions (reverse lookups)
CREATE INDEX idx_message_reactions_message_id ON message_reactions (message_id);
CREATE INDEX idx_message_reactions_user_id ON message_reactions (user_id);
CREATE INDEX idx_message_reactions_created_at ON message_reactions (created_at);

-- ============================================================================
-- 019_payment.sql - Payment System
-- ============================================================================

-- Customers
CREATE INDEX idx_customers_created_at ON customers (created_at);
CREATE INDEX idx_customers_updated_at ON customers (updated_at);

-- Subscriptions
CREATE INDEX idx_subscriptions_tier ON subscriptions (tier);
CREATE INDEX idx_subscriptions_status ON subscriptions (status);
CREATE INDEX idx_subscriptions_period_end ON subscriptions (current_period_end);
CREATE INDEX idx_subscriptions_cancel_at_period_end ON subscriptions (cancel_at_period_end);
CREATE INDEX idx_subscriptions_cancelled_at ON subscriptions (cancelled_at);
CREATE INDEX idx_subscriptions_created_at ON subscriptions (created_at);
CREATE INDEX idx_subscriptions_updated_at ON subscriptions (updated_at);
CREATE INDEX idx_subscriptions_user_status ON subscriptions (user_id, status);
CREATE INDEX idx_subscriptions_status_period_end ON subscriptions (status, current_period_end);
CREATE INDEX idx_subscriptions_provider ON subscriptions (provider); -- From 020_payments_store
CREATE INDEX idx_subscriptions_provider_subscription_id ON subscriptions (provider_subscription_id); -- From 020_payments_store

-- Payments
CREATE INDEX idx_payments_purpose ON payments (purpose);
CREATE INDEX idx_payments_status ON payments (status);
CREATE INDEX idx_payments_created_at ON payments (created_at);
CREATE INDEX idx_payments_updated_at ON payments (updated_at);
CREATE INDEX idx_payments_currency ON payments (currency);
CREATE INDEX idx_payments_user_status ON payments (user_id, status);
CREATE INDEX idx_payments_status_created ON payments (status, created_at);
CREATE INDEX idx_payments_metadata_gin ON payments USING GIN (metadata); -- For JSONB queries

-- Escrow Payments
CREATE INDEX idx_escrow_payments_payer_id ON escrow_payments (payer_id);
CREATE INDEX idx_escrow_payments_payee_id ON escrow_payments (payee_id);
CREATE INDEX idx_escrow_payments_status ON escrow_payments (status);
CREATE INDEX idx_escrow_payments_created_at ON escrow_payments (created_at);
CREATE INDEX idx_escrow_payments_release_at ON escrow_payments (release_at);
CREATE INDEX idx_escrow_payments_released_at ON escrow_payments (released_at);
CREATE INDEX idx_escrow_payments_status_release ON escrow_payments (status, release_at);
CREATE INDEX idx_escrow_payments_payer_status ON escrow_payments (payer_id, status);
CREATE INDEX idx_escrow_payments_payee_status ON escrow_payments (payee_id, status);

-- Payouts
CREATE INDEX idx_payouts_status ON payouts (status);
CREATE INDEX idx_payouts_initiated_at ON payouts (initiated_at);
CREATE INDEX idx_payouts_estimated_arrival ON payouts (estimated_arrival_at);
CREATE INDEX idx_payouts_currency ON payouts (currency);
CREATE INDEX idx_payouts_user_status ON payouts (user_id, status);
CREATE INDEX idx_payouts_status_arrival ON payouts (status, estimated_arrival_at);

-- Invoices
CREATE INDEX idx_invoices_subscription_id ON invoices (subscription_id);
CREATE INDEX idx_invoices_status ON invoices (status);
CREATE INDEX idx_invoices_due_date ON invoices (due_date);
CREATE INDEX idx_invoices_paid_at ON invoices (paid_at);
CREATE INDEX idx_invoices_created_at ON invoices (created_at);
CREATE INDEX idx_invoices_user_status ON invoices (user_id, status);
CREATE INDEX idx_invoices_status_due_date ON invoices (status, due_date);

-- ============================================================================
-- 019_gigs.sql - Gigs and Freelance System
-- ============================================================================

-- Gigs (some indexes already exist, adding missing ones)
CREATE INDEX idx_gigs_type ON gigs (type);
CREATE INDEX idx_gigs_is_hourly ON gigs (is_hourly);
CREATE INDEX idx_gigs_required_proficiency ON gigs (required_proficiency);
CREATE INDEX idx_gigs_deadline ON gigs (deadline);
CREATE INDEX idx_gigs_completed_at ON gigs (completed_at);
CREATE INDEX idx_gigs_created_at ON gigs (created_at);
CREATE INDEX idx_gigs_updated_at ON gigs (updated_at);
CREATE INDEX idx_gigs_status_created ON gigs (status, created_at DESC);
CREATE INDEX idx_gigs_status_deadline ON gigs (status, deadline);
CREATE INDEX idx_gigs_assigned_status ON gigs (assigned_to_id, status);
CREATE INDEX idx_gigs_creator_status ON gigs (creator_id, status);

-- Gig Skills (reverse lookup)
CREATE INDEX idx_gig_skills_skill_id ON gig_skills (skill_id);

-- Gig Applications (some indexes already exist, adding missing ones)
CREATE INDEX idx_gig_applications_status ON gig_applications (status);
CREATE INDEX idx_gig_applications_applied_at ON gig_applications (applied_at);
CREATE INDEX idx_gig_applications_responded_at ON gig_applications (responded_at);
CREATE INDEX idx_gig_applications_gig_status ON gig_applications (gig_id, status);
CREATE INDEX idx_gig_applications_freelancer_status ON gig_applications (freelancer_id, status);

-- Gig Work Submissions
CREATE INDEX idx_gig_work_submissions_gig_id ON gig_work_submissions (gig_id);
CREATE INDEX idx_gig_work_submissions_freelancer_id ON gig_work_submissions (freelancer_id);
CREATE INDEX idx_gig_work_submissions_revision ON gig_work_submissions (revision_number);
CREATE INDEX idx_gig_work_submissions_submitted_at ON gig_work_submissions (submitted_at);
CREATE INDEX idx_gig_work_submissions_gig_revision ON gig_work_submissions (gig_id, revision_number);
CREATE INDEX idx_gig_work_submissions_gig_submitted ON gig_work_submissions (gig_id, submitted_at DESC);

-- Gig Deliverables
CREATE INDEX idx_gig_deliverables_submission_id ON gig_deliverables (submission_id);
CREATE INDEX idx_gig_deliverables_uploaded_at ON gig_deliverables (uploaded_at);

-- Gig Reviews
CREATE INDEX idx_gig_reviews_creator_id ON gig_reviews (creator_id);
CREATE INDEX idx_gig_reviews_freelancer_id ON gig_reviews (freelancer_id);
CREATE INDEX idx_gig_reviews_rating ON gig_reviews (rating);
CREATE INDEX idx_gig_reviews_created_at ON gig_reviews (created_at);

-- ============================================================================
-- Additional Performance Optimizations
-- ============================================================================

-- Refresh the materialized view index (if needed)
-- Note: The materialized view user_skill_vectors should be refreshed periodically
-- CREATE INDEX CONCURRENTLY idx_user_skill_vectors_offered_vector ON user_skill_vectors USING GIN (offered_vector);
-- CREATE INDEX CONCURRENTLY idx_user_skill_vectors_wanted_vector ON user_skill_vectors USING GIN (wanted_vector);
