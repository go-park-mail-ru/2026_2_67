INSERT INTO "user" (login, password_hash)
VALUES
    ('seed.buyer.one@example.test', '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy'),
    ('seed.buyer.two@example.test', '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy'),
    ('seed.seller.one@example.test', '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy'),
    ('seed.seller.two@example.test', '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy'),
    ('seed.pickup.one@example.test', '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy'),
    ('seed.pickup.two@example.test', '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy');

INSERT INTO buyer (user_id, name, surname, birth_date, description, avatar_url, email, telephone)
SELECT users.id, profiles.name, profiles.surname, profiles.birth_date, profiles.description,
       profiles.avatar_url, profiles.email, profiles.telephone
FROM (VALUES
    ('seed.buyer.one@example.test', 'Name1', 'Surname1', DATE '1995-03-12', 'Descr1', NULL, 'buyer.one@example.test', '+79990000001'),
    ('seed.buyer.two@example.test', 'Name2', 'Surname2', DATE '1998-07-24', 'Descr2', NULL, 'buyer.two@example.test', '+79990000002')
) AS profiles(login, name, surname, birth_date, description, avatar_url, email, telephone)
JOIN "user" AS users ON LOWER(users.login) = LOWER(profiles.login);

INSERT INTO seller (user_id, name, description, avatar_url, email)
SELECT users.id, profiles.name, profiles.description, profiles.avatar_url, profiles.email
FROM (VALUES
    ('seed.seller.one@example.test', 'Seller1', 'Descr1', NULL, 'seller.one@example.test'),
    ('seed.seller.two@example.test', 'Seller2', 'Descr2', NULL, 'seller.two@example.test')
) AS profiles(login, name, description, avatar_url, email)
JOIN "user" AS users ON LOWER(users.login) = LOWER(profiles.login);

INSERT INTO pickup_point (user_id, longitude, latitude, start_time, end_time)
SELECT users.id, points.longitude, points.latitude, points.start_time, points.end_time
FROM (VALUES
    ('seed.pickup.one@example.test', 37.6173, 55.7558, TIME '09:00', TIME '21:00'),
    ('seed.pickup.two@example.test', 30.3159, 59.9398, TIME '10:00', TIME '20:00')
) AS points(login, longitude, latitude, start_time, end_time)
JOIN "user" AS users ON LOWER(users.login) = LOWER(points.login);
