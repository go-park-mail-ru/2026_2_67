DELETE FROM product AS p
USING product_category AS c, seller AS s, "user" AS u
WHERE p.category_id = c.id AND p.seller_id = s.user_id AND s.user_id = u.id
  AND EXISTS (
      SELECT 1 FROM generate_series(1, 80) AS n
      WHERE p.name = 'Seed product ' || n
        AND c.category = 'Seed category ' || (1 + ((n - 1) % 10))
        AND u.email = (ARRAY[
            'seed.seller.three@example.test', 'seed.seller.four@example.test',
            'seed.seller.five@example.test', 'seed.seller.six@example.test',
            'seed.seller.seven@example.test', 'seed.seller.eight@example.test'
        ])[1 + ((n - 1) % 6)]
  );
