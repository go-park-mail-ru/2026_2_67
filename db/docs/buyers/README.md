# Buyers: SQL-запросы

Контракт ручек: [`GET /api/v1/buyers/{id}/profile`](../../../backend/docs/buyers/profile/README.md) и [`/api/v1/buyers/{id}/basket`](../../../backend/docs/buyers/basket/README.md).

## Миграции

[`db/migrations/000001_create_user.up.sql`](../../migrations/000001_create_user.up.sql), [`db/migrations/000003_create_basket.up.sql`](../../migrations/000003_create_basket.up.sql).

## SQL repository/postgres

### `GET /api/v1/buyers/{id}/profile`

Получает данные покупателя по идентификатору.

```sql
SELECT user_id, name, surname, birth_date, description, avatar_url, telephone, created_at, updated_at FROM buyer WHERE user_id = $1;
```

### `POST /api/v1/auth/register` (создание профиля buyer)

Создаёт профиль покупателя.

```sql
INSERT INTO buyer (user_id, name, surname, birth_date, description, avatar_url, telephone) VALUES ($1, $2, $3, $4, $5, $6, $7);
```

### `GET /api/v1/buyers/{id}/profile` (join с user)

Получает профиль покупателя вместе с данными пользователя.

```sql
SELECT "user".id, "user".email, "user".password_hash, "user".created_at, "user".updated_at,
buyer.name, buyer.surname, buyer.birth_date, buyer.description, buyer.avatar_url, buyer.telephone, buyer.created_at, buyer.updated_at
FROM buyer LEFT JOIN "user" ON buyer.user_id = "user".id WHERE "user".id = $1;
```

### `GET /api/v1/buyers/{id}/basket`

Проверяет существование покупателя:

```sql
SELECT 1 FROM buyer WHERE user_id = $1;
```

Создаёт корзину, если её ещё нет:

```sql
INSERT INTO basket (buyer_id) VALUES ($1) ON CONFLICT (buyer_id) DO NOTHING;
```

Получает товары корзины с данными товаров, категорий и продавцов:

```sql
SELECT bp.product_id, c.category, p.name,
COALESCE((SELECT pp.picture_url FROM product_picture pp WHERE pp.product_id = p.id ORDER BY pp.created_at, pp.id LIMIT 1), ''),
p.price, p.discount, bp.count, p.available_count, s.name, p.category_id
FROM basket b
JOIN basket_product bp ON bp.basket_id = b.id
JOIN product p ON p.id = bp.product_id
JOIN product_category c ON c.id = p.category_id
JOIN seller s ON s.user_id = p.seller_id
WHERE b.buyer_id = $1
ORDER BY bp.created_at, bp.product_id;
```

### `POST /api/v1/buyers/{id}/basket/products`

Проверяет существование покупателя:

```sql
SELECT 1 FROM buyer WHERE user_id = $1;
```

Получает доступный остаток товара:

```sql
SELECT available_count FROM product WHERE id = $1;
```

Создаёт корзину, если её ещё нет:

```sql
INSERT INTO basket (buyer_id) VALUES ($1) ON CONFLICT (buyer_id) DO NOTHING;
```

Получает текущее количество товара в корзине:

```sql
SELECT bp.count FROM basket b JOIN basket_product bp ON bp.basket_id = b.id WHERE b.buyer_id = $1 AND bp.product_id = $2;
```

Добавляет товар или увеличивает его количество в корзине:

```sql
INSERT INTO basket_product (basket_id, product_id, count)
SELECT b.id, $2, $3 FROM basket b WHERE b.buyer_id = $1
ON CONFLICT (basket_id, product_id)
DO UPDATE SET count = basket_product.count + EXCLUDED.count, updated_at = NOW();
```

### `PATCH /api/v1/buyers/{buyer_id}/basket/products/{product_id}`

Получает доступный остаток товара в корзине:

```sql
SELECT p.available_count FROM basket b JOIN basket_product bp ON bp.basket_id = b.id JOIN product p ON p.id = bp.product_id WHERE b.buyer_id = $1 AND bp.product_id = $2;
```

Удаляет товар из корзины:

```sql
DELETE FROM basket_product WHERE basket_id = (SELECT id FROM basket WHERE buyer_id = $1) AND product_id = $2;
```

Устанавливает новое количество товара в корзине:

```sql
UPDATE basket_product SET count = $3, updated_at = NOW() WHERE basket_id = (SELECT id FROM basket WHERE buyer_id = $1) AND product_id = $2;
```

### `POST /api/v1/buyers/{id}/basket/preview`

Находит ближайший пункт выдачи:

```sql
SELECT user_id, longitude, latitude, start_time::text, end_time::text FROM pickup_point ORDER BY earth_distance(ll_to_earth(latitude, longitude), ll_to_earth($1, $2)) ASC LIMIT 1;
```

Находит действующий промокод без учёта регистра:

```sql
SELECT id, code, available_category_id, discount, start_datetime, end_datetime FROM promocode WHERE UPPER(code) = UPPER($1) AND NOW() BETWEEN start_datetime AND end_datetime;
```
