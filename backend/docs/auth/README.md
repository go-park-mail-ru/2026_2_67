## API
**Авторизация через JWT-токен**

**JSON поля** и **Cookie** пишутся в **camelCase**.

### POST /api/v1/auth/login

#### Request
`Body`:
```json
{
  "email": String,
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

- Если неверные данные (не существует email и т.п.):
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
```
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

### POST /api/v1/auth/password

Смена/сброс пароля авторизованного пользователя. Доступен также по алиасу `POST /api/v1/auth/reset-password`.

#### Request
`HTTP`-заголовок:
```
Authorization: Bearer $(accessToken)
```

`Body`:
```json
{
  "password": String
}
```
*(Также поддерживается ключ `"newPassword"`)*

#### Response
- Если `Body` имеет неверный формат или отсутствует заголовок `Authorization`:
  - `status_code`: 400

- Если невалидный `accessToken` (просрочен и т.п.):
  - `status_code`: 403

- Если пароль не прошел проверку (менее 8 символов, нет заглавных/строчных букв, цифр или спецсимволов):
  - `status_code`: 422
  - `Body`:
  ```json
  {
    "passwordErrMessage": String
  }
  ```

- Если пользователь не найден:
  - `status_code`: 404

- Иначе (пароль успешно изменен):
  - `status_code`: 200

### Примечание
Фронтенд отправляет у авторизованных пользователей в `HTTP`-заголовке `Authorization`:
```
Authorization: Bearer $(accessToken)
```
