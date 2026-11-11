# Sellers: SQL-запросы

Контракты ручек: [`GET /api/v1/sellers`](../../../backend/docs/sellers/README.md)

## Миграции

[`db/migrations/000001_create_user.up.sql`](../../migrations/000001_create_user.up.sql)

## SQL repository/postgres

Запрос выбирает продавцов по товарам, соответствующим фильтрам, и сортирует их
по числу найденных товаров и названию.

```sql
SELECT s.user_id, s.name, s.description, s.avatar_url, s.created_at, s.updated_at
FROM product p
JOIN seller s ON p.seller_id = s.user_id
JOIN category c ON p.category_id = c.id
WHERE LOWER(c.category) = LOWER($1)
  AND p.name ILIKE $2 ESCAPE E'\\'
  AND p.price >= $3
  AND p.price <= $4
GROUP BY s.user_id
ORDER BY COUNT(s.user_id) DESC, s.name ASC
OFFSET $5
LIMIT $6;
```
