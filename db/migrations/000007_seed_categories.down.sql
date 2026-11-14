DELETE FROM category AS c
WHERE c.category IN ('Техника', 'Одежда', 'Аксессуары', 'Товары для дома',
                     'Косметика', 'Парфюмерия', 'Мебель', 'Сантехника',
                     'Продукты питания', 'Товары для животных')
  AND NOT EXISTS (SELECT 1 FROM product AS p WHERE p.category_id = c.id);
