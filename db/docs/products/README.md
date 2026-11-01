# Products: SQL-запросы

Контракты ручек: [`GET /api/v1/products`](../../../backend/docs/products/README.md),

## Миграции

[`db/migrations/000002_create_product.up.sql`](../../migrations/000003_create_basket.up.sql);

## GET /api/v1/products

### SQL repository/postgres

Запрос выбирает товар, скидку, дополнительные изображения, рейтинг и число
отзывов. Полный пример со всеми фильтрами и курсором приведён ниже.

## Полный SQL и аргументы

Пример для `sortBy=popular`, двух продавцов и всех необязательных фильтров.
При `limit=20` в SQL передаётся 21.

```sql
PREPARE products_popular AS
SELECT p.id, p.category_id, p.seller_id, p.price, p.discount, p.name,
    p.description, p.available_count, p.created_at, p.updated_at,
    COALESCE((SELECT json_agg(pp.picture_url ORDER BY pp.created_at, pp.id)
        FROM product_picture pp WHERE pp.product_id = p.id), '[]'::json),
    reviews.rating, reviews.review_count
FROM product p
JOIN category c ON c.id = p.category_id
JOIN seller s ON s.user_id = p.seller_id
LEFT JOIN LATERAL (
    SELECT COALESCE(AVG(pr.rating), 0)::double precision AS rating, COUNT(*) AS review_count
    FROM product_review pr WHERE pr.product_id = p.id
) reviews ON TRUE
WHERE LOWER(c.category) = LOWER($1)
  AND LOWER(s.name) IN ($2, $3)
  AND p.name ILIKE $4 ESCAPE E'\\'
  AND p.price >= $5
  AND p.price <= $6
  AND (reviews.rating, p.created_at, p.id) < ($7, $8, $9)
ORDER BY reviews.rating DESC, p.created_at DESC, p.id DESC
LIMIT $10;
```
