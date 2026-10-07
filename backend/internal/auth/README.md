## API
**Авторизация через JWT-токен**

**JSON поля** и **Cookie** пишутся в **camelCase**.

### POST /api/v1/auth/login

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

- Если пользователь уже залогинин:
  - `status_code`: 200
  - `Body`:
  ```json
  {
    "accessToken": String
  }
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

### POST /api/v1/auth/refresh

#### Request
`Cookie`:
```
refreshToken: String
```

#### Response
- Если `Cookie` не соответствуют формату:
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
  - `Set-Cookie`:
  ```json
  refreshToken: String
  ```

### POST /api/v1/auth/register

#### Request
`Body`:
```json
{
  "email": String,
  "password": String
}
```

#### Response
- Если пользователь существует:
  - `status_code`: 409
  - `Body`:
  ```json
  {
    "emailErrMessage": String,
    "passwordErrMessage": String
  }
  ```

- Если пользователь залогинин, но хочет зарегистрироваться:
  - `status_code`: 400

- Если данные не прошли проверку (плохой пароль, email не поддерживается форматом):
  - `status_code`: 401
  - `Body`:
  ```json
  {
    "emailErrMessage": String,
    "passswordErrMessage": String
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

### POST /api/v1/auth/logout

#### Request
`Authorization`:
```json
"accessToken": String
```

#### Response
- Если невалидный `accessToken` (просрочен и т.п.) или отсутствует:
  - `status_code`: 403

- Иначе (всё верно):
  - `status_code`: 200
  - `Set-Cookie`:
  ```
  "Сбросить refreshToken"
  ```

### Примечание
Фронтенд отправляет у авторизованных пользователей в `HTTP`-заголовке `Authorization`:
```
Authorization: Bearer $(accessToken)
```

Для других ролей другие ручки:
- `POST /api/v1/profiles/sellers` - добавление профиля для продавца
```json
{
  "accessToken": String
}
```

`POST /api/v1/profiles/pickup-points` - добавление профиля для ПВЗ
```json
{
  "accessToken": String
}
```
