INSERT INTO "user" (email, password_hash)
VALUES
    ('seed.applevibe@example.test', '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy'),
    ('seed.nokiavibe@example.test', '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy'),
    ('seed.samsungvibe@example.test', '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy'),
    ('seed.laptopvibe@example.test', '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy'),
    ('seed.techvibe@example.test', '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy');

INSERT INTO seller (user_id, name, description, avatar_url)
SELECT users.id, profiles.name, profiles.description, profiles.avatar_url
FROM (VALUES
    ('seed.applevibe@example.test', 'AppleVibe', 'Seed seller AppleVibe', NULL),
    ('seed.nokiavibe@example.test', 'NokiaVibe', 'Seed seller NokiaVibe', NULL),
    ('seed.samsungvibe@example.test', 'SamsungVibe', 'Seed seller SamsungVibe', NULL),
    ('seed.laptopvibe@example.test', 'LaptopVibe', 'Seed seller LaptopVibe', NULL),
    ('seed.techvibe@example.test', 'TechVibe', 'Seed seller TechVibe', NULL)
) AS profiles(email, name, description, avatar_url)
JOIN "user" AS users ON LOWER(users.email) = LOWER(profiles.email);
