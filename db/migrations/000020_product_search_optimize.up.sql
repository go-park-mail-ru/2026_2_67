CREATE INDEX idx_product_available_id ON product (id DESC)
WHERE available_count > 0;

CREATE EXTENSION IF NOT EXISTS pg_trgm;

CREATE INDEX idx_product_name_trgm ON product USING GIN(name gin_trgm_ops);
