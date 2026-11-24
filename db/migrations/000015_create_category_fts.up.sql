ALTER TABLE category ADD COLUMN fts TSVECTOR;

UPDATE category SET fts =
      setweight(to_tsvector('ru', category),'A');

ALTER TABLE category ALTER COLUMN fts SET NOT NULL;
CREATE INDEX category_fts_idx ON category USING GIN (fts);

CREATE FUNCTION category_fts_update() RETURNS trigger AS $$
BEGIN
    NEW.fts :=
          setweight(to_tsvector('ru', NEW.category),'A');
    RETURN NEW;
END
$$ LANGUAGE plpgsql;

CREATE TRIGGER category_fts_trg
BEFORE INSERT OR UPDATE OF category ON category
FOR EACH ROW EXECUTE FUNCTION category_fts_update();
