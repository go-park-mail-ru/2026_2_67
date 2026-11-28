# OnlineConf: РЎС…РµРјР° Р‘Р” Рё SQL-Р·Р°РїСЂРѕСЃС‹

РљРѕРЅС‚СЂР°РєС‚ СЂСѓС‡РµРє: [`OnlineConf API`](../../../backend/docs/onlineconf/README.md)

## РўР°Р±Р»РёС†С‹

РўР°Р±Р»РёС†С‹ OnlineConf РёР·РѕР»РёСЂРѕРІР°РЅС‹ РѕС‚ Р±РёР·РЅРµСЃ-С‚Р°Р±Р»РёС† `vibe_market` (Р±РµР· `FOREIGN KEY` Рё `JOIN`), С‡С‚РѕР±С‹ РїСЂРё РІС‹РґРµР»РµРЅРёРё РјРёРєСЂРѕСЃРµСЂРІРёСЃРѕРІ Р±Р°Р·Сѓ РјРѕР¶РЅРѕ Р±С‹Р»Рѕ РїРµСЂРµРЅРµСЃС‚Рё Р±РµР· РёР·РјРµРЅРµРЅРёСЏ Р·Р°РїСЂРѕСЃРѕРІ.

### 1. `oc_config` вЂ” РєРѕРЅС„РёРіСѓСЂР°С†РёРѕРЅРЅС‹Рµ РїР°СЂР°РјРµС‚СЂС‹

- `id BIGSERIAL PRIMARY KEY`
- `key TEXT NOT NULL` вЂ” РєР»СЋС‡ РєРѕРЅС„РёРіР°
- `scope TEXT NOT NULL` вЂ” РѕР±Р»Р°СЃС‚СЊ (`frontend`, `backend`, `orders`, `products`, `auth`)
- `value_type TEXT NOT NULL` вЂ” С‚РёРї Р·РЅР°С‡РµРЅРёСЏ (`string`, `number`, `boolean`, `array`, `json`)
- `value JSONB NOT NULL` вЂ” Р·РЅР°С‡РµРЅРёРµ РІ С„РѕСЂРјР°С‚Рµ JSONB
- `description TEXT NOT NULL DEFAULT ''` вЂ” РѕРїРёСЃР°РЅРёРµ
- `created_by BIGINT NOT NULL` вЂ” `user_id` СЃРѕР·РґР°С‚РµР»СЏ
- `updated_by BIGINT NOT NULL` вЂ” `user_id` РїРѕСЃР»РµРґРЅРµРіРѕ СЂРµРґР°РєС‚РѕСЂР°
- `created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()`
- `updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()`
- `UNIQUE (scope, key)`

### 2. `oc_feature_flag` вЂ” С„РёС‡Р°-С‚РѕРіРіР»С‹

- `id BIGSERIAL PRIMARY KEY`
- `key TEXT NOT NULL` вЂ” РєР»СЋС‡ С‚РѕРіРіР»Р°
- `scope TEXT NOT NULL` вЂ” РѕР±Р»Р°СЃС‚СЊ (`frontend`, `backend`, `orders`, `products`, `auth`)
- `enabled BOOLEAN NOT NULL DEFAULT FALSE` вЂ” РіР»РѕР±Р°Р»СЊРЅС‹Р№ СЂСѓР±РёР»СЊРЅРёРє
- `rollout_percent INT NOT NULL DEFAULT 0 CHECK (rollout_percent BETWEEN 0 AND 100)` вЂ” РїСЂРѕС†РµРЅС‚ СЂР°СЃРєР°С‚РєРё
- `whitelist BIGINT[] NOT NULL DEFAULT '{}'` вЂ” `user_id` РґР»СЏ РїСЂРёРЅСѓРґРёС‚РµР»СЊРЅРѕРіРѕ РІРєР»СЋС‡РµРЅРёСЏ
- `blacklist BIGINT[] NOT NULL DEFAULT '{}'` вЂ” `user_id` РґР»СЏ РїСЂРёРЅСѓРґРёС‚РµР»СЊРЅРѕРіРѕ РІС‹РєР»СЋС‡РµРЅРёСЏ
- `description TEXT NOT NULL DEFAULT ''` вЂ” РѕРїРёСЃР°РЅРёРµ
- `created_by BIGINT NOT NULL`
- `updated_by BIGINT NOT NULL`
- `created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()`
- `updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()`
- `UNIQUE (scope, key)`

