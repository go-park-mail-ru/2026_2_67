DROP TRIGGER IF EXISTS product_fts_trg ON product;
DROP FUNCTION IF EXISTS product_fts_update();

DROP INDEX IF EXISTS product_fts_idx;

ALTER TABLE product DROP COLUMN IF EXISTS fts;
