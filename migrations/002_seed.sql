-- =============================================================
-- Seed Script: Random Data Generation
-- Generates sample users, boards, and scores for development/testing.
-- =============================================================

-- 1. Users — 50 random users
INSERT INTO users (id, username, created_at)
SELECT
    'user-' || gs                                         AS id,
    'player_' || substr(md5(random()::text), 1, 8)        AS username,
    NOW() - (random() * INTERVAL '90 days')               AS created_at
FROM generate_series(1, 50) AS gs
ON CONFLICT DO NOTHING;

-- 2. Boards — 5 boards with varying reset schedules
INSERT INTO boards (name, description, schedule_type, schedule_interval_seconds, created_at, updated_at)
VALUES
    ('Weekly Champions',  'Top players of the week',    'interval', 604800,  NOW(), NOW()),
    ('Daily Grind',       'Resets every day',            'interval', 86400,   NOW(), NOW()),
    ('All-Time Glory',    'Never resets',                NULL,       NULL,    NOW(), NOW()),
    ('Monthly Cup',       'Resets every 30 days',        'interval', 2592000, NOW(), NOW()),
    ('Weekend Warriors',  'Runs over the weekend',       'interval', 172800,  NOW(), NOW())
ON CONFLICT DO NOTHING;

-- 3. Scores — random scores for every user on every board
INSERT INTO scores (board_id, user_id, score, updated_at)
SELECT
    b.id                                                  AS board_id,
    'user-' || gs                                         AS user_id,
    round((random() * 9900 + 100)::numeric, 2)            AS score,
    NOW() - (random() * INTERVAL '30 days')               AS updated_at
FROM
    generate_series(1, 50) AS gs
    CROSS JOIN boards b
ON CONFLICT (board_id, user_id) DO UPDATE
    SET score      = EXCLUDED.score,
        updated_at = EXCLUDED.updated_at;
