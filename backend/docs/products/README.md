## API

### GET /api/v1/products[?searchParams=...]

#### Request
`searchParams` передаются как `QueryParameters`.
- `category` = `String` - название категории продуктов
- `seller` = `[String, String, String...]` - названия продавцов продуктов, если список пустой, то все возможные
- `query` = `String` - строка, которая написана в поисковом поле
- `cursor` = `String` - специальная позиция курсора в БД для скроллинга
- `limit` = `Int` - количество, сколько стоит отдать продуктов (на бэке ограничен)
- `priceMin` = `Float` - минимальная цена, если не указано, то 0.
- `priceMax` = `Float` - максимальная цена, если не указано, то +INF.
- `sortBy` = `enum(buy-count, asc-price, desc-price, high-rating, new, name)` - как именно должны быть отсортированы товары:
  - `buy-count` - первыми идут те, которые имеют самые большое количество покупок.
  - `asc-price` - первыми идут те, которые имеют маленькую стоимость.
  - `desc-price` - первыми идут те, которые имеют большую стоимость.
  - `high-rating` - первыми идут те, которые имеют высокий средний рейтинг.
  - `new` - первыми идут те, которые выложены недавно.
  - `name` - первыми идут те, которые лексиграфически меньше по названию.
  - по умолчанию `new`.

`searchParams` являются опциональными:
- Если не один не указан, возвращается список (критерий регулируется на беке)
- Если указан один, то запрос учитывает ТОЛЬКО его.
- Если указаны несколько, то эти критерии логически соединяются через **И** (то есть: наушники **И** Apple).

Если в результате получается пустой результат, то отдаётся пустой результат.

#### Response
- Всегда правильный:
  - `status_code`: 200
  - `Body`:
  ```json
  {
    "products": [
      {
        "id": Int,
        "name": String,
        "pictureUrls": [
          String...
        ],
        "price": Float,
        "discount": Int,
        "priceResult": Float,
        "rating": Float,
        "reviewsCount": Int,
        "sellerId": Int,
        "sellerName": String,
        "sellerAvatarUrl": String
      },
      "nextCursor": String,
    ]
  }
  ```
  Массив уже отсортирован по правильному критерию.

Если `nextCursor` является пустой строкой, то это страница является последней

### GET /api/v1/products/{id}

#### Request
**Пустой**

#### Response
- Если `id` не найден:
  - `status_code`: 404

- Иначе:
  - `status_code`: 200
  - `Body`:
  ```json
  {
    "id": Int,
    "name": String,
    "category": String,
    "pictureUrls": [
      String...
    ],
    "price": Float,
    "discount": Int,
    "priceResult": Float,
    "availableCount": Int,
    "description": String,

    "sellerName": String,
    "sellerAvatarUrl": String,

    "rating": Float,
    "marksCount": Int,
    "reviewsCount": Int,
  }
  ```
