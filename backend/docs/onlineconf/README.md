# OnlineConf: Сущности и API-контракты

Сервис динамических конфигураций и фича-тогглов (feature flags). Позволяет менять поведение Backend и Frontend без перезапуска и повторного деплоя (время применения изменений — до 15–30 секунд через фоновый кеш/поллинг).

## Сущности

### 1. Конфиг (`Config`)
Хранит параметр, который возвращает конкретное значение: строку, число, boolean, массив или JSON-объект.

| Поле | Тип | Описание |
|---|---|---|
| `id` | `Int` | Уникальный идентификатор |
| `key` | `String` | Имя ключа внутри скоупа (например, `promo_banner`, `delivery_pricing`) |
| `scope` | `String` | Область использования: `frontend`, `backend`, `orders`, `products`, `auth` и др. |
| `valueType` | `enum(string, number, boolean, array, json)` | Тип хранимого значения |
| `value` | `Any (JSON)` | Само значение конфига (валидируется согласно `valueType`) |
| `description` | `String` | Описание назначения конфига |
| `createdBy` | `Int` | `userId` администратора, создавшего ключ |
| `updatedBy` | `Int` | `userId` администратора, изменившего ключ |
| `createdAt` | `String` | Дата создания (RFC3339) |
| `updatedAt` | `String` | Дата обновления (RFC3339) |

### 2. Фича-тоггл / Фича-флаг (`FeatureFlag`)
Управляет включением функциональности для конкретных пользователей или процента аудитории. На выходе для каждого пользователя вычисляется в `Boolean`.

| Поле | Тип | Описание |
|---|---|---|
| `id` | `Int` | Уникальный идентификатор |
| `key` | `String` | Имя флага внутри скоупа (например, `express_checkout`, `smart_search`) |
| `scope` | `String` | Область использования (`frontend`, `backend`, `orders`, `products` и др.) |
| `enabled` | `Boolean` | Глобальный рубильник флага |
| `rolloutPercent` | `Int (0..100)` | Процент пользователей, для которых флаг активен |
| `whitelist` | `[Int]` | Список `userId`, для которых флаг **принудительно включён** |
| `blacklist` | `[Int]` | Список `userId`, для которых флаг **принудительно выключен** |
| `description` | `String` | Описание функциональности |
| `createdBy` | `Int` | `userId` создателя |
| `updatedBy` | `Int` | `userId` последнего редактора |
| `createdAt` | `String` | Дата создания (RFC3339) |
| `updatedAt` | `String` | Дата обновления (RFC3339) |

#### Правило вычисления флага для `userId`:
1. Если `userId` находится в `blacklist` → `false` (`reason: "blacklist"`).
2. Если `userId` находится в `whitelist` → `true` (`reason: "whitelist"`).
3. Если `enabled == false` → `false` (`reason: "disabled"`).
4. Иначе считается детерминированный бакет `bucket = fnv32a(scope + "." + key + ":" + userId) % 100`. Если `bucket < rolloutPercent` → `true`, иначе `false` (`reason: "rollout"`).

### 3. Запись истории (`HistoryEntry`)
Фиксирует каждое создание, обновление или удаление конфига/флага.

| Поле | Тип | Описание |
|---|---|---|
| `id` | `Int` | Идентификатор записи |
| `entityType` | `enum(config, flag)` | Тип сущности |
| `entityId` | `Int` | ID конфига или флага |
| `entityKey` | `String` | Ключ на момент изменения |
| `scope` | `String` | Скоуп |
| `action` | `enum(create, update, delete)` | Действие |
| `oldValue` | `Object \| null` | Состояние до изменения |
| `newValue` | `Object \| null` | Состояние после изменения |
| `changedBy` | `Int` | `userId` администратора |
| `createdAt` | `String` | Время изменения (RFC3339) |

### 4. Администратор (`Admin`)
| Поле | Тип | Описание |
|---|---|---|
| `userId` | `Int` | ID пользователя из `user` |
| `role` | `enum(superadmin, admin)` | Роль в админке OnlineConf |
| `createdAt` | `String` | Дата выдачи прав |

---

## API

Все админские ручки (`/api/v1/onlineconf/admin/*`) требуют валидный JWT и наличие `userId` в `oc_admins`.

