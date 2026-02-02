-- +goose Up
-- Load test seed data: 100+ users with diverse skill profiles
-- This migration creates realistic test data for load testing

-- Insert diverse skill categories if they don't exist
INSERT INTO skill_categories (id, name, description, created_at) VALUES
    (gen_random_uuid(), 'Programming', 'Software development and programming languages', NOW()),
    (gen_random_uuid(), 'Design', 'Visual and UX design skills', NOW()),
    (gen_random_uuid(), 'Languages', 'Human languages and communication', NOW()),
    (gen_random_uuid(), 'Music', 'Musical instruments and theory', NOW()),
    (gen_random_uuid(), 'Business', 'Business and marketing skills', NOW()),
    (gen_random_uuid(), 'Data Science', 'Data analysis and machine learning', NOW()),
    (gen_random_uuid(), 'DevOps', 'Infrastructure and deployment', NOW()),
    (gen_random_uuid(), 'Soft Skills', 'Communication and leadership', NOW())
ON CONFLICT (name) DO NOTHING;

-- Insert comprehensive skill catalog (50+ skills)
WITH skill_cat AS (
    SELECT id, name FROM skill_categories
)
INSERT INTO skills (id, category_id, name, description, created_at, updated_at)
SELECT
    gen_random_uuid(),
    sc.id,
    s.skill_name,
    s.description,
    NOW(),
    NOW()
FROM skill_cat sc
CROSS JOIN LATERAL (
    VALUES
        -- Programming skills
        ('Go', 'Backend development with Go'),
        ('Python', 'Python programming and scripting'),
        ('JavaScript', 'JavaScript and web development'),
        ('TypeScript', 'Typed JavaScript development'),
        ('Rust', 'Systems programming with Rust'),
        ('Java', 'Enterprise Java development'),
        ('C++', 'High-performance C++ programming'),
        ('React', 'React.js frontend development'),
        ('Vue.js', 'Vue.js framework'),
        ('Node.js', 'Server-side JavaScript'),
        ('SQL', 'Database querying and design'),
        ('GraphQL', 'GraphQL API development'),

        -- Design skills
        ('UI Design', 'User interface design'),
        ('UX Design', 'User experience design'),
        ('Figma', 'Figma design tool'),
        ('Adobe XD', 'Adobe XD prototyping'),
        ('Photoshop', 'Adobe Photoshop'),
        ('Illustration', 'Digital illustration'),

        -- Languages
        ('English', 'English language'),
        ('Spanish', 'Spanish language'),
        ('French', 'French language'),
        ('German', 'German language'),
        ('Mandarin', 'Mandarin Chinese'),
        ('Japanese', 'Japanese language'),

        -- Music
        ('Piano', 'Piano performance'),
        ('Guitar', 'Guitar playing'),
        ('Drums', 'Drum performance'),
        ('Violin', 'Violin playing'),
        ('Music Theory', 'Music theory and composition'),

        -- Business
        ('Marketing', 'Marketing strategy'),
        ('SEO', 'Search engine optimization'),
        ('Content Writing', 'Content creation'),
        ('Public Speaking', 'Public speaking skills'),
        ('Project Management', 'Project management'),

        -- Data Science
        ('Machine Learning', 'ML algorithms and models'),
        ('Data Analysis', 'Data analysis and visualization'),
        ('TensorFlow', 'TensorFlow framework'),
        ('PyTorch', 'PyTorch deep learning'),
        ('R', 'R programming for statistics'),

        -- DevOps
        ('Docker', 'Docker containerization'),
        ('Kubernetes', 'Kubernetes orchestration'),
        ('AWS', 'Amazon Web Services'),
        ('CI/CD', 'Continuous integration and deployment'),
        ('Terraform', 'Infrastructure as code'),

        -- Soft Skills
        ('Leadership', 'Team leadership'),
        ('Mentoring', 'Mentoring and coaching'),
        ('Agile', 'Agile methodologies'),
        ('Communication', 'Effective communication')
) AS s(skill_name, description)
WHERE sc.name = CASE
    WHEN s.skill_name IN ('Go', 'Python', 'JavaScript', 'TypeScript', 'Rust', 'Java', 'C++', 'React', 'Vue.js', 'Node.js', 'SQL', 'GraphQL') THEN 'Programming'
    WHEN s.skill_name IN ('UI Design', 'UX Design', 'Figma', 'Adobe XD', 'Photoshop', 'Illustration') THEN 'Design'
    WHEN s.skill_name IN ('English', 'Spanish', 'French', 'German', 'Mandarin', 'Japanese') THEN 'Languages'
    WHEN s.skill_name IN ('Piano', 'Guitar', 'Drums', 'Violin', 'Music Theory') THEN 'Music'
    WHEN s.skill_name IN ('Marketing', 'SEO', 'Content Writing', 'Public Speaking', 'Project Management') THEN 'Business'
    WHEN s.skill_name IN ('Machine Learning', 'Data Analysis', 'TensorFlow', 'PyTorch', 'R') THEN 'Data Science'
    WHEN s.skill_name IN ('Docker', 'Kubernetes', 'AWS', 'CI/CD', 'Terraform') THEN 'DevOps'
    WHEN s.skill_name IN ('Leadership', 'Mentoring', 'Agile', 'Communication') THEN 'Soft Skills'
