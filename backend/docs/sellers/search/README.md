## API

### GET /api/v1/sellers/search[?QueryParams=...]

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
    "sellers": [
      {
        "id": Int,
        "name": String,
        "avatarUrl": String
      }
    ]
  }
  ```
