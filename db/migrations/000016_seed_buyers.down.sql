DELETE FROM "user"
WHERE email IN (
    SELECT 'seed.review-buyer.' || lpad(n::text, 3, '0') || '@example.test'
    FROM generate_series(1, 100) AS n
);
