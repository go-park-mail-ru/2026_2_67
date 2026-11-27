CREATE EXTENSION IF NOT EXISTS pg_trgm;
CREATE INDEX seller_name_trgm_idx ON seller USING GIN (name gin_trgm_ops);
