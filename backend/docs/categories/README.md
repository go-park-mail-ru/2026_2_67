### GET /api/v1/categories[?QueryParameters=...]

#### Request
`QueryParameters`:
- `limit` = `Int` - какое количество нужно.
- `sortBy` = `enum(new, popular)` - критерий отдачи.
  - `new` - первыми идут те, которые вышли недавно.
  - `popular` - первыми идут популярные категории (условно, те у кого больше всего продуктов)
  - по умолчанию `new`.

Если не указан `limit` возвращает **все**, иначе `min{ДОСТУПНОЕ_КОЛИЧЕСТВО, limit}`.

#### Response
- Всегда правильный:
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
