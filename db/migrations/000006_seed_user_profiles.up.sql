INSERT INTO "user" (email, password_hash)
VALUES
    ('seed.buyer.one@example.test', '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy'),
    ('seed.buyer.two@example.test', '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy'),
    ('seed.seller.one@example.test', '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy'),
    ('seed.seller.two@example.test', '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy'),
    ('seed.seller.three@example.test', '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy'),
    ('seed.seller.four@example.test', '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy'),
    ('seed.seller.five@example.test', '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy'),
    ('seed.seller.six@example.test', '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy'),
    ('seed.seller.seven@example.test', '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy'),
    ('seed.seller.eight@example.test', '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy'),
    ('seed.pickup.one@example.test', '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy'),
    ('seed.pickup.two@example.test', '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy');

INSERT INTO buyer (user_id, name, surname, birth_date, description, avatar_url, telephone)
SELECT users.id, profiles.name, profiles.surname, profiles.birth_date, profiles.description,
       profiles.avatar_url, profiles.telephone
FROM (VALUES
    ('seed.buyer.one@example.test', 'Name1', 'Surname1', DATE '1995-03-12', 'Descr1', NULL, '+79990000001'),
    ('seed.buyer.two@example.test', 'Name2', 'Surname2', DATE '1998-07-24', 'Descr2', NULL, '+79990000002')
) AS profiles(email, name, surname, birth_date, description, avatar_url, telephone)
JOIN "user" AS users ON LOWER(users.email) = LOWER(profiles.email);

INSERT INTO seller (user_id, name, description, avatar_url)
SELECT users.id, profiles.name, profiles.description, profiles.avatar_url
FROM (VALUES
    ('seed.seller.one@example.test', 'Seller1', 'Descr1', NULL),
    ('seed.seller.two@example.test', 'Seller2', 'Descr2', NULL),
    ('seed.seller.three@example.test', 'Seller3', 'Descr3', NULL),
    ('seed.seller.four@example.test', 'Seller4', 'Descr4', NULL),
    ('seed.seller.five@example.test', 'Seller5', 'Descr5', NULL),
    ('seed.seller.six@example.test', 'Seller6', 'Descr6', NULL),
    ('seed.seller.seven@example.test', 'Seller7', 'Descr7', NULL),
    ('seed.seller.eight@example.test', 'Seller8', 'Descr8', NULL)
) AS profiles(email, name, description, avatar_url)
JOIN "user" AS users ON LOWER(users.email) = LOWER(profiles.email);

INSERT INTO pickup_point (user_id, longitude, latitude, start_time, end_time)
SELECT users.id, points.longitude, points.latitude, points.start_time, points.end_time
FROM (VALUES
    ('seed.pickup.one@example.test', 37.6173, 55.7558, TIME '09:00', TIME '21:00'),
    ('seed.pickup.two@example.test', 30.3159, 59.9398, TIME '10:00', TIME '20:00')
) AS points(email, longitude, latitude, start_time, end_time)
JOIN "user" AS users ON LOWER(users.email) = LOWER(points.email);