### GET /api/v1/onlineconf/admin/me

Проверяет права текущего пользователя для входа в админ-панель.

#### Request
`HTTP`-заголовок:
```
Authorization: Bearer $(accessToken)
```

#### Response
- Если `accessToken` не валиден:
  - `status_code`: 401
- Если пользователь не является администратором OnlineConf:
  - `status_code`: 403
  - `Body`:
  ```json
  {
    "errMessage": "нет доступа к панели администратора"
  }
  ```
- Если все успешно:
  - `status_code`: 200
  - `Body`:
  ```json
  {
    "userId": Int,
    "role": String
  }
  ```

---

### GET /api/v1/onlineconf/admin/configs[?scope=...&q=...]

Возвращает список конфигов с фильтрацией по `scope` и поиском по названию/описанию (`q`).

#### Request
Query-параметры:
- `scope` = `String` (опционально) — фильтр по скоупу.
- `q` = `String` (опционально) — поиск по `key` и `description` (`ILIKE`).

`HTTP`-заголовок:
```
Authorization: Bearer $(accessToken)
```

#### Response
- `status_code`: 401 / 403 при отсутствии токена или прав.
- `status_code`: 200
  - `Body`:
  ```json
  {
    "configs": [
      {
        "id": Int,
        "key": String,
        "scope": String,
        "valueType": String,
        "value": Any,
        "description": String,
        "createdBy": Int,
        "updatedBy": Int,
        "createdAt": String,
        "updatedAt": String
      }
    ]
  }
  ```

---

### POST /api/v1/onlineconf/admin/configs

Создает новый конфигурационный параметр и записывает событие в `oc_history`.

#### Request
`HTTP`-заголовок:
```
Authorization: Bearer $(accessToken)
```
`Body`:
```json
{
  "key": String,
  "scope": String,
  "valueType": String,
  "value": Any,
  "description": String
}
```

#### Response
- Если неверный формат `Body` или значение не соответствует `valueType`:
  - `status_code`: 400 / 422
  - `Body`:
  ```json
  {
    "errMessage": String
  }
  ```
- Если ключ в этом `scope` уже существует:
  - `status_code`: 409
  - `Body`:
  ```json
  {
    "errMessage": "конфиг с таким ключом в данном scope уже существует"
  }
  ```
- Если все успешно:
  - `status_code`: 201
  - `Body`: созданный объект `Config`.

---

### PUT /api/v1/onlineconf/admin/configs/{id}

Обновляет значение, тип или описание конфига и записывает событие в `oc_history`.

#### Request
`HTTP`-заголовок:
```
Authorization: Bearer $(accessToken)
```
`Body`:
```json
{
  "valueType": String,
  "value": Any,
  "description": String
}
```

#### Response
- `status_code`: 404 — если конфиг не найден.
- `status_code`: 422 — если `value` не соответствует `valueType`.
- `status_code`: 200 — обновленный объект `Config`.

---

### DELETE /api/v1/onlineconf/admin/configs/{id}

Удаляет конфиг и записывает событие в `oc_history`.

#### Request
`HTTP`-заголовок:
```
Authorization: Bearer $(accessToken)
```

#### Response
- `status_code`: 404 — если конфиг не найден.
- `status_code`: 204 — успешно удален.

---

### GET /api/v1/onlineconf/admin/flags[?scope=...&q=...]

Возвращает список фича-тогглов с фильтрацией по `scope` и поиском по названию (`q`).

#### Request
Query-параметры:
- `scope` = `String` (опционально)
- `q` = `String` (опционально)

`HTTP`-заголовок:
```
Authorization: Bearer $(accessToken)
```

#### Response
- `status_code`: 200
  - `Body`:
  ```json
  {
    "flags": [
      {
        "id": Int,
        "key": String,
        "scope": String,
        "enabled": Boolean,
        "rolloutPercent": Int,
        "whitelist": [Int],
        "blacklist": [Int],
        "description": String,
        "createdBy": Int,
        "updatedBy": Int,
        "createdAt": String,
        "updatedAt": String
      }
    ]
  }
  ```

---

### POST /api/v1/onlineconf/admin/flags

Создает новый фича-тоггл и записывает событие в `oc_history`.

