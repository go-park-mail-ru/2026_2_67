INSERT INTO category (category, picture_url)
SELECT 'Seed category ' || n,
       'https://example.test/categories/' || n || '.jpg'
FROM generate_series(1, 10) AS n;
