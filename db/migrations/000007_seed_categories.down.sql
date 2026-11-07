DELETE FROM category AS c
WHERE c.category IN ('Техника', 'Одежда', 'Уникальное')
  AND NOT EXISTS (SELECT 1 FROM product AS p WHERE p.category_id = c.id);
