CREATE OR REPLACE FUNCTION create_order_notification()
RETURNS TRIGGER
LANGUAGE plpgsql
AS $$
BEGIN
    IF TG_OP = 'INSERT' OR OLD.order_status_id IS DISTINCT FROM NEW.order_status_id THEN
        INSERT INTO order_notification (order_id, order_status_id, is_read)
        VALUES (NEW.id, NEW.order_status_id, FALSE);
    END IF;

    RETURN NEW;
END;
$$;

CREATE TRIGGER order_create_notification
AFTER INSERT OR UPDATE OF order_status_id ON "order"
FOR EACH ROW
EXECUTE FUNCTION create_order_notification();
