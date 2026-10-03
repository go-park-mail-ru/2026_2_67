## API
**Авторизация через JWT-токен**

**JSON поля** и **Cookie** пишутся в **camelCase**.

### POST /auth/login

#### Request
`Body`:
```json
{
  "loginOrEmail": String,
  "password": String
}
```

#### Response
- Если `Body` имеет неверный формат:
  - `status_code`: 400

- Если `user` уже залогинин:
  - `status_code`: 200
  - `Body`:
  ```json
  {
    "accessToken": String
  }
  ```
  - `Set-Cookie`:
  ```
  refreshToken: String
  ```

- Если неверные данные (не существует login, email и т.п.):
  - `status_code`: 401

- Иначе (всё верно):
  - `status_code`: 200
  - `Body`:
  ```json
  {
    "userId": Int,
    "accessToken": String
  }
  ```
  - `Set-Cookie`:
  ```
  refreshToken: String
  ```

### POST /auth/refresh

#### Request
`Cookie`:
```
refreshToken: String
```

#### Response
- Если `Body` не соответствует формату:
  - `status_code`: 400

- Если проблемы с `refreshToken` (истёк, не существует и т.д.):
 - `status_code`: 403

- Если всё нормально:
  - `status_code`: 200
  - `Body`:
  ```json
  {
    "accessToken": String
  }
  ```

### GET /api/v1/products

#### Request
**Пустой**

#### Response
- Всегда верен:
  - `status_code`: 200
  - `Body`:
  ```json
  {
    [
      {
        "productName": String,
        "productPictureUrls": [
          String...
        ]
        "productPrice": Int,
        "productRating": Float,
        "productReviewsCount": Int
      }
    ]
  }
  ```

### POST /auth/register

#### Request
`Body`:
```json
{
  "login": String,
  "email": String,
  "password": String
}
```

#### Response
- Если пользователь существует:
  - `status_code`: 409

- Если данные не прошли проверку (плохой пароль, email не поддерживается форматом):
  - `status_code`: 401
  - `Body`:
  ```json
  {
    "errMessage": String
  }
  ```

- Если пользователь создан успешно:
  - `status_code`: 200
  - `Body`:
  ```json
  {
    "userId": Int,
    "accessToken": String
  }
  ```
  - `Set-Cookie`:
  ```
  refreshToken: String
  ```

### Примечание
Фронтенд отправляет у авторизованных пользователей в `HTTP`-заголовке `Authorization`:
```
Authorization: Bearer $(accessToken)
```
