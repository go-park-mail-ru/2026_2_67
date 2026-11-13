DELETE FROM product
WHERE id IN (1, 2, 3, 4, 5, 6, 7, 8, 9)
  AND name IN ('Рюкзак Fjällräven Kånken', 'Камера Sony PXW-Z90',
               'Наушники Samsung Galaxy Buds2', 'Клавиатура Logitech K380',
               'Настольная лампа IKEA TERTIAL', 'Ноутбук Apple MacBook Air 13 M1',
               'Смартфон Samsung Galaxy S23', 'Умные часы Apple Watch Series 9',
               'Портативная колонка JBL Flip 6');
