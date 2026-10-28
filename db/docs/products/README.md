# Products: SQL-запросы

## Миграции

Основные таблицы создаются миграцией
`db/migrations/000002_create_product.up.sql`.

## GET /api/v1/products

Контракт написан в: `backend/docs/products/README.md`

### SQL-запрос

Запрос выбирает товар, скидку, дополнительные изображения, рейтинг и число
отзывов:

```sql
SELECT
    p.id,
    p.category_id,
    p.seller_id,
    p.price,
    p.discount,
    p.name,
    p.description,
    p.available_count,
    p.created_at,
    p.updated_at,
    COALESCE(
        (
            SELECT json_agg(pp.picture_url ORDER BY pp.created_at, pp.id)
            FROM product_picture pp
            WHERE pp.product_id = p.id
        ),
        '[]'::json
    ),
    reviews.rating,
    reviews.review_count
FROM product p
JOIN product_category c ON c.id = p.category_id
JOIN seller s ON s.user_id = p.seller_id
LEFT JOIN LATERAL (
    SELECT
        COALESCE(AVG(pr.rating), 0)::double precision AS rating,
        COUNT(*) AS review_count
    FROM product_review pr
    WHERE pr.product_id = p.id
) reviews ON TRUE
WHERE ...
ORDER BY ...
LIMIT ...;
```

### Фильтры

Фильтры добавляются только при наличии соответствующего параметра и
объединяются через `AND`.

#### category

```sql
LOWER(c.category) = LOWER($n)
```

Сравнение названия категории выполняется без учёта регистра.

#### seller

```sql
LOWER(s.name) = LOWER($n)
```

Сравнение названия продавца выполняется без учёта регистра.

#### query

```sql
p.name ILIKE $n ESCAPE E'\\'
```

В `SearchProducts` пользовательская строка экранируется, затем к ней
добавляется `%`. Поэтому поиск выполняется по префиксу, а символы `%`, `_` и
обратный слэш из ввода не становятся дополнительными wildcard-символами.

#### priceMin и priceMax

```sql
p.price >= $n
p.price <= $n
```

Минимальная цена всегда добавляется в запрос. Верхняя граница добавляется
только при непустом `priceMax`. Обе границы включаются в результат.

## Сортировка

Сортировка выполняется в базе данных.

Для `sortBy=new`:

```sql
ORDER BY p.created_at DESC, p.id DESC
```

Новые товары идут первыми. `id` используется как дополнительный стабильный
критерий, если у товаров одинаковый `created_at`.

Для `sortBy=popular`:

```sql
ORDER BY reviews.rating DESC, p.created_at DESC, p.id DESC
```

Популярность определяется средним рейтингом товара. При одинаковом рейтинге
используется порядок по дате создания и ID.

## Курсорная пагинация

Репозиторий запрашивает `limit + 1` строку. Дополнительная строка нужна, чтобы
определить наличие следующей страницы. В ответ включаются только первые
`limit` товаров.

Для `new` используется условие:

```sql
(p.created_at, p.id) < ($n, $n)
```

Для `popular` используется условие:

```sql
(reviews.rating, p.created_at, p.id) < ($n, $n, $n)
```

Последний возвращённый товар становится `NextCursor`. Для `popular` курсор
также содержит рейтинг.

Пример `Cursor`:
```json
{
  "created_at": "2026-10-27T12:00:00Z",
  "id": 42,
  "rating": 4.5
}
```

Если следующей страницы нет, `nextCursor` равен пустой строке.

## Лимит и ошибки

Если `limit == 0`, repository использует `ProductsDefaultLimit = 20`.
Значения выше `ProductsMaxLimit = 30` ограничиваются до 30.

## GET /api/v1/products/categories

Контракт написан в: `backend/docs/products/README.md`

### sortBy=new

```sql
SELECT id, category, picture_url
FROM product_category
ORDER BY id DESC
```

### `sortBy=popular`

```sql
SELECT
    c.id,
    c.category,
    c.picture_url
FROM product_category c
LEFT JOIN product p ON p.category_id = c.id
GROUP BY c.id, c.category, c.picture_url
ORDER BY COUNT(p.id) DESC, c.id DESC
```

`LEFT JOIN` сохраняет категории без товаров. Они получают количество товаров
`0`. При одинаковом количестве товаров используется `c.id DESC`.

Если `limit` не указан или равен `0`, возвращаются все категории. Положительный
`limit` передаётся в SQL как `$1`. Отрицательный `limit` отклоняется.
