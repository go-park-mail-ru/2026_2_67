# OnlineConf: Схема БД и SQL-запросы

Контракт ручек: [`OnlineConf API`](../../../backend/docs/onlineconf/README.md)

## Миграции

- [`db/migrations/000017_create_onlineconf.up.sql`](../../migrations/000017_create_onlineconf.up.sql)
- [`db/migrations/000018_seed_onlineconf.up.sql`](../../migrations/000018_seed_onlineconf.up.sql)

## Таблицы

Таблицы OnlineConf изолированы от бизнес-таблиц `vibe_market` (без `FOREIGN KEY` и `JOIN`), чтобы при выделении микросервисов базу можно было перенести без изменения запросов.

### 1. `oc_config` — конфигурационные параметры

- `id BIGSERIAL PRIMARY KEY`
- `key TEXT NOT NULL` — ключ конфига
- `scope TEXT NOT NULL` — область (`frontend`, `backend`, `orders`, `products`, `auth`)
- `value_type TEXT NOT NULL` — тип значения (`string`, `number`, `boolean`, `array`, `json`)
- `value JSONB NOT NULL` — значение в формате JSONB
- `description TEXT NOT NULL DEFAULT ''` — описание
- `created_by BIGINT NOT NULL` — `user_id` создателя
- `updated_by BIGINT NOT NULL` — `user_id` последнего редактора
- `created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()`
- `updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()`
- `UNIQUE (scope, key)`

### 2. `oc_feature_flag` — фича-тогглы

- `id BIGSERIAL PRIMARY KEY`
- `key TEXT NOT NULL` — ключ тоггла
- `scope TEXT NOT NULL` — область (`frontend`, `backend`, `orders`, `products`, `auth`)
- `enabled BOOLEAN NOT NULL DEFAULT FALSE` — глобальный рубильник
- `rollout_percent INT NOT NULL DEFAULT 0 CHECK (rollout_percent BETWEEN 0 AND 100)` — процент раскатки
- `whitelist BIGINT[] NOT NULL DEFAULT '{}'` — `user_id` для принудительного включения
- `blacklist BIGINT[] NOT NULL DEFAULT '{}'` — `user_id` для принудительного выключения
- `description TEXT NOT NULL DEFAULT ''` — описание
- `created_by BIGINT NOT NULL`
- `updated_by BIGINT NOT NULL`
- `created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()`
- `updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()`
- `UNIQUE (scope, key)`

### 3. `oc_history` — история изменений

- `id BIGSERIAL PRIMARY KEY`
- `entity_type TEXT NOT NULL` — `'config'` или `'flag'`
- `entity_id BIGINT NOT NULL` — ID сущности
- `entity_key TEXT NOT NULL` — ключ на момент изменения
- `scope TEXT NOT NULL` — скоуп
- `action TEXT NOT NULL` — `'create'`, `'update'` или `'delete'`
- `old_value JSONB` — состояние до изменения (`NULL` при создании)
- `new_value JSONB` — состояние после изменения (`NULL` при удалении)
- `changed_by BIGINT NOT NULL` — `user_id` администратора
- `created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()`

### 4. `oc_admin` — администраторы OnlineConf

- `user_id BIGINT PRIMARY KEY` — ID пользователя из `"user"`
- `role TEXT NOT NULL DEFAULT 'admin'` — `'superadmin'` или `'admin'`
- `created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()`

## SQL DDL

