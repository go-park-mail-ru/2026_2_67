INSERT INTO product (category_id, seller_id, price, discount, name, description, available_count)
SELECT c.id, s.user_id, 100.0 + n * 12.5, (n % 5) * 5,
       'Seed product ' || n, 'Example product ' || n, 1 + (n % 20)
FROM generate_series(1, 80) AS n
JOIN product_category AS c ON c.category = 'Seed category ' || (1 + ((n - 1) % 10))
JOIN "user" AS u ON u.email = (ARRAY[
    'seed.seller.three@example.test', 'seed.seller.four@example.test',
    'seed.seller.five@example.test', 'seed.seller.six@example.test',
    'seed.seller.seven@example.test', 'seed.seller.eight@example.test'
])[1 + ((n - 1) % 6)]
JOIN seller AS s ON s.user_id = u.id;
