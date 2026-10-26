## API

### GET /api/v1/products[?searchParams=...]

#### Request
`searchParams` передаются как `QueryParameters`.
- `category` = `String` - название категории товаров
- `seller` = `String` - название продавца товаров
- `query` = `String` - строка, которая написана в поисковом поле
- `cursor` = `String` - специальная позиция курсора в БД для скроллинга
- `limit` = `Int` - количество, сколько стоит отдать продуктов
- `price_min` = `Float` - минимальная цена, если не указано, то 0.
- `price_max` = `Float` - максимальная цена, если не указано, то +INF.

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
        "id": Int,
        "name": String,
        "pictureUrls": [
          String...
        ],
        "price": Float,
        "rating": Float,
        "reviewsCount": Int
      },
      "nextCursor": String,
    ]
  }
  ```

Если `nextCursor` является пустой строкой, то это страница является последней

### GET /api/v1/products/categories[?limit=...]

#### Request
`limit` = `Int` передаётся как `QueryParameters`.\
Если не указан `limit` возвращает **все**, иначе `min{ДОСТУПНОЕ_КОЛИЧЕСТВО, limit}`.

#### Response
- Всегда правильный:
  - `status_code`: 200
  - `Body`:
  ```json
  {
    "productCategories": [
      {
        "id": Int,
        "category": String
      }
    ]
  }
  ```