### 3. `oc_history` вЂ” РёСЃС‚РѕСЂРёСЏ РёР·РјРµРЅРµРЅРёР№

- `id BIGSERIAL PRIMARY KEY`
- `entity_type TEXT NOT NULL` вЂ” `'config'` РёР»Рё `'flag'`
- 
- `entity_key TEXT NOT NULL` вЂ” РєР»СЋС‡ РЅР° РјРѕРјРµРЅС‚ РёР·РјРµРЅРµРЅРёСЏ
- `scope TEXT NOT NULL` вЂ” СЃРєРѕСѓРї
- `action TEXT NOT NULL` вЂ” `'create'`, `'update'` РёР»Рё `'delete'`
- `old_value JSONB` вЂ” СЃРѕСЃС‚РѕСЏРЅРёРµ РґРѕ РёР·РјРµРЅРµРЅРёСЏ (`NULL` РїСЂРё СЃРѕР·РґР°РЅРёРё)
- `new_value JSONB` вЂ” СЃРѕСЃС‚РѕСЏРЅРёРµ РїРѕСЃР»Рµ РёР·РјРµРЅРµРЅРёСЏ (`NULL` РїСЂРё СѓРґР°Р»РµРЅРёРё)
- `changed_by BIGINT NOT NULL` вЂ” `user_id` Р°РґРјРёРЅРёСЃС‚СЂР°С‚РѕСЂР°
- `created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()`

### 4. `oc_admin` вЂ” Р°РґРјРёРЅРёСЃС‚СЂР°С‚РѕСЂС‹ OnlineConf

- `user_id BIGINT PRIMARY KEY` вЂ” ID РїРѕР»СЊР·РѕРІР°С‚РµР»СЏ РёР· `user`
- `role TEXT NOT NULL DEFAULT 'admin'` вЂ” `'superadmin'` РёР»Рё `'admin'`
- `created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()`

## SQL DDL

```sql
CREATE TABLE IF NOT EXISTS oc_config (
    id          BIGSERIAL PRIMARY KEY,
    key         TEXT NOT NULL,
    scope       TEXT NOT NULL,
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
    key             TEXT NOT NULL,
    scope           TEXT NOT NULL,
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

CREATE INDEX IF NOT EXISTS idx_oc_flags_scope ON oc_feature_flag (scope);

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

### РџСЂРѕРІРµСЂРєР° РїСЂР°РІ Р°РґРјРёРЅРёСЃС‚СЂР°С‚РѕСЂР° (`oc_admin`)

```sql
SELECT user_id, role, created_at
FROM oc_admin
WHERE user_id = $1;
```

### РџРѕРёСЃРє Рё СЃРїРёСЃРѕРє РєРѕРЅС„РёРіРѕРІ

```sql
SELECT id, key, scope, value_type, value, description, created_by, updated_by, created_at, updated_at
FROM oc_config
WHERE ($1 = '' OR scope = $1)
  AND ($2 = '' OR key ILIKE '%' || $2 || '%' OR description ILIKE '%' || $2 || '%')
ORDER BY scope ASC, key ASC;
```

### РЎРѕР·РґР°РЅРёРµ / РѕР±РЅРѕРІР»РµРЅРёРµ / СѓРґР°Р»РµРЅРёРµ РєРѕРЅС„РёРіР°

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

### РџРѕРёСЃРє Рё СЃРїРёСЃРѕРє С„РёС‡Р°-С‚РѕРіРіР»РѕРІ

```sql
SELECT id, key, scope, enabled, rollout_percent, whitelist, blacklist, description, created_by, updated_by, created_at, updated_at
FROM oc_feature_flag
WHERE ($1 = '' OR scope = $1)
  AND ($2 = '' OR key ILIKE '%' || $2 || '%' OR description ILIKE '%' || $2 || '%')
ORDER BY scope ASC, key ASC;
```

### РЎРѕР·РґР°РЅРёРµ / РѕР±РЅРѕРІР»РµРЅРёРµ / СѓРґР°Р»РµРЅРёРµ С„РёС‡Р°-С‚РѕРіРіР»Р°

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

### Р—Р°РїРёСЃСЊ Рё С‡С‚РµРЅРёРµ РёСЃС‚РѕСЂРёРё (`oc_history`)

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
