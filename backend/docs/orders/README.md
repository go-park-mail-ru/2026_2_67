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
  "pickupPointId": Int,
  "promocode": String
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
    "errPickupPointIdMessage": String,
    "promocodeErrMessage": String
  }
  ```

- Если при заказе произошла ошибка (к примеру, товары закончились):
  - `status_code`: 422
  - `Body`:
  ```json
  {
    "errMessage": "при создании заказа произошла ошибка",
    "productErrs": [
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

### GET /api/v1/orders

Возвращает все заказы авторизованного пользователя. Заказы отсортированы от новых к старым.

#### Request
`HTTP`-заголовок:
```
Authorization: Bearer $(accessToken)
```

#### Response
- Если не указан `Authorization`:
  - `status_code`: 400

- Если `accessToken` не валиден:
  - `status_code`: 403

- Если все успешно:
  - `status_code`: 200
  - `Body`:
  ```json
  {
    "orders": [
      {
        "id": Int,
        "number": Int,
        "status": String,

        "createdAt": String,
        "updatedAt": String,

        "pickupPoint": {
          "id": Int,
          "longitude": Float,
          "latitude": Float,
          "startTime": String,
          "endTime": String
        },

        "promocode": String,

        "products": [
          {
           "productId": Int,
            "name": String,
            "description": String,
            "category": String,
            "sellerName": String,
            "sellerAvatarUrl": String,
            "pictureUrls": [
              String...
            ],
            "count": Int,
            "unitPrice": Float,
            "totalPrice": Float
          }
        ],
        "totalCount": Int,
        "totalPrice": Float
      }
    ]
  }
  ```

Если у заказа нет промокода, поле `promocode` имеет значение `null`. Если у пользователя нет заказов, возвращается пустой массив `orders`.