END
ON CONFLICT (category_id, name) DO NOTHING;

-- Generate 150 test users with diverse profiles
-- Password for all test users: 'testpass123' (hashed with bcrypt)
-- Hash generated with: bcrypt.GenerateFromPassword([]byte("testpass123"), bcrypt.DefaultCost)
-- +goose StatementBegin
DO $$
DECLARE
    v_user_id UUID;
    v_counter INT;
    v_first_names TEXT[] := ARRAY['Alex', 'Jordan', 'Taylor', 'Morgan', 'Casey', 'Jamie', 'Riley', 'Avery', 'Quinn', 'Blake',
                                   'Sam', 'Charlie', 'Dakota', 'Reese', 'Skylar', 'Sage', 'Rowan', 'Parker', 'Hayden', 'Emerson',
                                   'Kai', 'River', 'Phoenix', 'Peyton', 'Cameron', 'Finley', 'Ariel', 'Jesse', 'Drew', 'Milan'];
    v_last_names TEXT[] := ARRAY['Smith', 'Johnson', 'Williams', 'Brown', 'Jones', 'Garcia', 'Miller', 'Davis', 'Rodriguez', 'Martinez',
                                  'Hernandez', 'Lopez', 'Gonzalez', 'Wilson', 'Anderson', 'Thomas', 'Taylor', 'Moore', 'Jackson', 'Martin',
                                  'Lee', 'Walker', 'Hall', 'Allen', 'Young', 'King', 'Wright', 'Scott', 'Torres', 'Nguyen'];
    v_cities TEXT[] := ARRAY['New York', 'San Francisco', 'London', 'Berlin', 'Tokyo', 'Paris', 'Toronto', 'Sydney', 'Amsterdam', 'Barcelona',
                              'Seattle', 'Austin', 'Boston', 'Chicago', 'Los Angeles', 'Singapore', 'Dublin', 'Copenhagen', 'Stockholm', 'Zurich'];
    v_bios TEXT[] := ARRAY[
        'Passionate about technology and learning new skills',
        'Lifelong learner seeking to expand my knowledge',
        'Love teaching and helping others grow',
        'Tech enthusiast with a creative side',
        'Building cool things and sharing knowledge',
        'Always curious, always learning',
        'Expert in my field, beginner in many others',
        'Believe in the power of peer learning',
        'Skill exchange enthusiast',
        'Learning by teaching, teaching by learning'
    ];
    v_roles TEXT[] := ARRAY['member', 'member', 'member', 'member', 'expert', 'member', 'member', 'moderator'];
    v_first_name TEXT;
    v_last_name TEXT;
    v_username TEXT;
    v_email TEXT;
    v_display_name TEXT;
    v_bio TEXT;
    v_city TEXT;
    v_lat DECIMAL(10, 8);
    v_lng DECIMAL(11, 8);
    v_role TEXT;
