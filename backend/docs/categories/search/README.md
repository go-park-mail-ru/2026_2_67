## API

### GET /api/v1/categories/search[?queryParams=...]

Ручка нужна для того, чтобы предоставить **категории**, название которых удовлетворяют `query`.

#### Request
`Query`-параметры включают в себя:
  - `limit` = `Int`: предельное количество категорий
  - `query` = `String`: запрос

#### Response
Всегда правильный:
  - `status_code`: 200
  - `Body`:
  ```json
  {
    "categories": [
      {
        "id": Int,
        "category": String,
        "pictureUrl": String
      }
    ]
  }
  ```
