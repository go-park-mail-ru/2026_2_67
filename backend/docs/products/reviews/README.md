## API

### GET /api/v1/products/{id}/reviews

#### Request
**Пустое тело**

#### Response
- Всегда положительный:
  - `status_code`: 200
  - `Body`:
  ```json
  {
    "reviews": [
      {
        "buyerName": String,
        "buyerSurname": String,
        "buyerAvatarUrl": String,
        "review": String | null,
        "raiting": Int,
        "createdAt": Date,
      }
    ],
    "avgRating": Float,
    "reviewsCount": Int,
    "raiting5Count": Int,
    "raiting4Count": Int,
    "raiting3Count": Int,
    "raiting2Count": Int,
    "raiting1Count": Int
  }
  ```

### POST /api/v1/products/{id}/reviews

#### Request
`HTTP`-заголовок:
`Authorization:`

```json
{
  "raiting": Int,
  "review": String
}
```

#### Response
- `id` не существует или не валидный:
  - `status_code`: 404

- `accessToken` не валиден или покупатель не покупал продукт `id`:
  - `status_code`: 403

- покупатель с `accessToken` уже оставлял отзыв на продукт `id`:
  - `status_code`: 422

- иначе, если всё хорошо:
  `status_code`: 200
