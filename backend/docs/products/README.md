## API

### GET /api/v1/products[?searchParams=...]

Ручка нужна для того, чтобы предоставить **продукты**, которые удовлетворяют `Query`-параметрам.

#### Request
`searchParams` передаются как `QueryParameters`.
- `category` = `[String, String, String...]` - название категории продуктов, если список пустой, то все возможные
- `seller` = `[String, String, String...]` - названия продавцов продуктов, если список пустой, то все возможные
- `query` = `String` - строка, которая написана в поисковом поле
- `cursor` = `String` - специальная позиция курсора в БД для скроллинга
- `limit` = `Int` - количество, сколько стоит отдать продуктов (на бэке ограничен)
- `priceMin` = `Float` - минимальная цена, если не указано, то 0.
- `priceMax` = `Float` - максимальная цена, если не указано, то +INF.
- `ratingMin` = `Float` - минимальный рейтинг, если не указано, то 0.
- `ratingMax` = `Float` - максимальный рейтинг, если не указано, то 5.
- `sortBy` = `enum(buy-count, asc-price, desc-price, desc-rating, asc-rating, new)` - как именно должны быть отсортированы товары:
  - `buy-count` - первыми идут те, которые имеют самые большое количество покупок.
  - `asc-price` - первыми идут те, которые имеют маленькую стоимость.
  - `desc-price` - первыми идут те, которые имеют большую стоимость.
  - `desc-rating` - первыми идут те, которые имеют высокий средний рейтинг.
  - `asc-rating` - первыми идут те, которые имеют низкий средний рейтинг.
  - `new` - первыми идут те, которые выложены недавно.
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
        "availableCount": Int,
        "sellerId": Int,
        "sellerName": String,
        "sellerAvatarUrl": String
      }
    ],
    priceResultMin: Float,
    priceResultMax: Float,
    "nextCursor": String,
  }
  ```
  Массив уже отсортирован по правильному критерию.
  `priceResultMin`, `priceResultMax` - это цена продуктов, с учётом их скидок, которые имеют краевые цены во всей БД.

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
    "reviewsCount": Int,
    "marksCount": Int,
    "rating5Count": Int,
    "rating4Count": Int,
    "rating3Count": Int,
    "rating2Count": Int,
    "rating1Count": Int,

    "canReview": Bool,
    "wasReview": Bool
  }
  ```

  `marksCount` - количество оценок, т.е: (звездочки) + (звездочки & написал текст). **текст** = отзыв, т.е. *Дада, предмет имба советую*\
  `reviewsCount` - количество отзывов, т.е. (звездочки & написал текст).\
  Всегда: `reviewsCount` <= `marksCount`.\
  `rating1Count` - количество покупателей, которые поставили рейтинг 1 (независимо писали ли они текст или нет).\
  `rating2Count` - количество покупателей, которые поставили рейтинг 2 (независимо писали ли они текст или нет).\
  ...\
  `rating5Count` - количество покупателей, которые поставили рейтинг 5 (независимо писали ли они текст или нет).\
  `review` - информация о `review` покупателя. Если покупатель уже оставлял `review`, то присылается `id` `review`, иначе `id` = 0.
