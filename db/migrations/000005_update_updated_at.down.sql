DROP TRIGGER order_updated_at ON "order";
DROP TRIGGER basket_product_updated_at ON basket_product;
DROP TRIGGER product_review_updated_at ON product_review;
DROP TRIGGER product_updated_at ON product;
DROP TRIGGER seller_updated_at ON seller;
DROP TRIGGER buyer_updated_at ON buyer;
DROP TRIGGER user_updated_at ON "user";

DROP FUNCTION update_updated_at_column();