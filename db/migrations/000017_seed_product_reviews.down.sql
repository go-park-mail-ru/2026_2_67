DELETE FROM product_review AS r
USING "user" AS u
WHERE r.buyer_id = u.id
  AND r.product_id BETWEEN 1 AND 42
  AND u.email IN (
      SELECT 'seed.review-buyer.' || lpad(n::text, 3, '0') || '@example.test'
      FROM generate_series(1, 100) AS n
  );
