### GET /api/v1/sellers[?sellerParams=...]

#### Request
`sellerParams` передается как `QueryParameters`.
- `limit` = `Int` - максимальное количество, возвращается: `min{ДОСТУПНОЕ_КОЛИЧЕСТВО, limit}`.
- `category` = `String` - продавцы, которые имеют товары в этой категории.

Если не указаны, то возвращается по критерию из бека.

#### Response
- Всегда правильный:
  - `status_code`: 200
  - `Body`:
  ```json
  {
    "sellers": [
      {
        "id": Int,
        "name": String
      }
    ]
  }
  ```
