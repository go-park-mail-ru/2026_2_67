INSERT INTO "user" (email, password_hash)
SELECT 'seed.review-buyer.' || lpad(n::text, 3, '0') || '@example.test',
       '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy'
FROM generate_series(1, 100) AS n;

INSERT INTO buyer (user_id, name, surname)
SELECT u.id,
       (ARRAY['Анна', 'Александр', 'Мария', 'Дмитрий', 'Елена', 'Иван', 'Ольга', 'Михаил', 'Наталья', 'Андрей'])[(n - 1) % 10 + 1],
       (ARRAY['Смирнова', 'Иванов', 'Кузнецова', 'Попов', 'Васильева', 'Петров', 'Соколова', 'Морозов', 'Волкова', 'Новиков'])[(n - 1) % 10 + 1]
FROM generate_series(1, 100) AS n
JOIN "user" AS u ON u.email = 'seed.review-buyer.' || lpad(n::text, 3, '0') || '@example.test';
