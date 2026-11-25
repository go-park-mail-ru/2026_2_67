## API

### GET /api/v1/products/{id}/reviews[?QueryParams=...]

#### Request
`Query`-параметры включают в себя:
  -  `soryBy` = `enum(desc-date, asc-date, desc-rating, asc-rating)` - сортировка отзывов:
    - `desc-date`:  первыми идут новые.
    - `asc-date`: первыми идут старые.
    - `desc-rating`: первыми идут с высоким рейтингом.
    - `asc-rating`: первыми идут с маленьким рейтингом.
  - `limit` = `Int` - количество отзывов.

#### Response
- Всегда положительный:
  - `status_code`: 200
  - `Body`:
  ```json
  {
    "reviews": [
      {
        "buyerId": Int,
        "buyerName": String,
        "buyerSurname": String,
        "buyerPatronymic": String,
        "buyerAvatarUrl": String,
        "review": String,
        "rating": Int,
        "createdAt": Date,
      }
    ],
    "nextCursor": String
  }
  ```
  `review` = "", если это оценка (отзыв не оставляли).

### POST /api/v1/products/{id}/reviews

#### Request
`HTTP`-заголовок:
`Authorization:`

```json
{
  "rating": Int,
  "review": String
}
```

#### Response
- `id` не существует или не валидный:
  - `status_code`: 404

- `accessToken` не валиден:
  - `status_code`: 401

- покупатель не покупал продукт с `id` или он ещё не получен или `rating` не валидный:
  - `status_code`: 422
  - `Body`:
  ```json
  {
    "errMessage": String
  }
  ```

- покупатель с `accessToken` уже оставлял отзыв на продукт `id`:
  - `status_code`: 409

- иначе, если всё хорошо:
  `status_code`: 200
