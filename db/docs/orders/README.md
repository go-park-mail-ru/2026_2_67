# Orders: SQL-запросы

Контракт ручек: [`POST /api/v1/orders`](../../../backend/docs/orders/README.md),

## Миграции

[`db/migrations/000002_create_order.up.sql`](../../migrations/000004_create_order.up.sql)

## SQL repository/postgres

### POST /api/v1/orders

```sql
SELECT
    bp.product_id,
    bp.count,
    p.name,
    p.price,
    p.discount,
    p.available_count,
    s.name,
    p.category_id
FROM basket b
JOIN basket_product bp ON bp.basket_id = b.id
JOIN product p ON p.id = bp.product_id
JOIN seller s ON s.user_id = p.seller_id
WHERE b.buyer_id = $1
FOR UPDATE OF p;
```

Находит ближайший пункт выдачи по координатам.

```sql
SELECT user_id
FROM pickup_point
ORDER BY earth_distance(ll_to_earth(latitude, longitude), ll_to_earth($1, $2)) ASC
LIMIT 1;
```

Находит действующий промокод; выполняется только после trim и только для непустого кода.

```sql
SELECT id, available_category_id, discount
FROM promocode
WHERE UPPER(code) = UPPER($1) AND NOW() BETWEEN start_datetime AND end_datetime;
```

Создаёт заказ и возвращает его идентификатор.

```sql
INSERT INTO "order" (buyer_id, order_status_id, promocode_id, pickup_point_id)
VALUES ($1, 1, $2, $3)
RETURNING id;
```

Создаёт позицию заказа для каждого товара корзины.

```sql
INSERT INTO order_product (order_id, product_id, count, price)
VALUES ($1, $2, $3, $4);
```

Уменьшает доступный остаток товара.

```sql
UPDATE product
SET available_count = available_count - $2, updated_at = NOW()
WHERE id = $1;
```

Создаёт непрочитанное уведомление о заказе.

```sql
INSERT INTO order_notification (order_id, order_status_id, is_read)
VALUES ($1, 1, FALSE);
```

Удаляет товары из корзины покупателя.

```sql
DELETE FROM basket_product
WHERE basket_id = (SELECT id FROM basket WHERE buyer_id = $1);
```
