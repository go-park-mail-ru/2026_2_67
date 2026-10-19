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
  "telephone": String,
  "email": String,
  "birthDate": String,
  "description": String
  "avatarUrl": String
}
```

#### Response

- Если `Body` имеет неверный формат или не указан `Authorization`:
  - `status_code`: 400

- Если покупатель с `id` не существует:
  - `status_code`: 404

- Если `id` из `accessToken` не соответствует или `accessToken` неверен:
  - `status_code`: 403

- Если изменённые поля невалидны:
  - `status_code`: 422
  - `Body`:
  ```json
  {
    "nameErrMessage": String,
    "surnameErrMessage": String,
    "telephoneErrMessage": String,
    "emailErrMessage": String,
    "birthDateErrMessage": String,
    "descriptionErrMessage": String,
    "avatarUrlErrMessage": String
  }
  ```

- Если изменение прошло успешно:
  - `status_code`: 200
  - `Body`:
  ```json
  {
    "name": String,
    "surname": String,
    "telephone": String,
    "email": String,
    "birthDate": String,
    "description": String
    "avatarUrl": String
  }
  ```
