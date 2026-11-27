DELETE FROM "user"
WHERE email IN (
    SELECT 'seed.market-seller.' || lpad(n::text, 3, '0') || '@example.test'
    FROM generate_series(1, 80) AS n
);
