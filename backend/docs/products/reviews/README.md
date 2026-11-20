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
        "buyerID": Int,
        "buyerName": String,
        "buyerSurname": String,
        "buyerAvatarUrl": String,
        "review": String | null,
        "rating": Int,
        "createdAt": Date,
      }
    ],
    "avgRating": Float,
    "reviewsCount": Int,
    "rating5Count": Int,
    "rating4Count": Int,
    "rating3Count": Int,
    "rating2Count": Int,
    "rating1Count": Int,
    "wasReview": Bool
  }
  ```
  `wasReview` - означает, оставлял ли покупатель отзыв или нет. Если пользователь не зарегистрирован, то оно `false`.

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

- `accessToken` не валиден или покупатель не покупал продукт `id`:
  - `status_code`: 403

- покупатель с `accessToken` уже оставлял отзыв на продукт `id` или `rating` не валидный:
  - `status_code`: 422

- иначе, если всё хорошо:
  `status_code`: 200
