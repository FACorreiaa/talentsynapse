-- +goose Up
-- Load test seed data: Sessions, Reviews, Matches, and Chat
-- Creates realistic interaction data for load testing

-- Generate sessions between random users
-- +goose StatementBegin
DO $$
DECLARE
    v_user_ids UUID[];
    v_initiator_id UUID;
    v_partner_id UUID;
    v_session_id UUID;
    v_counter INT;
    v_status TEXT;
    v_scheduled_start TIMESTAMPTZ;
    v_scheduled_end TIMESTAMPTZ;
    v_statuses TEXT[] := ARRAY['pending', 'confirmed', 'in_progress', 'completed', 'completed', 'completed', 'cancelled', 'no_show'];
    v_skill_ids UUID[];
    v_initiator_skills UUID[];
    v_partner_skills UUID[];
BEGIN
    -- Get all load test user IDs
    SELECT ARRAY_AGG(id) INTO v_user_ids FROM users WHERE email LIKE '%@loadtest.example.com';

    -- Create 200 sessions with various statuses
    FOR v_counter IN 1..200 LOOP
        -- Pick two random different users
        v_initiator_id := v_user_ids[1 + FLOOR(RANDOM() * array_length(v_user_ids, 1))];
        v_partner_id := v_user_ids[1 + FLOOR(RANDOM() * array_length(v_user_ids, 1))];

        -- Ensure different users
        WHILE v_initiator_id = v_partner_id LOOP
            v_partner_id := v_user_ids[1 + FLOOR(RANDOM() * array_length(v_user_ids, 1))];
        END LOOP;

        -- Random status (weighted towards completed)
        v_status := v_statuses[1 + FLOOR(RANDOM() * array_length(v_statuses, 1))];

        -- Generate realistic scheduled time (past 60 days to future 30 days)
        v_scheduled_start := NOW() + ((RANDOM() * 90 - 60) || ' days')::INTERVAL;
        v_scheduled_end := v_scheduled_start + INTERVAL '1 hour';

        -- Get skills from both users
        SELECT ARRAY_AGG(skill_id) INTO v_initiator_skills
        FROM user_skills
        WHERE user_id = v_initiator_id AND skill_type = 'offered'
        LIMIT 3;

        SELECT ARRAY_AGG(skill_id) INTO v_partner_skills
        FROM user_skills
        WHERE user_id = v_partner_id AND skill_type = 'offered'
        LIMIT 3;

        -- Insert session
        INSERT INTO sessions (
            initiator_id,
            partner_id,
            initiator_offers,
            partner_offers,
            status,
            scheduled_start,
            scheduled_end,
            actual_start,
            actual_end,
            meeting_url,
            notes,
            created_at,
            updated_at
        ) VALUES (
            v_initiator_id,
            v_partner_id,
            COALESCE(v_initiator_skills, ARRAY[]::UUID[])::TEXT[],
            COALESCE(v_partner_skills, ARRAY[]::UUID[])::TEXT[],
            v_status::session_status,
            v_scheduled_start,
            v_scheduled_end,
            CASE WHEN v_status IN ('in_progress', 'completed') THEN v_scheduled_start + (RANDOM() * INTERVAL '15 minutes') ELSE NULL END,
            CASE WHEN v_status = 'completed' THEN v_scheduled_start + INTERVAL '1 hour' + (RANDOM() * INTERVAL '30 minutes') ELSE NULL END,
            CASE WHEN RANDOM() > 0.3 THEN 'https://meet.example.com/' || gen_random_uuid() ELSE NULL END,
            CASE WHEN RANDOM() > 0.5 THEN 'Looking forward to learning together!' ELSE NULL END,
            v_scheduled_start - INTERVAL '2 days',
            NOW()
        )
        RETURNING id INTO v_session_id;

        -- Update user stats for completed sessions
        IF v_status = 'completed' THEN
            UPDATE user_stats
            SET
                total_sessions = total_sessions + 1,
                completed_sessions = completed_sessions + 1,
                last_updated_at = NOW()
            WHERE user_id IN (v_initiator_id, v_partner_id);
        END IF;

    END LOOP;
END $$;
-- +goose StatementEnd

-- Generate reviews for completed sessions
-- +goose StatementBegin
DO $$
DECLARE
    v_session RECORD;
    v_skill_id UUID;
    v_review_id UUID;
