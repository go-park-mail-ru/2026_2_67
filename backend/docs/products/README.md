## API

### GET /api/v1/products[?searchParams=...]

#### Request
`searchParams` передаются как `QueryParameters`.
- `category` = `String` - название категории товаров
- `seller` = `String` - название продавца товаров
- `cursor` = `String` - специальная позиция курсора в БД для скроллинга
- `limit` = `Int` - количество, сколько стоит отдать продуктов
- `price_min` = `Int` - минимальная цена, если не указано, то 0.
- `price_max` = `Int` - максимальная цена, если не указано, то +INF.

`searchParams` являются опциональными:
- Если не один не указан, возвращается весь список (критерий регулируется на беке)
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
        "name": String,
        "pictureUrls": [
          String...
        ],
        "price": Float,
        "rating": Float,
        "reviewsCount": Int
      },
      "nextCursor": String,
      "hasMore": Boolean
    ]
  }
  ```

### GET /api/v1/categories[?limit=...]

#### Request
`limit` = `Int` передаётся как `QueryParameters`.\
Если не указан `limit` возвращает **все**, иначе `min{ДОСТУПНОЕ_КОЛИЧЕСТВО, limit}`.

#### Response
- Всегда правильный:
  - `status_code`: 200
  - `Body`:
  ```json
  {
    "categories": [
      String...
    ]
  }
  ```

### GET /api/v1/sellers[?sellerParams=...]

#### Request
`sellerParams` передается как `QueryParameters`.
- `limit` = `Int` - максимальное количество, возвращается: `min{ДОСТУПНОЕ_КОЛИЧЕСТВО, limit}`.
- `category` = `String` - продавцы, которые имеют товары в этой категории.

#### Response
- Всегда правильный:
  - `status_code`: 200
  - `Body`:
  ```json
  {
    "sellers": [
      String...
    ]
  }
  ```
