# Categories: SQL-запросы

Контракт ручек: [`GET /api/v1/categories`](../../../backend/docs/categories/README.md).

## Миграции

[`db/migrations/000002_create_product.up.sql`](../../migrations/000002_create_product.up.sql).

## SQL repository/postgres

### `GET /api/v1/categories?limit=&sortBy=new`

Получает категории от новых к старым, без `limit`:

```sql
SELECT id, category, picture_url FROM category ORDER BY id DESC;
```

Получает категории от новых к старым с ограничением количества:

```sql
SELECT id, category, picture_url FROM category ORDER BY id DESC LIMIT $1;
```

### `GET /api/v1/categories?limit=&sortBy=popular`

Получает категории по количеству товаров, без `limit`:

```sql
SELECT c.id, c.category, c.picture_url FROM category c LEFT JOIN product p ON p.category_id = c.id
GROUP BY c.id, c.category, c.picture_url ORDER BY COUNT(p.id) DESC, c.id DESC;
```

Получает категории по количеству товаров с ограничением количества:

```sql
SELECT c.id, c.category, c.picture_url FROM category c LEFT JOIN product p ON p.category_id = c.id
GROUP BY c.id, c.category, c.picture_url ORDER BY COUNT(p.id) DESC, c.id DESC LIMIT $1;
```
