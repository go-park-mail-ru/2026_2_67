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

- Если при получении заказов произошла внутренняя ошибка:
  - `status_code`: 500

- Если все успешно:
  - `status_code`: 200
  - `Body`:
  ```json
  {
    "orders": [
      {
        "id": Int,
        "status": {
          "id": Int,
          "name": String
        },
        "createdAt": String,
        "updatedAt": String,
        "pickupPoint": {
          "id": Int,
          "longitude": Float,
          "latitude": Float,
          "startTime": String,
          "endTime": String
        },
        "promocode": {
          "id": Int,
          "code": String,
          "discount": Int,
          "availableCategoryId": Int | null,
          "startDatetime": String,
          "endDatetime": String
        },
        "products": [
          {
            "id": Int,
            "productId": Int,
            "name": String,
            "description": String | null,
            "category": {
              "id": Int,
              "name": String
            },
            "seller": {
              "id": Int,
              "name": String
            },
            "pictureUrls": [String],
            "count": Int,
            "unitPrice": Int,
            "totalPrice": Int
          }
        ],
        "notifications": [
          {
            "id": Int,
            "status": {
              "id": Int,
              "name": String
            },
            "isRead": Bool,
            "createdAt": String
          }
        ],
        "totalCount": Int,
        "totalPrice": Int
      }
    ]
  }
  ```

Если у заказа нет промокода, поле `promocode` имеет значение `null`. Если у пользователя нет заказов, возвращается пустой массив `orders`.

- `createdAt`, `updatedAt`, `startDatetime`, `endDatetime` — дата и время в формате RFC3339 (UTC).
- `startTime`, `endTime` — время работы пункта выдачи в формате `HH:MM:SS`.
- `products[].id` — идентификатор позиции заказа, `products[].productId` — идентификатор товара.
- `unitPrice` — сохраненная при оформлении цена единицы товара с учетом скидок. `products[].totalPrice` равен `count * unitPrice`.
- `totalCount` — суммарное количество единиц товаров, `totalPrice` — сумма стоимостей всех позиций заказа.
- `discount` — скидка промокода в процентах. Если `availableCategoryId` равен `null`, промокод не ограничен категорией.
- Описания, названия, изображения товаров, данные продавцов, категорий, ПВЗ и промокода возвращаются в текущем виде, а не на момент оформления заказа.
- Пустые коллекции `products`, `pictureUrls` и `notifications` возвращаются как `[]`. Получение заказов не помечает уведомления прочитанными.
