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
- Если указан неверный формат `Body`:
  - `status_code`: 400

- Если `accessToken` не валиден:
  - `status_code`: 401

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

### GET /api/v1/orders[?filterParams=...]

Возвращает все заказы авторизованного пользователя.

#### Request

`filterParams` включает в себя параметры:
- `status` = `String` - статус заказов.
- `sortBy` = `enum(desc-date, asc-date)`:
  - `desc-date`: сортирует заказы по убыванию даты создания заказов.
  - `asc-date`: сортирует заказы по возрастанию даты созданию заказов.
  - по умолчанию `desc-date`
- `limit` = `Int` - количество заказов.
- `cursor` = `String` - курсор для скролинга, т.е. информация о последнем полученном заказе.

Если аргументы не указаны, то возвращаются `sortBy` = `desc-date` и `limit` = `BackendDefaultLimit`.
Аргументы соединяются через логический **И**.

`HTTP`-заголовок:
```
Authorization: Bearer $(accessToken)
```

#### Response
- Если `accessToken` не валиден:
  - `status_code`: 401

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
