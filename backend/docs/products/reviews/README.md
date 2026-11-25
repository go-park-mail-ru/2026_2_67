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
    "rating": Float,
    "reviewsCount": Int,
    "marksCount": Int,
    "rating5Count": Int,
    "rating4Count": Int,
    "rating3Count": Int,
    "rating2Count": Int,
    "rating1Count": Int,
    "canReview": Bool,
    "nextCursor": String
  }
  ```
  `canReview` - означает, оставлял ли покупатель отзыв или нет. Если пользователь не авторизован, то оно `false`.\
  `marksCount` - количество оценок, т.е: (звездочки) + (звездочки & написал текст). **текст** = отзыв, т.е. *Дада, предмет имба советую*\
  `reviewsCount` - количество отзывов, т.е. (звездочки & написал текст).\
  Всегда: `reviewsCount` <= `marksCount`.\
  `rating1Count` - количество покупателей, которые поставили рейтинг 1 (независимо писали ли они текст или нет).\
  `rating2Count` - количество покупателей, которые поставили рейтинг 2 (независимо писали ли они текст или нет).\
  ...\
  `rating5Count` - количество покупателей, которые поставили рейтинг 5 (независимо писали ли они текст или нет).\

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
