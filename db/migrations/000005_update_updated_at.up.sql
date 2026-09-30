CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER user_updated_at
BEFORE UPDATE ON "user"
FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER buyer_updated_at
BEFORE UPDATE ON buyer
FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER seller_updated_at
BEFORE UPDATE ON seller
FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER product_updated_at
BEFORE UPDATE ON product
FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER product_review_updated_at
BEFORE UPDATE ON product_review
FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER basket_product_updated_at
BEFORE UPDATE ON basket_product
FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER order_updated_at
BEFORE UPDATE ON "order"
FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();