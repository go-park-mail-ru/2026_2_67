## API

### POST /api/v1/orders

#### Request
`HTTP`-заголовок:
```
Authorization: Bearer $(accessToken)
```
`Body`:
```json
{
  "longitude": Float,
  "latitude": Float,
  "promocode": Float
}
```

#### Response
- Если указан неверный формат `Body` или не указан `Authorization`:
  - `status_code`: 400

- Если `accessToken` не валиден:
  - `status_code`: 403

- Если неверные данные в `Body`:
  - `status_code`: 422
  - `Body`:
  ```json
  {
    "errMessage": "Body содержит неверные данные",
    "longitudeErr": String,
    "latitudeErr": String,
    "promocodeErr": String
  }
  ```

- Если при заказе произошла ошибка (к примеру, товары закончились):
  - `status_code`: 422
  - `Body`:
  ```json
  {
    "errMessage": "при создании заказа произошла ошибка",
    "productsErr": [
      {
        "productId": Int,
        "name": String,
        "seller": String,
        "errMessage": String
      }
    ]
  }
  ```

- Если все успешно:
  - `status_code`: 200
  - `Body`:
  ```json
  {
    "orderId": Int
  }
  ```
