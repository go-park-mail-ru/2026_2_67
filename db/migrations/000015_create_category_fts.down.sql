DROP TRIGGER IF EXISTS category_fts_trg ON category;
DROP FUNCTION IF EXISTS category_fts_update();

DROP INDEX IF EXISTS category_fts_idx;

ALTER TABLE category DROP COLUMN IF EXISTS fts;
