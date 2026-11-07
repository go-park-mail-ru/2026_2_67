DELETE FROM product
WHERE id IN (1, 2, 3, 4, 5, 6, 7, 8, 9)
  AND name IN ('Backpack', 'Camera', 'Headphones', 'Keyboard', 'Lamp',
               'Laptop', 'Smartphone', 'Smartwatch', 'Speaker');
