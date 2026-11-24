ALTER TABLE product ADD COLUMN fts TSVECTOR;

UPDATE product SET fts =
      setweight(to_tsvector('ru', name),                          'A')
   || setweight(to_tsvector('ru', coalesce(description, '')),     'B');

ALTER TABLE product ALTER COLUMN fts SET NOT NULL;
CREATE INDEX product_fts_idx ON product USING GIN (fts);

CREATE FUNCTION product_fts_update() RETURNS trigger AS $$
BEGIN
    NEW.fts :=
          setweight(to_tsvector('ru', NEW.name),                      'A')
       || setweight(to_tsvector('ru', coalesce(NEW.description, '')), 'B');
    RETURN NEW;
END
$$ LANGUAGE plpgsql;

CREATE TRIGGER product_fts_trg
BEFORE INSERT OR UPDATE OF name, description ON product
FOR EACH ROW EXECUTE FUNCTION product_fts_update();
