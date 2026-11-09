-- Тестовые аккаунты ПВЗ: Россия, Москва. Пароль: MoscowPickupSeed2026!
-- Время работы 09:00–21:00 по Москве.
-- Координаты WGS84 — точки внутри контуров зданий, не адреса реальных ПВЗ.
-- Источник контуров и адресов: © OpenStreetMap contributors, ODbL (openstreetmap.org).
-- В комментариях к строкам указаны адреса и идентификаторы OSM way.
WITH points (email, longitude, latitude) AS (
    VALUES
        -- Гоголевский бульвар, 29; OSM way 30153051.
        ('seed.pickup.moscow.01@example.test', 37.5995452, 55.7496036),
        -- Филипповский переулок, 18; OSM way 30153067.
        ('seed.pickup.moscow.02@example.test', 37.5987761, 55.7498038),
        -- Нижний Кисловский переулок, 3; OSM way 31496302.
        ('seed.pickup.moscow.03@example.test', 37.6046750, 55.7540128),
        -- Средний Кисловский переулок, 7/10; OSM way 31496319.
        ('seed.pickup.moscow.04@example.test', 37.6033074, 55.7557376),
        -- Средний Кисловский переулок, 5/6, стр. 14; OSM way 31496322.
        ('seed.pickup.moscow.05@example.test', 37.6039291, 55.7557317),
        -- Никитский бульвар, 17; OSM way 31497234.
        ('seed.pickup.moscow.06@example.test', 37.5985671, 55.7562583),
        -- Никитский бульвар, 25; OSM way 32045624.
        ('seed.pickup.moscow.07@example.test', 37.5980962, 55.7569080),
        -- Хлебный переулок, 9; OSM way 32045631.
        ('seed.pickup.moscow.08@example.test', 37.5962002, 55.7547659),
        -- Скатертный переулок, 8/1; OSM way 32045633.
        ('seed.pickup.moscow.09@example.test', 37.5961482, 55.7557785),
        -- Хлебный переулок, 6; OSM way 32045635.
        ('seed.pickup.moscow.10@example.test', 37.5961917, 55.7550528),
        -- улица Арбат, 1; OSM way 35814548.
        ('seed.pickup.moscow.11@example.test', 37.5992843, 55.7517095),
        -- улица Воздвиженка, 9; OSM way 37447151.
        ('seed.pickup.moscow.12@example.test', 37.6047929, 55.7525486)
), accounts AS (
    INSERT INTO "user" (email, password_hash)
    SELECT email, '$2b$10$FzTDzn1tTKS4qj.JSGtBCuII9Px3npeCotrQGKvgndSD3aIqo53tC'
    FROM points
    RETURNING id, email
)
INSERT INTO pickup_point (user_id, longitude, latitude, start_time, end_time)
SELECT accounts.id, points.longitude, points.latitude, TIME '09:00', TIME '21:00'
FROM accounts
JOIN points ON points.email = accounts.email;
