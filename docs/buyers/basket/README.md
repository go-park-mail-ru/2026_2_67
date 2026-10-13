## API

### GET /api/v1/buyers/{id}/basket

#### Request
`HTTP`-заголовок:
```
Authorization: Bearer $(accessToken)
```

#### Response
- Если `Body` имеет неверный формат или не указан `Authorization`:
  - `status_code`: 400

- Если пользователь с `id` не существует:
  - `status_code`: 404

- Если `id` не совпадает с `id` в `accessToken` или `accessToken` неверен:
  - `status_code`: 403

- Если все успешно:
  - `status_code`: 200
  - `Body`:
  ```json
  {
    "basketProducts": [
      {
        "category": String,
        "name": String,
        "pictureUrl": String,
        "price": Float,
        "count": Int
      }
    ],
    "priceSum": Float
  }
  ```

### POST /api/v1/buyers/{id}/basket/products

#### Request
`HTTP`-заголовок:
```
Authorization: Bearer $(accessToken)
```

`Body`:
```json
{
  "productId": Int,
  "count": Int
}
```

#### Response
- Если `Body` имеет неверный формат или не указан `Authorization`:
  - `status_code`: 400

- Если пользователь с `id` не существует:
  - `status_code`: 404
  - `Body`:
  ```json
  {
    "errMessage": "не найден пользователь с id = $(id)"
  }
  ```

- Если `id` не совпадает с `id` в `accessToken` или `accessToken` неверен:
  - `status_code`: 403

- Если `productId` неверен:
  - `status_code`: 404
  - `Body`:
  ```json
  {
    "errMessage": "не найден товар с id = $(id)"
  }
  ```

- Если `count` слишком большой:
  - `status_code`: 422

- Если все успешно:
  - `status_code`: 200


### PATCH /api/v1/buyers/{buyer_id}/basket/products/{product_id}

#### Request
`HTTP`-заголовок:
```
Authorization: Bearer $(accessToken)
```
`Body`:
```json
{
  "count": Int
}
```

`count` в `Body` - это новое количество товара, которое будет в корзине. Если требуется удалить, стоит установить `count` = `0`.

#### Response
- Если `Body` имеет неверный формат или не указан `Authorization`:
  - `status_code`: 400

- Если не существует `buyer_id`:
  - `status_code`: 404
  - `Body`:
  ```json
  {
    "errMessage": "не найден пользователь с id = $(buyer_id)"
  }
  ```

- Если не существует `product_id`:
  - `status_code`: 404
  - `Body`:
  ```json
  {
    "errMessage": "не найден товар с id = $(product_id)"
  }
  ```

- Если `count` в `Body` неверный: отрицательный или не соответствует реальному значению:
  - `status_code`: 422

- Если все успешно:
  - `status_code`: 200

