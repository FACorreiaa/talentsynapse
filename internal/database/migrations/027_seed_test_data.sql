-- +goose Up
-- Batch insert dummy data for testing purposes
-- Users password is 'secret' ($2a$10$EixZaYVK1fsbw1ZfbX3OXePaWxn96p36WQoeG6Lruj3vjPGga31lW)

-- 1. Insert Categories
INSERT INTO
    skill_categories (name, description)
VALUES (
        'Development',
        'Software engineering and coding'
    ),
    (
        'Languages',
        'Spoken languages'
    ),
    (
        'Music',
        'Musical instruments and theory'
    ) ON CONFLICT (name) DO NOTHING;

-- 2. Insert Skills
WITH
    dev AS (
        SELECT id
        FROM skill_categories
        WHERE
            name = 'Development'
    ),
    lang AS (
        SELECT id
        FROM skill_categories
        WHERE
            name = 'Languages'
    ),
    music AS (
        SELECT id
        FROM skill_categories
        WHERE
            name = 'Music'
    )
INSERT INTO
    skills (
        name,
        category_id,
        description
    )
VALUES (
        'Go',
        (
            SELECT id
            FROM dev
        ),
        'Go programming language'
    ),
    (
        'Python',
        (
            SELECT id
            FROM dev
        ),
        'Python programming language'
    ),
    (
        'JavaScript',
        (
            SELECT id
            FROM dev
        ),
        'Web development language'
    ),
    (
        'English',
        (
            SELECT id
            FROM lang
        ),
        'English language'
    ),
    (
        'Spanish',
        (
            SELECT id
            FROM lang
        ),
        'Spanish language'
    ),
    (
        'Piano',
        (
            SELECT id
            FROM music
        ),
        'Piano playing'
    ) ON CONFLICT (category_id, name) DO NOTHING;

-- 3. Insert Users
INSERT INTO
    users (
        username,
        email,
        display_name,
        hashed_password,
        role,
        is_active
    )
VALUES (
        'alice',
        'alice@example.com',
        'Alice Coder',
        '$2a$10$EixZaYVK1fsbw1ZfbX3OXePaWxn96p36WQoeG6Lruj3vjPGga31lW',
        'member',
        true
    ),
    (
        'bob',
        'bob@example.com',
        'Bob Polyglot',
        '$2a$10$EixZaYVK1fsbw1ZfbX3OXePaWxn96p36WQoeG6Lruj3vjPGga31lW',
        'member',
        true
    ),
    (
        'charlie',
        'charlie@example.com',
        'Charlie Musician',
        '$2a$10$EixZaYVK1fsbw1ZfbX3OXePaWxn96p36WQoeG6Lruj3vjPGga31lW',
        'member',
        true
    ) ON CONFLICT (email) DO NOTHING;

-- 4. Insert User Skills to create matches
-- Alice: Offers Go (Expert), Wants Spanish
-- Bob: Offers Spanish (Expert), Wants Go
-- Charlie: Offers Piano, Wants JavaScript

WITH
    u_alice AS (
        SELECT id
        FROM users
        WHERE
            username = 'alice'
    ),
    u_bob AS (
        SELECT id
        FROM users
        WHERE
            username = 'bob'
    ),
    u_charlie AS (
        SELECT id
        FROM users
        WHERE
            username = 'charlie'
    ),
    s_go AS (
        SELECT id
        FROM skills
        WHERE
            name = 'Go'
    ),
    s_js AS (
        SELECT id
        FROM skills
        WHERE
            name = 'JavaScript'
    ),
    s_span AS (
        SELECT id
        FROM skills
        WHERE
            name = 'Spanish'
    ),
    s_piano AS (
        SELECT id
        FROM skills
        WHERE
            name = 'Piano'
    )
INSERT INTO
    user_skills (
        user_id,
        skill_id,
        skill_type,
        proficiency
    )
VALUES
    -- Alice
    (
        (
            SELECT id
            FROM u_alice
        ),
        (
            SELECT id
            FROM s_go
        ),
        'offered',
        9
    ),
    (
        (
            SELECT id
            FROM u_alice
        ),
        (
            SELECT id
            FROM s_span
        ),
        'wanted',
        3
    ),
    -- Bob
    (
        (
            SELECT id
            FROM u_bob
        ),
        (
            SELECT id
            FROM s_span
        ),
        'offered',
        10
    ),
    (
        (
            SELECT id
            FROM u_bob
        ),
        (
            SELECT id
            FROM s_go
        ),
        'wanted',
        5
    ),
    -- Charlie
    (
        (
            SELECT id
            FROM u_charlie
        ),
        (
            SELECT id
            FROM s_piano
        ),
        'offered',
        8
    ),
    (
        (
            SELECT id
            FROM u_charlie
        ),
        (
            SELECT id
            FROM s_js
        ),
        'wanted',
        4
    ) ON CONFLICT (user_id, skill_id, skill_type) DO NOTHING;