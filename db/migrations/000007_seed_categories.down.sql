DELETE FROM product_category AS c
WHERE c.category IN (SELECT 'Seed category ' || n FROM generate_series(1, 10) AS n)
  AND NOT EXISTS (SELECT 1 FROM product AS p WHERE p.category_id = c.id);
