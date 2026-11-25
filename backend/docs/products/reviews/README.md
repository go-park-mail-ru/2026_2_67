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
```
Authorization: Bearer $(accessToken)
```

```json
{
  "rating": Int,
  "review": String
}
```
Если `review` = "", то это оценка (не отзыв), т.е. пустой текст.

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

### GET /api/v1/products/{product_id}/reviews/{buyer_id}

#### Request
**Пустое тело**

#### Response
- Если `product_id` или `buyer_id` не существует:
  - `status_code`: 404

- Иначе, если всё хорошо:
  - `status_code`: 200
  - `Body`:
  ```
  {
    "rating": Int,
    "review": String
  }
  ```
  Если покупатель ничего не писал, то `review` = ""

### PATCH /api/v1/products/{product_id}/reviews/{buyer_id}

#### Request
`HTTP`-заголовок:
```
Authorization: Bearer $(accessToken)
```

```json
{
  "rating": Int,
  "review": String
}
```

Если `review` = "", то это является оценкой (в отзывах не отображается).

#### Response

- Если `product_id` или `buyer_id` не существует:
  - `status_code`: 404

- Если `accessToken` невалиден:
  - `status_code`: 401

- Если `user_id` в `accessToken` не совпадает с `buyer_id`:
  - `status_code`: 403

- Если нарушено бизнес-правило: `rating` не верный или продукт `product_id` не куплен:
  - `status_code`: 422
  - `Body`:
  ```json
  {
    "errMessage": String
  }
  ```

- Иначе, если всё хорошо:
  - `status_code`: 200
  - `Body`:
  ```json
  {
    "rating": Int,
    "review": String
  }
  ```
