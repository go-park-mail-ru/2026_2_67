## API

### GET /api/v1/buyers/{id}/profile
Этот `endpoint` также сделан для того, чтобы можно было просмотреть профиль покупателя не только владельцу.

#### Request
**При любом запросе**

#### Response
- Если покупатель с указанным `id` НЕ существует:
  - `status_code`: 404

- Если пользователь с указанным `id` существует:
  - `status_code`: 200
  - `Body`:
  ```json
  {
    "name": String,
    "surname": String,
    "patronymic": String,
    "telephone": String,
    "email": String,
    "birthDate": String,
    "description": String
    "avatarUrl": String
  }
  ```

### PATCH /api/v1/buyers/{id}/profile

#### Request
`HTTP`-заголовок:
```
Authorization: Bearer $(accessToken)
```
`Body`:
```json
{
  "name": String,
  "surname": String,
  "patronymic": String,
  "telephone": String,
  "email": String,
  "birthDate": String,
  "description": String
  "avatar": Picture
}
```
**Подразумевается**, что отправленное `Body` будет полностью изменять существующее.

#### Response

- Если `Body` имеет неверный формат:
  - `status_code`: 400

- Если покупатель с `id` не существует:
  - `status_code`: 404

- `accessToken` неверен:
  - `status_code`: 401

- Если `id` не соответствует из `accessToken`:
  - `status_code`: 403

- Если изменённые поля невалидны:
  - `status_code`: 422
  - `Body`:
  ```json
  {
    "nameErrMessage": String,
    "surnameErrMessage": String,
    "patronymicErrMessage": String,
    "telephoneErrMessage": String,
    "emailErrMessage": String,
    "birthDateErrMessage": String,
    "descriptionErrMessage": String,
    "avatarErrMessage": String
  }
  ```

- Если изменение прошло успешно:
  - `status_code`: 200
  - `Body`:
  ```json
  {
    "name": String,
    "surname": String,
    "patronymic": String,
    "telephone": String,
    "email": String,
    "birthDate": String,
    "description": String
    "avatarUrl": String
  }
  ```