BEGIN
    -- Create reviews for ~60% of completed sessions
    FOR v_session IN
        SELECT
            s.id as session_id,
            s.initiator_id,
            s.partner_id
        FROM sessions s
        WHERE s.status = 'completed'
        AND EXISTS (SELECT 1 FROM users u WHERE u.id = s.initiator_id AND u.email LIKE '%@loadtest.example.com')
        AND RANDOM() > 0.4
    LOOP
        -- Pick a random skill
        SELECT id INTO v_skill_id FROM skills ORDER BY RANDOM() LIMIT 1;

        -- Insert review (both parties rating each other)
        INSERT INTO reviews (
            requester_id,
            reviewer_id,
            skill_id,
            type,
            depth,
            title,
            price,
            status,
            requester_rating,
            requester_comment,
            reviewer_rating,
            reviewer_comment,
            requested_at,
            completed_at
        ) VALUES (
            v_session.initiator_id,
            v_session.partner_id,
            v_skill_id,
            'peer_review',
            'quick',
            'Post-session review',
            0.00,
            'completed',
            4 + FLOOR(RANDOM() * 2)::INT, -- Rating 4-5
            CASE WHEN RANDOM() > 0.5 THEN 'Great learning experience!' ELSE 'Very helpful and knowledgeable.' END,
            4 + FLOOR(RANDOM() * 2)::INT, -- Rating 4-5
            CASE WHEN RANDOM() > 0.5 THEN 'Excellent session!' ELSE 'Patient and clear teacher.' END,
            NOW() - (RANDOM() * INTERVAL '30 days'),
            NOW() - (RANDOM() * INTERVAL '25 days')
        )
        RETURNING id INTO v_review_id;

        -- Update user stats
        UPDATE user_stats
        SET
            total_reviews = total_reviews + 1,
            last_updated_at = NOW()
        WHERE user_id IN (v_session.initiator_id, v_session.partner_id);

    END LOOP;
END $$;
-- +goose StatementEnd

-- Recalculate average ratings
UPDATE user_stats us
SET average_rating = COALESCE((
    SELECT AVG(rating)::NUMERIC(3,2)
    FROM (
        SELECT requester_rating as rating FROM reviews WHERE reviewer_id = us.user_id
        UNION ALL
        SELECT reviewer_rating as rating FROM reviews WHERE requester_id = us.user_id
    ) ratings
    WHERE rating IS NOT NULL
), 0.00);

-- Generate match history (who viewed whom, who matched)
-- +goose StatementBegin
DO $$
DECLARE
    v_user_ids UUID[];
    v_user_a UUID;
    v_user_b UUID;
    v_counter INT;
BEGIN
    SELECT ARRAY_AGG(id) INTO v_user_ids FROM users WHERE email LIKE '%@loadtest.example.com';

    -- Create 500 match history entries
    FOR v_counter IN 1..500 LOOP
        v_user_a := v_user_ids[1 + FLOOR(RANDOM() * array_length(v_user_ids, 1))];
        v_user_b := v_user_ids[1 + FLOOR(RANDOM() * array_length(v_user_ids, 1))];

        WHILE v_user_a = v_user_b LOOP
            v_user_b := v_user_ids[1 + FLOOR(RANDOM() * array_length(v_user_ids, 1))];
        END LOOP;

        -- Generate random match scores (weighted towards higher scores)
        INSERT INTO match_history (
            user_id_a,
            user_id_b,
            algorithm_used,
            match_score,
            interaction_initiated,
            created_at
        )
        VALUES (
            v_user_a,
            v_user_b,
            'cosine_similarity',
            0.5 + (RANDOM() * 0.5), -- Score between 0.5 and 1.0
            RANDOM() > 0.7, -- 30% initiated interaction
            NOW() - (RANDOM() * INTERVAL '60 days')
        )
        ON CONFLICT (user_id_a, user_id_b, created_at) DO NOTHING;

    END LOOP;
END $$;
-- +goose StatementEnd