#### Request
`HTTP`-заголовок:
```
Authorization: Bearer $(accessToken)
```
`Body`:
```json
{
  "key": String,
  "scope": String,
  "enabled": Boolean,
  "rolloutPercent": Int,
  "whitelist": [Int],
  "blacklist": [Int],
  "description": String
}
```

#### Response
- `status_code`: 422 — если `rolloutPercent` вне `0..100` или `key`/`scope` пустые.
- `status_code`: 409 — если флаг с таким `(scope, key)` уже существует.
- `status_code`: 201 — созданный объект `FeatureFlag`.

---

### PUT /api/v1/onlineconf/admin/flags/{id}

Обновляет параметры раскатки фича-тоггла и записывает событие в `oc_history`.

#### Request
`HTTP`-заголовок:
```
Authorization: Bearer $(accessToken)
```
`Body`:
```json
{
  "enabled": Boolean,
  "rolloutPercent": Int,
  "whitelist": [Int],
  "blacklist": [Int],
  "description": String
}
```

#### Response
- `status_code`: 404 — если флаг не найден.
- `status_code`: 422 — невалидные данные.
- `status_code`: 200 — обновленный объект `FeatureFlag`.

---

### DELETE /api/v1/onlineconf/admin/flags/{id}

Удаляет фича-тоггл и записывает событие в `oc_history`.

#### Response
- `status_code`: 404 — если флаг не найден.
- `status_code`: 204 — успешно удален.

---

### GET /api/v1/onlineconf/admin/flags/{id}/evaluate?userId={userId}

Проверяет результирующее состояние тоггла для конкретного пользователя с объяснением причины.

#### Request
Query-параметры:
- `userId` = `Int` (обязательно)

`HTTP`-заголовок:
```
Authorization: Bearer $(accessToken)
```

#### Response
- `status_code`: 200
  - `Body`:
  ```json
  {
    "flagId": Int,
    "key": String,
    "scope": String,
    "userId": Int,
    "enabled": Boolean,
    "reason": String,
    "bucket": Int
  }
  ```

---

### GET /api/v1/onlineconf/admin/history[?entityType=...&entityId=...&limit=50]

Возвращает историю изменений конфигов и флагов.

#### Request
Query-параметры:
- `entityType` = `enum(config, flag)` (опционально)
- `entityId` = `Int` (опционально)
- `limit` = `Int` (по умолчанию `50`)

`HTTP`-заголовок:
```
Authorization: Bearer $(accessToken)
```

#### Response
- `status_code`: 200
  - `Body`:
  ```json
  {
    "history": [
      {
        "id": Int,
        "entityType": String,
        "entityId": Int,
        "entityKey": String,
        "scope": String,
        "action": String,
        "oldValue": Any,
        "newValue": Any,
        "changedBy": Int,
        "createdAt": String
      }
    ]
  }
  ```

---

### GET /api/v1/onlineconf/runtime/frontend

Публичная ручка для продуктового Frontend. Читается из in-memory кеша `onlineconf` (без запросов в БД). Если передан валидный `Authorization: Bearer $(accessToken)`, флаги вычисляются для `userId` из токена; иначе используется `userId = 0` (или `X-Device-Id`).

#### Request
`HTTP`-заголовок (опционально):
```
Authorization: Bearer $(accessToken)
```

#### Response
- `status_code`: 200
  - `Body`:
  ```json
  {
    "configs": {
      "promo_banner": {
        "enabled": true,
        "text": "Скидка 15%"
      }
    },
    "flags": {
      "express_checkout": true,
      "discount_badge_v2": false
    }
  }
  ```

---

### GET /internal/v1/onlineconf/snapshot

Внутренняя ручка для основного Backend (`server`). Фоновый воркер монолита опрашивает её каждые `10–15` секунд и обновляет локальный in-memory кеш (`sync.RWMutex`).

#### Response
- `status_code`: 200
  - `Body`:
  ```json
  {
    "configs": [
      {
        "scope": String,
        "key": String,
        "valueType": String,
        "value": Any
      }
    ],
    "flags": [
      {
        "scope": String,
        "key": String,
        "enabled": Boolean,
        "rolloutPercent": Int,
        "whitelist": [Int],
        "blacklist": [Int]
      }
    ]
  }
  ```