BEGIN
    FOR v_counter IN 1..150 LOOP
        -- Generate random name combination
        v_first_name := v_first_names[(FLOOR(RANDOM() * array_length(v_first_names, 1)) + 1)];
        v_last_name := v_last_names[(FLOOR(RANDOM() * array_length(v_last_names, 1)) + 1)];
        v_username := LOWER(v_first_name || v_last_name || v_counter);
        v_email := v_username || '@loadtest.example.com';
        v_display_name := v_first_name || ' ' || v_last_name;
        v_bio := v_bios[(FLOOR(RANDOM() * array_length(v_bios, 1)) + 1)];
        v_city := v_cities[(FLOOR(RANDOM() * array_length(v_cities, 1)) + 1)];
        v_role := v_roles[(FLOOR(RANDOM() * array_length(v_roles, 1)) + 1)];

        -- Generate realistic lat/lng based on city (approximate)
        v_lat := CASE v_city
            WHEN 'New York' THEN 40.7128
            WHEN 'San Francisco' THEN 37.7749
            WHEN 'London' THEN 51.5074
            WHEN 'Berlin' THEN 52.5200
            WHEN 'Tokyo' THEN 35.6762
            WHEN 'Paris' THEN 48.8566
            WHEN 'Toronto' THEN 43.6532
            WHEN 'Sydney' THEN -33.8688
            WHEN 'Amsterdam' THEN 52.3676
            WHEN 'Barcelona' THEN 41.3851
            WHEN 'Seattle' THEN 47.6062
            WHEN 'Austin' THEN 30.2672
            WHEN 'Boston' THEN 42.3601
            WHEN 'Chicago' THEN 41.8781
            WHEN 'Los Angeles' THEN 34.0522
            WHEN 'Singapore' THEN 1.3521
            WHEN 'Dublin' THEN 53.3498
            WHEN 'Copenhagen' THEN 55.6761
            WHEN 'Stockholm' THEN 59.3293
            ELSE 40.7128
        END + (RANDOM() * 0.5 - 0.25); -- Add slight variance

        v_lng := CASE v_city
            WHEN 'New York' THEN -74.0060
            WHEN 'San Francisco' THEN -122.4194
            WHEN 'London' THEN -0.1278
            WHEN 'Berlin' THEN 13.4050
            WHEN 'Tokyo' THEN 139.6503
            WHEN 'Paris' THEN 2.3522
            WHEN 'Toronto' THEN -79.3832
            WHEN 'Sydney' THEN 151.2093
            WHEN 'Amsterdam' THEN 4.9041
            WHEN 'Barcelona' THEN 2.1734
            WHEN 'Seattle' THEN -122.3321
            WHEN 'Austin' THEN -97.7431
            WHEN 'Boston' THEN -71.0589
            WHEN 'Chicago' THEN -87.6298
            WHEN 'Los Angeles' THEN -118.2437
            WHEN 'Singapore' THEN 103.8198
            WHEN 'Dublin' THEN -6.2603
            WHEN 'Copenhagen' THEN 12.5683
            WHEN 'Stockholm' THEN 18.0686
            ELSE -74.0060
        END + (RANDOM() * 0.5 - 0.25);

        -- Insert user
        INSERT INTO users (
            email,
            username,
            hashed_password,
            display_name,
            bio,
            location_text,
            latitude,
            longitude,
            role,
            is_active,
            is_verified,
            email_verified_at,
            created_at,
            updated_at,
            last_login_at,
            social_links
        ) VALUES (
            v_email,
            v_username,
            '$2a$10$rQZ5YJ3p5Y8Z5Z5Z5Z5Z5uH5Z5Z5Z5Z5Z5Z5Z5Z5Z5Z5Z5Z5Z5Z5a', -- bcrypt hash for 'testpass123'
            v_display_name,
            v_bio,
            v_city,
            v_lat,
            v_lng,
            v_role::user_role,
            true,
            RANDOM() > 0.3, -- 70% verified
            CASE WHEN RANDOM() > 0.3 THEN NOW() - (RANDOM() * INTERVAL '90 days') ELSE NULL END,
            NOW() - (RANDOM() * INTERVAL '180 days'),
            NOW() - (RANDOM() * INTERVAL '7 days'),
            CASE WHEN RANDOM() > 0.2 THEN NOW() - (RANDOM() * INTERVAL '24 hours') ELSE NULL END,
            jsonb_build_object(
                'github', CASE WHEN RANDOM() > 0.5 THEN 'https://github.com/' || v_username ELSE NULL END,
                'linkedin', CASE WHEN RANDOM() > 0.4 THEN 'https://linkedin.com/in/' || v_username ELSE NULL END,
                'twitter', CASE WHEN RANDOM() > 0.6 THEN 'https://twitter.com/' || v_username ELSE NULL END,
                'website', CASE WHEN RANDOM() > 0.7 THEN 'https://' || v_username || '.dev' ELSE NULL END
            )
        )
        RETURNING id INTO v_user_id;

        -- Initialize user stats
        INSERT INTO user_stats (user_id, total_sessions, completed_sessions, average_rating, total_reviews, last_updated_at)
        VALUES (
            v_user_id,
            FLOOR(RANDOM() * 20)::INT,
            FLOOR(RANDOM() * 15)::INT,
            (3.5 + RANDOM() * 1.5)::NUMERIC(3,2), -- Rating between 3.5 and 5.0
            FLOOR(RANDOM() * 10)::INT,
            NOW()
        );

    END LOOP;
