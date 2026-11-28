DROP TABLE IF EXISTS oc_admins;
DROP TABLE IF EXISTS oc_history;
DROP TRIGGER IF EXISTS oc_feature_flags_updated_at ON oc_feature_flags;
DROP TABLE IF EXISTS oc_feature_flags;
DROP TRIGGER IF EXISTS oc_configs_updated_at ON oc_configs;
DROP TABLE IF EXISTS oc_configs;