-- Generate conversations and messages for chat load testing
-- +goose StatementBegin
DO $$
DECLARE
    v_user_ids UUID[];
    v_user_a UUID;
    v_user_b UUID;
    v_conversation_id UUID;
    v_counter INT;
    v_msg_counter INT;
    v_num_messages INT;
    v_sender_id UUID;
    v_last_message_id UUID;
    v_messages TEXT[] := ARRAY[
        'Hi! I saw your profile and would love to exchange skills.',
        'That sounds great! When are you available?',
        'How about this weekend?',
        'Sure, Saturday works for me.',
        'Perfect! Looking forward to it.',
        'Do you prefer video call or in-person?',
        'Video call would be easier for me.',
        'Great! I''ll send you a meeting link.',
        'Thanks! See you then.',
        'Just wanted to confirm we''re still on for tomorrow.',
        'Yes, absolutely!',
        'Awesome, talk soon!'
    ];
BEGIN
    SELECT ARRAY_AGG(id) INTO v_user_ids FROM users WHERE email LIKE '%@loadtest.example.com';

    -- Create 100 conversations with varying message counts
    FOR v_counter IN 1..100 LOOP
        -- Pick two random users
        v_user_a := v_user_ids[1 + FLOOR(RANDOM() * array_length(v_user_ids, 1))];
        v_user_b := v_user_ids[1 + FLOOR(RANDOM() * array_length(v_user_ids, 1))];

        WHILE v_user_a = v_user_b LOOP
            v_user_b := v_user_ids[1 + FLOOR(RANDOM() * array_length(v_user_ids, 1))];
        END LOOP;

        -- Ensure consistent ordering (user_a_id < user_b_id)
        IF v_user_a > v_user_b THEN
            DECLARE temp UUID;
            BEGIN
                temp := v_user_a;
                v_user_a := v_user_b;
                v_user_b := temp;
            END;
        END IF;

        -- Insert conversation
        INSERT INTO conversations (user_a_id, user_b_id, created_at, updated_at)
        VALUES (v_user_a, v_user_b, NOW() - (RANDOM() * INTERVAL '30 days'), NOW())
        RETURNING id INTO v_conversation_id;

        -- Initialize participant metadata
        INSERT INTO conversation_participants (conversation_id, user_id, last_read_at, is_archived, created_at)
        VALUES
            (v_conversation_id, v_user_a, NOW() - (RANDOM() * INTERVAL '1 day'), false, NOW()),
            (v_conversation_id, v_user_b, NOW() - (RANDOM() * INTERVAL '1 day'), false, NOW());

        -- Generate 5-25 messages per conversation
        v_num_messages := 5 + FLOOR(RANDOM() * 21)::INT;

        FOR v_msg_counter IN 1..v_num_messages LOOP
            -- Alternate between users
            v_sender_id := CASE WHEN v_msg_counter % 2 = 1 THEN v_user_a ELSE v_user_b END;

            INSERT INTO messages (
                conversation_id,
                sender_id,
                type,
                content,
                sent_at,
                is_deleted
            ) VALUES (
                v_conversation_id,
                v_sender_id,
                'text',
                v_messages[1 + (v_msg_counter % array_length(v_messages, 1))],
                NOW() - (RANDOM() * INTERVAL '30 days') + (v_msg_counter || ' hours')::INTERVAL,
                false
            )
            RETURNING id INTO v_last_message_id;

            -- Mark messages as read (90% of the time)
            IF RANDOM() > 0.1 THEN
                INSERT INTO message_read_status (message_id, user_id, read_at)
                VALUES
                    (v_last_message_id, v_user_a, NOW() - (RANDOM() * INTERVAL '1 day')),
                    (v_last_message_id, v_user_b, NOW() - (RANDOM() * INTERVAL '1 day'))
                ON CONFLICT DO NOTHING;
            END IF;

        END LOOP;

        -- Update conversation with last message
        UPDATE conversations
        SET
            last_message_id = v_last_message_id,
            last_message_at = NOW(),
            updated_at = NOW()
        WHERE id = v_conversation_id;

    END LOOP;
END $$;
-- +goose StatementEnd

-- Award badges to some users
-- +goose StatementBegin
DO $$
DECLARE
    v_user_ids UUID[];
    v_user_id UUID;
    v_badge_id UUID;
    v_counter INT;
