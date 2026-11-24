## API

### GET /api/v1/products/search?[QueryParams=...]

#### Request
`Query`-параметры включают в себя 2 параметры:
  - `limit` = `Int`: предельное количество продуктов
  - `query` = `String`: запрос

#### Response
Всегда правильный:
  - `status_code`: 200
  - `Body`:
  ```json
  {
    "products": [
      "id": Int,
      "name": String,
      "pictureUrl": String,
      "sellerName": String
    ]
  }
  ```
  Если картинок нет, то `pictureUrl` = "".
