# Orders: SQL-запросы

## Миграции

Таблицы создаются миграцией `db/migrations/000004_create_order.up.sql`.

## Выбор товаров из корзины

Контракт написан в: `backend/docs/orders/README.md`

Запрос выполняется первым:

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

## Поиск ближайшего пункта выдачи

```sql
SELECT user_id
FROM pickup_point
ORDER BY earth_distance(
    ll_to_earth(latitude, longitude),
    ll_to_earth($1, $2)
) ASC
LIMIT 1;
```

## Проверка промокода

Запрос выполняется только если передан непустой промокод. Перед запросом код
очищается от пробелов по краям.

```sql
SELECT
    id,
    available_category_id,
    discount
FROM promocode
WHERE UPPER(code) = UPPER($1)
  AND NOW() BETWEEN start_datetime AND end_datetime;
```

## Создание заказа

```sql
INSERT INTO "order"
    (buyer_id, order_status_id, promocode_id, pickup_point_id)
VALUES
    ($1, 1, $2, $3)
RETURNING id;
```

## Добавление позиции заказа

Для каждого товара из корзины выполняется отдельный запрос:

```sql
INSERT INTO order_product
    (order_id, product_id, count, price)
VALUES
    ($1, $2, $3, $4);
```

## Уменьшение остатка товара

После добавления позиции выполняется:

```sql
UPDATE product
SET
    available_count = available_count - $2,
    updated_at = NOW()
WHERE id = $1;
```

## Создание уведомления

```sql
INSERT INTO order_notification
    (order_id, order_status_id, is_read)
VALUES
    ($1, 1, FALSE);
```

Уведомление создаётся для начального статуса `Обрабатывается` и помечается как
непрочитанное.

## Очистка корзины

После успешного добавления всех позиций и уведомления корзина очищается:

```sql
DELETE FROM basket_product
WHERE basket_id = (
    SELECT id
    FROM basket
    WHERE buyer_id = $1
);
```

Удаляются все товары из корзины покупателя. Сам объект `basket` не удаляется.

## Транзакция

Все перечисленные SQL-операции выполняются в одной транзакции:

```text
BEGIN
  SELECT basket items ... FOR UPDATE
  SELECT nearest pickup point
  SELECT promocode, если передан
  INSERT order
  INSERT order_product для каждого товара
  UPDATE product для каждого товара
  INSERT order_notification
  DELETE basket_product
COMMIT
```

При ошибке любой операции транзакция откатывается, поэтому частично созданный
заказ, позиции, изменение остатков и очистка корзины не сохраняются.
