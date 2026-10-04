## API Товаров

**JSON поля** пишутся в **camelCase**.

### GET /api/v1/products

#### Request
**Пустой**

#### Response
- Всегда верен:
  - `status_code`: 200
  - `Body`:
  ```json
  [
    {
      "productName": String,
      "productPictureUrls": [
        String...
      ],
      "productPrice": Int,
      "productRating": Float,
      "productReviewsCount": Int
    }
  ]
  ```

### Примечание
Фронтенд отправляет у авторизованных пользователей в `HTTP`-заголовке `Authorization`:
```
Authorization: Bearer $(accessToken)
```
