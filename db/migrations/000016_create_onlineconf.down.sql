DROP TABLE IF EXISTS oc_admin;
DROP TABLE IF EXISTS oc_history;
DROP TRIGGER IF EXISTS oc_feature_flag_updated_at ON oc_feature_flag;
DROP TABLE IF EXISTS oc_feature_flag;
DROP TRIGGER IF EXISTS oc_config_updated_at ON oc_config;
DROP TABLE IF EXISTS oc_config;