BEGIN
    SELECT ARRAY_AGG(id) INTO v_user_ids FROM users WHERE email LIKE '%@loadtest.example.com';

    -- Award "First Session" badge to 50 random users
    FOR v_counter IN 1..50 LOOP
        v_user_id := v_user_ids[1 + FLOOR(RANDOM() * array_length(v_user_ids, 1))];

        SELECT id INTO v_badge_id FROM badges WHERE code = 'first_session' LIMIT 1;

        IF v_badge_id IS NOT NULL THEN
            INSERT INTO user_badges (user_id, badge_id, awarded_at)
            VALUES (v_user_id, v_badge_id, NOW() - (RANDOM() * INTERVAL '60 days'))
            ON CONFLICT DO NOTHING;
        END IF;
    END LOOP;

    -- Award "Mentor" badge to 20 expert users
    FOR v_counter IN 1..20 LOOP
        SELECT id INTO v_user_id
        FROM users
        WHERE email LIKE '%@loadtest.example.com' AND role = 'expert'
        ORDER BY RANDOM()
        LIMIT 1;

        SELECT id INTO v_badge_id FROM badges WHERE code = 'mentor' LIMIT 1;

        IF v_badge_id IS NOT NULL AND v_user_id IS NOT NULL THEN
            INSERT INTO user_badges (user_id, badge_id, awarded_at)
            VALUES (v_user_id, v_badge_id, NOW() - (RANDOM() * INTERVAL '60 days'))
            ON CONFLICT DO NOTHING;
        END IF;
    END LOOP;
END $$;
-- +goose StatementEnd

-- Add portfolio items for some users
-- +goose StatementBegin
DO $$
DECLARE
    v_user_ids UUID[];
    v_user_id UUID;
    v_counter INT;
    v_titles TEXT[] := ARRAY[
        'Personal Portfolio Website',
        'E-commerce Platform',
        'Mobile App Design',
        'Data Visualization Dashboard',
        'Open Source Contribution',
        'Blog Platform',
        'Machine Learning Project',
        'Design System'
    ];
BEGIN
    SELECT ARRAY_AGG(id) INTO v_user_ids FROM users WHERE email LIKE '%@loadtest.example.com';

    -- Add 1-3 portfolio items for 40% of users
    FOR v_user_id IN
        SELECT id FROM users WHERE email LIKE '%@loadtest.example.com' AND RANDOM() > 0.6
    LOOP
        FOR v_counter IN 1..(1 + FLOOR(RANDOM() * 3))::INT LOOP
            INSERT INTO portfolio_items (
                user_id,
                title,
                description,
                link_url,
                created_at,
                updated_at
            ) VALUES (
                v_user_id,
                v_titles[1 + FLOOR(RANDOM() * array_length(v_titles, 1))],
                'A showcase of my skills and experience in this area.',
                'https://example.com/project-' || gen_random_uuid(),
                NOW() - (RANDOM() * INTERVAL '180 days'),
                NOW() - (RANDOM() * INTERVAL '30 days')
            );
        END LOOP;
    END LOOP;
END $$;
-- +goose StatementEnd

-- +goose Down
-- Remove all load test session/review data
DELETE FROM message_read_status WHERE message_id IN (
    SELECT m.id FROM messages m
    JOIN conversations c ON c.id = m.conversation_id
    JOIN users u ON u.id = c.user_a_id
    WHERE u.email LIKE '%@loadtest.example.com'
);

DELETE FROM messages WHERE conversation_id IN (
    SELECT c.id FROM conversations c
    JOIN users u ON u.id = c.user_a_id
    WHERE u.email LIKE '%@loadtest.example.com'
);

DELETE FROM conversation_participants WHERE conversation_id IN (
    SELECT c.id FROM conversations c
    JOIN users u ON u.id = c.user_a_id
    WHERE u.email LIKE '%@loadtest.example.com'
);

DELETE FROM conversations WHERE user_a_id IN (
    SELECT id FROM users WHERE email LIKE '%@loadtest.example.com'
);

DELETE FROM match_history WHERE user_id_a IN (
    SELECT id FROM users WHERE email LIKE '%@loadtest.example.com'
);

DELETE FROM portfolio_items WHERE user_id IN (
    SELECT id FROM users WHERE email LIKE '%@loadtest.example.com'
);

DELETE FROM user_badges WHERE user_id IN (
    SELECT id FROM users WHERE email LIKE '%@loadtest.example.com'
);

DELETE FROM reviews WHERE requester_id IN (
    SELECT id FROM users WHERE email LIKE '%@loadtest.example.com'
);

DELETE FROM sessions WHERE initiator_id IN (
    SELECT id FROM users WHERE email LIKE '%@loadtest.example.com'
);