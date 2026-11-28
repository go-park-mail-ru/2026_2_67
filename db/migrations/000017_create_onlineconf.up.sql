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

CREATE TRIGGER oc_config_updated_at
BEFORE UPDATE ON oc_config
FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();

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
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (scope, key)
);

CREATE INDEX IF NOT EXISTS idx_oc_feature_flag_scope ON oc_feature_flag (scope);

CREATE TRIGGER oc_feature_flag_updated_at
BEFORE UPDATE ON oc_feature_flag
FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();

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

INSERT INTO oc_admin (user_id, role)
VALUES (1, 'superadmin')
ON CONFLICT (user_id) DO NOTHING;