```sql
CREATE TABLE IF NOT EXISTS oc_config (
    id          BIGSERIAL PRIMARY KEY,
    key         TEXT NOT NULL CHECK (LENGTH(TRIM(key)) > 0),
    scope       TEXT NOT NULL CHECK (LENGTH(TRIM(scope)) > 0),
    value_type  TEXT NOT NULL CHECK (value_type IN ('string', 'number', 'boolean', 'array', 'json')),
    value       JSONB NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    created_by  BIGINT NOT NULL,
    updated_by  BIGINT NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (scope, key)
);

CREATE INDEX IF NOT EXISTS idx_oc_config_scope ON oc_config (scope);

CREATE TABLE IF NOT EXISTS oc_feature_flag (
    id              BIGSERIAL PRIMARY KEY,
    key             TEXT NOT NULL CHECK (LENGTH(TRIM(key)) > 0),
    scope           TEXT NOT NULL CHECK (LENGTH(TRIM(scope)) > 0),
    enabled         BOOLEAN NOT NULL DEFAULT FALSE,
    rollout_percent INT NOT NULL DEFAULT 0 CHECK (rollout_percent BETWEEN 0 AND 100),
    whitelist       BIGINT[] NOT NULL DEFAULT '{}',
    blacklist       BIGINT[] NOT NULL DEFAULT '{}',
    description     TEXT NOT NULL DEFAULT '',
    created_by      BIGINT NOT NULL,
    updated_by      BIGINT NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (scope, key)
);

CREATE INDEX IF NOT EXISTS idx_oc_feature_flag_scope ON oc_feature_flag (scope);

CREATE TABLE IF NOT EXISTS oc_history (
    id          BIGSERIAL PRIMARY KEY,
    entity_type TEXT NOT NULL CHECK (entity_type IN ('config', 'flag')),
    entity_id   BIGINT NOT NULL,
    entity_key  TEXT NOT NULL,
    scope       TEXT NOT NULL,
    action      TEXT NOT NULL CHECK (action IN ('create', 'update', 'delete')),
    old_value   JSONB,
    new_value   JSONB,
    changed_by  BIGINT NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_oc_history_entity ON oc_history (entity_type, entity_id, created_at DESC);

CREATE TABLE IF NOT EXISTS oc_admin (
    user_id    BIGINT PRIMARY KEY,
    role       TEXT NOT NULL DEFAULT 'admin' CHECK (role IN ('superadmin', 'admin')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```

## SQL repository/postgres

### Проверка прав администратора (`oc_admin`)

```sql
SELECT user_id, role, created_at
FROM oc_admin
WHERE user_id = $1;
```

### Поиск и список конфигов

```sql
SELECT id, key, scope, value_type, value, description, created_by, updated_by, created_at, updated_at
FROM oc_config
WHERE ($1 = '' OR scope = $1)
  AND ($2 = '' OR key ILIKE '%' || $2 || '%' OR description ILIKE '%' || $2 || '%')
ORDER BY scope ASC, key ASC;
```

### Создание / обновление / удаление конфига

```sql
INSERT INTO oc_config (key, scope, value_type, value, description, created_by, updated_by)
VALUES ($1, $2, $3, $4, $5, $6, $6)
RETURNING id, key, scope, value_type, value, description, created_by, updated_by, created_at, updated_at;
```

```sql
UPDATE oc_config
SET value_type = $2,
    value = $3,
    description = $4,
    updated_by = $5,
    updated_at = NOW()
WHERE id = $1
RETURNING id, key, scope, value_type, value, description, created_by, updated_by, created_at, updated_at;
```

```sql
DELETE FROM oc_config
WHERE id = $1
RETURNING id, key, scope, value_type, value, description;
```

### Поиск и список фича-тогглов

```sql
SELECT id, key, scope, enabled, rollout_percent, whitelist, blacklist, description, created_by, updated_by, created_at, updated_at
FROM oc_feature_flag
WHERE ($1 = '' OR scope = $1)
  AND ($2 = '' OR key ILIKE '%' || $2 || '%' OR description ILIKE '%' || $2 || '%')
ORDER BY scope ASC, key ASC;
```

### Создание / обновление / удаление фича-тоггла

```sql
INSERT INTO oc_feature_flag (key, scope, enabled, rollout_percent, whitelist, blacklist, description, created_by, updated_by)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $8)
RETURNING id, key, scope, enabled, rollout_percent, whitelist, blacklist, description, created_by, updated_by, created_at, updated_at;
```

```sql
UPDATE oc_feature_flag
SET enabled = $2,
    rollout_percent = $3,
    whitelist = $4,
    blacklist = $5,
    description = $6,
    updated_by = $7,
    updated_at = NOW()
WHERE id = $1
RETURNING id, key, scope, enabled, rollout_percent, whitelist, blacklist, description, created_by, updated_by, created_at, updated_at;
```

```sql
DELETE FROM oc_feature_flag
WHERE id = $1
RETURNING id, key, scope, enabled, rollout_percent, whitelist, blacklist, description;
```

### Запись и чтение истории (`oc_history`)

```sql
INSERT INTO oc_history (entity_type, entity_id, entity_key, scope, action, old_value, new_value, changed_by)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8);
```

```sql
SELECT id, entity_type, entity_id, entity_key, scope, action, old_value, new_value, changed_by, created_at
FROM oc_history
WHERE ($1 = '' OR entity_type = $1)
  AND ($2::BIGINT = 0 OR entity_id = $2)
ORDER BY created_at DESC
LIMIT $3;
```
