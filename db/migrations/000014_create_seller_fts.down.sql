DROP TRIGGER IF EXISTS seller_fts_trg ON seller;
DROP FUNCTION IF EXISTS seller_fts_update();

DROP INDEX IF EXISTS seller_fts_idx;

ALTER TABLE seller DROP COLUMN IF EXISTS fts;
