ALTER TABLE seller ADD COLUMN fts TSVECTOR;

UPDATE seller SET fts =
      setweight(to_tsvector('ru', name),                          'A')
   || setweight(to_tsvector('ru', coalesce(description, '')),     'B');

ALTER TABLE seller ALTER COLUMN fts SET NOT NULL;
CREATE INDEX seller_fts_idx ON seller USING GIN (fts);

CREATE FUNCTION seller_fts_update() RETURNS trigger AS $$
BEGIN
    NEW.fts :=
          setweight(to_tsvector('ru', NEW.name),                      'A')
       || setweight(to_tsvector('ru', coalesce(NEW.description, '')), 'B');
    RETURN NEW;
END
$$ LANGUAGE plpgsql;

CREATE TRIGGER seller_fts_trg
BEFORE INSERT OR UPDATE OF name, description ON seller
FOR EACH ROW EXECUTE FUNCTION seller_fts_update();
