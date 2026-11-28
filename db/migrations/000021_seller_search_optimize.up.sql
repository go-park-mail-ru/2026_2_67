CREATE EXTENSION IF NOT EXISTS pg_trgm;
CREATE INDEX IF NOT EXISTS seller_name_trgm_idx ON seller USING GIN (name gin_trgm_ops);
