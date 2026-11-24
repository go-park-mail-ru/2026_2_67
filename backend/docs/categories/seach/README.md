## API

### GET /api/v1/categories/search[?queryParams=...]

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
