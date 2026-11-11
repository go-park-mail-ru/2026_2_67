INSERT INTO product (id, category_id, seller_id, price, discount, name, description, available_count)
SELECT products.id, c.id, s.user_id, products.price, products.discount,
       products.name, products.description, products.available_count
FROM (VALUES
    (1, 'Одежда', 'seed.applevibe@example.test', 999.99, 10, 'Backpack', 'Seed product Backpack', 12),
    (2, 'Техника', 'seed.nokiavibe@example.test', 499.99, 5, 'Camera', 'Seed product Camera', 8),
    (3, 'Техника', 'seed.samsungvibe@example.test', 149.99, 15, 'Headphones', 'Seed product Headphones', 20),
    (4, 'Техника', 'seed.laptopvibe@example.test', 79.99, 0, 'Keyboard', 'Seed product Keyboard', 25),
    (5, 'Уникальное', 'seed.techvibe@example.test', 59.99, 5, 'Lamp', 'Seed product Lamp', 10),
    (6, 'Техника', 'seed.laptopvibe@example.test', 1299.99, 10, 'Laptop', 'Seed product Laptop', 6),
    (7, 'Техника', 'seed.samsungvibe@example.test', 699.99, 20, 'Smartphone', 'Seed product Smartphone', 14),
    (8, 'Техника', 'seed.nokiavibe@example.test', 199.99, 0, 'Smartwatch', 'Seed product Smartwatch', 18),
    (9, 'Уникальное', 'seed.applevibe@example.test', 89.99, 5, 'Speaker', 'Seed product Speaker', 16)
) AS products(id, category, seller_email, price, discount, name, description, available_count)
JOIN category AS c ON c.category = products.category
JOIN "user" AS u ON u.email = products.seller_email
JOIN seller AS s ON s.user_id = u.id;

SELECT setval(pg_get_serial_sequence('product', 'id'), MAX(id), true)
FROM product;

INSERT INTO product_picture (product_id, picture_url)
VALUES
    (1, '/media/product-pictures/1/backpack-1.jpg'),
    (2, '/media/product-pictures/2/camera-1.jpg'),
    (3, '/media/product-pictures/3/headphones-1.jpg'),
    (3, '/media/product-pictures/3/headphones-2.jpg'),
    (4, '/media/product-pictures/4/keyboard-1.jpg'),
    (5, '/media/product-pictures/5/lamp-1.jpg'),
    (5, '/media/product-pictures/5/lamp-2.jpg'),
    (6, '/media/product-pictures/6/laptop-1.jpg'),
    (6, '/media/product-pictures/6/laptop-2.jpg'),
    (7, '/media/product-pictures/7/smartphone-1.jpg'),
    (7, '/media/product-pictures/7/smartphone-2.jpg'),
    (8, '/media/product-pictures/8/smartwatch-1.jpg'),
    (9, '/media/product-pictures/9/speaker-1.jpg');