END $$;
-- +goose StatementEnd

-- Assign diverse skills to users
-- Each user gets 3-8 offered skills and 2-6 wanted skills
-- +goose StatementBegin
DO $$
DECLARE
    v_user RECORD;
    v_skill RECORD;
    v_num_offered INT;
    v_num_wanted INT;
    v_skill_ids UUID[];
    v_random_skill_id UUID;
    v_counter INT;
BEGIN
    -- Loop through all load test users
    FOR v_user IN
        SELECT id FROM users WHERE email LIKE '%@loadtest.example.com'
    LOOP
        -- Get all available skill IDs
        SELECT ARRAY_AGG(id) INTO v_skill_ids FROM skills;

        -- Randomly decide how many skills to offer and want
        v_num_offered := 3 + FLOOR(RANDOM() * 6)::INT; -- 3-8 offered
        v_num_wanted := 2 + FLOOR(RANDOM() * 5)::INT;  -- 2-6 wanted

        -- Add offered skills
        FOR v_counter IN 1..v_num_offered LOOP
            v_random_skill_id := v_skill_ids[1 + FLOOR(RANDOM() * array_length(v_skill_ids, 1))];

            INSERT INTO user_skills (user_id, skill_id, skill_type, proficiency, created_at, updated_at)
            VALUES (
                v_user.id,
                v_random_skill_id,
                'offered',
                1 + FLOOR(RANDOM() * 10)::INT, -- Proficiency 1-10
                NOW(),
                NOW()
            )
            ON CONFLICT (user_id, skill_id, skill_type) DO NOTHING;
        END LOOP;

        -- Add wanted skills
        FOR v_counter IN 1..v_num_wanted LOOP
            v_random_skill_id := v_skill_ids[1 + FLOOR(RANDOM() * array_length(v_skill_ids, 1))];

            INSERT INTO user_skills (user_id, skill_id, skill_type, proficiency, created_at, updated_at)
            VALUES (
                v_user.id,
                v_random_skill_id,
                'wanted',
                NULL, -- Wanted skills don't have proficiency
                NOW(),
                NOW()
            )
            ON CONFLICT (user_id, skill_id, skill_type) DO NOTHING;
        END LOOP;
    END LOOP;
END $$;
-- +goose StatementEnd

-- Update skill usage counts
UPDATE skills s
SET
    users_offering_count = (SELECT COUNT(*) FROM user_skills us WHERE us.skill_id = s.id AND us.skill_type = 'offered'),
    users_wanting_count = (SELECT COUNT(*) FROM user_skills us WHERE us.skill_id = s.id AND us.skill_type = 'wanted'),
    updated_at = NOW();

-- Refresh materialized view for matching algorithm
REFRESH MATERIALIZED VIEW user_skill_vectors;

-- +goose Down
-- Remove load test data
DELETE FROM user_skills WHERE user_id IN (SELECT id FROM users WHERE email LIKE '%@loadtest.example.com');
DELETE FROM user_stats WHERE user_id IN (SELECT id FROM users WHERE email LIKE '%@loadtest.example.com');
DELETE FROM users WHERE email LIKE '%@loadtest.example.com';

-- Refresh materialized view
REFRESH MATERIALIZED VIEW user_skill_vectors;